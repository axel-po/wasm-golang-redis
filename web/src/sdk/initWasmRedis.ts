import { createWorkerClient } from "./client";
import { intersectKeys } from "./intersect";
import type {
  BatchResult,
  DbEntry,
  EngineConfig,
  FilterOp,
  Schema,
  SetOptions,
} from "./types";
import { CommandError, KeyNotFoundError } from "./types";
import {
  buildDeleteCommand,
  buildGetCommand,
  buildSetCommand,
  parseFilterValue,
  parseKey,
} from "./validation";

export interface Query<S extends Schema> {
  where: (op: FilterOp, value: string | number) => Query<S>;
  exec: () => Promise<DbEntry[]>;
}

export interface WasmRedis<S extends Schema> {
  set: <K extends keyof S & string>(
    key: K,
    value: S[K],
    options?: SetOptions,
  ) => Promise<void>;

  get: (<K extends keyof S & string>(key: K) => Promise<string>) &
    (() => Query<S>);
  delete: (key: keyof S & string) => Promise<void>;
  batch: (commands: string[]) => Promise<BatchResult[]>;
  cmd: {
    set: <K extends keyof S & string>(
      key: K,
      value: S[K],
      options?: SetOptions,
    ) => string;
    get: (key: keyof S & string) => string;
    delete: (key: keyof S & string) => string;
  };
  entries: () => Promise<DbEntry[]>;
  flush: () => Promise<void>;
  snapshot: () => Promise<void>;
  close: () => Promise<void>;
}

const DEFAULT_CONFIG: EngineConfig = {
  flushIntervalMs: 1000,
  snapshotIntervalMs: 120_000,
  sweepIntervalMs: 1000,
};

export const initWasmRedis = async <S extends Schema = Schema>(
  config?: Partial<EngineConfig>,
): Promise<WasmRedis<S>> => {
  const worker = new Worker(new URL("../worker/worker.ts", import.meta.url), {
    type: "module",
  });
  const client = createWorkerClient(worker);

  await client.request({
    type: "init",
    config: { ...DEFAULT_CONFIG, ...config },
  });

  const getMany = async (keys: string[]): Promise<DbEntry[]> => {
    const raw = (await client.request({ type: "getMany", keys })) as string;
    return normalizeEntries(JSON.parse(raw));
  };

  const entries = async (): Promise<DbEntry[]> => {
    const raw = (await client.request({ type: "entries" })) as string;
    return normalizeEntries(JSON.parse(raw));
  };

  const batch = async (commands: string[]): Promise<BatchResult[]> => {
    const results = (await client.request({ type: "batch", commands })) as {
      value?: string;
      error?: string;
    }[];
    return results.map((r) =>
      r.error !== undefined
        ? { ok: false, error: r.error }
        : { ok: true, value: r.value ?? "" },
    );
  };

  const runOne = async (command: string): Promise<string> => {
    const [result] = await batch([command]);
    if (!result.ok) throw new CommandError(result.error);
    return result.value;
  };

  const makeQuery = (filters: { op: FilterOp; value: string }[]): Query<S> => ({
    where: (op, value) =>
      makeQuery([...filters, { op, value: parseFilterValue(value) }]),
    exec: async () => {
      if (filters.length === 0) return entries();
      const keySets = (await client.request({
        type: "where",
        filters,
      })) as string[][];
      return getMany(intersectKeys(keySets));
    },
  });

  const get = (<K extends keyof S & string>(key?: K) => {
    if (key === undefined) return makeQuery([]);
    return getMany([parseKey(key)]).then((found) => {
      if (found.length === 0) throw new KeyNotFoundError(key);
      return found[0].value;
    });
  }) as WasmRedis<S>["get"];

  const api: WasmRedis<S> = {
    set: async (key, value, options) => {
      await runOne(buildSetCommand(key, value, options?.ex));
    },
    get,
    delete: async (key) => {
      await runOne(buildDeleteCommand(key));
    },
    batch,
    cmd: {
      set: (key, value, options) => buildSetCommand(key, value, options?.ex),
      get: (key) => buildGetCommand(key),
      delete: (key) => buildDeleteCommand(key),
    },
    entries,
    flush: async () => {
      await client.request({ type: "flush" });
    },
    snapshot: async () => {
      await client.request({ type: "snapshot" });
    },
    close: async () => {
      await client.request({ type: "close" });
      client.terminate();
    },
  };
  return Object.freeze(api);
};

const normalizeEntries = (
  raw: { key: string; value: string; expiresAt?: number }[],
): DbEntry[] =>
  raw.map((e) => ({
    key: e.key,
    value: e.value,
    expiresAt: e.expiresAt && e.expiresAt > 0 ? e.expiresAt : null,
  }));

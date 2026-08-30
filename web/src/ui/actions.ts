import { batch } from "@legendapp/state";
import { CONFIG } from "../config";
import {
  initWasmRedis,
  WasmRedisError,
  type Schema,
  type WasmRedis,
} from "../sdk";
import { insertSorted, store$ } from "./store";

let db: WasmRedis<Schema> | undefined;

const toast = (kind: "ok" | "error", text: string): void => {
  store$.toast.set({ kind, text });
  setTimeout(() => store$.toast.set(null), 3500);
};

const run = async (label: string, fn: () => Promise<string>): Promise<void> => {
  if (store$.busy.get() !== null) return;
  store$.busy.set(label);
  try {
    toast("ok", await fn());
  } catch (err) {
    toast("error", err instanceof WasmRedisError ? err.message : String(err));
  } finally {
    store$.busy.set(null);
  }
};

export const initApp = async (): Promise<void> => {
  try {
    db = await initWasmRedis({
      flushIntervalMs: CONFIG.flushIntervalMs,
      snapshotIntervalMs: CONFIG.snapshotIntervalMs,
      sweepIntervalMs: CONFIG.sweepIntervalMs,
    });
    await hydrate();
    store$.ready.set(true);

    window.addEventListener("beforeunload", () => {
      void db?.flush();
    });
  } catch (err) {
    toast(
      "error",
      `démarrage impossible: ${err instanceof Error ? err.message : String(err)}`,
    );
  }
};

export const hydrate = async (): Promise<void> => {
  if (!db) return;
  const all = await db.entries();

  const entries: Record<string, { value: string; expiresAt: number | null }> =
    {};
  const keys: string[] = new Array(all.length);
  for (let i = 0; i < all.length; i++) {
    entries[all[i].key] = { value: all[i].value, expiresAt: all[i].expiresAt };
    keys[i] = all[i].key;
  }
  batch(() => {
    store$.entries.set(entries);
    store$.keys.set(keys);
  });
};

export const upsertEntry = (
  key: string,
  value: string,
  ex?: number,
): Promise<void> =>
  run(`SET ${key}`, async () => {
    if (!db) throw new WasmRedisError("moteur non démarré");
    await db.set(key, value, ex ? { ex } : undefined);

    const expiresAt = ex ? Date.now() + ex * 1000 : null;
    batch(() => {

      store$.entries[key].set({ value, expiresAt });
      store$.keys.set(insertSorted(store$.keys.peek(), key));
    });
    return `SET ${key} ✓`;
  });

export const deleteEntry = (key: string): Promise<void> =>
  run(`DELETE ${key}`, async () => {
    if (!db) throw new WasmRedisError("moteur non démarré");
    await db.delete(key);

    batch(() => {
      store$.entries[key].delete();
      store$.keys.set(store$.keys.peek().filter((k) => k !== key));
    });
    return `DELETE ${key} ✓`;
  });

export const touchRandomEntry = (): Promise<void> =>
  run("update aléatoire", async () => {
    if (!db) throw new WasmRedisError("moteur non démarré");
    const keys = store$.keys.peek();
    if (keys.length === 0)
      throw new WasmRedisError("base vide : seedez d'abord");

    const key = keys[Math.floor(Math.random() * keys.length)];
    const value = String(Math.floor(Math.random() * 100_000));
    await db.set(key, value);

    store$.entries[key].value.set(value);
    return `${key} ← "${value}" (1 ligne sur ${keys.length.toLocaleString("fr-FR")} re-rendue)`;
  });

export const seed = (count: number): Promise<void> =>
  run(`seed ${count.toLocaleString("fr-FR")}`, async () => {
    if (!db) throw new WasmRedisError("moteur non démarré");
    store$.seedProgress.set(0);
    const started = performance.now();

    const random = mulberry32(42);
    const names = [
      "matt",
      "zoe",
      "lea",
      "marc",
      "emma",
      "hugo",
      "alice",
      "tom",
    ];
    const width = String(count - 1).length;

    for (let offset = 0; offset < count; offset += CONFIG.batchChunk) {
      const size = Math.min(CONFIG.batchChunk, count - offset);
      const commands = new Array<string>(size);
      for (let i = 0; i < size; i++) {
        const n = offset + i;
        const key = `item:${String(n).padStart(width, "0")}`;

        const value =
          n % 8 === 7
            ? names[Math.floor(random() * names.length)]
            : String(Math.floor(random() * 100_000));
        commands[i] = db.cmd.set(key, value);
      }
      const results = await db.batch(commands);
      const failed = results.find((r) => !r.ok);
      if (failed && !failed.ok) throw new WasmRedisError(failed.error);
      store$.seedProgress.set((offset + size) / count);
    }

    await hydrate();
    store$.seedProgress.set(null);
    const tookMs = Math.round(performance.now() - started);
    return `${count.toLocaleString("fr-FR")} entrées en ${tookMs} ms (batchs de ${CONFIG.batchChunk})`;
  });

export const runQuery = (
  op: "=" | ">" | ">=" | "<" | "<=" | "contains",
  value: string,
): Promise<void> =>
  run(`GET ${op}`, async () => {
    if (!db) throw new WasmRedisError("moteur non démarré");
    const started = performance.now();
    const found = await db.get().where(op, value).exec();
    const tookMs = Math.round((performance.now() - started) * 100) / 100;

    store$.queryResult.set({
      label: `GET ${op} "${value}"`,
      keys: found.map((e) => e.key),
      tookMs,
    });
    return `${found.length.toLocaleString("fr-FR")} clé(s) en ${tookMs} ms`;
  });

export const flushNow = (): Promise<void> =>
  run("flush", async () => {
    if (!db) throw new WasmRedisError("moteur non démarré");
    await db.flush();
    return "flush → aof.log ✓";
  });

export const snapshotNow = (): Promise<void> =>
  run("snapshot", async () => {
    if (!db) throw new WasmRedisError("moteur non démarré");
    await db.snapshot();
    return "snapshot.json écrit, AOF compacté ✓";
  });

const mulberry32 = (seed: number) => () => {
  seed |= 0;
  seed = (seed + 0x6d2b79f5) | 0;
  let t = Math.imul(seed ^ (seed >>> 15), 1 | seed);
  t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
  return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
};

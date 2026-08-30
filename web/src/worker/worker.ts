import "./wasm_exec.js";
import type { WorkerRequest, WorkerResponse } from "../sdk/protocol";
import { createOpfsStorage } from "./opfs";

interface WasmBridge {
  init(config: unknown): unknown;
  execBatch(commands: string[]): unknown;
  getWhere(op: string, value: string): unknown;
  getMany(keys: string[]): unknown;
  entries(): unknown;
  flush(): unknown;
  snapshot(): unknown;
  close(): unknown;
}

const globals = globalThis as unknown as {
  Go: new () => {
    importObject: WebAssembly.Imports;
    run(i: WebAssembly.Instance): Promise<void>;
  };
  __wasmredisStorage?: unknown;
  __wasmredis?: WasmBridge;
};

let bridge: WasmBridge | undefined;

const boot = async (config: unknown): Promise<unknown> => {
  globals.__wasmredisStorage = await createOpfsStorage();

  const go = new globals.Go();
  const { instance } = await WebAssembly.instantiateStreaming(
    fetch("/wasmredis.wasm"),
    go.importObject,
  );
  void go.run(instance);

  bridge = globals.__wasmredis;
  if (!bridge) {
    throw new Error("le moteur WASM n'a pas publié __wasmredis");
  }

  return unwrap(bridge.init(config));
};

const unwrap = (result: unknown): unknown => {
  if (
    result !== null &&
    typeof result === "object" &&
    "error" in result &&
    typeof (result as { error: unknown }).error === "string"
  ) {
    throw new Error((result as { error: string }).error);
  }
  return result;
};

const ready = (): WasmBridge => {
  if (!bridge) throw new Error("moteur non initialisé : init d'abord");
  return bridge;
};

const handle = async (request: WorkerRequest): Promise<unknown> => {
  switch (request.type) {
    case "init":
      return boot(request.config);
    case "batch":
      return unwrap(ready().execBatch(request.commands));
    case "where":
      return request.filters.map((f) =>
        unwrap(ready().getWhere(f.op, f.value)),
      );
    case "getMany":
      return unwrap(ready().getMany(request.keys));
    case "entries":
      return unwrap(ready().entries());
    case "flush":
      return unwrap(ready().flush());
    case "snapshot":
      return unwrap(ready().snapshot());
    case "close":
      return unwrap(ready().close());
  }
};

self.onmessage = async (event: MessageEvent<WorkerRequest>) => {
  const request = event.data;
  let response: WorkerResponse;
  try {
    response = { id: request.id, ok: true, data: await handle(request) };
  } catch (err) {
    response = {
      id: request.id,
      ok: false,
      error: err instanceof Error ? err.message : String(err),
    };
  }
  self.postMessage(response);
};

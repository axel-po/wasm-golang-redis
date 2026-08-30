import type { WorkerRequest, WorkerResponse } from "./protocol";
import { WasmRedisError } from "./types";

type Pending = {
  resolve: (data: unknown) => void;
  reject: (err: Error) => void;
};

type RequestBody = WorkerRequest extends infer R
  ? R extends WorkerRequest
    ? Omit<R, "id">
    : never
  : never;

export interface WorkerClient {
  request: (msg: RequestBody) => Promise<unknown>;
  terminate: () => void;
}

export const createWorkerClient = (worker: Worker): WorkerClient => {
  const pending = new Map<number, Pending>();
  let nextId = 1;

  worker.onmessage = (event: MessageEvent<WorkerResponse>) => {
    const response = event.data;
    const waiter = pending.get(response.id);
    if (!waiter) return;
    pending.delete(response.id);

    if (response.ok) {
      waiter.resolve(response.data);
    } else {
      waiter.reject(new WasmRedisError(response.error));
    }
  };

  worker.onerror = (event) => {
    const error = new WasmRedisError(`worker en erreur: ${event.message}`);
    pending.forEach((waiter) => waiter.reject(error));
    pending.clear();
  };

  return {
    request: (msg) =>
      new Promise((resolve, reject) => {
        const id = nextId++;
        pending.set(id, { resolve, reject });
        worker.postMessage({ ...msg, id });
      }),
    terminate: () => worker.terminate(),
  };
};

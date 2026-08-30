import type { EngineConfig, FilterOp } from "./types";

export type WorkerRequest =
  | { id: number; type: "init"; config: EngineConfig }
  | { id: number; type: "batch"; commands: string[] }
  | { id: number; type: "where"; filters: { op: FilterOp; value: string }[] }
  | { id: number; type: "getMany"; keys: string[] }
  | { id: number; type: "entries" }
  | { id: number; type: "flush" }
  | { id: number; type: "snapshot" }
  | { id: number; type: "close" };

export type WorkerResponse =
  | { id: number; ok: true; data: unknown }
  | { id: number; ok: false; error: string };

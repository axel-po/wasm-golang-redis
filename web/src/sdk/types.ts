export type Schema = Record<string, string | number>;

export type FilterOp = "=" | ">" | ">=" | "<" | "<=" | "contains";

export interface DbEntry {
  key: string;
  value: string;
  expiresAt: number | null;
}

export type BatchResult =
  | { ok: true; value: string }
  | { ok: false; error: string };

export interface SetOptions {
  ex?: number;
}

export interface EngineConfig {
  flushIntervalMs: number;
  snapshotIntervalMs: number;
  sweepIntervalMs: number;
}

export class WasmRedisError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "WasmRedisError";
  }
}

export class KeyNotFoundError extends WasmRedisError {
  constructor(readonly key: string) {
    super(`clé absente: "${key}"`);
    this.name = "KeyNotFoundError";
  }
}

export class ValidationError extends WasmRedisError {
  constructor(message: string) {
    super(message);
    this.name = "ValidationError";
  }
}

export class CommandError extends WasmRedisError {
  constructor(message: string) {
    super(message);
    this.name = "CommandError";
  }
}

const num = (raw: string | undefined, fallback: number): number => {
  const parsed = Number(raw);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
};

export const CONFIG = {
  flushIntervalMs: num(import.meta.env.VITE_FLUSH_INTERVAL_MS, 1000),
  snapshotIntervalMs: num(import.meta.env.VITE_SNAPSHOT_INTERVAL_MS, 120_000),
  sweepIntervalMs: num(import.meta.env.VITE_SWEEP_INTERVAL_MS, 1000),

  rowHeight: num(import.meta.env.VITE_ROW_HEIGHT, 36),
  viewportHeight: num(import.meta.env.VITE_VIEWPORT_HEIGHT, 560),
  overscan: num(import.meta.env.VITE_OVERSCAN, 8),

  seedSmall: num(import.meta.env.VITE_SEED_SMALL, 1000),
  seedLarge: num(import.meta.env.VITE_SEED_LARGE, 100_000),
  batchChunk: num(import.meta.env.VITE_BATCH_CHUNK, 5000),
} as const;

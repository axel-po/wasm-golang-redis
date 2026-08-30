import { observable } from "@legendapp/state";

export interface RowEntry {
  value: string;
  expiresAt: number | null;
}

export interface Toast {
  kind: "ok" | "error";
  text: string;
}

export const store$ = observable({
  entries: {} as Record<string, RowEntry>,
  keys: [] as string[],

  ready: false,
  busy: null as string | null,
  seedProgress: null as number | null,
  toast: null as Toast | null,

  queryResult: null as { label: string; keys: string[]; tookMs: number } | null,
});

export const insertSorted = (keys: string[], key: string): string[] => {
  let lo = 0;
  let hi = keys.length;
  while (lo < hi) {
    const mid = (lo + hi) >> 1;
    if (keys[mid] < key) lo = mid + 1;
    else hi = mid;
  }
  if (keys[lo] === key) return keys;
  return [...keys.slice(0, lo), key, ...keys.slice(lo)];
};

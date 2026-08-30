export const intersectKeys = (keySets: string[][]): string[] => {
  if (keySets.length === 0) return [];

  const [first, ...rest] = keySets;
  const sets = rest.map((keys) => new Set(keys));
  return first.filter((key) => sets.every((set) => set.has(key)));
};

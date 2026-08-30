import { describe, expect, it } from "vitest";
import { intersectKeys } from "./intersect";

describe("intersectKeys", () => {
  it("un seul filtre : ses clés telles quelles", () => {
    expect(intersectKeys([["a", "b", "c"]])).toEqual(["a", "b", "c"]);
  });

  it("deux filtres : seules les clés communes restent (AND)", () => {
    expect(
      intersectKeys([
        ["a", "b", "c"],
        ["b", "c", "d"],
      ]),
    ).toEqual(["b", "c"]);
  });

  it("préserve l'ordre (trié) du premier ensemble", () => {
    expect(
      intersectKeys([
        ["a", "b", "z"],
        ["z", "a"],
      ]),
    ).toEqual(["a", "z"]);
  });

  it("intersection vide quand rien ne matche partout", () => {
    expect(intersectKeys([["a"], ["b"]])).toEqual([]);
    expect(intersectKeys([["a", "b"], [], ["a"]])).toEqual([]);
  });

  it("aucun filtre : aucune clé (le SDK route ce cas vers entries())", () => {
    expect(intersectKeys([])).toEqual([]);
  });
});

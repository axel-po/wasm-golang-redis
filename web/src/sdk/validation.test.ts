import { describe, expect, it } from "vitest";
import {
  buildDeleteCommand,
  buildGetCommand,
  buildSetCommand,
  parseFilterValue,
} from "./validation";
import { ValidationError } from "./types";

describe("buildSetCommand", () => {
  it("fabrique un SET avec la valeur entre guillemets", () => {
    expect(buildSetCommand("name", "matt")).toBe('SET name "matt"');
  });

  it("supporte les valeurs avec espaces (le tokenizer Go respecte les guillemets)", () => {
    expect(buildSetCommand("msg", "hello world")).toBe('SET msg "hello world"');
  });

  it("convertit un number en string (le moteur compare numériquement)", () => {
    expect(buildSetCommand("age", 30)).toBe('SET age "30"');
  });

  it("ajoute EX quand une durée de vie est demandée", () => {
    expect(buildSetCommand("session", "abc", 60)).toBe(
      'SET session "abc" EX 60',
    );
  });

  it("rejette une clé vide", () => {
    expect(() => buildSetCommand("", "x")).toThrow(ValidationError);
  });

  it("rejette une clé avec espace (le protocole texte est découpé sur les espaces)", () => {
    expect(() => buildSetCommand("ma clé", "x")).toThrow(ValidationError);
  });

  it("rejette une clé avec guillemet", () => {
    expect(() => buildSetCommand('a"b', "x")).toThrow(ValidationError);
  });

  it("rejette une valeur avec guillemet (pas d'échappement \\\" côté moteur)", () => {
    expect(() => buildSetCommand("k", 'dire "bonjour"')).toThrow(
      ValidationError,
    );
  });

  it("rejette un EX non entier ou négatif", () => {
    expect(() => buildSetCommand("k", "v", 1.5)).toThrow(ValidationError);
    expect(() => buildSetCommand("k", "v", -3)).toThrow(ValidationError);
    expect(() => buildSetCommand("k", "v", 0)).toThrow(ValidationError);
  });
});

describe("buildGetCommand / buildDeleteCommand", () => {
  it("fabriquent les commandes simples", () => {
    expect(buildGetCommand("name")).toBe("GET name");
    expect(buildDeleteCommand("name")).toBe("DELETE name");
  });

  it("valident aussi la clé", () => {
    expect(() => buildGetCommand("a b")).toThrow(ValidationError);
    expect(() => buildDeleteCommand("")).toThrow(ValidationError);
  });
});

describe("parseFilterValue", () => {
  it("accepte string et number", () => {
    expect(parseFilterValue("matt")).toBe("matt");
    expect(parseFilterValue(25)).toBe("25");
  });

  it("rejette un pivot avec guillemet", () => {
    expect(() => parseFilterValue('a"b')).toThrow(ValidationError);
  });
});

import { z } from "zod";
import { ValidationError } from "./types";

const keySchema = z
  .string()
  .min(1, "la clé ne peut pas être vide")
  .refine((k) => !/[\s"]/.test(k), {
    message: "la clé ne peut contenir ni espace ni guillemet",
  });

const valueSchema = z
  .union([z.string(), z.number()])
  .transform(String)
  .refine((v) => !v.includes('"'), {
    message: 'la valeur ne peut pas contenir de guillemet (")',
  });

const exSchema = z.number().int().positive({
  message: "EX attend une durée en secondes, entière et positive",
});

const parse = <T>(
  schema: z.ZodType<T, z.ZodTypeDef, unknown>,
  input: unknown,
): T => {
  const result = schema.safeParse(input);
  if (!result.success) {
    throw new ValidationError(
      result.error.issues[0]?.message ?? "entrée invalide",
    );
  }
  return result.data;
};

export const buildSetCommand = (
  key: unknown,
  value: unknown,
  ex?: unknown,
): string => {
  const k = parse(keySchema, key);
  const v = parse(valueSchema, value);
  const suffix = ex === undefined ? "" : ` EX ${parse(exSchema, ex)}`;
  return `SET ${k} "${v}"${suffix}`;
};

export const buildGetCommand = (key: unknown): string =>
  `GET ${parse(keySchema, key)}`;

export const buildDeleteCommand = (key: unknown): string =>
  `DELETE ${parse(keySchema, key)}`;

export const parseFilterValue = (value: unknown): string =>
  parse(valueSchema, value);

export const parseKey = (key: unknown): string => parse(keySchema, key);

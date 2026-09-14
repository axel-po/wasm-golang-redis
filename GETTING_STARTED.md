# Démarrage rapide

Ce guide explique comment lancer le projet, du REPL en terminal jusqu'à
l'interface web. Pour comprendre *ce que fait* le projet et ses étapes, voir
**[README.md](./README.md)**.

## Prérequis

| Outil | Version utilisée | Pour |
|-------|------------------|------|
| Go    | 1.26+            | Le moteur, le REPL, la compilation WASM |
| Node  | 20+              | L'interface web |
| pnpm  | 10+              | Le gestionnaire de paquets du dossier `web/` |

Vérifier :

```bash
go version
node -v
pnpm -v
```

---

## 1. Lancer le REPL (le plus simple)

Le REPL est le moyen le plus direct de tester le moteur.

```bash
go run ./cmd/repl
```

Puis tape des commandes :

```
clone mini redis — SET k "v" [EX s] / GET k / DELETE k / GET WHERE <op> v / Ctrl+D
> SET name "Axel"
OK
> GET name
"Axel"
> SET age:1 30
OK
> SET age:2 42
OK
> GET WHERE > 35
age:2 = "42"
> SET session "token" EX 60
OK
> DELETE name
OK
> GET name
(nil)
```

Quitte avec **Ctrl+D**.

### Commandes disponibles

| Commande | Effet |
|----------|-------|
| `SET <clé> <valeur>` | Écrit une clé |
| `SET <clé> <valeur> EX <secondes>` | Écrit une clé qui expire après N secondes |
| `GET <clé>` | Lit une clé (`(nil)` si absente/expirée) |
| `DELETE <clé>` | Supprime une clé |
| `GET WHERE equals <valeur>` | Toutes les clés dont la valeur est exactement `<valeur>` |
| `GET WHERE contains <sous-chaîne>` | Valeurs contenant la sous-chaîne |
| `GET WHERE > <n>` (`>=`, `<`, `<=`) | Comparaison numérique |

### Persistance

Les données sont écrites dans le dossier `data/` (`aof.log` + `snapshot.json`).
Relance le REPL : tes clés sont toujours là. Pour repartir de zéro :

```bash
rm -rf data/
```

### Configuration (variables d'environnement)

Voir `.env.example`. Toutes optionnelles :

```bash
WASMREDIS_DATA_DIR=data           # dossier de persistance
WASMREDIS_FLUSH_INTERVAL=1s       # fréquence d'écriture de l'AOF
WASMREDIS_SNAPSHOT_INTERVAL=2m    # fréquence des snapshots
WASMREDIS_SWEEP_INTERVAL=10s      # fréquence de purge des clés expirées
```

Exemple :

```bash
WASMREDIS_DATA_DIR=/tmp/maredis go run ./cmd/repl
```

---

## 2. Lancer les tests Go

```bash
go test ./...
```

Un package précis, en mode verbeux :

```bash
go test ./internal/engine -v
```

---

## 3. Lancer l'interface web (WASM + React)

L'interface fait tourner le **même moteur Go**, compilé en WebAssembly, dans le
navigateur.

### Étape 1 — Compiler le moteur en WASM

Depuis la racine du projet :

```bash
./scripts/build-wasm.sh
```

Cela génère :

- `web/public/wasmredis.wasm` — le moteur compilé ;
- `web/src/worker/wasm_exec.js` — le runtime Go officiel pour WASM.

### Étape 2 — Installer les dépendances web

```bash
cd web
pnpm install
```

### Étape 3 — Lancer le serveur de dev

```bash
pnpm dev
```

Ouvre l'URL affichée par Vite (généralement <http://localhost:5173>).

> Raccourci : depuis `web/`, `pnpm wasm` relance la compilation WASM sans quitter
> le dossier.

### Scripts web disponibles (`web/`)

| Commande | Effet |
|----------|-------|
| `pnpm dev` | Serveur de développement Vite |
| `pnpm build` | Vérification TypeScript + build de production |
| `pnpm preview` | Prévisualise le build de production |
| `pnpm test` | Tests du SDK (Vitest) |
| `pnpm lint` | ESLint |
| `pnpm wasm` | Recompile le moteur Go en WASM |

---

## 4. Utiliser le SDK TypeScript

Le SDK expose une API typée par-dessus le moteur WASM (exécuté dans un Web Worker) :

```ts
import { initWasmRedis } from "./sdk";

// Un schéma optionnel type les clés/valeurs
const db = await initWasmRedis<{ "user:1": string; "age:1": string }>();

await db.set("user:1", "Axel");
await db.set("age:1", "30", { ex: 60 }); // expire dans 60 s

const name = await db.get("user:1");          // "Axel"
const olderThan25 = await db.get().where(">", 25).exec(); // GET WHERE chaîné

await db.snapshot(); // force un snapshot
await db.close();    // arrête le worker proprement
```

---

## Résumé des commandes

```bash
# REPL
go run ./cmd/repl

# Tests Go
go test ./...

# Web (depuis la racine)
./scripts/build-wasm.sh
cd web && pnpm install && pnpm dev

# Tests SDK
cd web && pnpm test
```

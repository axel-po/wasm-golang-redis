# wasmredis — un clone de Redis en Go, du REPL au navigateur

Un magasin clé/valeur en mémoire, écrit en Go, construit **par étapes**. On part
d'un moteur minimal piloté par un REPL, puis on ajoute la persistance sur disque,
l'exécution par lots, les requêtes filtrées avec index, l'expiration des clés (TTL),
et enfin une compilation en **WebAssembly** exposée via un **SDK TypeScript** et une
**interface web React**.

Le même moteur Go tourne à deux endroits :

- **en local** dans un terminal (REPL, persistance sur fichiers) ;
- **dans le navigateur** compilé en WASM, la persistance passant par l'OPFS
  (Origin Private File System) au lieu du disque.

```
                 ┌──────────────────────────────────────────────┐
                 │              internal/engine                  │
   command  ───► │  état (map) · index equals · index b-tree     │ ───► Result / Entry
   (parser)      │  TTL · AOF buffer · snapshot                   │
                 └───────────────┬───────────────┬──────────────┘
                                 │               │
                     Storage (interface)   background loop
                                 │          (flush/snapshot/sweep)
                     ┌───────────┴───────────┐
                     │                       │
              FileStorage (Go)         jsStorage → OPFS (WASM)
                     │                       │
              cmd/repl (terminal)      cmd/wasm → SDK TS → UI React
```

## Structure du dépôt

| Chemin | Rôle |
|--------|------|
| `internal/command/` | Tokenizer + parser des commandes (`SET`, `GET`, `DELETE`, `GET WHERE`) |
| `internal/engine/`  | Le moteur : état, index, TTL, AOF, snapshot, boucle de fond |
| `internal/storage/` | Persistance sur fichiers (`aof.log`, `snapshot.json`) |
| `internal/config/`  | Configuration via variables d'environnement |
| `cmd/repl/`         | Le REPL en ligne de commande |
| `cmd/wasm/`         | Le point d'entrée WebAssembly + pont OPFS |
| `web/src/sdk/`      | SDK TypeScript typé (worker, protocole, validation) |
| `web/src/ui/`       | Interface React (liste virtualisée, toolbar) |
| `scripts/build-wasm.sh` | Compile le moteur Go en `.wasm` pour le web |

> Pour lancer le projet, voir **[GETTING_STARTED.md](./GETTING_STARTED.md)**.

---

## Les étapes du projet

Chaque phase correspond à un commit et ajoute une capacité au-dessus de la
précédente, sans casser ce qui existe.

### PHASE 1 — Le moteur et le REPL

Le socle. On définit :

- un **modèle de commandes** typé (`command.Command` : `Set`, `Get`, `Delete`) ;
- un **parser** (`tokenize` + `Parse`) qui transforme une ligne texte en commande,
  en gérant les guillemets et les erreurs (`ERR ...`) ;
- un **moteur** (`engine.Engine`) qui stocke l'état dans une `map[string]record`
  protégée par un `sync.Mutex` ;
- un **REPL** (`cmd/repl`) qui lit stdin, parse, applique, affiche le résultat.

```
> SET name "Axel"
OK
> GET name
"Axel"
> DELETE name
OK
> GET name
(nil)
```

### PHASE 2 — La persistance sur fichiers (AOF + snapshot)

Le moteur devient durable. On introduit une interface `Storage` et une
implémentation `FileStorage` qui écrit dans un dossier `data/` :

- **AOF (Append-Only File)** — chaque écriture (`SET`/`DELETE`) est journalisée
  dans `aof.log`. C'est le journal des opérations.
- **Snapshot** — périodiquement, l'état complet est sérialisé en JSON
  (`snapshot.json`) via une écriture atomique (`.tmp` puis `rename`), et l'AOF
  est vidé.
- **Restore** — au démarrage, le moteur recharge le snapshot puis rejoue l'AOF
  par-dessus pour retrouver l'état exact.
- **Boucle de fond** (`StartBackground`) — des tickers déclenchent le *flush* du
  buffer AOF et les *snapshots* sans bloquer les requêtes.

Résultat : on peut arrêter et relancer le REPL sans perdre les données.

### PHASE 3 — L'exécution par lots (`Batch`)

On ajoute `Engine.Batch([]command.Command) []BatchResult` : appliquer plusieurs
commandes en une seule fois et récupérer un résultat par commande. C'est la brique
qui rendra les insertions massives (des milliers de clés) efficaces plus tard,
notamment côté WASM où chaque aller-retour JS↔Go a un coût.

### PHASE 4 — Les requêtes filtrées `GET WHERE` + index

Le magasin sait maintenant **chercher par valeur**, pas seulement par clé. Nouvelle
commande `GET WHERE <op> <valeur>` avec plusieurs opérateurs :

- `equals` — égalité exacte, servie par un **index inversé** (`map[valeur] → clés`) ;
- `contains` — sous-chaîne (scan linéaire) ;
- `>`, `>=`, `<`, `<=` — comparaisons numériques, servies par un **index trié**.

```
> SET age:1 30
OK
> SET age:2 42
OK
> GET WHERE > 35
age:2 = "42"
```

Les index sont maintenus à jour à chaque `SET`/`DELETE` et reconstruits au *restore*.

### PHASE 4.5 — Refactor : l'index trié devient un b-tree

L'index de plage initial est remplacé par une implémentation **b-tree**
(`btree.go`, derrière l'interface `rangeIndex`). Objectif : des requêtes de plage
et des insertions performantes même avec beaucoup de clés numériques, sans re-trier
un slice à chaque écriture.

### PHASE 5 — L'expiration des clés (TTL)

Les clés peuvent désormais avoir une durée de vie, via l'option `EX` sur `SET` :

```
> SET session:1 "token" EX 60
OK          # expire dans 60 secondes
```

Trois mécanismes cohabitent :

- **expiration paresseuse** — une clé expirée lue via `GET` est supprimée à la volée ;
- **balayage périodique** (`sweepExpired`) — la boucle de fond purge régulièrement
  les clés mortes (intervalle `SweepInterval`) ;
- **persistance du TTL** — la date d'expiration est stockée dans l'AOF et le
  snapshot, donc préservée après un redémarrage.

### PHASE 6 — WASM, SDK TypeScript et UI React

Le moteur quitte le terminal pour le navigateur.

- **Cible WASM** (`cmd/wasm`) — le moteur est compilé en WebAssembly
  (`GOOS=js GOARCH=wasm`). Il expose des fonctions globales (`init`, `execBatch`,
  `getWhere`, `getMany`, `entries`, `flush`, `snapshot`, `close`) sous
  `__wasmredis`. La persistance utilise l'**OPFS** du navigateur au lieu du disque.
- **SDK TypeScript** (`web/src/sdk`) — une API typée et ergonomique par-dessus le
  WASM, exécutée dans un **Web Worker** pour ne pas bloquer l'UI. On récupère un
  client typé par un schéma :

  ```ts
  const db = await initWasmRedis<{ "user:1": string }>();
  await db.set("user:1", "Axel", { ex: 60 });
  const value = await db.get("user:1");
  const rows = await db.get().where(">", 35).exec(); // GET WHERE chaîné
  ```

- **UI React** (`web/src/ui`) — une interface pour explorer la base : liste
  **virtualisée** capable d'afficher des dizaines de milliers de clés, toolbar
  pour semer des données, filtrer et vider.

---

## Concepts clés à retenir

- **Un seul moteur, deux hôtes.** L'interface `Storage` découple le moteur du
  support de persistance : `FileStorage` sur disque en Go, `jsStorage`→OPFS en WASM.
- **Durabilité par AOF + snapshot.** Le journal capture chaque écriture ; le
  snapshot compacte l'état ; le restore rejoue les deux.
- **Recherche par valeur indexée.** Index inversé pour l'égalité, b-tree pour les
  plages numériques.
- **TTL à trois niveaux.** Lecture paresseuse, balayage de fond, persistance.
- **Concurrence maîtrisée.** Un mutex protège l'état ; une goroutine de fond gère
  flush/snapshot/sweep.

## Tests

Chaque brique a ses tests (`*_test.go` côté Go, `*.test.ts` côté SDK). Voir
[GETTING_STARTED.md](./GETTING_STARTED.md) pour les commandes.

## Benchmark

Rapport chiffré et reproductible : [BENCHMARK.md](./BENCHMARK.md) (latences
SET/GET p50/p95, GET filtré equals/contains/range par taille de base, temps de
restore, gain du batch, + procédure UI FPS/render granulaire).

```bash
go run ./cmd/bench                                  # tableaux de latences
go test -bench=. -benchmem -run='^$' ./internal/engine/   # débit + allocations
```

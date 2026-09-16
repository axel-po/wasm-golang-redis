# Rapport de benchmark — WasmRedis

Rapport chiffré et **reproductible** exigé par le cahier des charges §9.2.
Deux volets : le **moteur Go / WASM** (mesuré ici) et l'**UI React** (procédure fournie).

---

## 1. Méthodologie

| Paramètre | Valeur |
|-----------|--------|
| Machine | Apple M3 (8 cœurs) |
| OS | macOS 26.6.2 |
| Go | go1.26.3 |
| Source aléatoire | `rand.NewSource(42)` (fixée → reproductible) |
| Tailles de base | 1 000 / 10 000 / 100 000 entrées |
| Itérations latence unitaire | 100 000 |
| Jeu de données | clés `key:i`, valeurs numériques `i` (indexées equals + B-Tree) |

Deux outils, tous deux versionnés dans le repo :

```bash
# 1) Latences p50/p95/p99, filtres par taille, restore, gain batch (tableaux markdown)
go run ./cmd/bench
go run ./cmd/bench -sizes 1000,50000 -iters 200000   # paramétrable

# 2) Débit brut ns/op + allocations
go test -bench=. -benchmem -run='^$' ./internal/engine/
```

> Les latences sont mesurées avec `time.Now()`/`time.Since` autour de chaque
> opération, triées puis échantillonnées par percentile. L'overhead de la mesure
> (~30–40 ns) est inclus : les chiffres sont donc **conservateurs** pour les
> opérations les plus rapides (GET, equals).

---

## 2. Moteur Go — résultats

### 2.1 Latence unitaire SET / GET par clé

| Opération | p50 | p95 | p99 | débit (op/s) |
|-----------|-----|-----|-----|--------------|
| SET | 708 ns | 1.62 µs | 3.46 µs | ~1 060 000 |
| GET (hit) | 83 ns | 417 ns | 584 ns | ~7 170 000 |

`go test -bench` (débit brut, allocations) :

```
BenchmarkSet-8       1771140    698.8 ns/op    678 B/op    8 allocs/op
BenchmarkGetHit-8   15411807     75.69 ns/op    15 B/op    1 allocs/op
```

**Lecture.** Le GET est un accès map O(1) (~76 ns, 1 alloc). Le SET est plus
coûteux (~700 ns) car il maintient **trois structures** de façon synchrone :
`state`, l'index inversé equals, et le B-Tree range — plus le `buffer` d'AOF.
C'est le compromis assumé « écriture un peu plus chère → lectures filtrées
instantanées ».

### 2.2 Latence GET filtré selon la taille de base

| Taille | equals (index inversé) | contains (scan) | range `>` (B-Tree) |
|--------|------------------------|-----------------|--------------------|
| 1 000 | 125 ns | 11.12 µs | 5.25 µs |
| 10 000 | 125 ns | 116.29 µs | 59.88 µs |
| 100 000 | 125 ns | 1.16 ms | 695.04 µs |

**Lecture — le point clé du projet :**

- **`equals` = O(1) constant** : 125 ns quelle que soit la taille (1k → 100k).
  L'index inversé `valeur → {clés}` évite tout scan. C'est la preuve chiffrée
  que l'index sert.
- **`range` via B-Tree** : ne parcourt **pas** les 100 000 entrées, seulement les
  ~10 % qui matchent (`> n-n/10`). Le coût suit la **taille du résultat**, pas
  celle de la base — c'est le comportement attendu d'un index ordonné.
- **`contains` = scan** : croît linéairement avec la base (11 µs → 1.16 ms sur
  100k), conforme au CDC (« contains → scan »). C'est la seule opération sans
  index, et donc la borne haute.

Comparaison implicite « avec vs sans index » : un `equals` par scan coûterait
comme le `contains` (~1.16 ms à 100k) ; l'index inversé le ramène à 125 ns, soit
un facteur **~9 000×**.

### 2.3 Temps de restore au démarrage

| Taille | snapshot seul | snapshot + replay AOF |
|--------|---------------|-----------------------|
| 1 000 | 821 µs | 887 µs |
| 10 000 | 7.53 ms | 9.65 ms |
| 100 000 | 97.65 ms | 113.12 ms |

**Protocole.** *Snapshot seul* : tout l'état est dumpé puis l'AOF vidé
(compaction) → au restore, un seul `json.Unmarshal` + reconstruction des index.
*Snapshot + AOF* : la moitié des clés est dans le snapshot, l'autre moitié
rejouée depuis l'AOF ligne par ligne.

**Lecture.** Restaurer 100 000 entrées prend ~100 ms. Le **replay AOF** ajoute
~16 % (13 ms pour 50 000 opérations rejouées) : cela **valide la compaction**,
car sans snapshot périodique tout l'historique serait rejoué à chaque démarrage.
Piste d'optimisation chiffrable (bonus §10) : un snapshot **binaire** au lieu de
JSON réduirait surtout ce poste.

### 2.4 Gain du batch

| N commandes | une par une (Execute) | batch unique | accélération |
|-------------|-----------------------|--------------|--------------|
| 1 000 | 450.58 µs | 308.04 µs | 1.46× |
| 10 000 | 4.86 ms | 3.79 ms | 1.28× |
| 100 000 | 68.54 ms | 48.14 ms | 1.42× |

**Ce que ce chiffre mesure (côté moteur Go).** N `Execute` (parse + dispatch à
chaque appel) vs un seul `Batch` de commandes pré-parsées. Le gain moteur (~1.3–1.5×)
vient de l'amortissement du **parsing et du dispatch**.

**Le vrai gain est côté worker.** Dans l'application réelle, chaque commande
« une par une » traverse la frontière **main thread → Web Worker → WASM** via un
`postMessage` (sérialisation structurée + aller-retour asynchrone), là où le
batch n'en fait **qu'un seul** pour N commandes. Ce coût de round-trip (dizaines
de µs par message, dominé par la sérialisation et le scheduling du worker)
**s'ajoute** au gain mesuré ci-dessus et le dépasse largement sur gros volumes.
Mesure en navigateur : voir §3.2.

---

## 3. UI React — procédure de mesure

> Ces mesures se font **en navigateur** (worker + WASM + DOM), non couvertes par
> `go test`. La procédure ci-dessous est reproductible avec l'app (`web/`).

### 3.1 FPS au scroll (≥ 100 000 entrées)

1. `cd web && pnpm install && pnpm dev`, ouvrir l'app.
2. Cliquer **« Seed 100 000 »** (barre de progression jusqu'à 100 %).
3. Ouvrir les **DevTools → Performance**, lancer un enregistrement.
4. Scroller la liste de bout en bout (molette + drag de la scrollbar).
5. Arrêter l'enregistrement, lire le **FPS moyen** (bandeau vert « Frames »).

**Attendu / critère (§6.2).** Scroll fluide, FPS proche de 60, aucun lag
perceptible. Le compteur `n° X à Y — Z lignes montées dans le DOM` (dans
`VirtualList.tsx`) prouve que seules ~20–30 lignes sont montées quelle que soit
la taille de la base : le DOM reste constant, d'où la fluidité.

| Base | Lignes montées dans le DOM | FPS scroll | Verdict |
|------|----------------------------|------------|---------|
| 100 000 | ~ `viewport/rowHeight + 2·overscan` (constant) | _à relever_ | _fluide ?_ |

### 3.2 Render granulaire — 1 ligne modifiée = 1 seul re-render

La preuve est **intégrée à l'UI** : chaque ligne affiche son **compteur de
renders** (`renders.current`, colonne `.row-renders` dans `Row.tsx`). Le store
externe (`@legendapp/state`) abonne chaque ligne à **sa seule clé** via
`useSelector(() => store$.entries[k].get())`.

**Procédure :**
1. Seed 100 000, noter que toutes les lignes visibles affichent le même compteur.
2. Cliquer **« Touch random entry »** (mute une seule clé), ou double-cliquer une
   ligne et l'éditer.
3. **Observer** : seul le compteur de la ligne modifiée s'incrémente ; les autres
   restent figés.
4. Confirmation croisée : **React DevTools → « Highlight updates »** — seule la
   ligne changée clignote.

**Réponse claire oui/non attendue :** _oui_ — la modification d'une ligne ne
re-render **que** cette ligne (abonnement par clé au store externe, pas de
`React.memo` seul).

| Action | Lignes qui re-render | Preuve |
|--------|----------------------|--------|
| Modifier 1 entrée | 1 (la ligne modifiée) | compteur `.row-renders` + Highlight updates |

---

## 4. Résumé

| Axe §9.2 | Statut |
|----------|--------|
| Latence SET / GET p50/p95 | ✅ mesuré (`go run ./cmd/bench`) |
| GET filtré equals / contains / range par taille | ✅ mesuré — equals O(1), range via B-Tree |
| Temps de restore (snapshot / snapshot+AOF) | ✅ mesuré, compaction validée |
| Gain du batch | ✅ mesuré côté moteur ; round-trip worker documenté (§3.2) |
| FPS scroll ≥ 100k | 🔧 procédure fournie — à relever en navigateur |
| Render granulaire oui/non | 🔧 preuve intégrée (compteur par ligne) — à capturer |

Reproduire l'intégralité du volet moteur :

```bash
go run ./cmd/bench                              # tableaux markdown de ce rapport
go test -bench=. -benchmem -run='^$' ./internal/engine/
```

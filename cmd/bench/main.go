package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/axel-po/project-go-clone-redis/internal/command"
	"github.com/axel-po/project-go-clone-redis/internal/engine"
	"github.com/axel-po/project-go-clone-redis/internal/storage"
)

const seed = 42

func main() {
	sizesFlag := flag.String("sizes", "1000,10000,100000", "tailles de base à mesurer (séparées par des virgules)")
	iters := flag.Int("iters", 100000, "nombre d'itérations pour les latences unitaires SET/GET")
	flag.Parse()

	sizes := parseSizes(*sizesFlag)

	fmt.Println("# Résultats du benchmark — moteur WasmRedis (Go natif)")
	fmt.Println()
	fmt.Printf("Go %s · seed=%d · itérations latence unitaire=%d\n\n", goVersion(), seed, *iters)

	benchSetGetLatency(*iters)
	benchFilterLatency(sizes)
	benchRestore(sizes)
	benchBatchGain(sizes)
}

func benchSetGetLatency(iters int) {
	rng := rand.New(rand.NewSource(seed))
	db := engine.New()

	keys := make([]string, iters)
	for i := range keys {
		keys[i] = "key:" + strconv.Itoa(i)
	}

	setLat := make([]time.Duration, iters)
	for i := 0; i < iters; i++ {
		value := strconv.Itoa(rng.Intn(1_000_000))
		start := time.Now()
		db.Set(keys[i], value)
		setLat[i] = time.Since(start)
	}

	getLat := make([]time.Duration, iters)
	for i := 0; i < iters; i++ {
		k := keys[rng.Intn(iters)]
		start := time.Now()
		_, _ = db.Get(k)
		getLat[i] = time.Since(start)
	}

	fmt.Println("## Latence unitaire SET / GET par clé")
	fmt.Println()
	fmt.Println("| Opération | p50 | p95 | p99 | débit (op/s) |")
	fmt.Println("|-----------|-----|-----|-----|--------------|")
	printLatencyRow("SET", setLat)
	printLatencyRow("GET (hit)", getLat)
	fmt.Println()
}

func benchFilterLatency(sizes []int) {
	fmt.Println("## Latence GET filtré selon la taille de base")
	fmt.Println()
	fmt.Println("| Taille | equals (index inversé) | contains (scan) | range `>` (B-Tree) |")
	fmt.Println("|--------|------------------------|-----------------|--------------------|")

	for _, n := range sizes {
		db := seededEngine(n)

		equals := medianOf(200, func() { db.Filter(command.GetWhere{Op: command.OpEquals, Value: strconv.Itoa(n / 2)}) })
		contains := medianOf(50, func() { db.Filter(command.GetWhere{Op: command.OpContains, Value: "999"}) })
		rangeGT := medianOf(200, func() { db.Filter(command.GetWhere{Op: command.OpGT, Value: strconv.Itoa(n - n/10)}) })

		fmt.Printf("| %s | %s | %s | %s |\n",
			humanCount(n), fmtDur(equals), fmtDur(contains), fmtDur(rangeGT))
	}
	fmt.Println()
}

func benchRestore(sizes []int) {
	fmt.Println("## Temps de restore au démarrage")
	fmt.Println()
	fmt.Println("| Taille | snapshot seul | snapshot + replay AOF |")
	fmt.Println("|--------|---------------|-----------------------|")

	for _, n := range sizes {
		fmt.Printf("| %s | %s | %s |\n",
			humanCount(n), fmtDur(restoreSnapshotOnly(n)), fmtDur(restoreSnapshotPlusAOF(n)))
	}
	fmt.Println()
}

func restoreSnapshotOnly(n int) time.Duration {
	dir, _ := tempDir()
	store := mustStore(dir)
	db := engine.NewWithStorage(store)
	fillSequential(db, n)
	_ = db.Snapshot()

	fresh := engine.NewWithStorage(mustStore(dir))
	start := time.Now()
	_ = fresh.Restore()
	return time.Since(start)
}

func restoreSnapshotPlusAOF(n int) time.Duration {
	dir, _ := tempDir()
	store := mustStore(dir)
	db := engine.NewWithStorage(store)

	fillRange(db, 0, n/2)
	_ = db.Snapshot()
	fillRange(db, n/2, n)
	_ = db.Flush()

	fresh := engine.NewWithStorage(mustStore(dir))
	start := time.Now()
	_ = fresh.Restore()
	return time.Since(start)
}

func benchBatchGain(sizes []int) {
	fmt.Println("## Gain du batch (moteur Go)")
	fmt.Println()
	fmt.Println("> N commandes `Execute` (parse + apply à chaque appel) vs un seul `Batch` de N commandes pré-parsées.")
	fmt.Println("> Le gain côté worker s'y ajoute : 1 `postMessage` au lieu de N (voir §UI dans BENCHMARK.md).")
	fmt.Println()
	fmt.Println("| N commandes | une par une (Execute) | batch unique | accélération |")
	fmt.Println("|-------------|-----------------------|--------------|--------------|")

	for _, n := range sizes {
		raw := make([]string, n)
		parsed := make([]command.Command, n)
		for i := 0; i < n; i++ {
			raw[i] = fmt.Sprintf("SET key:%d %q", i, strconv.Itoa(i))
			parsed[i] = command.Set{Key: "key:" + strconv.Itoa(i), Value: strconv.Itoa(i)}
		}

		dbSeq := engine.New()
		startSeq := time.Now()
		for _, cmd := range raw {
			_, _ = dbSeq.Execute(cmd)
		}
		seqDur := time.Since(startSeq)

		dbBatch := engine.New()
		startBatch := time.Now()
		dbBatch.Batch(parsed)
		batchDur := time.Since(startBatch)

		fmt.Printf("| %s | %s | %s | %.2f× |\n",
			humanCount(n), fmtDur(seqDur), fmtDur(batchDur),
			float64(seqDur)/float64(batchDur))
	}
	fmt.Println()
}

func seededEngine(n int) *engine.Engine {
	db := engine.New()
	fillSequential(db, n)
	return db
}

func fillSequential(db *engine.Engine, n int) { fillRange(db, 0, n) }

func fillRange(db *engine.Engine, from, to int) {
	for i := from; i < to; i++ {
		db.Set("key:"+strconv.Itoa(i), strconv.Itoa(i))
	}
}

func medianOf(runs int, fn func()) time.Duration {
	lat := make([]time.Duration, runs)
	for i := 0; i < runs; i++ {
		start := time.Now()
		fn()
		lat[i] = time.Since(start)
	}
	sort.Slice(lat, func(i, j int) bool { return lat[i] < lat[j] })
	return lat[len(lat)/2]
}

func printLatencyRow(name string, lat []time.Duration) {
	sorted := append([]time.Duration(nil), lat...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	var total time.Duration
	for _, d := range sorted {
		total += d
	}
	throughput := float64(len(sorted)) / total.Seconds()

	fmt.Printf("| %s | %s | %s | %s | %s |\n",
		name,
		fmtDur(percentile(sorted, 0.50)),
		fmtDur(percentile(sorted, 0.95)),
		fmtDur(percentile(sorted, 0.99)),
		humanCount(int(throughput)),
	)
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func fmtDur(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%d ns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.2f µs", float64(d.Nanoseconds())/1e3)
	default:
		return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1e6)
	}
}

func humanCount(n int) string {
	s := strconv.Itoa(n)
	var out []byte
	for i, c := range []byte(s) {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, c)
	}
	return string(out)
}

func parseSizes(csv string) []int {
	parts := strings.Split(csv, ",")
	sizes := make([]int, 0, len(parts))
	for _, p := range parts {
		if n, err := strconv.Atoi(strings.TrimSpace(p)); err == nil && n > 0 {
			sizes = append(sizes, n)
		}
	}
	return sizes
}

func mustStore(dir string) *storage.FileStorage {
	store, err := storage.NewFileStorage(dir)
	if err != nil {
		panic(err)
	}
	return store
}

func tempDir() (string, error) {
	return os.MkdirTemp("", "wasmredis-bench-*")
}

func goVersion() string { return runtime.Version() }

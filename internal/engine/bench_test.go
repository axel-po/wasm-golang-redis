package engine

import (
	"strconv"
	"testing"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

func benchSeed(n int) *Engine {
	db := New()
	for i := 0; i < n; i++ {
		db.Set("key:"+strconv.Itoa(i), strconv.Itoa(i))
	}
	return db
}

func BenchmarkSet(b *testing.B) {
	db := New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Set("key:"+strconv.Itoa(i%10000), strconv.Itoa(i))
	}
}

func BenchmarkGetHit(b *testing.B) {
	db := benchSeed(10000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = db.Get("key:" + strconv.Itoa(i%10000))
	}
}

func BenchmarkFilterEquals(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		db := benchSeed(n)
		target := strconv.Itoa(n / 2)
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				db.Filter(command.GetWhere{Op: command.OpEquals, Value: target})
			}
		})
	}
}

func BenchmarkFilterContains(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		db := benchSeed(n)
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				db.Filter(command.GetWhere{Op: command.OpContains, Value: "999"})
			}
		})
	}
}

func BenchmarkFilterRange(b *testing.B) {
	for _, n := range []int{1000, 10000, 100000} {
		db := benchSeed(n)
		target := strconv.Itoa(n - n/10)
		b.Run(strconv.Itoa(n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				db.Filter(command.GetWhere{Op: command.OpGT, Value: target})
			}
		})
	}
}

func BenchmarkBatch(b *testing.B) {
	cmds := make([]command.Command, 1000)
	for i := range cmds {
		cmds[i] = command.Set{Key: "key:" + strconv.Itoa(i), Value: strconv.Itoa(i)}
	}
	db := New()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db.Batch(cmds)
	}
}

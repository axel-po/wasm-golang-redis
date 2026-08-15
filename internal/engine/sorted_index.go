package engine

import (
	"sort"

	"github.com/axel-po/project-go-clone-redis/internal/command"
	"github.com/samber/lo"
)

type numKey struct {
	n   float64
	key string
}

type sortedIndex struct {
	items []numKey
}

func newSortedIndex() *sortedIndex {
	return &sortedIndex{}
}

func (s *sortedIndex) add(key string, n float64) {
	i := sort.Search(len(s.items), func(i int) bool { return s.items[i].n >= n })
	s.items = append(s.items, numKey{})
	copy(s.items[i+1:], s.items[i:])
	s.items[i] = numKey{n: n, key: key}
}

func (s *sortedIndex) remove(key string, n float64) {
	i := sort.Search(len(s.items), func(i int) bool { return s.items[i].n >= n })
	for ; i < len(s.items) && s.items[i].n == n; i++ {
		if s.items[i].key == key {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return
		}
	}
}

func (s *sortedIndex) query(op command.Operator, target float64) []string {
	firstGE := sort.Search(len(s.items), func(i int) bool { return s.items[i].n >= target })
	firstGT := sort.Search(len(s.items), func(i int) bool { return s.items[i].n > target })

	var slice []numKey
	switch op {
	case command.OpGT:
		slice = s.items[firstGT:]
	case command.OpGTE:
		slice = s.items[firstGE:]
	case command.OpLT:
		slice = s.items[:firstGE]
	case command.OpLTE:
		slice = s.items[:firstGT]
	}

	return lo.Map(slice, func(item numKey, _ int) string { return item.key })
}

package engine

import (
	"strings"

	"github.com/axel-po/project-go-clone-redis/internal/command"
	"github.com/samber/lo"
)

type Entry struct {
	Key   string
	Value string
}

func (e *Engine) Filter(q command.GetWhere) []Entry {
	e.mu.Lock()
	defer e.mu.Unlock()

	switch q.Op {
	case command.OpEquals:
		return e.filterEquals(q.Value)
	case command.OpContains:
		return e.filterContains(q.Value)
	default:
		return e.filterRange(q.Op, q.Value)
	}
}

func (e *Engine) filterEquals(value string) []Entry {
	now := e.clock()
	var entries []Entry
	for key := range e.equalsIndex[value] {
		if rec, ok := e.state[key]; ok && !rec.expired(now) {
			entries = append(entries, Entry{Key: key, Value: value})
		}
	}
	return entries
}

func (e *Engine) filterContains(sub string) []Entry {
	now := e.clock()
	var entries []Entry
	for key, rec := range e.state {
		if !rec.expired(now) && strings.Contains(rec.value, sub) {
			entries = append(entries, Entry{Key: key, Value: rec.value})
		}
	}
	return entries
}

func (e *Engine) filterRange(op command.Operator, target string) []Entry {
	n, ok := parseNumber(target)
	if !ok {
		return nil
	}
	now := e.clock()
	keys := e.rangeIdx.query(op, n)

	return lo.FilterMap(keys, func(key string, _ int) (Entry, bool) {
		rec, ok := e.state[key]
		if !ok || rec.expired(now) {
			return Entry{}, false
		}
		return Entry{Key: key, Value: rec.value}, true
	})
}

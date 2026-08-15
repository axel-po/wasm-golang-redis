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
	entries := make([]Entry, 0, len(e.equalsIndex[value]))
	for key := range e.equalsIndex[value] {
		entries = append(entries, Entry{Key: key, Value: value})
	}
	return entries
}

func (e *Engine) filterContains(sub string) []Entry {
	var entries []Entry
	for key, value := range e.state {
		if strings.Contains(value, sub) {
			entries = append(entries, Entry{Key: key, Value: value})
		}
	}
	return entries
}

func (e *Engine) filterRange(op command.Operator, target string) []Entry {
	n, ok := parseNumber(target)
	if !ok {
		return nil
	}
	keys := e.rangeIdx.query(op, n)
	return lo.Map(keys, func(key string, _ int) Entry {
		return Entry{Key: key, Value: e.state[key]}
	})
}

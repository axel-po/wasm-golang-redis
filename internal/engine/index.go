package engine

import (
	"strconv"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

type rangeIndex interface {
	add(key string, n float64)
	remove(key string, n float64)
	query(op command.Operator, target float64) []string
}

func (e *Engine) indexAdd(key, value string) {
	set := e.equalsIndex[value]
	if set == nil {
		set = make(map[string]struct{})
		e.equalsIndex[value] = set
	}
	set[key] = struct{}{}

	if n, ok := parseNumber(value); ok {
		e.rangeIdx.add(key, n)
	}
}

func (e *Engine) indexRemove(key, value string) {
	if set := e.equalsIndex[value]; set != nil {
		delete(set, key)
		if len(set) == 0 {
			delete(e.equalsIndex, value)
		}
	}

	if n, ok := parseNumber(value); ok {
		e.rangeIdx.remove(key, n)
	}
}

func (e *Engine) rebuildIndexes() {
	e.equalsIndex = make(map[string]map[string]struct{})
	e.rangeIdx = newSortedIndex()
	for key, value := range e.state {
		e.indexAdd(key, value)
	}
}

func parseNumber(s string) (float64, bool) {
	n, err := strconv.ParseFloat(s, 64)
	return n, err == nil
}

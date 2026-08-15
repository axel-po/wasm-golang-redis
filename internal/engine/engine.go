package engine

import "sync"

type Engine struct {
	mu sync.Mutex

	state       map[string]string
	buffer      []Operation
	equalsIndex map[string]map[string]struct{}
	rangeIdx    rangeIndex

	storage Storage
	done    chan struct{}
}

func New() *Engine {
	return &Engine{
		state:       make(map[string]string),
		equalsIndex: make(map[string]map[string]struct{}),
		rangeIdx:    newBTree(btreeMinDegree),
	}
}

func NewWithStorage(store Storage) *Engine {
	e := New()
	e.storage = store
	return e
}

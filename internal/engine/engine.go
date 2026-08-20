package engine

import (
	"sync"
	"time"
)

type Engine struct {
	mu sync.Mutex

	state       map[string]record
	buffer      []Operation
	equalsIndex map[string]map[string]struct{}
	rangeIdx    rangeIndex

	storage Storage
	done    chan struct{}
	clock   func() time.Time
}

func New() *Engine {
	return &Engine{
		state:       make(map[string]record),
		equalsIndex: make(map[string]map[string]struct{}),
		rangeIdx:    newBTree(btreeMinDegree),
		clock:       time.Now,
	}
}

func NewWithStorage(store Storage) *Engine {
	e := New()
	e.storage = store
	return e
}

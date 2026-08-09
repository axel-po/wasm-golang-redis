package engine

import "sync"

type Engine struct {
	mu sync.Mutex

	state  map[string]string
	buffer []Operation

	storage Storage
	done    chan struct{}
}

func New() *Engine {
	return &Engine{state: make(map[string]string)}
}

func NewWithStorage(store Storage) *Engine {
	e := New()
	e.storage = store
	return e
}

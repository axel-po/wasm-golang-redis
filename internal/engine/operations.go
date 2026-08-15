package engine

import "fmt"

func (e *Engine) Set(key, value string) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if old, ok := e.state[key]; ok {
		e.indexRemove(key, old)
	}
	e.state[key] = value
	e.indexAdd(key, value)
	e.record(Operation{Kind: opSet, Key: key, Value: value})
}

func (e *Engine) Get(key string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	value, ok := e.state[key]
	if !ok {
		return "", fmt.Errorf("get %q: %w", key, ErrKeyNotFound)
	}
	return value, nil
}

func (e *Engine) Delete(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	old, existed := e.state[key]
	if existed {
		e.indexRemove(key, old)
	}
	delete(e.state, key)
	e.record(Operation{Kind: opDelete, Key: key})
	return existed
}

func (e *Engine) record(op Operation) {
	if e.storage == nil {
		return
	}
	e.buffer = append(e.buffer, op)
}

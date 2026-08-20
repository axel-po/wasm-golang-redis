package engine

import (
	"fmt"
	"time"
)

func (e *Engine) Set(key, value string) {
	e.setRecord(key, record{value: value})
}

func (e *Engine) SetEX(key, value string, ttl time.Duration) {
	e.setRecord(key, record{value: value, expiresAt: e.clock().Add(ttl)})
}

func (e *Engine) setRecord(key string, rec record) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if old, ok := e.state[key]; ok {
		e.indexRemove(key, old.value)
	}
	e.state[key] = rec
	e.indexAdd(key, rec.value)
	e.record(Operation{Kind: opSet, Key: key, Value: rec.value, ExpiresAt: unixNano(rec.expiresAt)})
}

func (e *Engine) Get(key string) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	rec, ok := e.state[key]
	if ok && rec.expired(e.clock()) {
		e.deleteLocked(key, rec.value)
		ok = false
	}
	if !ok {
		return "", fmt.Errorf("get %q: %w", key, ErrKeyNotFound)
	}
	return rec.value, nil
}

func (e *Engine) Delete(key string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	old, existed := e.state[key]
	if existed {
		e.deleteLocked(key, old.value)
	}
	return existed
}

func (e *Engine) deleteLocked(key, value string) {
	e.indexRemove(key, value)
	delete(e.state, key)
	e.record(Operation{Kind: opDelete, Key: key})
}

func (e *Engine) record(op Operation) {
	if e.storage == nil {
		return
	}
	e.buffer = append(e.buffer, op)
}

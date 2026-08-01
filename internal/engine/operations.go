package engine

import "fmt"

func (e *Engine) Set(key, value string) {
	e.state[key] = value
}

func (e *Engine) Get(key string) (string, error) {
	value, ok := e.state[key]
	if !ok {
		return "", fmt.Errorf("get %q: %w", key, ErrKeyNotFound)
	}
	return value, nil
}

func (e *Engine) Delete(key string) bool {
	_, existed := e.state[key]
	delete(e.state, key)
	return existed
}





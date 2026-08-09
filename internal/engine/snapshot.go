package engine

import "encoding/json"

func (e *Engine) Snapshot() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.storage == nil {
		return nil
	}

	data, err := json.Marshal(e.state)
	if err != nil {
		return err
	}
	if err := e.storage.WriteSnapshot(data); err != nil {
		return err
	}

	e.buffer = e.buffer[:0]
	return e.storage.ClearLog()
}

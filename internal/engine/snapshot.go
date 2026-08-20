package engine

import "encoding/json"

func (e *Engine) Snapshot() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.storage == nil {
		return nil
	}

	dump := make(map[string]persistedRecord, len(e.state))
	for key, rec := range e.state {
		dump[key] = persistedRecord{Value: rec.value, ExpiresAt: unixNano(rec.expiresAt)}
	}
	data, err := json.Marshal(dump)
	if err != nil {
		return err
	}
	if err := e.storage.WriteSnapshot(data); err != nil {
		return err
	}

	e.buffer = e.buffer[:0]
	return e.storage.ClearLog()
}

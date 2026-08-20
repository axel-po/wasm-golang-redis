package engine

import "encoding/json"

func (e *Engine) Restore() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.storage == nil {
		return nil
	}

	snap, err := e.storage.ReadSnapshot()
	if err != nil {
		return err
	}
	if snap != nil {
		var dump map[string]persistedRecord
		if err := json.Unmarshal(snap, &dump); err != nil {
			return err
		}
		for key, pr := range dump {
			e.state[key] = record{value: pr.Value, expiresAt: fromUnixNano(pr.ExpiresAt)}
		}
	}

	lines, err := e.storage.ReadLog()
	if err != nil {
		return err
	}
	for _, line := range lines {
		op, err := decodeOp(line)
		if err != nil {
			return err
		}
		e.applyOp(op)
	}
	e.rebuildIndexes()
	return nil
}

func (e *Engine) applyOp(op Operation) {
	switch op.Kind {
	case opSet:
		e.state[op.Key] = record{value: op.Value, expiresAt: fromUnixNano(op.ExpiresAt)}
	case opDelete:
		delete(e.state, op.Key)
	}
}

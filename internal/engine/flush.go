package engine

func (e *Engine) Flush() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.storage == nil || len(e.buffer) == 0 {
		return nil
	}

	lines := make([][]byte, 0, len(e.buffer))
	for _, op := range e.buffer {
		line, err := encodeOp(op)
		if err != nil {
			return err
		}
		lines = append(lines, line)
	}

	if err := e.storage.AppendLog(lines); err != nil {
		return err
	}

	e.buffer = e.buffer[:0]
	return nil
}

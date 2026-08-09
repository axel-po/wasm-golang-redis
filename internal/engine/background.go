package engine

import "time"

func (e *Engine) StartBackground(flushInterval, snapshotInterval time.Duration) {
	e.done = make(chan struct{})
	go e.loop(flushInterval, snapshotInterval)
}

func (e *Engine) loop(flushInterval, snapshotInterval time.Duration) {
	flushTicker := time.NewTicker(flushInterval)
	snapshotTicker := time.NewTicker(snapshotInterval)
	defer flushTicker.Stop()
	defer snapshotTicker.Stop()

	for {
		select {
		case <-flushTicker.C:
			_ = e.Flush()
		case <-snapshotTicker.C:
			_ = e.Snapshot()
		case <-e.done:
			return
		}
	}
}

func (e *Engine) Close() error {
	if e.done != nil {
		close(e.done)
		e.done = nil
	}
	return e.Flush()
}

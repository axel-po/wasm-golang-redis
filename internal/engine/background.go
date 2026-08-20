package engine

import "time"

func (e *Engine) StartBackground(flushInterval, snapshotInterval, sweepInterval time.Duration) {
	e.done = make(chan struct{})
	go e.loop(flushInterval, snapshotInterval, sweepInterval)
}

func (e *Engine) loop(flushInterval, snapshotInterval, sweepInterval time.Duration) {
	flushTicker := time.NewTicker(flushInterval)
	snapshotTicker := time.NewTicker(snapshotInterval)
	sweepTicker := time.NewTicker(sweepInterval)
	defer flushTicker.Stop()
	defer snapshotTicker.Stop()
	defer sweepTicker.Stop()

	for {
		select {
		case <-flushTicker.C:
			_ = e.Flush()
		case <-snapshotTicker.C:
			_ = e.Snapshot()
		case <-sweepTicker.C:
			e.sweepExpired()
		case <-e.done:
			return
		}
	}
}

func (e *Engine) sweepExpired() {
	e.mu.Lock()
	defer e.mu.Unlock()

	now := e.clock()
	for key, rec := range e.state {
		if rec.expired(now) {
			e.deleteLocked(key, rec.value)
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

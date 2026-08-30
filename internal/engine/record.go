package engine

import "time"

type record struct {
	value     string
	expiresAt time.Time
}

func (r record) expired(now time.Time) bool {
	return !r.expiresAt.IsZero() && !now.Before(r.expiresAt)
}

func (r record) expiresAtMillis() int64 {
	if r.expiresAt.IsZero() {
		return 0
	}
	return r.expiresAt.UnixMilli()
}

type persistedRecord struct {
	Value     string `json:"value"`
	ExpiresAt int64  `json:"expires_at,omitempty"`
}

func unixNano(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano()
}

func fromUnixNano(n int64) time.Time {
	if n == 0 {
		return time.Time{}
	}
	return time.Unix(0, n)
}

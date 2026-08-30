package engine

import (
	"errors"
	"testing"
	"time"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

func fixedClock(t *time.Time) func() time.Time {
	return func() time.Time { return *t }
}

func TestTTLExpirationLazy(t *testing.T) {
	now := time.Unix(1000, 0)
	db := New()
	db.clock = fixedClock(&now)

	db.SetEX("k", "v", 60*time.Second)

	if v, err := db.Get("k"); err != nil || v != "v" {
		t.Fatalf("avant expiration: got %q, %v", v, err)
	}

	now = time.Unix(1061, 0)
	if _, err := db.Get("k"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("après expiration: erreur = %v, attendu %v", err, ErrKeyNotFound)
	}
}

func TestTTLBalayageActif(t *testing.T) {
	now := time.Unix(1000, 0)
	db := New()
	db.clock = fixedClock(&now)

	db.SetEX("temp", "30", 10*time.Second)
	db.Set("permanent", "x")

	now = time.Unix(2000, 0)
	db.sweepExpired()

	if _, ok := db.state["temp"]; ok {
		t.Error("temp toujours dans le state après balayage")
	}

	if got := db.Filter(command.GetWhere{Op: command.OpGTE, Value: "0"}); len(got) != 0 {
		t.Errorf("temp toujours indexée après balayage: %v", got)
	}
	if _, err := db.Get("permanent"); err != nil {
		t.Errorf("permanent supprimée à tort: %v", err)
	}
}

func TestTTLFiltreIgnoreExpirees(t *testing.T) {
	now := time.Unix(1000, 0)
	db := New()
	db.clock = fixedClock(&now)

	db.SetEX("a", "30", 10*time.Second)

	now = time.Unix(2000, 0)
	if got := db.Filter(command.GetWhere{Op: command.OpEquals, Value: "30"}); len(got) != 0 {
		t.Errorf("equals renvoie une entrée expirée: %v", got)
	}
}

func TestTTLSurvitAuRestore(t *testing.T) {
	dir := t.TempDir()
	now := time.Unix(1000, 0)

	db := NewWithStorage(newStore(t, dir))
	db.clock = fixedClock(&now)
	db.SetEX("k", "v", 60*time.Second)
	if err := db.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	db2 := NewWithStorage(newStore(t, dir))
	db2.clock = fixedClock(&now)
	if err := db2.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	now = time.Unix(1061, 0)
	if _, err := db2.Get("k"); !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("après restore + expiration: erreur = %v, attendu %v", err, ErrKeyNotFound)
	}
}

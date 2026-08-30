package engine

import (
	"errors"
	"testing"

	"github.com/axel-po/project-go-clone-redis/internal/storage"
)

func TestRestoreDepuisJournal(t *testing.T) {
	dir := t.TempDir()

	store := newStore(t, dir)
	db := NewWithStorage(store)
	db.Set("name", "matt")
	db.Set("age", "30")
	db.Delete("age")
	if err := db.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	db2 := NewWithStorage(newStore(t, dir))
	if err := db2.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if got, err := db2.Get("name"); err != nil || got != "matt" {
		t.Errorf(`Get("name") = %q, %v ; attendu "matt", nil`, got, err)
	}
	if _, err := db2.Get("age"); !errors.Is(err, ErrKeyNotFound) {
		t.Errorf(`Get("age") = %v ; attendu absente (delete rejoué)`, err)
	}
}

func TestRestoreSnapshotPuisJournal(t *testing.T) {
	dir := t.TempDir()

	db := NewWithStorage(newStore(t, dir))
	db.Set("a", "1")
	if err := db.Snapshot(); err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	db.Set("b", "2")
	if err := db.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	db2 := NewWithStorage(newStore(t, dir))
	if err := db2.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	if got, _ := db2.Get("a"); got != "1" {
		t.Errorf(`Get("a") = %q ; attendu "1" (via snapshot)`, got)
	}
	if got, _ := db2.Get("b"); got != "2" {
		t.Errorf(`Get("b") = %q ; attendu "2" (via journal)`, got)
	}
}

func newStore(t *testing.T, dir string) *storage.FileStorage {
	t.Helper()
	store, err := storage.NewFileStorage(dir)
	if err != nil {
		t.Fatalf("NewFileStorage: %v", err)
	}
	return store
}

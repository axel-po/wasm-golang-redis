package engine

import (
	"errors"
	"testing"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

func TestBatch(t *testing.T) {
	db := New()

	results := db.Batch([]command.Command{
		command.Set{Key: "a", Value: "1"},
		command.Get{Key: "a"},
		command.Get{Key: "absente"},
		command.Delete{Key: "a"},
	})

	if len(results) != 4 {
		t.Fatalf("len = %d, attendu 4 (résultats alignés sur les commandes)", len(results))
	}
	if results[0].Err != nil {
		t.Errorf("set: erreur inattendue %v", results[0].Err)
	}
	if results[1].Result.Value != "1" {
		t.Errorf("get = %q, attendu %q", results[1].Result.Value, "1")
	}
	if !errors.Is(results[2].Err, ErrKeyNotFound) {
		t.Errorf("get absente: erreur = %v, attendu %v", results[2].Err, ErrKeyNotFound)
	}
	if !results[3].Result.Found {
		t.Error("delete: Found = false, attendu true")
	}
}

func TestBatchVide(t *testing.T) {
	db := New()

	if results := db.Batch(nil); len(results) != 0 {
		t.Errorf("Batch(nil) = %d résultats, attendu 0", len(results))
	}
}

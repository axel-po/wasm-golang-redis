package engine

import (
	"errors"
	"testing"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

func TestSetGet(t *testing.T) {
	db := New()

	db.Set("name", "matt")

	got, err := db.Get("name")
	if err != nil {
		t.Fatalf("Get erreur inattendue: %v", err)
	}
	if got != "matt" {
		t.Errorf("Get = %q, attendu %q", got, "matt")
	}
}

func TestSetEcrase(t *testing.T) {
	db := New()

	db.Set("name", "matt")
	db.Set("name", "alice")

	got, _ := db.Get("name")
	if got != "alice" {
		t.Errorf("Get = %q, attendu %q après écrasement", got, "alice")
	}
}

func TestGetCleAbsente(t *testing.T) {
	db := New()

	_, err := db.Get("inconnue")
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("Get erreur = %v, attendu %v", err, ErrKeyNotFound)
	}
}

func TestDelete(t *testing.T) {
	db := New()
	db.Set("name", "matt")

	if existed := db.Delete("name"); !existed {
		t.Error("Delete = false, attendu true sur une clé existante")
	}
	if _, err := db.Get("name"); !errors.Is(err, ErrKeyNotFound) {
		t.Errorf("Get après Delete = %v, attendu %v", err, ErrKeyNotFound)
	}
}

func TestDeleteCleAbsente(t *testing.T) {
	db := New()

	if existed := db.Delete("inconnue"); existed {
		t.Error("Delete = true, attendu false sur une clé absente")
	}
}

func TestApply(t *testing.T) {
	tests := []struct {
		name    string
		seed    map[string]string
		cmd     command.Command
		want    Result
		wantErr error
	}{
		{
			name: "set",
			cmd:  command.Set{Key: "k", Value: "v"},
			want: Result{},
		},
		{
			name: "get trouve",
			seed: map[string]string{"k": "v"},
			cmd:  command.Get{Key: "k"},
			want: Result{Value: "v", Found: true},
		},
		{
			name:    "get absent",
			cmd:     command.Get{Key: "k"},
			wantErr: ErrKeyNotFound,
		},
		{
			name: "delete effectif",
			seed: map[string]string{"k": "v"},
			cmd:  command.Delete{Key: "k"},
			want: Result{Found: true},
		},
		{
			name: "delete sans effet",
			cmd:  command.Delete{Key: "k"},
			want: Result{Found: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := New()
			for k, v := range tt.seed {
				db.Set(k, v)
			}

			got, err := db.Apply(tt.cmd)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Apply erreur = %v, attendu %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Apply erreur inattendue: %v", err)
			}
			if got != tt.want {
				t.Errorf("Apply = %+v, attendu %+v", got, tt.want)
			}
		})
	}
}

func TestExecute(t *testing.T) {
	db := New()

	if _, err := db.Execute(`SET msg "hello world"`); err != nil {
		t.Fatalf("Execute SET erreur: %v", err)
	}

	got, err := db.Execute(`GET msg`)
	if err != nil {
		t.Fatalf("Execute GET erreur: %v", err)
	}
	if got.Value != "hello world" {
		t.Errorf("Execute GET = %q, attendu %q", got.Value, "hello world")
	}
}

func TestExecutePropageErreurParsing(t *testing.T) {
	db := New()

	_, err := db.Execute(`FOO bar`)
	if !errors.Is(err, command.ErrUnknownCommand) {
		t.Fatalf("Execute erreur = %v, attendu %v", err, command.ErrUnknownCommand)
	}
}

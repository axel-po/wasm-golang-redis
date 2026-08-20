package command

import (
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestParseSetEX(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Command
		wantErr error
	}{
		{name: "sans ttl", input: `SET name matt`, want: Set{Key: "name", Value: "matt"}},
		{name: "avec ex", input: `SET name matt EX 60`, want: Set{Key: "name", Value: "matt", TTL: 60 * time.Second}},
		{name: "ex casse melangee", input: `SET k v ex 5`, want: Set{Key: "k", Value: "v", TTL: 5 * time.Second}},

		{name: "ex sans nombre", input: `SET k v EX abc`, wantErr: ErrInvalidTTL},
		{name: "ex negatif", input: `SET k v EX -1`, wantErr: ErrInvalidTTL},
		{name: "ex zero", input: `SET k v EX 0`, wantErr: ErrInvalidTTL},
		{name: "mot cle inconnu", input: `SET k v FOO 60`, wantErr: ErrWrongArgCount},
		{name: "trop d'arguments", input: `SET k v EX 60 extra`, wantErr: ErrWrongArgCount},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Parse(%q) erreur = %v, attendu %v", tt.input, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse(%q) erreur inattendue: %v", tt.input, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse(%q) = %#v, attendu %#v", tt.input, got, tt.want)
			}
		})
	}
}

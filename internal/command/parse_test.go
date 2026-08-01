package command

import (
	"errors"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Command
		wantErr error
	}{
		{name: "set simple", input: `SET name matt`, want: Set{Key: "name", Value: "matt"}},
		{name: "set avec espaces", input: `SET msg "hello world"`, want: Set{Key: "msg", Value: "hello world"}},
		{name: "set valeur vide", input: `SET k ""`, want: Set{Key: "k", Value: ""}},
		{name: "get", input: `GET name`, want: Get{Key: "name"}},
		{name: "delete", input: `DELETE name`, want: Delete{Key: "name"}},

		{name: "verbe minuscule", input: `set name matt`, want: Set{Key: "name", Value: "matt"}},
		{name: "verbe casse melangee", input: `SeT name matt`, want: Set{Key: "name", Value: "matt"}},
		{name: "cle et valeur gardent leur casse", input: `SET Name Matt`, want: Set{Key: "Name", Value: "Matt"}},

		{name: "entree vide", input: ``, wantErr: ErrEmptyInput},
		{name: "que des espaces", input: `   `, wantErr: ErrEmptyInput},
		{name: "commande inconnue", input: `FOO bar`, wantErr: ErrUnknownCommand},
		{name: "set sans valeur", input: `SET k`, wantErr: ErrWrongArgCount},
		{name: "set argument en trop", input: `SET k v extra`, wantErr: ErrWrongArgCount},
		{name: "get sans cle", input: `GET`, wantErr: ErrWrongArgCount},
		{name: "get argument en trop", input: `GET a b`, wantErr: ErrWrongArgCount},
		{name: "delete sans cle", input: `DELETE`, wantErr: ErrWrongArgCount},
		{name: "guillemet non ferme", input: `SET k "oops`, wantErr: ErrUnclosedQuote},
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

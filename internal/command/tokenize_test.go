package command

import (
	"errors"
	"reflect"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr error
	}{
		{name: "simple", input: `SET name matt`, want: []string{"SET", "name", "matt"}},
		{name: "valeur entre guillemets", input: `SET msg "hello world"`, want: []string{"SET", "msg", "hello world"}},
		{name: "valeur vide", input: `SET k ""`, want: []string{"SET", "k", ""}},
		{name: "espaces multiples", input: `SET   k    v`, want: []string{"SET", "k", "v"}},
		{name: "espaces autour", input: `  GET k  `, want: []string{"GET", "k"}},
		{name: "tabulation", input: "GET\tk", want: []string{"GET", "k"}},
		{name: "utf8", input: `SET clé "été"`, want: []string{"SET", "clé", "été"}},
		{name: "entree vide", input: ``, want: nil},
		{name: "que des espaces", input: `   `, want: nil},
		{name: "guillemet non ferme", input: `SET k "oops`, wantErr: ErrUnclosedQuote},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tokenize(tt.input)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("erreur = %v, attendu %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("erreur inattendue: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tokenize(%q) = %#v, attendu %#v", tt.input, got, tt.want)
			}
		})
	}
}

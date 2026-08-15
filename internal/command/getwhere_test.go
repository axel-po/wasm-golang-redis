package command

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseGetWhere(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Command
		wantErr error
	}{
		{name: "equals", input: `GET WHERE equals 30`, want: GetWhere{Op: OpEquals, Value: "30"}},
		{name: "contains", input: `GET WHERE contains ma`, want: GetWhere{Op: OpContains, Value: "ma"}},
		{name: "gt", input: `GET WHERE > 18`, want: GetWhere{Op: OpGT, Value: "18"}},
		{name: "gte", input: `GET WHERE >= 18`, want: GetWhere{Op: OpGTE, Value: "18"}},
		{name: "lt", input: `GET WHERE < 18`, want: GetWhere{Op: OpLT, Value: "18"}},
		{name: "lte", input: `GET WHERE <= 18`, want: GetWhere{Op: OpLTE, Value: "18"}},
		{name: "where minuscule", input: `get where equals 30`, want: GetWhere{Op: OpEquals, Value: "30"}},
		{name: "valeur avec espaces", input: `GET WHERE contains "hello world"`, want: GetWhere{Op: OpContains, Value: "hello world"}},

		{name: "get simple inchangé", input: `GET name`, want: Get{Key: "name"}},

		{name: "operateur inconnu", input: `GET WHERE foo 1`, wantErr: ErrUnknownOperator},
		{name: "manque la valeur", input: `GET WHERE equals`, wantErr: ErrWrongArgCount},
		{name: "trop d'arguments", input: `GET WHERE equals 1 2`, wantErr: ErrWrongArgCount},
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

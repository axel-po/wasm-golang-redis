package engine

import (
	"sort"
	"testing"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

func keysOf(entries []Entry) []string {
	keys := make([]string, 0, len(entries))
	for _, e := range entries {
		keys = append(keys, e.Key)
	}
	sort.Strings(keys)
	return keys
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func seed(db *Engine) {
	db.Set("alice", "30")
	db.Set("bob", "30")
	db.Set("carol", "42")
	db.Set("dave", "18")
	db.Set("erin", "matthieu")
}

func TestFilterEquals(t *testing.T) {
	db := New()
	seed(db)

	got := keysOf(db.Filter(command.GetWhere{Op: command.OpEquals, Value: "30"}))
	if want := []string{"alice", "bob"}; !equalStrings(got, want) {
		t.Errorf("equals 30 = %v, attendu %v", got, want)
	}
}

func TestFilterContains(t *testing.T) {
	db := New()
	seed(db)

	got := keysOf(db.Filter(command.GetWhere{Op: command.OpContains, Value: "matt"}))
	if want := []string{"erin"}; !equalStrings(got, want) {
		t.Errorf("contains matt = %v, attendu %v", got, want)
	}
}

func TestFilterRange(t *testing.T) {
	tests := []struct {
		op   command.Operator
		val  string
		want []string
	}{
		{command.OpGT, "30", []string{"carol"}},
		{command.OpGTE, "30", []string{"alice", "bob", "carol"}},
		{command.OpLT, "30", []string{"dave"}},
		{command.OpLTE, "30", []string{"alice", "bob", "dave"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.op), func(t *testing.T) {
			db := New()
			seed(db)

			got := keysOf(db.Filter(command.GetWhere{Op: tt.op, Value: tt.val}))
			if !equalStrings(got, tt.want) {
				t.Errorf("%s %s = %v, attendu %v", tt.op, tt.val, got, tt.want)
			}
		})
	}
}

func TestFilterApresMiseAJour(t *testing.T) {
	db := New()
	db.Set("alice", "30")
	db.Set("alice", "40")

	if got := keysOf(db.Filter(command.GetWhere{Op: command.OpEquals, Value: "30"})); len(got) != 0 {
		t.Errorf("equals 30 = %v, attendu vide après changement", got)
	}
	if got := keysOf(db.Filter(command.GetWhere{Op: command.OpEquals, Value: "40"})); !equalStrings(got, []string{"alice"}) {
		t.Errorf("equals 40 = %v, attendu [alice]", got)
	}

	db.Delete("alice")
	if got := db.Filter(command.GetWhere{Op: command.OpGTE, Value: "0"}); len(got) != 0 {
		t.Errorf("range après delete = %v, attendu vide", got)
	}
}

func TestFilterApresRestore(t *testing.T) {
	dir := t.TempDir()

	db := NewWithStorage(newStore(t, dir))
	seed(db)
	if err := db.Flush(); err != nil {
		t.Fatalf("Flush: %v", err)
	}

	db2 := NewWithStorage(newStore(t, dir))
	if err := db2.Restore(); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	got := keysOf(db2.Filter(command.GetWhere{Op: command.OpGTE, Value: "30"}))
	if want := []string{"alice", "bob", "carol"}; !equalStrings(got, want) {
		t.Errorf("range après restore = %v, attendu %v", got, want)
	}
}

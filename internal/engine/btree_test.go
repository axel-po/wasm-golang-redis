package engine

import (
	"strconv"
	"testing"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

func TestBTreeRangeSurGrosVolume(t *testing.T) {
	db := New()
	const n = 1000
	for i := 0; i < n; i++ {
		db.Set("k"+strconv.Itoa(i), strconv.Itoa(i))
	}

	cases := []struct {
		op   command.Operator
		val  string
		want int
	}{
		{command.OpGTE, "990", 10}, // 990..999
		{command.OpGT, "990", 9},   // 991..999
		{command.OpLT, "5", 5},     // 0..4
		{command.OpLTE, "5", 6},    // 0..5
		{command.OpGTE, "0", n},    // tout
	}

	for _, c := range cases {
		got := db.Filter(command.GetWhere{Op: c.op, Value: c.val})
		if len(got) != c.want {
			t.Errorf("%s %s = %d entrées, attendu %d", c.op, c.val, len(got), c.want)
		}
	}
}

func TestBTreeValeursPartageesEtSuppression(t *testing.T) {
	db := New()
	db.Set("a", "30")
	db.Set("b", "30")
	db.Set("c", "30")

	if got := db.Filter(command.GetWhere{Op: command.OpGTE, Value: "30"}); len(got) != 3 {
		t.Fatalf("gte 30 = %d, attendu 3", len(got))
	}

	db.Delete("b")
	got := keysOf(db.Filter(command.GetWhere{Op: command.OpGTE, Value: "30"}))
	if want := []string{"a", "c"}; !equalStrings(got, want) {
		t.Errorf("gte 30 après delete = %v, attendu %v", got, want)
	}
}

package engine

import (
	"math"
	"sort"

	"github.com/axel-po/project-go-clone-redis/internal/command"
)

const btreeMinDegree = 4

type btreeEntry struct {
	n    float64
	keys map[string]struct{}
}

type btreeNode struct {
	leaf     bool
	entries  []btreeEntry
	children []*btreeNode
}

type bTree struct {
	root   *btreeNode
	degree int
}

func newBTree(degree int) *bTree {
	return &bTree{root: &btreeNode{leaf: true}, degree: degree}
}

func (t *bTree) add(key string, n float64) {
	if entry := t.search(n); entry != nil {
		entry.keys[key] = struct{}{}
		return
	}

	root := t.root
	if len(root.entries) == 2*t.degree-1 {
		newRoot := &btreeNode{children: []*btreeNode{root}}
		newRoot.splitChild(0, t.degree)
		t.root = newRoot
		newRoot.insertNonFull(n, key, t.degree)
		return
	}
	root.insertNonFull(n, key, t.degree)
}

func (t *bTree) remove(key string, n float64) {
	if entry := t.search(n); entry != nil {
		delete(entry.keys, key)
	}
}

func (t *bTree) query(op command.Operator, target float64) []string {
	var low, high float64
	switch op {
	case command.OpGT, command.OpGTE:
		low, high = target, math.Inf(1)
	case command.OpLT, command.OpLTE:
		low, high = math.Inf(-1), target
	default:
		return nil
	}

	var out []string
	t.root.rangeScan(low, high, op, target, &out)
	return out
}

func (t *bTree) search(n float64) *btreeEntry {
	return t.root.search(n)
}

func (node *btreeNode) search(n float64) *btreeEntry {
	i := sort.Search(len(node.entries), func(i int) bool { return node.entries[i].n >= n })
	if i < len(node.entries) && node.entries[i].n == n {
		return &node.entries[i]
	}
	if node.leaf {
		return nil
	}
	return node.children[i].search(n)
}

func (node *btreeNode) insertNonFull(n float64, key string, degree int) {
	i := sort.Search(len(node.entries), func(i int) bool { return node.entries[i].n >= n })

	if node.leaf {
		entry := btreeEntry{n: n, keys: map[string]struct{}{key: {}}}
		node.entries = append(node.entries, btreeEntry{})
		copy(node.entries[i+1:], node.entries[i:])
		node.entries[i] = entry
		return
	}

	if len(node.children[i].entries) == 2*degree-1 {
		node.splitChild(i, degree)
		if n > node.entries[i].n {
			i++
		}
	}
	node.children[i].insertNonFull(n, key, degree)
}

func (parent *btreeNode) splitChild(i, degree int) {
	full := parent.children[i]
	mid := degree - 1

	right := &btreeNode{leaf: full.leaf}
	right.entries = append(right.entries, full.entries[mid+1:]...)
	median := full.entries[mid]
	full.entries = full.entries[:mid]

	if !full.leaf {
		right.children = append(right.children, full.children[mid+1:]...)
		full.children = full.children[:mid+1]
	}

	parent.entries = append(parent.entries, btreeEntry{})
	copy(parent.entries[i+1:], parent.entries[i:])
	parent.entries[i] = median

	parent.children = append(parent.children, nil)
	copy(parent.children[i+2:], parent.children[i+1:])
	parent.children[i+1] = right
}

func (node *btreeNode) rangeScan(low, high float64, op command.Operator, target float64, out *[]string) {
	for i := 0; i <= len(node.entries); i++ {
		if !node.leaf {
			leftBound, rightBound := math.Inf(-1), math.Inf(1)
			if i > 0 {
				leftBound = node.entries[i-1].n
			}
			if i < len(node.entries) {
				rightBound = node.entries[i].n
			}
			if leftBound <= high && rightBound >= low {
				node.children[i].rangeScan(low, high, op, target, out)
			}
		}

		if i < len(node.entries) && matchNumber(op, node.entries[i].n, target) {
			for k := range node.entries[i].keys {
				*out = append(*out, k)
			}
		}
	}
}

func matchNumber(op command.Operator, n, target float64) bool {
	switch op {
	case command.OpGT:
		return n > target
	case command.OpGTE:
		return n >= target
	case command.OpLT:
		return n < target
	case command.OpLTE:
		return n <= target
	default:
		return false
	}
}

package main

import (
	"sort"

	"github.com/shaia/succincter"
)

// Louds is a trie stored as two bit vectors and one byte per node, in the
// level-order unary degree sequence (LOUDS) layout.
//
// Nodes are numbered breadth first, root = 0. Children of a node get
// consecutive ids, in label order.
//
//	shape   for every node in id order: one 0 per child, then a 1.
//	        Node k's block is the bits after the k-th 1, up to the (k+1)-th.
//	        The j-th 0 overall stands for node j, so the children of k are
//	        the ids (start-k+1) .. (start-k+degree), where start is where
//	        the block begins.
//	labels  labels[id] is the byte on the edge into node id.
//	final   bit id is 1 when a key ends at node id. Final(id) numbers the
//	        keys in breadth-first order, which is why NewLouds returns order.
type Louds struct {
	shape  succincter.RankSelector
	final  succincter.RankSelector
	labels []byte
	nodes  int
	keys   int
}

// Bits is how a Louds bit vector is built: the plain Succincter, or RRR.
type Bits func([]bool) succincter.RankSelector

// PlainBits builds the uncompressed Succincter (~2.06 bits per bit).
func PlainBits(b []bool) succincter.RankSelector {
	return succincter.NewSuccincter(b, func(x bool) bool { return x })
}

// RRRBits builds the RRR encoding, which pays off on sparse vectors.
func RRRBits(b []bool) succincter.RankSelector {
	return succincter.NewRRR(b, func(x bool) bool { return x })
}

// NewLouds builds the trie of keys, which must be sorted and unique.
// order[r] is the index in keys of the key whose final node has rank r,
// so callers can lay out per-key data in the trie's own key order.
func NewLouds(keys []string, shapeBits, finalBits Bits) (*Louds, []int32) {
	// A node is the run of sorted keys sharing its first depth bytes.
	type span struct{ lo, hi, depth int32 }
	queue := []span{{0, int32(len(keys)), 0}}
	var shape, final []bool
	var labels []byte
	var order []int32
	labels = append(labels, 0) // the root has no incoming edge

	for head := 0; head < len(queue); head++ {
		s := queue[head]
		lo := s.lo
		// Sorted order puts the key that ends here first.
		if lo < s.hi && int32(len(keys[lo])) == s.depth {
			final = append(final, true)
			order = append(order, lo)
			lo++
		} else {
			final = append(final, false)
		}
		for lo < s.hi {
			c := keys[lo][s.depth]
			hi := lo + 1
			for hi < s.hi && keys[hi][s.depth] == c {
				hi++
			}
			queue = append(queue, span{lo, hi, s.depth + 1})
			labels = append(labels, c)
			shape = append(shape, false)
			lo = hi
		}
		shape = append(shape, true)
	}
	t := &Louds{
		shape:  shapeBits(shape),
		final:  finalBits(final),
		labels: append([]byte(nil), labels...), // without append's spare capacity
		nodes:  len(queue),
		keys:   len(order),
	}
	return t, order
}

// children returns the id of node k's first child and how many it has.
func (t *Louds) children(k int) (first, n int) {
	start := 0
	if k > 0 {
		start = t.shape.Select(k) + 1
	}
	end := t.shape.Select(k + 1)
	return start - k + 1, end - start
}

// child follows the edge labelled c out of node k, or returns -1.
func (t *Louds) child(k int, c byte) int {
	first, n := t.children(k)
	ls := t.labels[first : first+n]
	i := sort.Search(n, func(i int) bool { return ls[i] >= c })
	if i < n && ls[i] == c {
		return first + i
	}
	return -1
}

// Find returns the node reached by spelling prefix from the root, or -1.
func (t *Louds) Find(prefix string) int {
	k := 0
	for i := 0; i < len(prefix) && k >= 0; i++ {
		k = t.child(k, prefix[i])
	}
	return k
}

// IsFinal reports whether a key ends at node k.
func (t *Louds) IsFinal(k int) bool {
	return t.final.Rank(k+1) != t.final.Rank(k)
}

// KeyID is the rank of a final node among the final nodes.
func (t *Louds) KeyID(k int) int { return t.final.Rank(k) }

// Ranges calls fn once per level below node k with the key ids [lo, hi) that
// end on that level. Breadth-first numbering makes a subtree's nodes on any
// one level a run of consecutive ids, and their children the next run, so
// a whole subtree costs two selects and two ranks per level, not per node.
func (t *Louds) Ranges(k int, fn func(lo, hi int)) {
	a, b := k, k+1 // the subtree's nodes on this level are [a, b)
	for a < b {
		if lo, hi := t.final.Rank(a), t.final.Rank(b); lo < hi {
			fn(lo, hi)
		}
		// Blocks a..b-1 run from just after the a-th 1 to the b-th 1, and
		// hold b-1-a of their own terminating 1s; the rest are children.
		start := 0
		if a > 0 {
			start = t.shape.Select(a) + 1
		}
		end := t.shape.Select(b)
		first := start - a + 1
		a, b = first, first+(end-start)-(b-1-a)
	}
}

// Each calls fn for every key under node k, in sorted order, with the key's
// bytes after the prefix that led to k. It stops when fn returns false.
func (t *Louds) Each(k int, fn func(id int, suffix []byte) bool) {
	var buf []byte
	var walk func(k int) bool
	walk = func(k int) bool {
		if t.IsFinal(k) && !fn(t.KeyID(k), buf) {
			return false
		}
		first, n := t.children(k)
		for c := first; c < first+n; c++ {
			buf = append(buf, t.labels[c])
			ok := walk(c)
			buf = buf[:len(buf)-1]
			if !ok {
				return false
			}
		}
		return true
	}
	walk(k)
}

// parent returns the parent of node c > 0. Node c is the c-th 0 in shape,
// and its parent p is the number of 1s before that 0, so the 0 sits at
// c-1+p. Succincter selects 1s only, so p is found by binary search on
// Rank: it is the smallest p for which the first c+p bits hold only p ones
// (and therefore c zeros). Before that point they hold more than p.
func (t *Louds) parent(c int) int {
	lo, hi := 0, c-1
	for lo < hi {
		p := (lo + hi) / 2
		if t.shape.Rank(c+p) > p {
			lo = p + 1
		} else {
			hi = p
		}
	}
	return lo
}

// Key spells key id back out of the trie by walking from its final node to
// the root. This is how the trie stores the text, not just indexes it.
func (t *Louds) Key(id int, buf []byte) []byte {
	buf = buf[:0]
	for k := t.final.Select(id + 1); k > 0; k = t.parent(k) {
		buf = append(buf, t.labels[k])
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return buf
}

// Nodes is the node count, root included.
func (t *Louds) Nodes() int { return t.nodes }

// Keys is the number of keys stored.
func (t *Louds) Keys() int { return t.keys }

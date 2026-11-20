package main

// pnode is the textbook trie: one heap object per node, children found by
// binary search over a sorted label slice. It is here to be measured.
type pnode struct {
	labels []byte
	kids   []*pnode
	final  bool
}

// newPointerTrie builds the trie of sorted, unique keys.
func newPointerTrie(keys []string) (*pnode, int) {
	nodes := 0
	var build func(keys []string, depth int) *pnode
	build = func(keys []string, depth int) *pnode {
		nodes++
		n := &pnode{}
		if len(keys) > 0 && len(keys[0]) == depth {
			n.final = true
			keys = keys[1:]
		}
		for len(keys) > 0 {
			c := keys[0][depth]
			j := 1
			for j < len(keys) && keys[j][depth] == c {
				j++
			}
			n.labels = append(n.labels, c)
			n.kids = append(n.kids, build(keys[:j], depth+1))
			keys = keys[j:]
		}
		return n
	}
	return build(keys, 0), nodes
}

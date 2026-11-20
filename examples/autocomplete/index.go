package main

import (
	"encoding/binary"
	"math/bits"
	"sort"
	"strings"
)

// Entry is one ICD-10-CM code and its description.
type Entry struct {
	Code, Desc string
}

// Result is one completion.
type Result struct {
	Code, Desc string
}

// Searcher answers "every description containing a word that starts with
// each typed token", best first. Both implementations below agree on the
// answer exactly; they differ only in how they store the vocabulary.
type Searcher interface {
	Search(query string, k int) []Result
}

// tokens lower-cases s and splits it on anything that is not a letter or a
// digit. ICD-10-CM descriptions are plain ASCII.
func tokens(s string) []string {
	return strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
}

// rankOrder sorts descriptions shortest first, then alphabetically. It is
// the order a LOUDS trie numbers its keys in (breadth first is length
// first), and it doubles as the ranking: a shorter description is the more
// general diagnosis, so it is offered first.
func rankOrder(entries []Entry) []Entry {
	out := append([]Entry(nil), entries...)
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Desc, out[j].Desc
		if len(a) != len(b) {
			return len(a) < len(b)
		}
		return a < b
	})
	return out
}

// wordPostings maps each distinct word to the sorted ids of the
// descriptions that contain it. ids index the rank-ordered entries.
func wordPostings(ranked []Entry) ([]string, [][]uint32) {
	m := map[string][]uint32{}
	for id, e := range ranked {
		seen := map[string]bool{}
		for _, w := range tokens(e.Desc) {
			if !seen[w] {
				seen[w] = true
				m[w] = append(m[w], uint32(id))
			}
		}
	}
	words := make([]string, 0, len(m))
	for w := range m {
		words = append(words, w)
	}
	sort.Strings(words)
	lists := make([][]uint32, len(words))
	for i, w := range words {
		lists[i] = m[w]
	}
	return words, lists
}

// bitset is one bit per description, reused across keystrokes.
type bitset []uint64

func (b bitset) clear() {
	for i := range b {
		b[i] = 0
	}
}

// intersect keeps in acc only the ids also in cur.
func (b bitset) intersect(cur bitset) {
	for i := range b {
		b[i] &= cur[i]
	}
}

// first appends the k smallest set ids to out.
func (b bitset) first(k int, out []uint32) []uint32 {
	for i, w := range b {
		for w != 0 && len(out) < k {
			out = append(out, uint32(i*64+bits.TrailingZeros64(w)))
			w &= w - 1
		}
		if len(out) == k {
			break
		}
	}
	return out
}

// match runs the query given a way to OR every posting list of the words
// that start with a token into a bitset. It returns ids, best first.
func match(query string, k int, acc, cur bitset, orPrefix func(tok string, into bitset) bool) []uint32 {
	toks := tokens(query)
	if len(toks) == 0 {
		return nil
	}
	for i, tok := range toks {
		into := acc
		if i > 0 {
			into = cur
		}
		into.clear()
		if !orPrefix(tok, into) {
			return nil
		}
		if i > 0 {
			acc.intersect(cur)
		}
	}
	return acc.first(k, nil)
}

// ---- the plain Go version: strings, slices, binary search ----

// Plain keeps the descriptions as Go strings and finds the words that start
// with a token by binary search over the sorted word list.
type Plain struct {
	codes, descs []string
	words        []string
	lists        [][]uint32
	acc, cur     bitset
}

// NewPlain copies every string it keeps, so that it is charged for them.
func NewPlain(entries []Entry) *Plain {
	ranked := rankOrder(entries)
	p := &Plain{}
	for _, e := range ranked {
		p.codes = append(p.codes, strings.Clone(e.Code))
		p.descs = append(p.descs, strings.Clone(e.Desc))
	}
	p.words, p.lists = wordPostings(ranked)
	p.acc = make(bitset, (len(ranked)+63)/64)
	p.cur = make(bitset, (len(ranked)+63)/64)
	return p
}

func (p *Plain) Search(query string, k int) []Result {
	ids := match(query, k, p.acc, p.cur, func(tok string, into bitset) bool {
		i := sort.SearchStrings(p.words, tok)
		found := false
		for ; i < len(p.words) && strings.HasPrefix(p.words[i], tok); i++ {
			for _, id := range p.lists[i] {
				into[id/64] |= 1 << (id % 64)
			}
			found = true
		}
		return found
	})
	out := make([]Result, len(ids))
	for i, id := range ids {
		out[i] = Result{p.codes[id], p.descs[id]}
	}
	return out
}

// ---- the succinct version: two LOUDS tries ----

// Succinct stores every description inside one LOUDS trie and the word
// list in another. The word trie's key ids index the posting lists; the
// description trie's key ids are the description ids, because a LOUDS trie
// numbers its keys shortest first, then alphabetically -- rankOrder.
type Succinct struct {
	text     *Louds
	words    *Louds
	codes    []byte   // codeWidth bytes per description id, space padded
	postOff  []uint32 // word id -> start of its list in post
	post     []byte   // each list as gaps between sorted ids, in varints
	acc, cur bitset
	buf      []byte
}

const codeWidth = 7

func NewSuccinct(entries []Entry, shape, final Bits) *Succinct {
	ranked := rankOrder(entries)
	sorted := make([]string, len(ranked))
	for i, e := range ranked {
		sorted[i] = e.Desc
	}
	sort.Strings(sorted)
	s := &Succinct{}
	// The trie's key order is rankOrder, so ranked[id] is key id already;
	// the test checks that the two orders agree.
	s.text, _ = NewLouds(sorted, shape, final)
	s.codes = make([]byte, codeWidth*len(ranked))
	for id, e := range ranked {
		copy(s.codes[id*codeWidth:], e.Code+strings.Repeat(" ", codeWidth-len(e.Code)))
	}

	words, lists := wordPostings(ranked)
	var worder []int32
	s.words, worder = NewLouds(words, shape, final)
	s.postOff = make([]uint32, len(words)+1)
	for wid, w := range worder {
		prev := uint32(0)
		for i, id := range lists[w] {
			gap := id - prev
			if i == 0 {
				gap = id
			}
			s.post = binary.AppendUvarint(s.post, uint64(gap))
			prev = id
		}
		s.postOff[wid+1] = uint32(len(s.post))
	}
	s.post = append([]byte(nil), s.post...) // drop append's spare capacity
	s.acc = make(bitset, (len(ranked)+63)/64)
	s.cur = make(bitset, (len(ranked)+63)/64)
	return s
}

// orList decodes one posting list and sets its ids in into.
func orList(list []byte, into bitset) {
	id := uint32(0)
	for i := 0; i < len(list); {
		var gap uint32
		for shift := 0; ; shift += 7 {
			b := list[i]
			i++
			gap |= uint32(b&0x7f) << shift
			if b < 0x80 {
				break
			}
		}
		id += gap
		into[id/64] |= 1 << (id % 64)
	}
}

func (s *Succinct) Search(query string, k int) []Result {
	ids := match(query, k, s.acc, s.cur, func(tok string, into bitset) bool {
		node := s.words.Find(tok)
		if node < 0 {
			return false
		}
		s.words.Ranges(node, func(lo, hi int) {
			for wid := lo; wid < hi; wid++ {
				orList(s.post[s.postOff[wid]:s.postOff[wid+1]], into)
			}
		})
		return true
	})
	out := make([]Result, len(ids))
	for i, id := range ids {
		s.buf = s.text.Key(int(id), s.buf)
		code := strings.TrimRight(string(s.codes[int(id)*codeWidth:int(id+1)*codeWidth]), " ")
		out[i] = Result{code, string(s.buf)}
	}
	return out
}

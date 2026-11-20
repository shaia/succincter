package main

import (
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
)

func randomKeys(r *rand.Rand, n int) []string {
	set := map[string]bool{"": r.Intn(2) == 0}
	for len(set) < n {
		b := make([]byte, r.Intn(9))
		for i := range b {
			b[i] = "abcde"[r.Intn(5)]
		}
		set[string(b)] = true
	}
	var keys []string
	for k, ok := range set {
		if ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// byRank is the order LOUDS numbers keys in: length, then bytes.
func byRank(keys []string) []string {
	out := append([]string(nil), keys...)
	sort.Slice(out, func(i, j int) bool {
		if len(out[i]) != len(out[j]) {
			return len(out[i]) < len(out[j])
		}
		return out[i] < out[j]
	})
	return out
}

func checkLouds(t *testing.T, keys []string, shape, final Bits) {
	t.Helper()
	tr, order := NewLouds(keys, shape, final)
	ranked := byRank(keys)
	if tr.Keys() != len(keys) {
		t.Fatalf("keys %d, want %d", tr.Keys(), len(keys))
	}
	var buf []byte
	for id, want := range ranked {
		if keys[order[id]] != want {
			t.Fatalf("order[%d] = %q, want %q", id, keys[order[id]], want)
		}
		if buf = tr.Key(id, buf); string(buf) != want {
			t.Fatalf("Key(%d) = %q, want %q", id, buf, want)
		}
		k := tr.Find(want)
		if k < 0 || !tr.IsFinal(k) || tr.KeyID(k) != id {
			t.Fatalf("Find(%q) = %d, final %v", want, k, k >= 0 && tr.IsFinal(k))
		}
	}
	// Every prefix of every key, plus some strings that are not prefixes.
	prefixes := map[string]bool{"zz": true, "q": true}
	for _, k := range keys {
		for i := 0; i <= len(k); i++ {
			prefixes[k[:i]] = true
		}
		prefixes[k+"z"] = true
	}
	for p := range prefixes {
		var want []string
		for _, k := range keys {
			if strings.HasPrefix(k, p) {
				want = append(want, k)
			}
		}
		var got []string
		if n := tr.Find(p); n >= 0 {
			var each, ranges []int
			tr.Each(n, func(id int, suffix []byte) bool {
				if ranked[id] != p+string(suffix) {
					t.Fatalf("id %d is %q, suffix gives %q", id, ranked[id], p+string(suffix))
				}
				got = append(got, p+string(suffix))
				each = append(each, id)
				return true
			})
			tr.Ranges(n, func(lo, hi int) {
				for id := lo; id < hi; id++ {
					ranges = append(ranges, id)
				}
			})
			sort.Ints(each)
			if fmt.Sprint(each) != fmt.Sprint(ranges) {
				t.Fatalf("prefix %q: Each ids %v, Ranges ids %v", p, each, ranges)
			}
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("prefix %q: got %v want %v", p, got, want)
		}
	}
}

func TestLoudsRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for trial := 0; trial < 300; trial++ {
		keys := randomKeys(r, 1+r.Intn(200))
		checkLouds(t, keys, PlainBits, PlainBits)
		checkLouds(t, keys, RRRBits, RRRBits)
	}
}

func TestLoudsTiny(t *testing.T) {
	checkLouds(t, []string{"ileum", "ileus", "iliac", "ilium"}, PlainBits, PlainBits)
	checkLouds(t, []string{"a"}, PlainBits, PlainBits)
}

// TestICD runs against the real file when ICD10CM names it.
func TestICD(t *testing.T) {
	path := os.Getenv("ICD10CM")
	if path == "" {
		t.Skip("set ICD10CM to the icd10cm-codes file")
	}
	entries, err := load(path)
	if err != nil {
		t.Fatal(err)
	}
	var descs []string
	for _, e := range entries {
		descs = append(descs, e.Desc)
	}
	sort.Strings(descs)
	checkLouds(t, descs[:5000], PlainBits, RRRBits)
	tr, _ := NewLouds(descs, PlainBits, RRRBits)
	var buf []byte
	for id, want := range byRank(descs) {
		if buf = tr.Key(id, buf); string(buf) != want {
			t.Fatalf("Key(%d) = %q, want %q", id, buf, want)
		}
	}
	p, s := NewPlain(entries), NewSuccinct(entries, PlainBits, RRRBits)
	r := rand.New(rand.NewSource(2))
	for i := 0; i < 3000; i++ {
		e := entries[r.Intn(len(entries))]
		toks := tokens(e.Desc)
		q := ""
		for j := 0; j < 1+r.Intn(3); j++ {
			w := toks[r.Intn(len(toks))]
			q += " " + w[:1+r.Intn(len(w))]
		}
		if a, b := p.Search(q, 10), s.Search(q, 10); fmt.Sprint(a) != fmt.Sprint(b) {
			t.Fatalf("%q:\n plain    %v\n succinct %v", q, a, b)
		}
	}
}

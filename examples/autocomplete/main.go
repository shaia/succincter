// Command autocomplete builds a diagnosis search box over ICD-10-CM on top of
// succincter's rank and select, and measures it against plain Go.
//
// Get the data (public domain, from the US National Center for Health
// Statistics) and point the program at the codes file:
//
//	https://ftp.cdc.gov/pub/Health_Statistics/NCHS/Publications/ICD10CM/2027/icd10cm-code-descriptions-2027.zip
//	go run ./examples/autocomplete -data icd10cm-codes-2027.txt
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"
)

func load(path string) ([]Entry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var out []Entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		// Fixed layout: code padded to 8 columns, then the description.
		line := strings.TrimRight(sc.Text(), "\r")
		if len(line) < 9 {
			continue
		}
		out = append(out, Entry{strings.TrimSpace(line[:8]), line[8:]})
	}
	return out, sc.Err()
}

// heap reports how many bytes build leaves live on the heap.
func heap(build func() any) (any, int64) {
	var a, b runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&a)
	v := build()
	runtime.GC()
	runtime.ReadMemStats(&b)
	runtime.KeepAlive(v)
	return v, int64(b.HeapAlloc) - int64(a.HeapAlloc)
}

func kb(n int64) string { return fmt.Sprintf("%7.0f KB", float64(n)/1024) }

// typed is what a user might type; every prefix of each is one keystroke.
var typed = []string{
	"type 2 diabetes with kidney",
	"essential hypertension",
	"fracture shaft right femur closed",
	"crohn disease small intestine",
	"acute myocardial infarction",
	"asthma exacerbation",
	"otitis media acute left",
	"pneumonia unspecified",
	"migraine without aura",
	"low back pain",
	"heart failure",
	"sepsis",
}

func keystrokes() []string {
	var out []string
	for _, q := range typed {
		for i := 1; i <= len(q); i++ {
			out = append(out, q[:i])
		}
	}
	return out
}

// perCall repeats one search until at least min has passed and returns the
// mean. The Windows clock this was written on ticks in steps of about half
// a millisecond, far coarser than one search, so single calls are not timed.
func perCall(s Searcher, q string, min time.Duration) time.Duration {
	n := 0
	t := time.Now()
	for time.Since(t) < min {
		for i := 0; i < 16; i++ {
			s.Search(q, 10)
		}
		n += 16
	}
	return time.Since(t) / time.Duration(n)
}

func us(d time.Duration) string { return fmt.Sprintf("%7.1f µs", float64(d)/1e3) }

func main() {
	data := flag.String("data", "icd10cm-codes-2027.txt", "ICD-10-CM codes file")
	min := flag.Duration("min", 20*time.Millisecond, "time each keystroke for at least this long")
	flag.Parse()
	entries, err := load(*data)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("go %s %s/%s\n\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)

	// ---- the vocabulary ----
	descBytes, wordCount, longest, seventh := 0, 0, "", 0
	endings := map[string]int{"encounter": 0, ", sequela": 0}
	for _, e := range entries {
		descBytes += len(e.Desc)
		wordCount += len(tokens(e.Desc))
		if len(e.Desc) > len(longest) {
			longest = e.Desc
		}
		if len(e.Code) == 7 {
			seventh++
		}
		if strings.Contains(e.Desc, "encounter") {
			endings["encounter"]++
		}
		if strings.HasSuffix(e.Desc, ", sequela") {
			endings[", sequela"]++
		}
	}
	ranked := rankOrder(entries)
	words, lists := wordPostings(ranked)
	postings, wordBytes := 0, 0
	for i, l := range lists {
		postings += len(l)
		wordBytes += len(words[i])
	}
	fmt.Printf("codes %d (%d of 7 characters), description bytes %d (mean %.1f), longest %d\n",
		len(entries), seventh, descBytes, float64(descBytes)/float64(len(entries)), len(longest))
	fmt.Printf("descriptions with \"encounter\" %d, ending \", sequela\" %d\n", endings["encounter"], endings[", sequela"])
	fmt.Printf("words %d (mean %.1f per description), distinct %d (%d bytes), postings %d\n",
		wordCount, float64(wordCount)/float64(len(entries)), len(words), wordBytes, postings)
	fmt.Printf("shortest: %q, %q, %q\n", ranked[0].Desc, ranked[1].Desc, ranked[2].Desc)
	fmt.Printf("longest:  %q\n\n", longest)

	sortedDescs := func() []string {
		out := make([]string, len(entries))
		for i, e := range entries {
			out[i] = e.Desc
		}
		sort.Strings(out)
		return out
	}

	// ---- memory: each structure built from nothing but entries ----
	fmt.Println("the descriptions alone:")
	_, b := heap(func() any {
		out := sortedDescs()
		for i := range out {
			out[i] = strings.Clone(out[i])
		}
		return out
	})
	fmt.Printf("  %-26s %s\n", "[]string", kb(b))
	_, b = heap(func() any { r, _ := newPointerTrie(sortedDescs()); return r })
	_, pnodes := newPointerTrie(sortedDescs())
	fmt.Printf("  %-26s %s  %.1f bytes/node\n", "pointer trie", kb(b), float64(b)/float64(pnodes))
	variants := []struct {
		name         string
		shape, final Bits
	}{{"plain/plain", PlainBits, PlainBits}, {"plain/RRR", PlainBits, RRRBits}, {"RRR/RRR", RRRBits, RRRBits}}
	for _, v := range variants {
		t, b := heap(func() any { t, _ := NewLouds(sortedDescs(), v.shape, v.final); return t })
		n := t.(*Louds).Nodes()
		fmt.Printf("  %-26s %s  %.2f bits/node, %.2f bits/char  (%d nodes)\n", "LOUDS "+v.name,
			kb(b), float64(b*8)/float64(n), float64(b*8)/float64(descBytes), n)
	}
	rev := sortedDescs()
	for i, d := range rev {
		r := []byte(d)
		for a, z := 0, len(r)-1; a < z; a, z = a+1, z-1 {
			r[a], r[z] = r[z], r[a]
		}
		rev[i] = string(r)
	}
	sort.Strings(rev)
	rt, _ := NewLouds(rev, PlainBits, PlainBits)
	fmt.Printf("  %-26s %d nodes (the same text spelled backwards)\n", "reversed trie", rt.Nodes())
	wt, b := heap(func() any { t, _ := NewLouds(words, PlainBits, RRRBits); return t })
	fmt.Printf("  word trie, plain/RRR       %s  (%d nodes)\n\n", kb(b), wt.(*Louds).Nodes())

	fmt.Println("the whole index, everything needed to answer and display:")
	pl, b := heap(func() any { return NewPlain(entries) })
	fmt.Printf("  %-26s %s\n", "plain Go", kb(b))
	built := map[string]*Succinct{}
	for _, v := range variants {
		s, b := heap(func() any { return NewSuccinct(entries, v.shape, v.final) })
		built[v.name] = s.(*Succinct)
		fmt.Printf("  %-26s %s\n", "succinct "+v.name, kb(b))
	}
	s := built["plain/RRR"]
	fmt.Printf("  of which: codes %s, postings %s + offsets %s, bitsets %s\n\n",
		kb(int64(len(s.codes))), kb(int64(len(s.post))), kb(int64(4*len(s.postOff))), kb(int64(16*len(s.acc))))

	searchers := []struct {
		name string
		s    Searcher
	}{{"plain", pl.(*Plain)}, {"plain/plain", built["plain/plain"]}, {"plain/RRR", built["plain/RRR"]}, {"RRR/RRR", built["RRR/RRR"]}}

	// ---- they must agree ----
	ks := keystrokes()
	for _, q := range ks {
		want := fmt.Sprint(searchers[0].s.Search(q, 10))
		for _, x := range searchers[1:] {
			if fmt.Sprint(x.s.Search(q, 10)) != want {
				fmt.Printf("MISMATCH %s on %q\n", x.name, q)
				os.Exit(1)
			}
		}
	}
	fmt.Printf("all %d keystrokes give identical top-10 lists in all four\n\n", len(ks))

	// ---- what a user sees ----
	for _, q := range []string{"type 2 diab kid", "fem shaft frac", "crohn", "hypert", "mi", "heart attack", "myocard inf", "ileum", "ilium"} {
		fmt.Printf("%q\n", q)
		for _, r := range s.Search(q, 5) {
			fmt.Printf("    %-8s %s\n", r.Code, r.Desc)
		}
	}
	fmt.Println()

	// ---- what spelling a description back out costs ----
	fmt.Println("spelling every description back out of the trie, per description:")
	for _, v := range variants {
		t := built[v.name].text
		var buf []byte
		start := time.Now()
		for id := 0; id < t.Keys(); id++ {
			buf = t.Key(id, buf)
		}
		d := time.Since(start)
		fmt.Printf("  %-12s %s  (%.1f ns per character)\n", v.name,
			us(d/time.Duration(t.Keys())), float64(d)/float64(descBytes))
	}
	fmt.Println()

	// ---- time per keystroke, interleaved so all four see the same machine ----
	times := make([][]time.Duration, len(searchers))
	fmt.Println("per keystroke, \"fracture shaft right femur\":")
	q := "fracture shaft right femur"
	for _, k := range ks {
		line := fmt.Sprintf("  %-30q", k)
		for i, x := range searchers {
			d := perCall(x.s, k, *min)
			times[i] = append(times[i], d)
			line += " " + us(d)
		}
		shown := 0
		for _, r := range searchers[0].s.Search(k, 10) {
			shown += len(r.Desc)
		}
		line += fmt.Sprintf("  %4d chars shown", shown)
		if strings.HasPrefix(q, k) {
			fmt.Println(line)
		}
	}
	fmt.Printf("\nover all %d keystrokes:\n", len(ks))
	for i, x := range searchers {
		ds := append([]time.Duration(nil), times[i]...)
		sort.Slice(ds, func(a, b int) bool { return ds[a] < ds[b] })
		worst := ""
		for j, d := range times[i] {
			if d == ds[len(ds)-1] {
				worst = ks[j]
			}
		}
		pct := func(p float64) time.Duration { return ds[int(p*float64(len(ds)-1)+0.5)] }
		fmt.Printf("  %-12s median %s  p90 %s  max %s (%q)\n", x.name, us(pct(0.5)), us(pct(0.9)), us(ds[len(ds)-1]), worst)
	}
	fmt.Println()

	// ---- one typo away ----
	neighbours(words)
	runtime.KeepAlive(entries)
}

// neighbours counts the alphabetic words of 4+ letters that have another
// such word one edit away (insert, delete or substitute a letter).
func neighbours(words []string) {
	var ws []string
	for _, w := range words {
		if len(w) >= 4 && strings.Trim(w, "abcdefghijklmnopqrstuvwxyz") == "" {
			ws = append(ws, w)
		}
	}
	byLen := map[int][]string{}
	for _, w := range ws {
		byLen[len(w)] = append(byLen[len(w)], w)
	}
	pairs := 0
	has := map[string]bool{}
	var list []string
	for _, a := range ws {
		for d := 0; d <= 1; d++ {
			for _, b := range byLen[len(a)+d] {
				if (d == 1 || a < b) && oneEdit(a, b) {
					pairs++
					has[a], has[b] = true, true
					list = append(list, a+" / "+b)
				}
			}
		}
	}
	fmt.Printf("alphabetic words of 4+ letters: %d; with a neighbour one edit away: %d (%.1f%%); pairs %d\n",
		len(ws), len(has), 100*float64(len(has))/float64(len(ws)), pairs)
	if os.Getenv("PAIRS") != "" {
		for _, p := range list {
			fmt.Println("  ", p)
		}
	}
}

// oneEdit reports whether b is a with exactly one letter substituted, or
// with one letter inserted (len(b) == len(a)+1).
func oneEdit(a, b string) bool {
	if len(a) == len(b) {
		diff := 0
		for i := 0; i < len(a); i++ {
			if a[i] != b[i] {
				diff++
				if diff > 1 {
					return false
				}
			}
		}
		return diff == 1
	}
	i := 0
	for i < len(a) && a[i] == b[i] {
		i++
	}
	return a[i:] == b[i+1:]
}

// Example: DNA sequence analysis using Succincter
//
// Bioinformatics applications frequently need to answer:
// - "How many adenines (A) occur before position X?"
// - "Where is the 1000th guanine (G)?"
//
// Run with: go run ./examples/dna
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/shaia/succincter"
)

type Nucleotide byte

const (
	A Nucleotide = 'A' // Adenine
	C Nucleotide = 'C' // Cytosine
	G Nucleotide = 'G' // Guanine
	T Nucleotide = 'T' // Thymine
)

func main() {
	fmt.Println("=== DNA Sequence Analysis Example ===")

	// Generate a random DNA sequence (1 million base pairs)
	sequenceLen := 1_000_000
	fmt.Printf("\nGenerating %d base pair sequence...\n", sequenceLen)
	sequence := generateDNASequence(sequenceLen)

	// Build indices for each nucleotide
	fmt.Println("Building nucleotide indices...")
	start := time.Now()

	adenineIndex := succincter.NewSuccincter(sequence, func(n Nucleotide) bool { return n == A })
	cytosineIndex := succincter.NewSuccincter(sequence, func(n Nucleotide) bool { return n == C })
	guanineIndex := succincter.NewSuccincter(sequence, func(n Nucleotide) bool { return n == G })
	thymineIndex := succincter.NewSuccincter(sequence, func(n Nucleotide) bool { return n == T })

	fmt.Printf("Indices built in %v\n\n", time.Since(start))

	// Nucleotide composition
	fmt.Println("--- Sequence Composition ---")
	total := len(sequence)
	aCount := adenineIndex.Rank(total)
	cCount := cytosineIndex.Rank(total)
	gCount := guanineIndex.Rank(total)
	tCount := thymineIndex.Rank(total)

	fmt.Printf("Adenine (A):  %6d (%.2f%%)\n", aCount, float64(aCount)*100/float64(total))
	fmt.Printf("Cytosine (C): %6d (%.2f%%)\n", cCount, float64(cCount)*100/float64(total))
	fmt.Printf("Guanine (G):  %6d (%.2f%%)\n", gCount, float64(gCount)*100/float64(total))
	fmt.Printf("Thymine (T):  %6d (%.2f%%)\n", tCount, float64(tCount)*100/float64(total))
	fmt.Printf("GC Content:   %.2f%%\n", float64(gCount+cCount)*100/float64(total))

	// Find specific nucleotide positions
	fmt.Println("\n--- Position Queries ---")
	fmt.Printf("Position of 1000th Adenine:  %d\n", adenineIndex.Select(1000))
	fmt.Printf("Position of 5000th Guanine:  %d\n", guanineIndex.Select(5000))
	fmt.Printf("Position of 10000th Cytosine: %d\n", cytosineIndex.Select(10000))

	// Count nucleotides in a region (simulating a gene)
	geneStart, geneEnd := 100000, 105000
	fmt.Printf("\n--- Gene Region [%d, %d) ---\n", geneStart, geneEnd)
	fmt.Printf("Adenines in region:  %d\n", adenineIndex.Rank(geneEnd)-adenineIndex.Rank(geneStart))
	fmt.Printf("Cytosines in region: %d\n", cytosineIndex.Rank(geneEnd)-cytosineIndex.Rank(geneStart))
	fmt.Printf("Guanines in region:  %d\n", guanineIndex.Rank(geneEnd)-guanineIndex.Rank(geneStart))
	fmt.Printf("Thymines in region:  %d\n", thymineIndex.Rank(geneEnd)-thymineIndex.Rank(geneStart))

	// Find CpG islands (regions with high CG content)
	fmt.Println("\n--- CpG Island Detection ---")
	windowSize := 1000
	threshold := 0.6 // 60% GC content
	islands := findCpGIslands(guanineIndex, cytosineIndex, total, windowSize, threshold)
	fmt.Printf("Found %d potential CpG islands (>%.0f%% GC in %d bp windows)\n",
		len(islands), threshold*100, windowSize)
	if len(islands) > 0 {
		fmt.Printf("First island at position: %d\n", islands[0])
	}
}

func generateDNASequence(n int) []Nucleotide {
	nucleotides := []Nucleotide{A, C, G, T}
	sequence := make([]Nucleotide, n)
	for i := range sequence {
		sequence[i] = nucleotides[rand.Intn(4)]
	}
	return sequence
}

func findCpGIslands(gIndex, cIndex *succincter.Succincter, seqLen, windowSize int, threshold float64) []int {
	var islands []int
	for pos := 0; pos+windowSize <= seqLen; pos += windowSize / 2 {
		gcCount := (gIndex.Rank(pos+windowSize) - gIndex.Rank(pos)) +
			(cIndex.Rank(pos+windowSize) - cIndex.Rank(pos))
		if float64(gcCount)/float64(windowSize) >= threshold {
			islands = append(islands, pos)
		}
	}
	return islands
}

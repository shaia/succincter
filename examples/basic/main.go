// Example: Basic boolean array operations
//
// Run with: go run ./examples/basic
package main

import (
	"fmt"

	"github.com/shaia/succincter"
)

func main() {
	fmt.Println("=== Basic Succincter Example ===")

	// Simple boolean slice
	flags := []bool{true, false, true, true, false, false, true, false, true, true}
	fmt.Printf("Input: %v\n\n", flags)

	// Create Succincter - for booleans, predicate just returns the value
	s := succincter.NewSuccincter(flags, func(b bool) bool { return b })

	// Rank queries: count true values before position
	fmt.Println("Rank queries (count of true before position):")
	for _, pos := range []int{0, 1, 3, 5, 10} {
		fmt.Printf("  Rank(%d) = %d\n", pos, s.Rank(pos))
	}

	// Select queries: find position of Nth true value
	fmt.Println("\nSelect queries (position of Nth true):")
	for i := 1; i <= 6; i++ {
		pos := s.Select(i)
		if pos == -1 {
			fmt.Printf("  Select(%d) = -1 (not found)\n", i)
		} else {
			fmt.Printf("  Select(%d) = %d (flags[%d] = %v)\n", i, pos, pos, flags[pos])
		}
	}

	// Demonstrate rank-select duality
	fmt.Println("\nRank-Select duality:")
	for i := 1; i <= 6; i++ {
		pos := s.Select(i)
		if pos != -1 {
			rank := s.Rank(pos)
			fmt.Printf("  Select(%d) = %d, Rank(%d) = %d (rank = i-1)\n", i, pos, pos, rank)
		}
	}
}

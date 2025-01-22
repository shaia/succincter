package succincter

import (
	"fmt"
	"testing"
)

func BenchmarkCompareImplementations(b *testing.B) {
	benchCases := []struct {
		name    string
		size    int
		pattern string
	}{
		{"Small_Sparse_1K", 1000, "sparse"},
		{"Small_Dense_1K", 1000, "dense"},
		{"Medium_Sparse_10K", 10000, "sparse"},
		{"Medium_Dense_10K", 10000, "dense"},
	}

	createTestData := func(size int, pattern string) []bool {
		data := make([]bool, size)
		switch pattern {
		case "sparse": // 1% ones
			for i := 0; i < size; i += 100 {
				data[i] = true
			}
		case "dense": // 90% ones
			for i := range data {
				if i%10 != 0 {
					data[i] = true
				}
			}
		}
		return data
	}

	for _, bc := range benchCases {
		data := createTestData(bc.size, bc.pattern)

		// Create instances outside the benchmark loop
		succincter := NewSuccincter(data, func(b bool) bool { return b })

		// Test positions for Rank operations
		positions := []int{
			0,
			bc.size / 4,
			bc.size / 2,
			(bc.size * 3) / 4,
			bc.size - 1,
		}

		// Calculate number of ones for Select operations
		onesCount := succincter.Rank(bc.size)
		ranks := []int{
			1,
			onesCount / 4,
			onesCount / 2,
			(onesCount * 3) / 4,
			onesCount,
		}

		// Benchmark Rank operations
		b.Run(fmt.Sprintf("Rank_Succincter_%s", bc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, pos := range positions {
					_ = succincter.Rank(pos)
				}
			}
		})

		// Benchmark Select operations
		b.Run(fmt.Sprintf("Select_Succincter_%s", bc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, rank := range ranks {
					_ = succincter.Select(rank)
				}
			}
		})

		// Benchmark construction time
		b.Run(fmt.Sprintf("Build_Succincter_%s", bc.name), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = NewSuccincter(data, func(b bool) bool { return b })
			}
		})
	}
}

// Correctness tests comparing both implementations
func TestCompareImplementations(t *testing.T) {
	tests := []struct {
		name  string
		input []bool
	}{
		{"Empty", []bool{}},
		{"Single_True", []bool{true}},
		{"Single_False", []bool{false}},
		{"Small_Mixed", []bool{true, false, true, true, false}},
		{"Ten_Alternating", []bool{true, false, true, false, true, false, true, false, true, false}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			succincter := NewSuccincter(tt.input, func(b bool) bool { return b })

			// Test Rank operations
			for pos := 0; pos <= len(tt.input); pos++ {
				succRank := succincter.Rank(pos)
				simpleRank := succincter.Rank(pos)
				if succRank != simpleRank {
					t.Errorf("Rank(%d) mismatch: Succincter=%d, Simple=%d", pos, succRank, simpleRank)
				}
			}

			// Test Select operations
			maxOnes := succincter.Rank(len(tt.input))
			for rank := 1; rank <= maxOnes+1; rank++ {
				succSelect := succincter.Select(rank)
				simpleSelect := succincter.Select(rank)
				if succSelect != simpleSelect {
					t.Errorf("Select(%d) mismatch: Succincter=%d, Simple=%d", rank, succSelect, simpleSelect)
				}
			}
		})
	}
}

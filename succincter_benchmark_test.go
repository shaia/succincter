package succincter

import (
	"fmt"
	"testing"
)

func BenchmarkSuccincter(b *testing.B) {
	// Helper function to create large arrays with different patterns
	createLargeArray := func(size int, pattern string) []bool {
		arr := make([]bool, size)
		switch pattern {
		case "sparse": // 1% ones
			for i := 0; i < size; i += 100 {
				arr[i] = true
			}
		case "dense": // 90% ones
			for i := range arr {
				if i%10 != 0 {
					arr[i] = true
				}
			}
		case "alternating": // 50% ones
			for i := range arr {
				arr[i] = i%2 == 0
			}
		}
		return arr
	}

	benchCases := []struct {
		name    string
		size    int
		pattern string
	}{
		{"Small_Sparse_1K", 1000, "sparse"},
		{"Small_Dense_1K", 1000, "dense"},
		{"Small_Alternating_1K", 1000, "alternating"},
		{"Medium_Sparse_100K", 100000, "sparse"},
		{"Medium_Dense_100K", 100000, "dense"},
		{"Medium_Alternating_100K", 100000, "alternating"},
		{"Large_Sparse_1M", 1000000, "sparse"},
		{"Large_Dense_1M", 1000000, "dense"},
		{"Large_Alternating_1M", 1000000, "alternating"},
	}

	for _, bc := range benchCases {
		data := createLargeArray(bc.size, bc.pattern)
		s := NewSuccincter(data, func(b bool) bool { return b })

		b.Run("Build_"+bc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewSuccincter(data, func(b bool) bool { return b })
			}
		})

		b.Run("Rank_"+bc.name, func(b *testing.B) {
			positions := []int{
				0,
				bc.size / 4,
				bc.size / 2,
				(bc.size * 3) / 4,
				bc.size - 1,
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, pos := range positions {
					_ = s.Rank(pos)
				}
			}
		})

		b.Run("Select_"+bc.name, func(b *testing.B) {
			// Calculate actual number of ones for realistic Select operations
			onesCount := s.Rank(bc.size)
			ranks := []int{
				1,
				onesCount / 4,
				onesCount / 2,
				(onesCount * 3) / 4,
				onesCount,
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, rank := range ranks {
					_ = s.Select(rank)
				}
			}
		})
	}
}

// Memory usage benchmark
func BenchmarkSuccincterMemory(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Memory_%d", size), func(b *testing.B) {
			data := make([]bool, size)
			for i := range data {
				data[i] = i%2 == 0
			}

			b.ResetTimer()
			var s *Succincter
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s = NewSuccincter(data, func(b bool) bool { return b })
			}
			_ = s
		})
	}
}

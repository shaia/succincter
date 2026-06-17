package succincter

import (
	"fmt"
	"testing"
	"unsafe"
)

func BenchmarkRRR(b *testing.B) {
	createLargeArray := func(size int, pattern string) []bool {
		arr := make([]bool, size)
		switch pattern {
		case "sparse":
			for i := 0; i < size; i += 100 {
				arr[i] = true
			}
		case "dense":
			for i := range arr {
				if i%10 != 0 {
					arr[i] = true
				}
			}
		case "alternating":
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
		r := NewRRR(data, func(b bool) bool { return b })

		b.Run("Build_RRR_"+bc.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				NewRRR(data, func(b bool) bool { return b })
			}
		})

		b.Run("Rank_RRR_"+bc.name, func(b *testing.B) {
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
					_ = r.Rank(pos)
				}
			}
		})

		b.Run("Select_RRR_"+bc.name, func(b *testing.B) {
			onesCount := r.Rank(bc.size)
			if onesCount == 0 {
				b.Skip("no ones in input")
			}
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
					if rank > 0 {
						_ = r.Select(rank)
					}
				}
			}
		})
	}
}

func BenchmarkRRRMemory(b *testing.B) {
	sizes := []int{1000, 10000, 100000, 1000000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Memory_RRR_%d", size), func(b *testing.B) {
			data := make([]bool, size)
			for i := range data {
				data[i] = i%2 == 0
			}

			b.ResetTimer()
			var r *RRR
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r = NewRRR(data, func(b bool) bool { return b })
			}
			_ = r
		})
	}
}

// buildDensityArray spreads `size * density` ones uniformly across the array via
// modular striding, giving every block a comparable popcount. This matches the
// "sparse/alternating" style of the existing Succincter benchmarks and gives RRR
// a realistic per-block class distribution rather than a step function.
func buildDensityArray(size int, density float64) []bool {
	arr := make([]bool, size)
	if density <= 0 {
		return arr
	}
	stride := int(1.0 / density)
	if stride < 1 {
		stride = 1
	}
	for i := 0; i < size; i += stride {
		arr[i] = true
	}
	return arr
}

func BenchmarkRRRVsSuccincter(b *testing.B) {
	size := 100000
	densities := []struct {
		name    string
		density float64
	}{
		{"1pct", 0.01},
		{"10pct", 0.10},
		{"50pct", 0.50},
	}

	for _, d := range densities {
		data := buildDensityArray(size, d.density)

		r := NewRRR(data, func(b bool) bool { return b })
		s := NewSuccincter(data, func(b bool) bool { return b })
		pos := size / 2

		b.Run("RRR_Rank_"+d.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = r.Rank(pos)
			}
		})

		b.Run("Succincter_Rank_"+d.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.Rank(pos)
			}
		})
	}
}

func BenchmarkRRRSpace(b *testing.B) {
	size := 100000
	densities := []struct {
		name    string
		density float64
	}{
		{"1pct", 0.01},
		{"10pct", 0.10},
		{"50pct", 0.50},
	}

	for _, d := range densities {
		data := buildDensityArray(size, d.density)

		r := NewRRR(data, func(b bool) bool { return b })
		s := NewSuccincter(data, func(b bool) bool { return b })

		rrrBytes := int(unsafe.Sizeof(*r))
		rrrBytes += len(r.classes) * 8
		rrrBytes += len(r.offsets) * 8
		rrrBytes += len(r.superBlocks) * 8

		succincterBytes := int(unsafe.Sizeof(*s))
		succincterBytes += len(s.data) * 8
		succincterBytes += len(s.blockRanks) * 8
		succincterBytes += len(s.superBlocks) * 8

		rrrBitsPerElem := float64(rrrBytes*8) / float64(size)
		succincterBitsPerElem := float64(succincterBytes*8) / float64(size)

		b.Run("Space_"+d.name, func(b *testing.B) {
			b.ReportMetric(float64(rrrBytes), "RRR_bytes")
			b.ReportMetric(float64(succincterBytes), "Succincter_bytes")
			b.ReportMetric(rrrBitsPerElem, "RRR_bits/elem")
			b.ReportMetric(succincterBitsPerElem, "Succincter_bits/elem")
			b.ReportMetric(float64(succincterBytes)/float64(rrrBytes), "space_ratio")
		})
	}
}

package succincter

import (
	"fmt"
	"testing"
)

func TestRRRConstruction(t *testing.T) {
	tests := []struct {
		name  string
		input []bool
	}{
		{"empty", []bool{}},
		{"single true", []bool{true}},
		{"single false", []bool{false}},
		{"all true", []bool{true, true, true, true, true}},
		{"all false", []bool{false, false, false, false, false}},
		{"mixed", []bool{true, false, true, true, false, true, false, false, true}},
		{"1000 elements", func() []bool {
			arr := make([]bool, 1000)
			for i := range arr {
				arr[i] = i%3 == 0
			}
			return arr
		}()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRRR(tt.input, func(b bool) bool { return b })

			expectedOnes := 0
			for _, v := range tt.input {
				if v {
					expectedOnes++
				}
			}

			if r.totalOnes != expectedOnes {
				t.Errorf("totalOnes = %d; want %d", r.totalOnes, expectedOnes)
			}

			expectedBlocks := (len(tt.input) + 14) / 15
			if r.numBlocks != expectedBlocks {
				t.Errorf("numBlocks = %d; want %d", r.numBlocks, expectedBlocks)
			}
		})
	}
}

func TestRRRBoundaryLengths(t *testing.T) {
	lengths := []int{14, 15, 16, 29, 30, 31, 44, 45, 46}

	for _, length := range lengths {
		t.Run(fmt.Sprintf("len=%d", length), func(t *testing.T) {
			input := make([]bool, length)
			for i := range input {
				input[i] = i%2 == 0
			}

			r := NewRRR(input, func(b bool) bool { return b })

			expectedOnes := 0
			for _, v := range input {
				if v {
					expectedOnes++
				}
			}

			if r.totalOnes != expectedOnes {
				t.Errorf("length %d: totalOnes = %d; want %d", length, r.totalOnes, expectedOnes)
			}
		})
	}
}

func TestRRRDensities(t *testing.T) {
	size := 1000
	tests := []struct {
		name    string
		density float64
	}{
		{"1pct", 0.01},
		{"10pct", 0.10},
		{"50pct", 0.50},
		{"90pct", 0.90},
		{"99pct", 0.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := make([]bool, size)
			expectedOnes := 0
			for i := range input {
				if float64(i) < float64(size)*tt.density {
					input[i] = true
					expectedOnes++
				}
			}

			r := NewRRR(input, func(b bool) bool { return b })

			if r.totalOnes != expectedOnes {
				t.Errorf("%s density: totalOnes = %d; want %d", tt.name, r.totalOnes, expectedOnes)
			}
		})
	}
}

func TestRRRGenericInput(t *testing.T) {
	input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	r := NewRRR(input, func(n int) bool { return n%2 == 0 })

	if r.totalOnes != 5 {
		t.Errorf("totalOnes = %d; want 5", r.totalOnes)
	}

	input2 := []string{"a", "bb", "ccc", "d", "ee"}
	r2 := NewRRR(input2, func(s string) bool { return len(s) > 1 })

	if r2.totalOnes != 3 {
		t.Errorf("totalOnes = %d; want 3", r2.totalOnes)
	}
}

func TestRRRRankBasic(t *testing.T) {
	input := []bool{true, false, true, true, false, true, false, false, true}
	r := NewRRR(input, func(b bool) bool { return b })

	tests := []struct {
		pos      int
		expected int
	}{
		{0, 0},
		{1, 1},
		{3, 2},
		{5, 3},
		{9, 5},
	}

	for _, tt := range tests {
		got := r.Rank(tt.pos)
		if got != tt.expected {
			t.Errorf("Rank(%d) = %d; want %d", tt.pos, got, tt.expected)
		}
	}
}

func TestRRRSelectBasic(t *testing.T) {
	input := []bool{true, false, true, true, false, true, false, false, true}
	r := NewRRR(input, func(b bool) bool { return b })

	tests := []struct {
		rank     int
		expected int
	}{
		{0, -1},
		{1, 0},
		{2, 2},
		{3, 3},
		{5, 8},
		{6, -1},
	}

	for _, tt := range tests {
		got := r.Select(tt.rank)
		if got != tt.expected {
			t.Errorf("Select(%d) = %d; want %d", tt.rank, got, tt.expected)
		}
	}
}

func TestRRRVsSuccincter(t *testing.T) {
	sizes := []int{100, 500, 1000}
	densities := []float64{0.01, 0.10, 0.50, 0.90, 0.99}

	for _, size := range sizes {
		for _, density := range densities {
			t.Run(fmt.Sprintf("size=%d/density=%.2f", size, density), func(t *testing.T) {
				input := make([]bool, size)
				for i := range input {
					if float64(i) < float64(size)*density {
						input[i] = true
					}
				}

				r := NewRRR(input, func(b bool) bool { return b })
				s := NewSuccincter(input, func(b bool) bool { return b })

				for pos := 0; pos <= size; pos++ {
					rRank := r.Rank(pos)
					sRank := s.Rank(pos)
					if rRank != sRank {
						t.Errorf("Rank(%d): RRR=%d, Succincter=%d", pos, rRank, sRank)
					}
				}

				totalOnes := r.Rank(size)
				for rank := 0; rank <= totalOnes+1; rank++ {
					rSel := r.Select(rank)
					sSel := s.Select(rank)
					if rSel != sSel {
						t.Errorf("Select(%d): RRR=%d, Succincter=%d", rank, rSel, sSel)
					}
				}
			})
		}
	}
}

func TestRRRBlockBoundaries(t *testing.T) {
	input := make([]bool, 240)
	for i := range input {
		input[i] = i%3 == 0
	}

	r := NewRRR(input, func(b bool) bool { return b })
	s := NewSuccincter(input, func(b bool) bool { return b })

	boundaries := []int{0, 14, 15, 16, 29, 30, 31, 239, 240}
	for _, pos := range boundaries {
		rRank := r.Rank(pos)
		sRank := s.Rank(pos)
		if rRank != sRank {
			t.Errorf("Rank(%d): RRR=%d, Succincter=%d", pos, rRank, sRank)
		}
	}
}

func TestRRRSuperBlockBoundaries(t *testing.T) {
	input := make([]bool, 500)
	for i := range input {
		input[i] = i%5 == 0
	}

	r := NewRRR(input, func(b bool) bool { return b })
	s := NewSuccincter(input, func(b bool) bool { return b })

	boundaries := []int{0, 239, 240, 241, 479, 480, 500}
	for _, pos := range boundaries {
		rRank := r.Rank(pos)
		sRank := s.Rank(pos)
		if rRank != sRank {
			t.Errorf("Rank(%d): RRR=%d, Succincter=%d", pos, rRank, sRank)
		}
	}
}

func TestRRREdgeCases(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		r := NewRRR([]bool{}, func(b bool) bool { return b })
		if r.Rank(0) != 0 {
			t.Errorf("Rank(0) on empty = %d; want 0", r.Rank(0))
		}
		if r.Rank(10) != 0 {
			t.Errorf("Rank(10) on empty = %d; want 0", r.Rank(10))
		}
		if r.Select(1) != -1 {
			t.Errorf("Select(1) on empty = %d; want -1", r.Select(1))
		}
	})

	t.Run("negative", func(t *testing.T) {
		r := NewRRR([]bool{true, false, true}, func(b bool) bool { return b })
		if r.Rank(-5) != 0 {
			t.Errorf("Rank(-5) = %d; want 0", r.Rank(-5))
		}
		if r.Select(-3) != -1 {
			t.Errorf("Select(-3) = %d; want -1", r.Select(-3))
		}
	})

	t.Run("interface satisfaction", func(t *testing.T) {
		var _ RankSelector = (*RRR)(nil)
	})
}

package succincter

import "testing"

func FuzzRRRRank(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1})
	f.Add([]byte{0})
	f.Add([]byte{1, 1, 1, 1, 1})
	f.Add([]byte{0, 0, 0, 0, 0})
	f.Add([]byte{1, 0, 1, 0, 1, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}

		input := make([]bool, len(data))
		expectedOnes := 0
		for i, b := range data {
			input[i] = (b % 2) == 1
			if input[i] {
				expectedOnes++
			}
		}

		r := NewRRR(input, func(b bool) bool { return b })
		s := NewSuccincter(input, func(b bool) bool { return b })

		prevRank := r.Rank(0)
		if prevRank != 0 {
			t.Errorf("Rank(0) = %d; must be exactly 0", prevRank)
			return
		}

		for pos := 1; pos <= len(input); pos++ {
			rank := r.Rank(pos)

			if rank < 0 {
				t.Errorf("Rank(%d) = %d; must be non-negative", pos, rank)
				return
			}

			if rank < prevRank {
				t.Errorf("Monotonicity violated: Rank(%d) = %d < Rank(%d) = %d", pos, rank, pos-1, prevRank)
				return
			}

			sRank := s.Rank(pos)
			if rank != sRank {
				t.Errorf("Cross-validation failed at pos=%d: RRR=%d, Succincter=%d", pos, rank, sRank)
				return
			}

			prevRank = rank
		}

		if r.Rank(len(input)) != expectedOnes {
			t.Errorf("Rank(len) = %d; want %d", r.Rank(len(input)), expectedOnes)
		}
	})
}

func FuzzRRRSelect(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1})
	f.Add([]byte{0})
	f.Add([]byte{1, 1, 1, 1, 1})
	f.Add([]byte{0, 0, 0, 0, 0})
	f.Add([]byte{1, 0, 1, 0, 1, 0})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 || len(data) > 50 {
			return
		}

		input := make([]bool, len(data))
		totalOnes := 0
		for i, b := range data {
			input[i] = (b % 2) == 1
			if input[i] {
				totalOnes++
			}
		}

		r := NewRRR(input, func(b bool) bool { return b })
		s := NewSuccincter(input, func(b bool) bool { return b })

		for rank := 1; rank <= totalOnes+5; rank++ {
			rPos := r.Select(rank)
			sPos := s.Select(rank)

			if rPos != sPos {
				t.Errorf("Cross-validation failed at rank=%d: RRR=%d, Succincter=%d", rank, rPos, sPos)
				return
			}

			if rPos != -1 {
				if rPos < 0 || rPos >= len(input) {
					t.Errorf("Select(%d) = %d; out of bounds [0, %d)", rank, rPos, len(input))
					return
				}

				if !input[rPos] {
					t.Errorf("Select(%d) = %d; but input[%d] is false", rank, rPos, rPos)
					return
				}

				computedRank := r.Rank(rPos + 1)
				if computedRank != rank {
					t.Errorf("Inverse violated: Select(%d) = %d, but Rank(%d) = %d", rank, rPos, rPos+1, computedRank)
					return
				}
			}
		}
	})
}

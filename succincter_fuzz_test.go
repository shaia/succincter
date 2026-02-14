package succincter

import "testing"

func FuzzRank(f *testing.F) {
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
		for i, b := range data {
			input[i] = (b % 2) == 1
		}

		s := NewSuccincter(input, func(b bool) bool { return b })

		prevRank := s.Rank(0)
		if prevRank != 0 {
			t.Errorf("Rank(0) = %d; must be exactly 0", prevRank)
			return
		}

		for pos := 1; pos <= len(input); pos++ {
			rank := s.Rank(pos)

			if rank < 0 {
				t.Errorf("Rank(%d) = %d; must be non-negative", pos, rank)
				return
			}

			if rank < prevRank {
				t.Errorf("Monotonicity violated: Rank(%d) = %d < Rank(%d) = %d", pos, rank, pos-1, prevRank)
				return
			}

			prevRank = rank
		}
	})
}

func FuzzSelect(f *testing.F) {
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

		s := NewSuccincter(input, func(b bool) bool { return b })

		for rank := 1; rank <= totalOnes+5; rank++ {
			pos := s.Select(rank)

			if pos != -1 {
				if pos < 0 || pos >= len(input) {
					t.Errorf("Select(%d) = %d; out of bounds [0, %d)", rank, pos, len(input))
					return
				}

				if !input[pos] {
					t.Errorf("Select(%d) = %d; but input[%d] is false", rank, pos, pos)
					return
				}
			}
		}
	})
}

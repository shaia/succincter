package succincter

import (
	"math/rand"
	"sync"
	"testing"
)

func TestSuccincter(t *testing.T) {
	tests := []struct {
		name      string
		input     []bool
		rankTests []struct {
			pos      int
			expected int
		}
		selectTests []struct {
			rank     int
			expected int
		}
	}{

		{
			name:  "Rank test",
			input: []bool{true, false, true, true, false, true, false, false, true},
			rankTests: []struct {
				pos      int
				expected int
			}{
				{5, 3},
			},
		},
		{
			name:  "Single true value",
			input: []bool{true},
			rankTests: []struct {
				pos      int
				expected int
			}{
				{0, 0},
				{1, 1},
				{2, 1},
			},
			selectTests: []struct {
				rank     int
				expected int
			}{
				{1, 0},
				{2, -1},
			},
		},
		{
			name:  "Large sparse input",
			input: make([]bool, 1000), // Initialize with all false
			rankTests: []struct {
				pos      int
				expected int
			}{
				{500, 0},
				{999, 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSuccincter(tt.input, func(b bool) bool { return b })

			t.Logf("Testing %s with input length %d", tt.name, len(tt.input))

			// Test Rank operations
			for _, rt := range tt.rankTests {
				got := s.Rank(rt.pos)
				if got != rt.expected {
					t.Errorf("Rank(%d) = %d; want %d", rt.pos, got, rt.expected)
				}
				t.Logf("Rank(%d) = %d", rt.pos, got)
			}

			// Test Select operations
			for _, st := range tt.selectTests {
				got := s.Select(st.rank)
				if got != st.expected {
					t.Errorf("Select(%d) = %d; want %d", st.rank, got, st.expected)
				}
				t.Logf("Select(%d) = %d", st.rank, got)
			}
		})
	}
}

// IntSuccincter is a simple wrapper that converts int slice to bool slice
type IntSuccincter struct {
	*Succincter
}

func NewIntSuccincter(input []int, predicate func(int) bool) *IntSuccincter {
	boolArr := make([]bool, len(input))
	for i, v := range input {
		boolArr[i] = predicate(v)
	}
	return &IntSuccincter{NewSuccincter(boolArr, func(b bool) bool { return b })}
}

func TestNonBooleanInputs(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		predicate func(int) bool
		rankTests []struct {
			pos      int
			expected int
		}
		selectTests []struct {
			rank     int
			expected int
		}
	}{
		{
			name:  "Even numbers",
			input: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			predicate: func(n int) bool {
				return n%2 == 0
			},
			rankTests: []struct {
				pos      int
				expected int
			}{
				{0, 0},  // No even numbers before pos 0
				{2, 1},  // One even number (2) before pos 2
				{4, 2},  // Two even numbers (2,4) before pos 4
				{10, 5}, // Five even numbers total
			},
			selectTests: []struct {
				rank     int
				expected int
			}{
				{1, 1},  // First even number at pos 1 (2)
				{2, 3},  // Second even number at pos 3 (4)
				{5, 9},  // Fifth even number at pos 9 (10)
				{6, -1}, // No sixth even number
			},
		},
		{
			name:  "Numbers greater than 5",
			input: []int{3, 7, 1, 9, 4, 6, 8, 2, 5, 10},
			predicate: func(n int) bool {
				return n > 5
			},
			rankTests: []struct {
				pos      int
				expected int
			}{
				{0, 0},  // No numbers > 5 before pos 0
				{2, 1},  // One number > 5 before pos 2 (7)
				{5, 2},  // Two numbers > 5 before pos 5 (7,9)
				{10, 5}, // Five numbers > 5 total
			},
			selectTests: []struct {
				rank     int
				expected int
			}{
				{1, 1},  // First number > 5 at pos 1 (7)
				{3, 5},  // Third number > 5 at pos 5 (6)
				{5, 9},  // Fifth number > 5 at pos 9 (10)
				{6, -1}, // No sixth number > 5
			},
		},
		{
			name:  "Prime numbers",
			input: []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			predicate: func(n int) bool {
				if n < 2 {
					return false
				}
				for i := 2; i*i <= n; i++ {
					if n%i == 0 {
						return false
					}
				}
				return true
			},
			rankTests: []struct {
				pos      int
				expected int
			}{
				{0, 0},  // No primes before pos 0
				{2, 2},  // Two primes before pos 2 (2,3)
				{6, 4},  // Four primes before pos 6 (2,3,5,7)
				{10, 5}, // Five primes total (2,3,5,7,11)
			},
			selectTests: []struct {
				rank     int
				expected int
			}{
				{1, 0},  // First prime at pos 0 (2)
				{3, 3},  // Third prime at pos 3 (5)
				{5, 9},  // Fifth prime at pos 9 (11)
				{6, -1}, // No sixth prime
			},
		},
		{
			name:  "Empty array",
			input: []int{},
			predicate: func(n int) bool {
				return n > 0
			},
			rankTests: []struct {
				pos      int
				expected int
			}{
				{0, 0},
				{1, 0},
			},
			selectTests: []struct {
				rank     int
				expected int
			}{
				{1, -1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewIntSuccincter(tt.input, tt.predicate)

			t.Logf("Testing %s with input length %d", tt.name, len(tt.input))

			// Test Rank operations
			for _, rt := range tt.rankTests {
				got := s.Rank(rt.pos)
				if got != rt.expected {
					t.Errorf("Rank(%d) = %d; want %d", rt.pos, got, rt.expected)
				}
				t.Logf("Rank(%d) = %d", rt.pos, got)
			}

			// Test Select operations
			for _, st := range tt.selectTests {
				got := s.Select(st.rank)
				if got != st.expected {
					t.Errorf("Select(%d) = %d; want %d", st.rank, got, st.expected)
				}
				t.Logf("Select(%d) = %d", st.rank, got)
			}
		})
	}
}

func TestRankProperties(t *testing.T) {
	input := make([]bool, 200)
	for i := range input {
		if i%3 == 0 {
			input[i] = true
		}
	}
	s := NewSuccincter(input, func(b bool) bool { return b })

	if s.Rank(0) != 0 {
		t.Errorf("Rank(0) = %d; want 0", s.Rank(0))
	}

	for i := 0; i < len(input)-1; i++ {
		rank_i := s.Rank(i + 1)
		rank_i_plus_1 := s.Rank(i + 2)

		if rank_i > rank_i_plus_1 {
			t.Errorf("Monotonicity violated: Rank(%d) = %d > Rank(%d) = %d", i+1, rank_i, i+2, rank_i_plus_1)
		}
	}

	for i := 1; i <= len(input); i++ {
		rank := s.Rank(i)
		if rank > i {
			t.Errorf("Bounds violated: Rank(%d) = %d > %d", i, rank, i)
		}
	}

	for i := 0; i < len(input); i++ {
		rank_i := s.Rank(i)
		rank_i_plus_1 := s.Rank(i + 1)
		diff := rank_i_plus_1 - rank_i

		if input[i] {
			if diff != 1 {
				t.Errorf("Differential property violated: bit %d is set, but Rank(%d) - Rank(%d) = %d; want 1", i, i+1, i, diff)
			}
		} else {
			if diff != 0 {
				t.Errorf("Differential property violated: bit %d is unset, but Rank(%d) - Rank(%d) = %d; want 0", i, i+1, i, diff)
			}
		}
	}
}

func TestSelectProperties(t *testing.T) {
	input := make([]bool, 200)
	for i := range input {
		if i%3 == 0 {
			input[i] = true
		}
	}
	s := NewSuccincter(input, func(b bool) bool { return b })

	totalOnes := 0
	for _, v := range input {
		if v {
			totalOnes++
		}
	}

	if s.Select(0) != -1 {
		t.Errorf("Select(0) = %d; want -1", s.Select(0))
	}

	if s.Select(totalOnes+1) != -1 {
		t.Errorf("Select(%d) = %d; want -1", totalOnes+1, s.Select(totalOnes+1))
	}

	for i := 1; i < totalOnes; i++ {
		pos_i := s.Select(i)
		pos_i_plus_1 := s.Select(i + 1)

		if pos_i >= pos_i_plus_1 {
			t.Errorf("Monotonicity violated: Select(%d) = %d >= Select(%d) = %d", i, pos_i, i+1, pos_i_plus_1)
		}
	}

	for i := 1; i <= totalOnes; i++ {
		pos := s.Select(i)
		if pos < 0 || pos >= len(input) {
			t.Errorf("Select(%d) = %d; out of bounds", i, pos)
			continue
		}
		if !input[pos] {
			t.Errorf("Select(%d) = %d; but input[%d] is false", i, pos, pos)
		}
	}
}

func TestConcurrentReads(t *testing.T) {
	input := make([]bool, 1000)
	for i := range input {
		if i%3 == 0 {
			input[i] = true
		}
	}
	s := NewSuccincter(input, func(b bool) bool { return b })

	var wg sync.WaitGroup
	for g := 0; g < 10; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				pos := rand.Intn(len(input)) + 1
				s.Rank(pos)

				rank := rand.Intn(100) + 1
				s.Select(rank)
			}
		}()
	}
	wg.Wait()
}

func TestSelectRankInverse(t *testing.T) {
	tests := []struct {
		name  string
		input []bool
	}{
		{
			name:  "Single block",
			input: []bool{true, false, true, true, false, true, false, false, true},
		},
		{
			name:  "Multiple blocks",
			input: func() []bool {
				arr := make([]bool, 200)
				for i := range arr {
					arr[i] = i%3 == 0
				}
				return arr
			}(),
		},
		{
			name: "Block boundary",
			input: func() []bool {
				arr := make([]bool, 128)
				arr[63] = true
				arr[64] = true
				arr[65] = true
				return arr
			}(),
		},
		{
			name: "Superblock boundary",
			input: func() []bool {
				arr := make([]bool, 2048)
				arr[1023] = true
				arr[1024] = true
				arr[1025] = true
				return arr
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSuccincter(tt.input, func(b bool) bool { return b })

			for i := 0; i < len(tt.input); i++ {
				if tt.input[i] {
					rank := s.Rank(i + 1)
					selectedPos := s.Select(rank)

					if selectedPos != i {
						t.Errorf("Select(Rank(%d)) = Select(%d) = %d; want %d", i+1, rank, selectedPos, i)
					}
				}
			}
		})
	}
}

func TestEdgeCases(t *testing.T) {
	t.Run("Rank at position 0", func(t *testing.T) {
		input := []bool{true, false, true, true, false}
		s := NewSuccincter(input, func(b bool) bool { return b })

		if got := s.Rank(0); got != 0 {
			t.Errorf("Rank(0) = %d; want 0", got)
		}
	})

	t.Run("Select with rank 0", func(t *testing.T) {
		input := []bool{true, false, true, true, false}
		s := NewSuccincter(input, func(b bool) bool { return b })

		if got := s.Select(0); got != -1 {
			t.Errorf("Select(0) = %d; want -1", got)
		}
	})

	t.Run("Select beyond total ones", func(t *testing.T) {
		input := []bool{true, false, true}
		s := NewSuccincter(input, func(b bool) bool { return b })

		if got := s.Select(10); got != -1 {
			t.Errorf("Select(10) on 3-element array = %d; want -1", got)
		}
	})

	t.Run("Negative positions", func(t *testing.T) {
		input := []bool{true, false, true}
		s := NewSuccincter(input, func(b bool) bool { return b })

		if got := s.Rank(-5); got != 0 {
			t.Errorf("Rank(-5) = %d; want 0", got)
		}
		if got := s.Select(-3); got != -1 {
			t.Errorf("Select(-3) = %d; want -1", got)
		}
	})

	t.Run("Empty array operations", func(t *testing.T) {
		s := NewSuccincter([]bool{}, func(b bool) bool { return b })

		if got := s.Rank(0); got != 0 {
			t.Errorf("Rank(0) on empty = %d; want 0", got)
		}
		if got := s.Rank(1000); got != 0 {
			t.Errorf("Rank(1000) on empty = %d; want 0", got)
		}
		if got := s.Select(1); got != -1 {
			t.Errorf("Select(1) on empty = %d; want -1", got)
		}
	})
}

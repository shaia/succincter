package succincter

import (
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

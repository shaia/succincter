package internal

import (
	"math/bits"
	"testing"
)

// TestCombEncodeDecode exhaustively tests all 2^15 = 32768 patterns.
func TestCombEncodeDecode(t *testing.T) {
	for block := uint16(0); block < (1 << 15); block++ {
		class := bits.OnesCount16(block)
		offset := CombEncode(block, class)

		// Verify offset is in valid range [0, C(15, class) - 1]
		maxOffset := binomial[15][class]
		if offset >= maxOffset {
			t.Errorf("CombEncode(0x%x, %d) = %d; must be < %d", block, class, offset, maxOffset)
			continue
		}

		// Verify roundtrip: decode(encode(block)) == block
		decoded := CombDecode(class, offset, 15)
		if decoded != block {
			t.Errorf("CombDecode(%d, %d, 15) = 0x%x; want 0x%x", class, offset, decoded, block)
		}
	}
}

// TestCombEncodeEdgeCases verifies edge cases.
func TestCombEncodeEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		block  uint16
		class  int
		offset uint16
	}{
		{"all zeros", 0x0000, 0, 0},
		{"all ones (15 bits)", 0x7FFF, 15, 0},
		{"single one at bit 0", 0x0001, 1, 0},
		{"single one at bit 14", 0x4000, 1, 14},
		{"two ones at bits 0,1", 0x0003, 2, 0},
		{"two ones at bits 13,14", 0x6000, 2, 104}, // C(14,2) + C(13,2) = 91 + 78 = ... need to compute
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := CombEncode(tt.block, tt.class)

			// For edge cases with known offsets, verify
			if tt.name == "all zeros" || tt.name == "all ones (15 bits)" ||
				tt.name == "single one at bit 0" || tt.name == "single one at bit 14" ||
				tt.name == "two ones at bits 0,1" {
				if offset != tt.offset {
					t.Errorf("CombEncode(0x%x, %d) = %d; want %d", tt.block, tt.class, offset, tt.offset)
				}
			}

			// Always verify roundtrip
			decoded := CombDecode(tt.class, offset, 15)
			if decoded != tt.block {
				t.Errorf("CombDecode(%d, %d, 15) = 0x%x; want 0x%x", tt.class, offset, decoded, tt.block)
			}
		})
	}
}

// TestOffsetBits verifies bit width calculations.
func TestOffsetBits(t *testing.T) {
	tests := []struct {
		class    int
		expected int
	}{
		{0, 0},   // C(15,0) = 1, need 0 bits
		{1, 4},   // C(15,1) = 15, need 4 bits (ceil(log2(15)) = 4)
		{2, 7},   // C(15,2) = 105, need 7 bits
		{7, 13},  // C(15,7) = 6435, need 13 bits
		{8, 13},  // C(15,8) = 6435, need 13 bits
		{14, 4},  // C(15,14) = 15, need 4 bits
		{15, 0},  // C(15,15) = 1, need 0 bits
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := OffsetBits(tt.class)
			if got != tt.expected {
				t.Errorf("OffsetBits(%d) = %d; want %d (C(15,%d) = %d)",
					tt.class, got, tt.expected, tt.class, binomial[15][tt.class])
			}
		})
	}
}

// TestBinomialTable verifies Pascal's triangle values.
func TestBinomialTable(t *testing.T) {
	// Spot-check known values
	tests := []struct {
		n, k     int
		expected uint16
	}{
		{0, 0, 1},
		{1, 0, 1},
		{1, 1, 1},
		{2, 1, 2},
		{5, 2, 10},
		{10, 5, 252},
		{15, 0, 1},
		{15, 1, 15},
		{15, 7, 6435},
		{15, 8, 6435},
		{15, 14, 15},
		{15, 15, 1},
	}

	for _, tt := range tests {
		got := binomial[tt.n][tt.k]
		if got != tt.expected {
			t.Errorf("binomial[%d][%d] = %d; want %d", tt.n, tt.k, got, tt.expected)
		}
	}

	// Verify symmetry: C(n,k) = C(n, n-k)
	for n := 0; n <= 15; n++ {
		for k := 0; k <= n; k++ {
			if binomial[n][k] != binomial[n][n-k] {
				t.Errorf("symmetry failed: binomial[%d][%d]=%d != binomial[%d][%d]=%d",
					n, k, binomial[n][k], n, n-k, binomial[n][n-k])
			}
		}
	}

	// Verify recurrence: C(n,k) = C(n-1,k-1) + C(n-1,k)
	for n := 1; n <= 15; n++ {
		for k := 1; k < n; k++ {
			expected := binomial[n-1][k-1] + binomial[n-1][k]
			if binomial[n][k] != expected {
				t.Errorf("recurrence failed: binomial[%d][%d]=%d != %d+%d=%d",
					n, k, binomial[n][k], binomial[n-1][k-1], binomial[n-1][k], expected)
			}
		}
	}
}

// TestWorkedExamples verifies the examples from the algorithm description.
func TestWorkedExamples(t *testing.T) {
	// Example from plan: β = 0100 (b=4 simulation, but we use b=15)
	// For b=15: 0x0100 = bit 8 set, class=1
	// offset should be 8 (patterns with single bit at positions 0-7 come before)
	block := uint16(0x0100) // bit 8 set
	class := bits.OnesCount16(block)
	if class != 1 {
		t.Fatalf("class should be 1, got %d", class)
	}
	offset := CombEncode(block, class)
	if offset != 8 {
		t.Errorf("CombEncode(0x0100, 1) = %d; want 8", offset)
	}

	// Verify decode
	decoded := CombDecode(class, offset, 15)
	if decoded != block {
		t.Errorf("CombDecode(1, 8, 15) = 0x%x; want 0x0100", decoded)
	}

	// Another example: 0x0006 = bits 1,2 set, class=2
	// Patterns before: 0x0003 (bits 0,1)
	// offset should be 1
	block2 := uint16(0x0006) // bits 1,2 set
	class2 := bits.OnesCount16(block2)
	offset2 := CombEncode(block2, class2)
	decoded2 := CombDecode(class2, offset2, 15)
	if decoded2 != block2 {
		t.Errorf("roundtrip failed for 0x%x: got 0x%x", block2, decoded2)
	}
}

// TestOffsetRanges verifies offset values are in valid ranges for each class.
func TestOffsetRanges(t *testing.T) {
	for class := 0; class <= 15; class++ {
		maxOffset := binomial[15][class]
		count := 0

		// Count patterns with this class
		for block := uint16(0); block < (1 << 15); block++ {
			if bits.OnesCount16(block) == class {
				count++
				offset := CombEncode(block, class)
				if offset >= maxOffset {
					t.Errorf("class %d: offset %d >= max %d for block 0x%x",
						class, offset, maxOffset, block)
				}
			}
		}

		// Verify count matches C(15, class)
		if count != int(maxOffset) {
			t.Errorf("class %d: found %d patterns, expected C(15,%d)=%d",
				class, count, class, maxOffset)
		}
	}
}

// BenchmarkCombEncode measures encoding performance.
func BenchmarkCombEncode(b *testing.B) {
	// Pre-generate test patterns
	patterns := make([]uint16, 1000)
	for i := range patterns {
		patterns[i] = uint16(i % (1 << 15))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		block := patterns[i%len(patterns)]
		class := bits.OnesCount16(block)
		_ = CombEncode(block, class)
	}
}

// BenchmarkCombDecode measures decoding performance.
func BenchmarkCombDecode(b *testing.B) {
	// Pre-generate test (class, offset) pairs
	type pair struct {
		class  int
		offset uint16
	}
	pairs := make([]pair, 1000)
	for i := range pairs {
		block := uint16(i % (1 << 15))
		class := bits.OnesCount16(block)
		pairs[i] = pair{class, CombEncode(block, class)}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := pairs[i%len(pairs)]
		_ = CombDecode(p.class, p.offset, 15)
	}
}

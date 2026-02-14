// Package internal provides low-level encoding primitives for RRR compression.
// Combinatorial number system encodes bit patterns as (class, offset) pairs.
package internal

import "math/bits"

// Binomial coefficient table: binomial[n][k] = C(n,k).
// Precomputed at init to avoid repeated computation on query hot path.
var binomial [16][16]uint16

func init() {
	// Build Pascal's triangle for n=0..15 using recurrence C(n,k) = C(n-1,k-1) + C(n-1,k).
	// Base cases: C(n,0) = C(n,n) = 1.
	for n := 0; n <= 15; n++ {
		binomial[n][0] = 1
		binomial[n][n] = 1
		for k := 1; k < n; k++ {
			binomial[n][k] = binomial[n-1][k-1] + binomial[n-1][k]
		}
	}
}

// CombEncode returns the combinatorial rank of a bit pattern.
// Given a block with `class` 1-bits, returns its index among all C(15, class) patterns.
//
// Algorithm: Iterate bits high-to-low. For each 1-bit at position p with r ones remaining,
// add C(p, r) to offset. This counts patterns lexicographically smaller than the current pattern.
//
// Example: block=0b0101 (class=2) -> offset is the count of 2-bit patterns less than 0b0101.
func CombEncode(block uint16, class int) uint16 {
	// Class 0 and class 15 have only one pattern each (all zeros, all ones).
	if class == 0 || class == 15 {
		return 0
	}

	offset := uint16(0)
	onesLeft := class

	// Traverse bits from MSB to LSB. Each 1-bit contributes C(bitPos, onesLeft) to the offset.
	for bitPos := 14; bitPos >= 0 && onesLeft > 0; bitPos-- {
		if (block & (1 << bitPos)) != 0 {
			if onesLeft > 0 {
				// C(bitPos, onesLeft) counts patterns with a 0 at this position.
				offset += binomial[bitPos][onesLeft]
			}
			onesLeft--
		}
	}

	return offset
}

// CombDecode reconstructs a bit pattern from its combinatorial rank.
// Given class and offset, returns the unique block with `class` 1-bits at rank `offset`.
//
// Algorithm: Greedy bit placement from MSB to LSB. For each position, if C(pos, onesLeft) <= offset,
// place a 1-bit and subtract the coefficient. This reverses the encoding process.
func CombDecode(class int, offset uint16, b int) uint16 {
	// Edge cases: class 0 -> all zeros, class b -> all ones.
	if class == 0 {
		return 0
	}
	if class == b {
		return (1 << b) - 1
	}

	result := uint16(0)
	onesLeft := class

	// Greedily place 1-bits from MSB to LSB.
	for bitPos := b - 1; bitPos >= 0 && onesLeft > 0; bitPos-- {
		coeff := binomial[bitPos][onesLeft]
		if offset >= coeff {
			// Place 1-bit at this position.
			result |= (1 << bitPos)
			offset -= coeff
			onesLeft--
		}
	}

	return result
}

// OffsetBits returns the number of bits required to store an offset for a given class.
// Returns ceil(log2(C(15, class))). Zero for class 0 and 15 since they have offset=0.
func OffsetBits(class int) int {
	if class == 0 || class == 15 {
		return 0
	}
	// bits.Len16(x) returns floor(log2(x)) + 1 = ceil(log2(x+1)).
	// For C(15,class) possible values, we need ceil(log2(C(15,class))) bits.
	return bits.Len16(binomial[15][class] - 1)
}

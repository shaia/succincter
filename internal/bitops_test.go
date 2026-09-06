package internal

import (
	"math"
	"math/bits"
	"testing"
)

func TestBinarySearchHighBits(t *testing.T) {
	pack := func(rank uint64, offsetBits uint64) uint64 {
		return rank<<32 | offsetBits
	}

	tests := []struct {
		name   string
		array  []uint64
		target int
		want   int
	}{
		{
			name:   "empty array",
			array:  nil,
			target: 5,
			want:   -1,
		},
		{
			name:   "single element, target below",
			array:  []uint64{pack(10, 0xDEADBEEF)},
			target: 5,
			want:   -1,
		},
		{
			name:   "single element, target equal (strict less)",
			array:  []uint64{pack(10, 0xDEADBEEF)},
			target: 10,
			want:   -1,
		},
		{
			name:   "single element, target above",
			array:  []uint64{pack(10, 0xDEADBEEF)},
			target: 11,
			want:   0,
		},
		{
			name:   "target equal to element's high bits returns previous index",
			array:  []uint64{pack(0, 1), pack(5, 2), pack(10, 3), pack(15, 4)},
			target: 10,
			want:   1,
		},
		{
			name:   "target between elements",
			array:  []uint64{pack(0, 1), pack(5, 2), pack(10, 3), pack(15, 4)},
			target: 7,
			want:   1,
		},
		{
			name:   "target below all elements",
			array:  []uint64{pack(5, 1), pack(10, 2), pack(15, 3)},
			target: 1,
			want:   -1,
		},
		{
			name:   "target above all elements",
			array:  []uint64{pack(5, 1), pack(10, 2), pack(15, 3)},
			target: 100,
			want:   2,
		},
		{
			name:   "low 32 bits ignored when comparing",
			array:  []uint64{pack(0, 0xFFFFFFFF), pack(10, 0xFFFFFFFF), pack(20, 0xFFFFFFFF)},
			target: 15,
			want:   1,
		},
		{
			name:   "zero target returns -1",
			array:  []uint64{pack(0, 1), pack(5, 2)},
			target: 0,
			want:   -1,
		},
		{
			name:   "negative target returns -1",
			array:  []uint64{pack(0, 1), pack(5, 2)},
			target: -5,
			want:   -1,
		},
		{
			// A rank above MaxInt32 must not wrap negative where int is
			// 32 bits, which would make it compare as less than target.
			name:   "high bits above MaxInt32 compare as unsigned",
			array:  []uint64{pack(0, 1), pack(3000000000, 2)},
			target: 2000000000,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BinarySearchHighBits(tt.array, tt.target)
			if got != tt.want {
				t.Errorf("BinarySearchHighBits(%v, %d) = %d; want %d", tt.array, tt.target, got, tt.want)
			}
		})
	}
}

// A target above MaxUint32 must not be narrowed to uint32, which would
// truncate it and understate every element. Only reachable where int is
// 64 bits.
func TestBinarySearchHighBitsTargetAboveMaxUint32(t *testing.T) {
	if bits.UintSize < 64 {
		t.Skip("requires a 64-bit int")
	}
	// Built through a variable so the conversion is not a constant
	// expression, which would not compile on 32-bit builds.
	var maxU32 uint64 = math.MaxUint32
	target := int(maxU32) + 4

	array := []uint64{5 << 32, 10 << 32}
	if got := BinarySearchHighBits(array, target); got != 1 {
		t.Errorf("BinarySearchHighBits(array, %d) = %d; want 1", target, got)
	}
}

package internal

import "testing"

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

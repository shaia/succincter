package internal

import "math/bits"

// Popcount returns the number of 1-bits in x.
// Uses hardware POPCNT instruction where available.
func Popcount(x uint64) int {
	return bits.OnesCount64(x)
}

// SelectInBlock returns the position of the rank-th 1-bit within a 64-bit block.
// Returns -1 if the block has fewer than rank 1-bits.
func SelectInBlock(block uint64, rank int) int {
	count := 0
	for i := 0; i < 64; i++ {
		if (block & (uint64(1) << i)) != 0 {
			count++
			if count == rank {
				return i
			}
		}
	}
	return -1
}

// BinarySearch returns the index of the last element strictly less than target.
// Returns -1 if no element is less than target.
func BinarySearch(array []uint64, target int) int {
	low, high := 0, len(array)-1
	for low <= high {
		mid := (low + high) / 2
		if int(array[mid]) < target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return low - 1
}

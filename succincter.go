package succincter

import "github.com/shaia/succincter/internal"

type Succincter struct {
	data           []uint64 // Data stored in blocks
	blockRanks     []uint32 // Precomputed rank information for each block
	superBlocks    []uint32 // Precomputed rank for super-blocks
	blockSize      int      // Size of a block in bits
	superBlockSize int      // Size of a super-block in bits
}

func NewSuccincter[T any](input []T, predicate func(T) bool) *Succincter {
	// Initialize block and superblock sizes
	blockSize := 64        // For example, a block is 64 bits
	superBlockSize := 1024 // A super-block contains multiple blocks

	// Convert input to compact representation
	data := internal.CompressToBitVector(input, predicate)

	// Precompute rank data for blocks and super-blocks
	blockRanks, superBlocks := precomputeRank(data, superBlockSize)

	return &Succincter{
		data:           data,
		blockRanks:     blockRanks,
		superBlocks:    superBlocks,
		blockSize:      blockSize,
		superBlockSize: superBlockSize,
	}
}

func (s *Succincter) Rank(pos int) int {
	blockIndex := pos / s.blockSize
	offset := pos % s.blockSize
	rank := int(s.superBlocks[blockIndex/s.superBlockSize])
	rank += int(s.blockRanks[blockIndex])
	rank += popcount(s.data[blockIndex] & ((1 << offset) - 1))
	return rank
}

func (s *Succincter) Select(rank int) int {
	// Binary search for the superblock and block containing the rank
	superBlockIndex := binarySearch(s.superBlocks, rank)
	blockIndex := binarySearch(s.blockRanks[superBlockIndex:], rank)
	blockRank := rank - int(s.blockRanks[blockIndex])
	return blockIndex*s.blockSize + selectInBlock(s.data[blockIndex], blockRank)
}

func precomputeRank(data []uint64, superBlockSize int) ([]uint32, []uint32) {
	var blockRanks []uint32
	var superBlocks []uint32
	currentRank := uint32(0)

	for i, block := range data {
		if i%superBlockSize == 0 {
			superBlocks = append(superBlocks, currentRank)
		}
		blockRanks = append(blockRanks, currentRank)
		currentRank += uint32(popcount(block)) // Count 1-bits in the block
	}

	return blockRanks, superBlocks
}

func popcount(x uint64) int {
	count := 0
	for x != 0 {
		count += int(x & 1)
		x >>= 1
	}
	return count
}

func selectInBlock(block uint64, rank int) int {
	count := 0
	for i := 0; i < 64; i++ {
		if (block & (1 << i)) != 0 {
			count++
			if count == rank {
				return i
			}
		}
	}
	return -1 // Not found
}

func binarySearch(array []uint32, target int) int {
	low, high := 0, len(array)-1
	for low <= high {
		mid := (low + high) / 2
		if int(array[mid]) <= target {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return low - 1
}

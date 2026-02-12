package succincter

import "github.com/shaia/succincter/internal"

// RankSelector is the interface for data structures supporting rank and select queries.
type RankSelector interface {
	Rank(pos int) int
	Select(rank int) int
}

// Succincter is a succinct data structure for O(1) rank and O(log n) select queries
// on boolean arrays, with ~1.5 bits per element overhead.
type Succincter struct {
	data               []uint64
	blockRanks         []uint64
	superBlocks        []uint64
	blockSize          int
	superBlockSize     int
	blocksPerSuperBlock int
	totalOnes          int
}

// NewSuccincter constructs a Succincter from any slice using a predicate to determine 1-bits.
// Construction is O(n).
func NewSuccincter[T any](input []T, predicate func(T) bool) *Succincter {
	blockSize := 64
	superBlockSize := 1024
	blocksPerSuperBlock := superBlockSize / blockSize

	data := internal.CompressToBitVector(input, predicate)
	blockRanks, superBlocks, totalOnes := precomputeRank(data, blocksPerSuperBlock)

	return &Succincter{
		data:               data,
		blockRanks:         blockRanks,
		superBlocks:        superBlocks,
		blockSize:          blockSize,
		superBlockSize:     superBlockSize,
		blocksPerSuperBlock: blocksPerSuperBlock,
		totalOnes:          totalOnes,
	}
}

// Rank returns the count of 1-bits before position pos. O(1) time.
// Returns 0 for pos <= 0 or empty arrays.
func (s *Succincter) Rank(pos int) int {
	if pos <= 0 || len(s.data) == 0 {
		return 0
	}
	maxPos := len(s.data) * s.blockSize
	if pos >= maxPos {
		pos = maxPos - 1
	}
	blockIndex := pos / s.blockSize
	offset := pos % s.blockSize
	rank := int(s.blockRanks[blockIndex])
	rank += internal.Popcount(s.data[blockIndex] & ((1 << offset) - 1))
	return rank
}

// Select returns the position of the rank-th 1-bit (1-indexed). O(log n) time.
// Returns -1 for invalid ranks or empty arrays.
func (s *Succincter) Select(rank int) int {
	if rank <= 0 || len(s.data) == 0 {
		return -1
	}
	if rank > s.totalOnes {
		return -1
	}

	superBlockIndex := internal.BinarySearch(s.superBlocks, rank)

	startBlock := superBlockIndex * s.blocksPerSuperBlock
	endBlock := startBlock + s.blocksPerSuperBlock
	if endBlock > len(s.blockRanks) {
		endBlock = len(s.blockRanks)
	}
	relativeBlockIndex := internal.BinarySearch(s.blockRanks[startBlock:endBlock], rank)
	if relativeBlockIndex < 0 {
		relativeBlockIndex = 0
	}
	absoluteBlockIndex := startBlock + relativeBlockIndex

	blockRank := rank - int(s.blockRanks[absoluteBlockIndex])
	return absoluteBlockIndex*s.blockSize + internal.SelectInBlock(s.data[absoluteBlockIndex], blockRank)
}

func precomputeRank(data []uint64, blocksPerSuperBlock int) ([]uint64, []uint64, int) {
	var blockRanks []uint64
	var superBlocks []uint64
	currentRank := uint64(0)

	for i, block := range data {
		if i%blocksPerSuperBlock == 0 {
			superBlocks = append(superBlocks, currentRank)
		}
		blockRanks = append(blockRanks, currentRank)
		currentRank += uint64(internal.Popcount(block))
	}

	return blockRanks, superBlocks, int(currentRank)
}

package succincter

import (
	"math"
	"math/bits"

	"github.com/shaia/succincter/internal"
)

// RRR is a zero-order compressed rank/select structure using Raman-Raman-Rao encoding.
// Achieves nH0(B) + o(n) space with O(1) rank and O(log n) select queries.
// Rank queries are slower than Succincter (~50-200ns vs ~3ns) due to combinatorial decoding.
// Use RRR when memory is constrained and bitvector density is far from 50%.
// Immutable after construction; safe for concurrent reads without synchronization.
type RRR struct {
	classes            []uint64
	offsets            []uint64
	superBlocks        []uint64
	blockSize          int
	numBlocks          int
	totalOnes          int
	superBlockInterval int
}

// NewRRR constructs an RRR-compressed rank/select structure from input.
// The predicate determines which elements map to 1-bits. Construction is O(n).
// Panics if input produces more than 2^32 ones or exceeds uint32 offset capacity.
func NewRRR[T any](input []T, predicate func(T) bool) *RRR {
	blockSize := 15
	superBlockInterval := 16

	if len(input) == 0 {
		return &RRR{
			classes:            []uint64{},
			offsets:            []uint64{},
			superBlocks:        []uint64{0},
			blockSize:          blockSize,
			numBlocks:          0,
			totalOnes:          0,
			superBlockInterval: superBlockInterval,
		}
	}

	numBlocks := (len(input) + blockSize - 1) / blockSize
	numClassWords := (numBlocks*4 + 63) / 64
	classes := make([]uint64, numClassWords)

	blockClasses := make([]int, numBlocks)
	blockOffsets := make([]uint16, numBlocks)
	totalOffsetBits := 0
	totalOnes := 0

	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		start := blockIdx * blockSize
		end := start + blockSize
		if end > len(input) {
			end = len(input)
		}

		block := uint16(0)
		for i := start; i < end; i++ {
			if predicate(input[i]) {
				block |= 1 << (i - start)
			}
		}

		class := bits.OnesCount16(block)
		offset := internal.CombEncode(block, class)

		blockClasses[blockIdx] = class
		blockOffsets[blockIdx] = offset
		totalOnes += class
		totalOffsetBits += internal.OffsetBits(class)

		packClass(classes, blockIdx, class)
	}

	if totalOnes > math.MaxUint32 {
		panic("succincter: RRR does not support more than 2^32 ones; use Succincter for very large dense datasets")
	}
	if totalOffsetBits > math.MaxUint32 {
		panic("succincter: RRR offset storage exceeds uint32 capacity")
	}

	numOffsetWords := (totalOffsetBits + 63) / 64
	offsets := make([]uint64, numOffsetWords)
	bitPtr := 0
	for blockIdx := 0; blockIdx < numBlocks; blockIdx++ {
		class := blockClasses[blockIdx]
		offset := blockOffsets[blockIdx]
		width := internal.OffsetBits(class)
		if width > 0 {
			packOffset(offsets, bitPtr, offset, width)
			bitPtr += width
		}
	}

	numSuperBlocks := (numBlocks + superBlockInterval - 1) / superBlockInterval
	superBlocks := make([]uint64, numSuperBlocks)
	cumulativeRank := 0
	cumulativeOffsetBits := 0
	for sbIdx := 0; sbIdx < numSuperBlocks; sbIdx++ {
		superBlocks[sbIdx] = uint64(cumulativeRank)<<32 | uint64(cumulativeOffsetBits)

		endBlock := (sbIdx + 1) * superBlockInterval
		if endBlock > numBlocks {
			endBlock = numBlocks
		}
		for blockIdx := sbIdx * superBlockInterval; blockIdx < endBlock; blockIdx++ {
			cumulativeRank += blockClasses[blockIdx]
			cumulativeOffsetBits += internal.OffsetBits(blockClasses[blockIdx])
		}
	}

	return &RRR{
		classes:            classes,
		offsets:            offsets,
		superBlocks:        superBlocks,
		blockSize:          blockSize,
		numBlocks:          numBlocks,
		totalOnes:          totalOnes,
		superBlockInterval: superBlockInterval,
	}
}

// Rank returns the count of 1-bits at positions < pos.
// Returns 0 for pos <= 0 and totalOnes for pos >= length.
func (r *RRR) Rank(pos int) int {
	if pos <= 0 || r.numBlocks == 0 {
		return 0
	}

	maxPos := r.numBlocks * r.blockSize
	if pos >= maxPos {
		return r.totalOnes
	}

	targetBlock := pos / r.blockSize
	offsetInBlock := pos % r.blockSize

	sbIdx := targetBlock / r.superBlockInterval
	sb := r.superBlocks[sbIdx]
	rank := int(sb >> 32)
	offsetBitPtr := int(sb & 0xFFFFFFFF)

	startBlock := sbIdx * r.superBlockInterval
	for blockIdx := startBlock; blockIdx < targetBlock; blockIdx++ {
		class := readClass(r.classes, blockIdx)
		rank += class
		offsetBitPtr += internal.OffsetBits(class)
	}

	targetClass := readClass(r.classes, targetBlock)
	if targetClass == 0 {
		return rank
	}

	targetWidth := internal.OffsetBits(targetClass)
	targetOffset := readOffset(r.offsets, offsetBitPtr, targetWidth)
	block := internal.CombDecode(targetClass, targetOffset, r.blockSize)

	mask := (uint16(1) << offsetInBlock) - 1
	rank += bits.OnesCount16(block & mask)

	return rank
}

// Select returns the position of the rank-th 1-bit (1-indexed).
// Returns -1 for rank <= 0 or rank > totalOnes.
func (r *RRR) Select(rank int) int {
	if rank <= 0 || r.numBlocks == 0 {
		return -1
	}
	if rank > r.totalOnes {
		return -1
	}

	sbIdx := 0
	for sbIdx < len(r.superBlocks)-1 {
		nextSBRank := int(r.superBlocks[sbIdx+1] >> 32)
		if nextSBRank >= rank {
			break
		}
		sbIdx++
	}

	sb := r.superBlocks[sbIdx]
	cumulativeRank := int(sb >> 32)
	offsetBitPtr := int(sb & 0xFFFFFFFF)

	startBlock := sbIdx * r.superBlockInterval
	endBlock := startBlock + r.superBlockInterval
	if endBlock > r.numBlocks {
		endBlock = r.numBlocks
	}

	blockIdx := startBlock
	for blockIdx < endBlock {
		class := readClass(r.classes, blockIdx)
		if cumulativeRank+class >= rank {
			break
		}
		cumulativeRank += class
		offsetBitPtr += internal.OffsetBits(class)
		blockIdx++
	}

	class := readClass(r.classes, blockIdx)
	offset := readOffset(r.offsets, offsetBitPtr, internal.OffsetBits(class))
	block := internal.CombDecode(class, offset, r.blockSize)

	return blockIdx*r.blockSize + internal.SelectInBlock(uint64(block), rank-cumulativeRank)
}

func packClass(classes []uint64, index int, class int) {
	wordIdx := (index * 4) / 64
	bitOffset := (index * 4) % 64
	classes[wordIdx] &= ^(uint64(0xF) << bitOffset)
	classes[wordIdx] |= uint64(class&0xF) << bitOffset
}

func readClass(classes []uint64, index int) int {
	wordIdx := (index * 4) / 64
	bitOffset := (index * 4) % 64
	return int((classes[wordIdx] >> bitOffset) & 0xF)
}

func packOffset(offsets []uint64, bitPtr int, value uint16, width int) {
	if width == 0 {
		return
	}
	wordIdx := bitPtr / 64
	bitOffset := bitPtr % 64

	if bitOffset+width <= 64 {
		mask := uint64((1 << width) - 1)
		offsets[wordIdx] &= ^(mask << bitOffset)
		offsets[wordIdx] |= uint64(value) << bitOffset
		return
	}

	bitsInFirstWord := 64 - bitOffset
	bitsInSecondWord := width - bitsInFirstWord

	mask1 := uint64((1 << bitsInFirstWord) - 1)
	offsets[wordIdx] &= ^(mask1 << bitOffset)
	offsets[wordIdx] |= (uint64(value) & mask1) << bitOffset

	mask2 := uint64((1 << bitsInSecondWord) - 1)
	offsets[wordIdx+1] &= ^mask2
	offsets[wordIdx+1] |= (uint64(value) >> bitsInFirstWord) & mask2
}

func readOffset(offsets []uint64, bitPtr int, width int) uint16 {
	if width == 0 {
		return 0
	}
	wordIdx := bitPtr / 64
	bitOffset := bitPtr % 64

	if bitOffset+width <= 64 {
		mask := uint64((1 << width) - 1)
		return uint16((offsets[wordIdx] >> bitOffset) & mask)
	}

	bitsInFirstWord := 64 - bitOffset
	bitsInSecondWord := width - bitsInFirstWord

	mask1 := uint64((1 << bitsInFirstWord) - 1)
	part1 := (offsets[wordIdx] >> bitOffset) & mask1

	mask2 := uint64((1 << bitsInSecondWord) - 1)
	part2 := offsets[wordIdx+1] & mask2

	return uint16(part1 | (part2 << bitsInFirstWord))
}

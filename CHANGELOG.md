# Changelog

All notable changes to the Succincter library will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Breaking Changes

- **uint64 Migration**: `blockRanks` and `superBlocks` fields changed from `[]uint32` to `[]uint64`
  - **Reason**: Prevents integer overflow for arrays with more than 537 million elements
  - **Impact**: Binary serialization format incompatible with previous versions
  - **Migration**: Recompile your code. No API changes required.

### Performance

- **Hardware popcount**: Replaced software bit counting with `math/bits.OnesCount64`
  - Uses hardware POPCNT instruction where available, automatic software fallback otherwise

### Added

- **Combinatorial encoding (Z1)**: `internal/combinatorial.go` with `CombEncode`, `CombDecode`, `OffsetBits`
  - Foundation for RRR zero-order compression
  - Precomputed binomial coefficient table for b=15
  - Exhaustive tests for all 32768 15-bit patterns

### Fixed

- Empty array panic: `Rank()` and `Select()` now handle empty arrays gracefully
- Select offset bug: Fixed binary search index calculation across block boundaries
- Superblock indexing: Fixed `precomputeRank` to group blocks correctly (every 16 blocks, not every 1024)
- Negative position handling: `Rank()` returns 0 for negative/zero positions
- Invalid rank handling: `Select()` returns -1 for invalid ranks
- **32-bit overflow**: Fixed `(1 << offset)` to use `uint64(1) << offset` in `Rank()` and `SelectInBlock()`
  - Prevents undefined behavior on 32-bit platforms when offset >= 32
- **Rank boundary fix**: `Rank(pos)` now returns `totalOnes` when pos >= maxPos instead of incorrect calculation
- **Defensive bounds check**: Added guard for `superBlockIndex < 0` in `Select()`
- **Fuzz test invariant**: Strengthened `Rank(0)` check from `>= 0` to `== 0`

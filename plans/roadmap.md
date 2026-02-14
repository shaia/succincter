# Succincter Library Production Roadmap

## Progress

### Bug Fixes

- [x] **M1: Fix Empty Array Panic** — Bounds checking in `Rank()` and `Select()`
- [x] **M2: Fix Select Offset Bug** — Binary search index conversion (relative → absolute)
- [x] **M3: Fix uint32 Overflow** — Migrated `blockRanks`/`superBlocks` to `[]uint64`
- [x] **M4: Optimize Popcount** — Replaced software loop with `math/bits.OnesCount64`

### Testing

- [x] **M5: Comprehensive Test Coverage** — 95.0% coverage, fuzz tests, property tests, concurrent reads
  - [x] Fuzz tests (`FuzzRank`, `FuzzSelect`) pass 10s+ without failure
  - [x] Property tests: monotonicity, differential, inverse
  - [x] Concurrent reads: 10 goroutines, no races
  - [ ] `go test -race` (requires CGO on Windows — not validated)

### Documentation

- [x] **M6: API Documentation** — Godoc on all exported types/functions
- [x] **M7: Root README** — Overview, install, quick start, API, performance, thread-safety
- [x] **M8: Alternative Evaluation** — GO decision documented in Decision Log section of this roadmap

### Performance
- [x] **M9: Performance Validation** — Benchmarks confirm O(1) rank (~13ns), O(log n) select (~100-120ns)
  - [x] Scalability benchmarks across 1K–10M elements
  - [x] Memory overhead validated at ~1.5 bits/element

### Refactoring
- [x] **R1: Extract bit operations** — `internal/bitops.go` (Popcount, SelectInBlock, BinarySearch)
- [x] **R2: RankSelector interface** — Shared contract for Succincter and SimpleArray
- [x] **R3: Eliminate redundant O(n) pass** — `precomputeRank` returns totalOnes directly
- [x] **R4: Cache blocksPerSuperBlock** — Computed once in constructor
- [x] **R5: Fix comparison test** — `TestCompareImplementations` now compares Succincter vs SimpleArray

### Zero-Order Compression (RRR)

- [x] **Z1: Combinatorial encoding** — `internal/combinatorial.go`: binomial table, CombEncode, CombDecode, OffsetBits
- [ ] **Z2: RRR construction** — `rrr.go`: NewRRR with block size 15, class/offset packing, superblock index
- [ ] **Z3: RRR Rank/Select** — O(1) rank and O(log n) select on compressed bitvector
- [ ] **Z4: RRR tests** — Exhaustive combinatorial tests, cross-validation vs Succincter, boundary tests
- [ ] **Z5: RRR benchmarks** — Build/Rank/Select benchmarks, space measurement, comparison vs Succincter
- [ ] **Z6: RRR fuzz tests** — FuzzRRRRank, FuzzRRRSelect with cross-validation
- [ ] **Z7: Encode/decode walkthrough** — Document combinatorial encoding algorithm with worked example (e.g., β=0100 → o=2)

### Higher-Order Compression (Hk)

- [ ] **H1: HkOptions and constants** — `hk.go`: ContextOrder enum (K0/K1/K2/KAdaptive), HkOptions struct, context-indexed binomial tables
- [ ] **H2: Hk construction (k=0,1,2)** — NewHk constructor, per-block context state, superblock index with context snapshot
- [ ] **H3: Hk Rank/Select** — O(1) rank and O(log n) select with context tracking per block
- [ ] **H4: Adaptive mode** — `internal/entropy.go`: local H₀ estimation, per-superblock k selection (thresholds 0.9/0.7)
- [ ] **H5: Hk tests and benchmarks** — Property-based tests, fuzz tests, cross-validation vs Succincter/RRR, space measurement
- [ ] **H6: Hk documentation** — docs/higher-order-compression.md (Hk vs H₀ tradeoffs), README example

### Remaining
- [ ] **License file** — Add MIT or Apache 2.0
- [ ] **go test -race** — Validate with CGO enabled
- [ ] **v0.1.0 tag** — Tag first pre-release after final review

---

## Decision Log

| Decision                                          | Reasoning                                                                                  |
|---------------------------------------------------|--------------------------------------------------------------------------------------------|
| Hybrid parallel workstreams over sequential       | Parallel allows early course correction; Week 3 decision point enables pivot               |
| Fix existing library over find/build alternative  | Isolated bugs in predictable patterns; 3-4 weeks vs 4-6 weeks for new implementation       |
| uint64 for rank arrays over uint32 limitations    | uint32 limits to 537M elements with silent data corruption; uint64 is correct solution     |
| 95% coverage achieved (85% target)                | Target was 85% (diminishing returns); exceeded to 95% via fuzz testing edge cases          |
| Fuzz testing over only example-based              | Example tests found 0 of 4 bugs; fuzz testing explores input space systematically          |
| Hardware popcount (math/bits) over software       | Software popcount 10-50x slower; math/bits provides automatic fallback                     |
| Pre-1.0 semver policy                             | Breaking changes needed (uint64 migration); pre-1.0 signals API instability                |
| Channel-based sync rejected                       | Library is read-only after construction; "read-safe, write requires external sync"         |

## Constraints

- **Technical**: Go 1.21+, zero external dependencies, pure Go stdlib, generic API preserved
- **Performance**: Must maintain O(1) rank and O(log n) select complexity
- **Dependencies**: Only `math/bits` (stdlib); test dependencies OK

## Architecture

```
Construction Phase (NewSuccincter):
  Input []T + Predicate → CompressToBitVector → []uint64 data blocks
  → precomputeRank → Two-level index:
    - blockRanks []uint64 (cumulative rank per block)
    - superBlocks []uint64 (cumulative rank per superblock, every 16 blocks)

Query Phase:
  Rank(pos) → O(1): blockRank lookup + popcount remaining bits
  Select(rank) → O(log n): binary search superblocks → blocks → scan bits
```

### Invariants

1. **Immutability**: After construction, all fields are read-only (enables lock-free concurrent reads)
2. **Rank monotonicity**: blockRanks[i] <= blockRanks[i+1]
3. **Block alignment**: data[i] corresponds to blockRanks[i]
4. **Super-block divisibility**: superBlocks index every 16 blocks (1024/64)
5. **Rank bounds**: Rank(pos) <= pos
6. **Select inverse**: For position of a 1-bit, Select(Rank(pos)) == pos

### Design Tradeoffs

- **Memory vs Speed**: ~1.5 bits/element overhead for O(1) rank (vs O(n) naive)
- **Construction vs Query**: One-time O(n) construction for amortized O(1) queries
- **Generic API vs Performance**: Generic predicate adds negligible overhead vs flexibility gained
- **uint64 vs uint32 Ranks**: uint64 doubles rank memory but prevents silent overflow at >537M elements

## Milestone Dependencies

```
M1 (Empty Array) ──┐
M2 (Select Offset) ─┼──→ M5 (Test Coverage) ──→ M9 (Performance)
M3 (uint32 Overflow)─┘
M4 (Popcount) ──────────→ M9 (Performance)
M6 (Godoc) ─────────┐
M8 (Alternatives) ──┴──→ M7 (README + Docs)
```

## Known Risks

| Risk                                      | Mitigation                                            |
|-------------------------------------------|-------------------------------------------------------|
| Fixes reveal deeper architectural issues  | Week 3 GO/NO-GO decision point allows pivot           |
| uint64 migration breaks existing users    | Pre-1.0 semver; CHANGELOG documents breaking change   |
| Fuzz testing discovers unbounded issues   | Time-box to 2 hours; triage by severity               |
| Race detector finds concurrency issues    | Document as "read-only after construction"            |

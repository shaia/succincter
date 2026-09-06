# Benchmark results

Raw measurements backing the numbers quoted in
[Finding Errors in Log Streams](https://slow-is-smooth.io/blog/finding-errors-in-log-streams/).

**Environment**

| | |
|---|---|
| CPU | 13th Gen Intel(R) Core(TM) i9-13980HX (32 logical cores) |
| OS / arch | windows / amd64 |
| Go | go1.25.5 |
| Date | 2026-09-05 |

**Methodology.** Benchmarks ran from a separate module that imports the library
through a `replace` directive, so the repository's `go.mod` was untouched. Every
timed loop accumulates its result into a package-level `var sink int`. This
matters: without it the compiler may inline `Rank` and discard the call, and the
measured time collapses toward zero.

Log entries are generated with a fixed seed (`rand.NewSource(42)`) at a 5% error
rate, so runs are reproducible. The `LogEntry` struct is deliberately lean (one
`string` field). A fatter struct makes the naive scan slower through worse cache
behaviour, which would inflate the speedup — this is the conservative choice.

---

## Two Rank numbers, and why they differ

The repository quoted both "~2ns" (post, `examples/README.md`) and "~13ns"
(`README.md`, `plans/roadmap.md`). **Both are correct.** They measure different
things, and the difference is entirely cache behaviour:

| Access pattern | 1M entries | What it represents |
|---|---|---|
| Same position, every iteration | **2.56 ns** | Best case. The three cache lines involved never leave L1. |
| Scattered positions (1024-entry ring) | **16.56 ns** | Realistic. A dashboard queries different positions. |

The scattered figure includes ~3 ns of ring-buffer indexing overhead, measured
separately below. The post quotes the scattered numbers.

---

## Rank — same position (hot cache)

```
BenchmarkRank/Naive/10K-32                331960          3601 ns/op       0 B/op   0 allocs/op
BenchmarkRank/Succincter/10K-32        473400606             2.606 ns/op   0 B/op   0 allocs/op
BenchmarkRank/Naive/100K-32                 8030        161311 ns/op       0 B/op   0 allocs/op
BenchmarkRank/Succincter/100K-32       482023147             2.546 ns/op   0 B/op   0 allocs/op
BenchmarkRank/Naive/1M-32                    597       1878963 ns/op       0 B/op   0 allocs/op
BenchmarkRank/Succincter/1M-32         457384180             2.562 ns/op   0 B/op   0 allocs/op
BenchmarkRank/Naive/10M-32                    51      20600498 ns/op       0 B/op   0 allocs/op
BenchmarkRank/Succincter/10M-32        454813346             2.608 ns/op   0 B/op   0 allocs/op
```

## Rank — scattered positions (realistic)

```
BenchmarkRankScattered/RingOverhead/10K-32     798635397    1.434 ns/op
BenchmarkRankScattered/Succincter/10K-32       462607618    2.715 ns/op
BenchmarkRankScattered/RingOverhead/100K-32    849758455    2.733 ns/op
BenchmarkRankScattered/Succincter/100K-32      100000000   13.97  ns/op
BenchmarkRankScattered/RingOverhead/1M-32      437276350    3.178 ns/op
BenchmarkRankScattered/Succincter/1M-32         78592666   16.56  ns/op
BenchmarkRankScattered/RingOverhead/10M-32     376306488    3.095 ns/op
BenchmarkRankScattered/Succincter/10M-32        70134423   17.97  ns/op
```

Note the naive scan's *per-entry* cost also grows with n, because the scan falls
out of successive cache levels: 0.72 ns/entry at 10K, 3.2 at 100K, 3.8 at 1M,
4.1 at 10M. Naive is O(n), but with a worsening constant.

## Select

```
BenchmarkSelect/Naive/10K-32              299092      4445 ns/op
BenchmarkSelect/Succincter/10K-32       44289267        26.46 ns/op
BenchmarkSelect/Naive/100K-32               7219    161533 ns/op
BenchmarkSelect/Succincter/100K-32      50476582        23.99 ns/op
BenchmarkSelect/Naive/1M-32                  615   1966941 ns/op
BenchmarkSelect/Succincter/1M-32        80082751        16.67 ns/op
BenchmarkSelect/Naive/10M-32                  49  22322494 ns/op
BenchmarkSelect/Succincter/10M-32       50638680        30.02 ns/op

BenchmarkSelectScattered/Succincter/10K-32     6258124   228.1 ns/op
BenchmarkSelectScattered/Succincter/100K-32    3781868   314.8 ns/op
BenchmarkSelectScattered/Succincter/1M-32      2775885   423.4 ns/op
BenchmarkSelectScattered/Succincter/10M-32     2140524   533.0 ns/op
```

Select is far more cache-sensitive than Rank — 17–30 ns for a repeated rank
versus 228–533 ns for scattered ranks. Two reasons: `BinarySearch` over the
superblock array walks ~13 cold cache lines at 10M, and `SelectInBlock`
([internal/bitops.go:13](../../internal/bitops.go#L13)) is a plain linear scan of
up to 64 bit positions rather than a constant-time bit trick.

## Build

```
BenchmarkBuild/10K-32      30058     38338 ns/op      5728 B/op   16 allocs/op
BenchmarkBuild/100K-32      2662    488960 ns/op     55264 B/op   23 allocs/op
BenchmarkBuild/1M-32         241   5114798 ns/op    645088 B/op   34 allocs/op
BenchmarkBuild/10M-32         21  52691510 ns/op   8227424 B/op   51 allocs/op
```

---

## Memory

Three different numbers, often confused. All for a 5% error rate:

| n | Analytical | Retained heap | `B/op` churn | `[]bool` |
|---|---|---|---|---|
| 100K | 25,792 B | 29,016 B | 55,264 B | 100,000 B |
| 1M | 257,816 B | 272,544 B | 645,088 B | 1,000,000 B |
| 10M | 2,578,128 B | 2,752,656 B | 8,227,424 B | 10,000,000 B |

| | bits/element at 1M |
|---|---|
| Analytical (exact array lengths) | **2.0625** |
| Retained heap (measured, includes `append` slack) | **2.18** |
| `[]bool` baseline | 8.00 |

**Analytical** follows directly from the constructor
([succincter.go:26-31](../../succincter.go#L26-L31)), where `blockSize=64` and
`superBlockSize=1024`. With `B = ceil(n/64)`:

```
data        = B * 8 bytes                 = 1.0    bits/element
blockRanks  = B * 8 bytes                 = 1.0    bits/element
superBlocks = ceil(B/16) * 8 bytes        = 0.0625 bits/element
                                          ---------------------
                                            2.0625 bits/element
```

**Retained** is `runtime.HeapAlloc` delta across construction with a forced GC
either side. It exceeds analytical by ~6% because
[succincter.go:92-93](../../succincter.go#L92-L93) grows `blockRanks` and
`superBlocks` with `append` and no preallocation, leaving capacity slack.
(The 10K row is unreliable at this size — GC noise exceeded the signal — so it is
omitted.)

**Churn** is `-benchmem` `B/op`, which counts *every* intermediate allocation
during append growth, not the resting footprint. It is a construction cost, not
a memory overhead. Preallocating with `make([]uint64, 0, B)` would remove most
of it.

The claim of "~1.5 bits/element" that appeared throughout the repository
described the pre-`uint64` layout; see the `[]uint32` → `[]uint64` migration in
[CHANGELOG.md](../../CHANGELOG.md). Under that layout the total was
`1 + 0.5 + 0.03125 = 1.53` bits/element.

---

## Example transcript

`go run ./examples/loganalysis`, after the timing fix described above:

```
=== Succincter Log Analysis Example ===
Generating 1000000 log entries (5% errors)...
Building Succincter index...
Index built in 8.0052ms


--- Query Examples ---
1. Errors before position 500000: 25150 (query time: 0s)
2. Position of error #1000: 19322 (query time: 0s)
   Verification: logs[19322].Level = "ERROR"
3. Errors in range [100000, 200000): 5072 (query time: 0s)

4. First 5 errors:
   Error 1 at position 10: Log message 10
   Error 2 at position 46: Log message 46
   Error 3 at position 55: Log message 55
   Error 4 at position 58: Log message 58
   Error 5 at position 76: Log message 76

--- Performance Comparison ---
Rank(500000):
  Naive:      2.185981ms avg (100 iterations)
  Succincter: 2ns avg (1000000 iterations)
  Speedup:    1092990x
  (checksum 25152515000, printed so the timed loops cannot be optimized away)
```

Two things to read carefully here:

- `query time: 0s` is a **clock-resolution artifact**, not a claim of zero time.
  Windows' monotonic clock cannot resolve a single ~3 ns operation.
- `Succincter: 2ns` is the *hot-cache, same-position* figure, and
  `time.Duration` truncates 2.6 ns to `2ns`. The `1092990x` speedup follows from
  it and is therefore a best case, not a typical one.

The error counts vary between runs: the example seeds `math/rand` implicitly, so
the generated data differs each time.

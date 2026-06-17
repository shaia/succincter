# Succincter

[![Test](https://github.com/shaia/succincter/actions/workflows/test.yml/badge.svg)](https://github.com/shaia/succincter/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/shaia/succincter.svg)](https://pkg.go.dev/github.com/shaia/succincter)

A Go library implementing succinct data structures for efficient rank and select queries on boolean arrays.

## Overview

Succincter provides O(1) rank queries and O(log n) select queries on compressed boolean arrays with only ~1.5 bits per element overhead. It uses a generic constructor that accepts any slice type and a predicate function.

## Installation

```bash
go get github.com/shaia/succincter
```

Requires Go 1.21+.

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/shaia/succincter"
)

func main() {
    numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    s := succincter.NewSuccincter(numbers, func(n int) bool {
        return n%2 == 0
    })

    // Rank: count true elements before position 7
    fmt.Println(s.Rank(7)) // 3

    // Select: find position of 3rd true element
    fmt.Println(s.Select(3)) // 5

    // Version
    fmt.Println(succincter.Version) // "0.1.0"
}
```

## API

### Types

#### `RankSelector` interface

```go
type RankSelector interface {
    Rank(pos int) int
    Select(rank int) int
}
```

Both `Succincter` and `RRR` implement this interface, so callers can swap implementations without changing query sites.

### Constructors

#### `NewSuccincter[T any](input []T, predicate func(T) bool) *Succincter`

Creates a Succincter (plain bitvector + rank/select index) from any slice using a predicate to determine 1-bits. O(n) construction, ~1.5 bits/element overhead.

#### `NewRRR[T any](input []T, predicate func(T) bool) *RRR`

Creates an RRR-compressed structure with the same predicate-based API. Achieves `nH₀(B) + o(n)` space using block size b=15 with combinatorial (class, offset) encoding. Rank is O(1) and Select is O(log n) — same complexity as Succincter but with a higher rank constant (~50–200ns) due to combinatorial decoding on each query.

### Methods

The same methods exist on `*Succincter` and `*RRR` with identical signatures and edge-case behavior.

#### `Rank(pos int) int`

Returns the count of 1-bits before position `pos`. O(1) time.

Returns 0 for `pos <= 0` or empty arrays.

#### `Select(rank int) int`

Returns the position of the `rank`-th 1-bit (1-indexed). O(log n) time.

Returns -1 for invalid ranks or empty arrays.

### Choosing Between `Succincter` and `RRR`

- **Use `Succincter`** when density is near 50%, or when rank latency is the priority — no decode step on the hot path.
- **Use `RRR`** when the bitvector is sparse or dense (≪ 50% or ≫ 50% ones) and memory is the bottleneck — space approaches the zero-order entropy `nH₀(B)`, which can be a fraction of Succincter's 1.5n bits.
- Both are immutable after construction and safe for concurrent reads with no synchronization.

### Version

```go
const Version = "0.1.0"

func FullVersion() string  // Returns version with prerelease tag if set
```

## Performance

| Operation    | Time       | Space Overhead       |
|-------------|------------|---------------------|
| Construction | O(n)       | ~1.5 bits/element   |
| Rank         | O(1)       | —                   |
| Select       | O(log n)   | —                   |

### Benchmarks

Measured on Intel Core Ultra 9 (see `go test -bench=.` for your system):

| Dataset Size | Naive Rank | Succincter Rank | Speedup |
|--------------|------------|-----------------|---------|
| 10K          | ~3µs       | ~13ns           | ~220x   |
| 100K         | ~30µs      | ~13ns           | ~2,300x |

Speedup scales linearly with data size (naive is O(n), Succincter is O(1)).

Run benchmarks:

```bash
go test -bench=. -benchmem
```

## Examples

See [examples/loganalysis](examples/loganalysis) for a complete example demonstrating:

- Building an index on log entries
- Counting errors before a position
- Finding the Nth error
- Pagination of filtered results

Run the example:

```bash
go run ./examples/loganalysis
```

## Documentation

**RRR encoding (combinatorial number system):** each 15-bit block is stored as a `(class, offset)` pair, where `class` is the popcount and `offset` is the block's index among `C(15, class)` patterns. The offset shrinks from 15 bits at class = 7/8 down to 9 bits at class = 3 and 4 bits at class = 1 — that compression vs. raw bits is where the `nH₀(B)` space bound comes from.

## Thread Safety

Safe for concurrent reads after construction. No synchronization needed for read-only access.

## License

MIT

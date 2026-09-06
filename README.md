# Succincter

[![Test](https://github.com/shaia/succincter/actions/workflows/test.yml/badge.svg)](https://github.com/shaia/succincter/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/shaia/succincter.svg)](https://pkg.go.dev/github.com/shaia/succincter)

A Go library implementing succinct data structures for efficient rank and select queries on boolean arrays.

## Overview

Succincter provides O(1) rank queries and O(log n) select queries on compressed boolean arrays, costing ~2.06 bits per element — bit vector and rank index combined, under a third of what a Go `[]bool` spends on the booleans alone. It uses a generic constructor that accepts any slice type and a predicate function.

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

`Succincter` implements this interface.

### Constructor

#### `NewSuccincter[T any](input []T, predicate func(T) bool) *Succincter`

Creates a Succincter from any slice using a predicate to determine 1-bits. O(n) construction.

### Methods

#### `Rank(pos int) int`

Returns the count of 1-bits before position `pos`. O(1) time.

Returns 0 for `pos <= 0` or empty arrays.

#### `Select(rank int) int`

Returns the position of the `rank`-th 1-bit (1-indexed). O(log n) time.

Returns -1 for invalid ranks or empty arrays.

### Version

```go
const Version = "0.1.0"

func FullVersion() string  // Returns version with prerelease tag if set
```

## Performance

| Operation    | Time       | Space               |
|--------------|------------|---------------------|
| Construction | O(n)       | ~2.06 bits/element  |
| Rank         | O(1)       | —                   |
| Select       | O(log n)   | —                   |

Space breaks down as 1 bit/element for the packed bit vector, 1 bit/element for
the per-word rank index, and 1/16 bit/element for the superblock index. Measured
heap is ~2.18 bits/element, since both index arrays are grown with `append` and
no preallocation.

### Benchmarks

Measured on a 13th Gen Intel Core i9-13980HX, Go 1.25.5, querying **scattered**
positions (see `go test -bench=.` for your system):

| Dataset Size | Naive Rank | Succincter Rank | Speedup     |
|--------------|------------|-----------------|-------------|
| 10K          | 3.60µs     | 2.7ns           | ~1,300x     |
| 100K         | 161µs      | 14.0ns          | ~11,500x    |
| 1M           | 1.88ms     | 16.6ns          | ~113,000x   |
| 10M          | 20.6ms     | 18.0ns          | ~1,145,000x |

Naive is O(n) with a constant that worsens as the scan falls out of cache;
Succincter's rank is O(1).

Note that repeatedly ranking the *same* position measures ~2.6ns, because its
cache lines never leave L1. That best case is not what a real workload sees — the
table above uses scattered positions. Select is more cache-sensitive still:
~17-30ns for a repeated rank against ~228-533ns for scattered ranks.

Full measurements and methodology: [docs/bench/results.md](docs/bench/results.md).

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

- [Finding Errors in Log Streams](https://slow-is-smooth.io/blog/finding-errors-in-log-streams/) - Real-world usage tutorial
- [Benchmark results](docs/bench/results.md) - Raw measurements and methodology

## Thread Safety

Safe for concurrent reads after construction. No synchronization needed for read-only access.

## License

MIT

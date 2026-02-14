# Succincter Examples

This directory contains examples demonstrating how to use Succincter for various real-world and educational use cases.

## Quick Start

Run any example:

```bash
go run ./examples/<name>
```

## Examples

### Simple Examples

#### [basic](basic/)

**Basic Boolean Array Operations**

The simplest possible example showing core Succincter functionality with a boolean slice.

```bash
go run ./examples/basic
```

Demonstrates:
- Creating a Succincter from a boolean slice
- Rank queries: counting true values before a position
- Select queries: finding the position of the Nth true value
- Rank-select duality: `Rank(Select(i)) == i-1`

Output:
```
=== Basic Succincter Example ===
Input: [true false true true false false true false true true]

Rank queries (count of true before position):
  Rank(0) = 0
  Rank(5) = 3
  Rank(10) = 6

Select queries (position of Nth true):
  Select(1) = 0 (flags[0] = true)
  Select(3) = 3 (flags[3] = true)
```

---

#### [primes](primes/)

**Prime Number Indexing**

Uses Succincter to index prime numbers for efficient counting and lookup.

```bash
go run ./examples/primes
```

Demonstrates:
- Custom predicate function (`isPrime`)
- Counting primes in ranges
- Finding the Nth prime by rank
- Range-based counting with `Rank(end) - Rank(start)`

Output:
```
=== Prime Numbers Example ===
Prime numbers in [0, 100): 25

First 10 primes:
  Prime #1 = 2
  Prime #2 = 3
  ...

Primes in ranges:
  Primes in [0, 10): 4
  Primes in [50, 100): 10

The 25th prime is: 97
```

---

### Real-World Examples

#### [loganalysis](loganalysis/)

**Log Stream Error Analysis**

Processes log entries to efficiently find and count errors. The canonical example from the blog post.

```bash
go run ./examples/loganalysis
```

Demonstrates:
- Indexing structured data (log entries) by a field predicate
- Counting errors before a timestamp/position
- Finding the Nth error for pagination
- Performance comparison vs naive O(n) scanning (~50,000x speedup)

Use cases:
- Log monitoring dashboards
- Error pagination in admin interfaces
- Time-range error counting

---

#### [dna](dna/)

**DNA Sequence Analysis**

Bioinformatics example analyzing a 1 million base pair DNA sequence.

```bash
go run ./examples/dna
```

Demonstrates:
- Multiple indices on the same data (one per nucleotide: A, C, G, T)
- Sequence composition analysis (GC content)
- Finding specific nucleotide positions
- Region-based analysis (gene regions)
- CpG island detection using sliding window queries

Output:
```
=== DNA Sequence Analysis Example ===

Generating 1000000 base pair sequence...
Building nucleotide indices...
Indices built in 19ms

--- Sequence Composition ---
Adenine (A):  250256 (25.03%)
Cytosine (C): 250823 (25.08%)
Guanine (G):  249720 (24.97%)
Thymine (T):  249201 (24.92%)
GC Content:   50.05%

--- Position Queries ---
Position of 1000th Adenine:  3907
Position of 5000th Guanine:  20018
```

Use cases:
- Genome analysis tools
- Sequence alignment preprocessing
- Variant calling pipelines

---

#### [timeseries](timeseries/)

**IoT Sensor Anomaly Detection**

Analyzes 24 hours of simulated sensor data (86,400 readings) for temperature anomalies.

```bash
go run ./examples/timeseries
```

Demonstrates:
- Multiple anomaly indices (cold, hot, any)
- Anomaly summary statistics
- Finding first N anomalies
- Hourly distribution analysis
- Pagination through anomalies
- Time-range queries

Output:
```
=== Time Series Anomaly Detection Example ===

Generating 86400 sensor readings (24 hours @ 1/sec)...
Building anomaly indices...
Indices built in 1.5ms

--- Anomaly Summary ---
Cold anomalies (<20°C):  828 (0.96%)
Hot anomalies (>80°C):   856 (0.99%)
Total anomalies:         1684 (1.95%)

--- Hourly Anomaly Distribution ---
  00:00-04:00: ███████████████████████████ 274
  04:00-08:00: ████████████████████████████ 282
  ...
```

Use cases:
- IoT monitoring dashboards
- Industrial sensor analysis
- Alert system backends
- Time-series databases

---

#### [useractivity](useractivity/)

**User Activity Tracking and Leaderboards**

Manages 100,000 users with multiple segment indices for gaming/social platform scenarios.

```bash
go run ./examples/useractivity
```

Demonstrates:
- Multiple segment indices (online, premium, high-scorers, recently active)
- Segment statistics and percentages
- Leaderboard queries (find Nth online user)
- Pagination through user segments
- Distribution analysis across ID ranges (useful for sharding)
- Combining multiple predicates

Output:
```
=== User Activity Tracking Example ===

Generating 100000 users...
Building user indices...
Indices built in 4.5ms

--- User Segments ---
Online users:         15042 (15.0%)
Premium users:         9771 (9.8%)
High scorers (1000+):  49850 (49.9%)

--- Leaderboard Queries ---
1st online user:    BraveLion8 (ID: 5)
100th online user:  SwiftBear911 (ID: 651)
1000th online user: MightyTiger942 (ID: 7181)
```

Use cases:
- Gaming leaderboards
- Social platform user management
- Premium user dashboards
- Sharding analysis for distributed systems

---

## Performance Notes

All examples demonstrate O(1) rank and O(log n) select queries, compared to O(n) for naive filtering approaches:

| Dataset Size | Naive Filter | Succincter Rank | Speedup |
|--------------|--------------|-----------------|---------|
| 100K         | ~10µs        | ~2ns            | 5,000x  |
| 1M           | ~100µs       | ~2ns            | 50,000x |
| 10M          | ~1ms         | ~2ns            | 500,000x|

Construction is O(n) and happens once. Queries are nearly instant regardless of data size.

## Creating Your Own

To create a Succincter for your data:

```go
// 1. Define your data type
type MyRecord struct {
    ID     int
    Status string
    Value  float64
}

// 2. Create your dataset
records := loadRecords()

// 3. Build index with a predicate
activeIndex := succincter.NewSuccincter(records, func(r MyRecord) bool {
    return r.Status == "active"
})

// 4. Query efficiently
countBefore := activeIndex.Rank(position)  // O(1)
nthActive := activeIndex.Select(n)          // O(log n)
countInRange := activeIndex.Rank(end) - activeIndex.Rank(start)  // O(1)
```

## Further Reading

- [Finding Errors in Log Streams](../docs/posts/finding-errors-in-log-streams.md) - Detailed tutorial
- [Combinatorial Encoding](../docs/posts/combinatorial-encoding-for-compression.md) - Compression internals

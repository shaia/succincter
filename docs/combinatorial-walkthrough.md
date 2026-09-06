# Combinatorial Encoding: A Tiny Worked Example

A minimal walkthrough of how `CombEncode` and `CombDecode` map a bit pattern to its index among all patterns with the same popcount, and back. This page uses a 4-bit example so the whole derivation fits on one screen. For the canonical 15-bit case as it appears in production code, see [combinatorial-encoding-for-compression.md](posts/combinatorial-encoding-for-compression.md).

## The Problem

Given a `b`-bit block β with `c` 1-bits (its **class**), there are exactly `C(b, c)` possible patterns. Assigning each pattern a unique index in `[0, C(b, c) - 1]` (its **offset**) lets us replace the `b` raw bits with `(class, offset)` — and the offset takes only `⌈log₂ C(b, c)⌉` bits, which is smaller than `b` whenever the pattern is sparse or dense.

This file walks the algorithm on β = `0100` (b = 4, class = 1). The answer is o = 2.

## Encode: 0100 → 2

`CombEncode` scans bit positions from MSB (position 3) down to LSB (position 0). At each 1-bit it adds `C(p, onesLeft)` to the running offset and decrements `onesLeft`. That coefficient counts the patterns lexicographically smaller than β that have a 0 at this position.

| bitPos `p` | bit | onesLeft `r` | C(p, r) added | offset after |
|-----------:|----:|-------------:|--------------:|-------------:|
| 3 | 0 | 1 | — (skip) | 0 |
| 2 | 1 | 1 | C(2, 1) = **2** | 2 |
| 1 | 0 | 0 | — (done, onesLeft = 0) | 2 |
| 0 | 0 | 0 | — (done) | 2 |

Result: **o = 2**.

Sanity check by enumerating all class-1 patterns of width 4 in increasing offset order:

| offset | pattern |
|------:|--------:|
| 0 | `0001` |
| 1 | `0010` |
| 2 | `0100` |
| 3 | `1000` |

Offset 2 ↔ `0100`. ✓

## Decode: (class = 1, offset = 2) → 0100

`CombDecode` reverses the process. It scans positions MSB-to-LSB and greedily places a 1-bit at position `p` whenever `offset ≥ C(p, onesLeft)`, then subtracts the coefficient.

| bitPos `p` | C(p, onesLeft) | offset ≥ coeff? | action | offset after |
|-----------:|---------------:|:----------------|:-------|-------------:|
| 3 | C(3, 1) = 3 | 2 ≥ 3? **no** | skip | 2 |
| 2 | C(2, 1) = 2 | 2 ≥ 2? **yes** | set bit 2 | 0 |
| 1 | — | onesLeft = 0 | done | 0 |
| 0 | — | onesLeft = 0 | done | 0 |

Result: bit 2 set, all others 0 → **`0100`**. ✓

## Why C(p, r) Counts Smaller Patterns

When the encoder sees a 1-bit at position `p` with `r` ones still to place (counting the current one), every pattern that has a 0 at position `p` and the same `r` ones spread across positions `0..p-1` comes before β in offset order. There are exactly `C(p, r)` such patterns — choose `r` positions for the remaining ones out of the `p` lower positions.

The full lexicographic-ordering argument (treating bit positions as a `c`-subset of `{0, …, b−1}`) is laid out in [Why This Works](posts/combinatorial-encoding-for-compression.md#why-this-works).

## Edge Cases

`CombEncode` and `CombDecode` short-circuit when the class is extremal:

- **class = 0** → only one possible pattern (all zeros). Offset is 0, decode returns `0`. See [internal/combinatorial.go](../internal/combinatorial.go) lines 32–34 and 60–62.
- **class = b** → only one possible pattern (all ones). Offset is 0, decode returns `(1 << b) - 1`. See [internal/combinatorial.go](../internal/combinatorial.go) lines 32–34 and 63–65.

`OffsetBits(class)` returns 0 for both, so no bits at all are written for these classes — the class itself (stored separately in 4 bits) is the entire encoding.

## See Also

- [combinatorial-encoding-for-compression.md](posts/combinatorial-encoding-for-compression.md) — full 15-bit canonical example (β = `0b010110000000000` → o = 351), space analysis, performance numbers.
- [zero-order-compression.md](zero-order-compression.md) — how RRR uses combinatorial encoding to achieve `nH₀(B) + o(n)` space.
- [internal/combinatorial.go](../internal/combinatorial.go) — Go implementation of `CombEncode`, `CombDecode`, `OffsetBits`.
- [rrr.go](../rrr.go) — how `NewRRR` calls `CombEncode` at construction and `Rank`/`Select` call `CombDecode` at query time.

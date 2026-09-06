# Combinatorial Encoding: A Tiny Worked Example

A minimal walkthrough of how `CombEncode` and `CombDecode` map a bit pattern to its index among all patterns with the same popcount, and back. This page uses a 4-bit example so the whole derivation fits on one screen. For the canonical 15-bit case as it appears in production code, see [internal/combinatorial.go](../internal/combinatorial.go).

## The Problem

Given a `b`-bit block β with `c` 1-bits (its **class**), there are exactly `C(b, c)` possible patterns. Assigning each pattern a unique index in `[0, C(b, c) - 1]` (its **offset**) lets us replace the `b` raw bits with `(class, offset)` — and the offset takes only `⌈log₂ C(b, c)⌉` bits, which is smaller than `b` whenever the pattern is sparse or dense.

This file walks the algorithm on β = `0100` (b = 4, class = 1). The answer is o = 2.

## Encode: 0100 → 2

`CombEncode(block, class)` takes no width parameter — it always scans from bit 14 down to bit 0, since the coefficient `C(p, r)` depends only on the position and the ones remaining, never on `b`. The high bits of a 4-bit pattern are zero, so the trace below starts at position 3. (`CombDecode(class, offset, b)` *does* take `b`; only the decoder needs to know where to start.)

At each 1-bit the encoder adds `C(p, onesLeft)` to the running offset and decrements `onesLeft`. That coefficient counts the patterns lexicographically smaller than β that have a 0 at this position.

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

The full argument treats the set bit positions as a `c`-subset of `{0, …, b−1}` and ranks those subsets lexicographically.

## Edge Cases

An extremal class admits only one pattern, so its offset carries no information. The short-circuits for this are written against the **production block size of 15**, not against a general `b`:

- **class = 0** → all zeros. `CombEncode` returns 0; `CombDecode` returns `0`.
- **class = 15** → all ones. `CombEncode` returns 0; `CombDecode` returns `(1 << b) - 1` via its `class == b` test.

`OffsetBits(class)` likewise special-cases only 0 and 15, and otherwise sizes the offset from `C(15, class)`. So at b = 15 no offset bits are written for either extreme, and the class alone (packed separately in 4 bits) is the entire encoding.

**This does not generalize to the 4-bit example above.** `CombEncode`'s guard tests `class == 15` literally, so `CombEncode(0b1111, 4)` does not short-circuit — it falls through the loop and returns 0 only because the coefficients it reads are unpopulated (zero) table entries. And `OffsetBits(4)` returns 11, sizing an offset for `C(15, 4)` rather than the single 4-bit pattern. The toy width is a teaching device for the ranking itself; the extremal-class handling is 15-specific.

## See Also

- [README.md](../README.md#documentation) — how RRR uses combinatorial encoding to reach the `nH₀(B) + o(n)` space bound.
- [internal/combinatorial.go](../internal/combinatorial.go) — Go implementation of `CombEncode`, `CombDecode`, `OffsetBits`.
- [rrr.go](../rrr.go) — how `NewRRR` calls `CombEncode` at construction and `Rank`/`Select` call `CombDecode` at query time.

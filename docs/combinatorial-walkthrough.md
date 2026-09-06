# Combinatorial Encoding: A Tiny Worked Example

RRR replaces each `b`-bit block with a `(class, offset)` pair: the **class** is the block's popcount, the **offset** is its index among all blocks sharing that popcount. This page derives both directions by hand on a 4-bit block, then repeats the derivation at the production width of 15 and shows what it costs.

Implementation: [internal/combinatorial.go](../internal/combinatorial.go).

## The Problem

Given a `b`-bit block β with `c` 1-bits, exactly `C(b, c)` patterns share that class. Number them `0 … C(b, c) - 1` and β can be stored as `(c, offset)` instead of `b` raw bits — the offset needs only `⌈log₂ C(b, c)⌉` bits, far fewer than `b` when the block is mostly 0s or mostly 1s.

The numbering is the **combinatorial number system**, which yields an invariant worth holding onto:

> Within a class, offset order *is* numeric order. The numerically smallest block of class `c` has offset 0; the largest has offset `C(b, c) - 1`.

Worked example: β = `0110`, b = 4, class = 2. The answer is offset 2.

## Encode: `0110` → 2

`CombEncode(block, class)` takes no width parameter — it always scans from bit 14 down to bit 0, because the coefficient `C(p, r)` depends only on the position `p` and the ones remaining `r`, never on `b`. A 4-bit pattern is zero above bit 3, so the trace starts there. (`CombDecode(class, offset, b)` *does* take `b`; only the decoder needs to know where to begin.)

At each 1-bit the encoder adds `C(p, onesLeft)` to the running offset, then decrements `onesLeft`. The loop stops as soon as `onesLeft` hits 0:

| bitPos `p` | bit | onesLeft `r` | C(p, r) added | offset after |
|-----------:|----:|-------------:|--------------:|-------------:|
| 3 | 0 | 2 | — (skip) | 0 |
| 2 | 1 | 2 | C(2, 2) = **1** | 1 |
| 1 | 1 | 1 | C(1, 1) = **1** | 2 |
| 0 | — | 0 | — (loop ends) | 2 |

Result: **offset = 2**.

Sanity check — all `C(4, 2) = 6` class-2 patterns, in offset order:

| offset | pattern |
|------:|--------:|
| 0 | `0011` |
| 1 | `0101` |
| 2 | `0110` |
| 3 | `1001` |
| 4 | `1010` |
| 5 | `1100` |

`0110` sits at offset 2. ✓ Note the patterns ascend numerically (3, 5, 6, 9, 10, 12) — that is the invariant above.

## Decode: (class 2, offset 2) → `0110`

`CombDecode` reverses the walk. It scans MSB-to-LSB and greedily places a 1-bit wherever `offset ≥ C(p, onesLeft)`, subtracting the coefficient it consumes:

| bitPos `p` | C(p, onesLeft) | offset ≥ coeff? | action | offset after |
|-----------:|---------------:|:----------------|:-------|-------------:|
| 3 | C(3, 2) = 3 | 2 ≥ 3? **no** | skip | 2 |
| 2 | C(2, 2) = 1 | 2 ≥ 1? **yes** | set bit 2 | 1 |
| 1 | C(1, 1) = 1 | 1 ≥ 1? **yes** | set bit 1 | 0 |
| 0 | — | onesLeft = 0 | loop ends | 0 |

Bits 2 and 1 set → **`0110`**. ✓

Greedy works because the coefficients strictly decrease as `p` falls: taking the largest affordable coefficient at each step is forced, exactly as in positional notation.

## Why C(p, r) Counts Smaller Patterns

When the encoder meets a 1-bit at position `p` with `r` ones still to place (counting this one), consider every pattern that agrees with β above `p` but has a **0** at position `p`. Each must fit all `r` remaining ones into the `p` positions below, and every one of them is numerically smaller than β. There are exactly `C(p, r)` ways to choose those positions — so the coefficient is precisely the count of class-mates skipped over by setting this bit.

Summing over the set bits ranks β among the `c`-subsets of `{0, …, b−1}`.

## At Production Scale: b = 15

The same derivation on a real block. β = `000000101000010`, class 3, with bits set at positions 8, 6 and 1:

| bitPos `p` | onesLeft `r` | C(p, r) added | offset after |
|-----------:|-------------:|--------------:|-------------:|
| 8 | 3 | C(8, 3) = **56** | 56 |
| 6 | 2 | C(6, 2) = **15** | 71 |
| 1 | 1 | C(1, 1) = **1** | 72 |

Offset 72, which `OffsetBits(3) = 9` bits hold comfortably. So 15 raw bits become a 4-bit class plus a 9-bit offset — **13 bits**.

## What It Buys, and When It Doesn't

`OffsetBits(c)` is `⌈log₂ C(15, c)⌉`. Adding the 4-bit class gives the true per-block cost:

| class `c` | C(15, c) | offset bits | + 4-bit class | vs. 15 raw |
|---------:|---------:|------------:|--------------:|-----------:|
| 0 | 1 | 0 | 4 | −11 |
| 1 | 15 | 4 | 8 | −7 |
| 2 | 105 | 7 | 11 | −4 |
| 3 | 455 | 9 | 13 | −2 |
| 4 | 1365 | 11 | 15 | 0 |
| 5 | 3003 | 12 | 16 | **+1** |
| 6 | 5005 | 13 | 17 | **+2** |
| 7 | 6435 | 13 | 17 | **+2** |
| 8 | 6435 | 13 | 17 | **+2** |
| 9 | 5005 | 13 | 17 | **+2** |
| 10 | 3003 | 12 | 16 | **+1** |
| 11 | 1365 | 11 | 15 | 0 |
| 12 | 455 | 9 | 13 | −2 |
| 13 | 105 | 7 | 11 | −4 |
| 14 | 15 | 4 | 8 | −7 |
| 15 | 1 | 0 | 4 | −11 |

The table is symmetric, and mid-range classes **cost more than the raw bits they replace**. A balanced block (class 5–10) expands by 1–2 bits; only skewed blocks pay for themselves, and the extremes save 11.

That is the `nH₀(B)` bound made concrete: the win comes from the *distribution* of classes, not from any single block. On uniformly random bits the classes cluster at 7 and 8 and RRR is a net loss. On sparse or dense data the mass sits at the ends of the table and it wins big — which is why the README recommends RRR only when density is far from 50%.

## Where the Pair Lands in Memory

[rrr.go](../rrr.go) stores the two halves separately, which is what makes the offsets reachable at all:

- **Classes** — 4 bits per block (0–15 fits exactly), packed into a `[]uint64`. Fixed width, so block `i`'s class is a direct index.
- **Offsets** — variable width, bit-packed back to back with no padding. This *cannot* be indexed directly: block `i`'s offset begins wherever the previous `i` offsets happened to end.
- **Superblocks** — one `uint64` per 16 blocks, packing cumulative rank in the high 32 bits and the cumulative *offset-bit position* in the low 32.

The superblock's low half is what rescues the offsets: `Rank` jumps to the nearest superblock, then walks at most 16 classes, summing `OffsetBits` until it arrives at the target block's offset. A bounded scan, hence O(1).

## Edge Cases

An extremal class admits only one pattern, so its offset carries no information. The short-circuits are written against the **production block size of 15**, not a general `b`:

- **class = 0** → all zeros. `CombEncode` returns 0; `CombDecode` returns `0`.
- **class = 15** → all ones. `CombEncode` returns 0; `CombDecode` returns `(1 << b) - 1` via its `class == b` test.

`OffsetBits` special-cases only 0 and 15, sizing every other class from `C(15, class)`. At b = 15 that means no offset bits at all for either extreme — the 4-bit class is the entire encoding, the −11 rows in the table above.

**This does not generalize to the 4-bit example.** `CombEncode`'s guard tests `class == 15` literally, so `CombEncode(0b1111, 4)` does not short-circuit — it falls through the loop and returns 0 only because the coefficients it reads are unpopulated (zero) table entries. And `OffsetBits(4)` returns 11, sizing an offset for `C(15, 4)` rather than for the single 4-bit pattern. The toy width teaches the ranking; the extremal-class handling is 15-specific.

## Verify It Yourself

The `0110` example and the offset widths above are both asserted in the test suite:

```sh
go test ./internal/ -run 'TestWorkedExamples|TestOffsetBits|TestCombEncodeDecode' -v
```

`TestCombEncodeDecode` round-trips all 32768 blocks at b = 15.

## See Also

- [README.md](../README.md#documentation) — how RRR uses combinatorial encoding to reach the `nH₀(B) + o(n)` space bound.
- [internal/combinatorial.go](../internal/combinatorial.go) — Go implementation of `CombEncode`, `CombDecode`, `OffsetBits`.
- [rrr.go](../rrr.go) — how `NewRRR` calls `CombEncode` at construction and `Rank`/`Select` call `CombDecode` at query time.

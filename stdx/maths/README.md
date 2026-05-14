# stdx/maths

Import path: `github.com/InsideGallery/core/stdx/maths`

## Overview

`stdx/maths` provides Cantor pairing, rounding, clamping, weighted random selection, and 128-bit integer split
helpers.

## Main APIs

- `CantorPair` and `CantorUnpair` map two `uint64` values to and from one `uint64` value.
- `RoundWithPrecision`, `Round`, `Clamp`, and `ApproximatelyEqual` provide small float helpers.
- `IntStringToBigInt`, `BigIntToHighAndLow`, and `HighAndLowToBigInt` convert decimal strings and high/low
  `uint64` pairs.
- `WeightIndex` selects a key from a `map[interface{}]uint64` using the weights as probabilities.
- `RandomDigitString` returns a random string of the requested length.
- `DefaultPrecision` is used by `RoundWithPrecision` when the provided precision is approximately zero.

## Usage

```go
pair := maths.CantorPair(7, 11)
left, right := maths.CantorUnpair(pair)
clamped := maths.Clamp(120, 0, 100)

_ = left
_ = right
_ = clamped
```

## Notes

`WeightIndex` returns nil for an empty map. Despite its name, `RandomDigitString` fills the string with random
byte values from `math/rand`, not only ASCII digits and not cryptographic randomness.

# stdx/slices

Import path: `github.com/InsideGallery/core/stdx/slices`

## Overview

`stdx/slices` contains generic slice batching and string shingling helpers.

## Main APIs

- `BatchSlice[K any](size int, result []K)` returns a channel that yields consecutive batches.
- `Shingle(text string, k int)` returns a `set.GenericDataSet[string]` containing unique `k`-length shingles.

## Usage

```go
for batch := range coreslices.BatchSlice(2, []int{1, 2, 3, 4, 5}) {
	_ = batch
}

shingles := coreslices.Shingle("testing", 3)
_ = shingles
```

## Notes

`BatchSlice` treats non-positive sizes as `1`, starts a goroutine, and closes the channel after all batches are
sent. Returned batches are slices of the original input. `Shingle` returns an empty set for empty text or
non-positive `k`; when the text is shorter than `k`, it returns a set containing the whole text. Shingling uses
byte indexes.

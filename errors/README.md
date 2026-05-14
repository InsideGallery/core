# errors

Import path: `github.com/InsideGallery/core/errors`

## Overview

`errors` provides compatibility helpers for building and combining errors while preserving `errors.Is` and
`errors.As` checks across a cause/effect pair.

## Main APIs

- `New(text string)` delegates to the standard library error constructor.
- `Wrap(cause, effect error)` combines two distinct errors into `MultipleError`.
- `Wrapf(err error, format string, args ...any)` adds formatted context to an existing error.
- `Combine(errs ...error)` folds several errors into one.
- `MultipleError` stores `Cause` and `Effect`, joins messages with `": "`, unwraps to `Cause`, and checks both
  sides in `Is` and `As`.
- `BoundaryError` and `WrapBoundary(kind, operation, err)` wrap infrastructure or SDK errors at package
  boundaries.

## Usage

```go
var closeErr error
var writeErr error

err := coreerrors.Combine(writeErr, closeErr)
if err != nil && errors.Is(err, closeErr) {
	return err
}
```

## Notes

Import this package with an alias such as `coreerrors` when the standard library `errors` package is also used.
`Wrap` returns nil when both inputs are nil, returns the non-nil input when only one exists, and deduplicates two
errors with the same message by returning the cause.

package errors //nolint:revive

import "strings"

const boundaryErrorParts = 3

// BoundaryError wraps an infrastructure or SDK error at a core-owned package boundary.
type BoundaryError struct {
	Kind      string
	Operation string
	Err       error
}

// Error returns the boundary error message.
func (e BoundaryError) Error() string {
	parts := make([]string, 0, boundaryErrorParts)

	if e.Kind != "" {
		parts = append(parts, e.Kind)
	}

	if e.Operation != "" {
		parts = append(parts, e.Operation)
	}

	if e.Err != nil {
		parts = append(parts, e.Err.Error())
	}

	return strings.Join(parts, ": ")
}

// Unwrap returns the wrapped SDK or infrastructure error.
func (e BoundaryError) Unwrap() error {
	return e.Err
}

// WrapBoundary returns nil for nil errors or a BoundaryError for non-nil errors.
func WrapBoundary(kind string, operation string, err error) error {
	if err == nil {
		return nil
	}

	return BoundaryError{
		Kind:      kind,
		Operation: operation,
		Err:       err,
	}
}

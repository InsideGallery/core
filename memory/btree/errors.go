package btree

import "errors"

// All kind of errors
var (
	ErrInvalidOrder = errors.New("invalid order, should be at least 3")
)

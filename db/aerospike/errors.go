package aerospike

import "errors"

// All kind of errors for aerospike
var (
	ErrConnectionIsNotSet = errors.New("connection is not set")
)

package mongodb

import "errors"

// All kind of errors for mongo
var (
	ErrConnectionIsNotSet = errors.New("connection is not set")
)

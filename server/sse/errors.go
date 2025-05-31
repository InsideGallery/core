package sse

import "errors"

// All kind of errors
var (
	ErrInvalidUserID              = errors.New("invalid user id")
	ErrNotFoundConnectedUser      = errors.New("not found connected user")
	ErrResponseWriterIsNotFlusher = errors.New("response does not support flush")
)

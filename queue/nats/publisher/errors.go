package publisher

import (
	"errors"
	"fmt"
)

// All kind of errors for nats
var (
	ErrWrongCountOfArguments = errors.New("error wrong count of arguments")
	// ErrHandlerFail is the sentinel error for handlerError wrapped errors
	ErrHandlerFail = errors.New("handler error")
)

// handlerError is an error wrapper that returns true when compared by errors.Is
// both to the ErrHandlerFail sentinel error and the wrapped error (except for
// nil-ness checks). It's meant for avoiding the fmt.Errorf("%s: %w", err,
// sentinelErr) idiom.
type handlerError struct {
	E error
}

func (e *handlerError) Error() string {
	if e == nil || e.E == nil {
		return "<nil>"
	}

	return fmt.Sprintf("handler error: %s:", e.E)
}

func (e *handlerError) Is(target error) bool {
	if e == nil || e.E == nil {
		return target == nil
	}

	if errors.Is(target, ErrHandlerFail) {
		return true
	}

	return errors.Is(e.E, target)
}

func handlerErr(msg string) error {
	return &handlerError{E: errors.New(msg)}
}

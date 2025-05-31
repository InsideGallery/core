package errors

import (
	nativeErrors "errors"
	"fmt"
	"strings"
)

// MultipleError type for wrap error around error
type MultipleError struct {
	Cause  error
	Effect error
}

func (s MultipleError) Unwrap() error {
	return s.Cause
}

// As finds the first error in err's chain that matches target, and if so, sets
// target to that error value and returns true. Otherwise, it returns false.
func (s MultipleError) As(target interface{}) bool {
	if nativeErrors.As(s.Cause, &target) {
		return true
	}

	return nativeErrors.As(s.Effect, &target)
}

// Is reports whether any error in err's chain matches target.
func (s MultipleError) Is(err error) bool {
	if nativeErrors.Is(s.Cause, err) {
		return true
	}

	return nativeErrors.Is(s.Effect, err)
}

// Combine receive multiple errors and return one
func Combine(errs ...error) (err error) {
	if len(errs) == 0 {
		return nil
	}

	if len(errs) > 1 {
		for _, e := range errs {
			err = Wrap(err, e)
		}
	} else if len(errs) > 0 {
		err = errs[0]
	}

	return err
}

// Error return string based on error
func (s MultipleError) Error() string {
	return strings.Join([]string{s.Cause.Error(), s.Effect.Error()}, ": ")
}

// New return new error
func New(text string) error {
	return nativeErrors.New(text)
}

// Wrap wrap error with error
func Wrap(cause error, effect error) error {
	if cause == nil && effect == nil {
		return nil
	}

	if effect == nil {
		return cause
	}

	if cause == nil {
		return effect
	}

	if cause.Error() == effect.Error() { // case when context of both errors are same
		return cause
	}

	return MultipleError{
		Cause:  cause,
		Effect: effect,
	}
}

// Wrapf wrap by format
func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return &MultipleError{
		Cause:  err,
		Effect: fmt.Errorf(format, args...),
	}
}

package testutils

import (
	"errors"
	"fmt"
	"testing"

	werrors "github.com/pkg/errors"
)

func TestEqual(t *testing.T) {
	Equal(t, "test msg", "test msg")
	NotEqual(t, "test msg2", "test msg")
	Equal(t, 0.654, 0.654)
	NotEqual(t, 0.653, 0.654)
	Equal(t, nil, nil)
	NotEqual(t, nil, errors.New("err"))
	Equal(t, errors.New("err"), errors.New("err"))
	NotEqual(t, errors.New("err2"), errors.New("err"))

	err := errors.New("test error")
	err2 := errors.New("test error")

	Equal(t, werrors.Wrap(err, "additional data"), err)
	NotEqual(t, werrors.Wrap(err2, "additional data"), err)
	Equal(t, fmt.Errorf("test: %w", err), err)
	NotEqual(t, fmt.Errorf("test: %w", err2), err)
	Equal(t, fmt.Errorf("test2: %w", fmt.Errorf("test: %w", err)), err)
	Equal(t, fmt.Errorf("test2: %w", fmt.Errorf("test: %w", err)), fmt.Errorf("test2: %w", fmt.Errorf("test: %w", err)))
}

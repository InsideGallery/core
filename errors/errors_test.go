package errors

import (
	nativeErrors "errors"
	"testing"

	"github.com/pkg/errors"

	"github.com/InsideGallery/core/testutils"
)

func TestErrors(t *testing.T) {
	testcases := []struct {
		name   string
		result error
		input  []error
	}{
		{
			name:  "should return nil for nil error",
			input: nil,
		},
		{
			name:  "should return nil for multiple nil errors",
			input: []error{nil, nil},
		},
		{
			name: "should return nil if not errors",
		},
		{
			name:   "should return single error of one error",
			result: errors.New("test string error"),
			input:  []error{errors.New("test string error")},
		},
		{
			name:   "should return error of multiple errors",
			result: errors.New("test string error: test string error2: test string error3"),
			input:  []error{errors.New("test string error"), errors.New("test string error2"), errors.New("test string error3")},
		},
	}
	for _, test := range testcases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			testutils.Equal(t, Combine(test.input...), test.result)
		})
	}
}

func TestMultipleError(t *testing.T) {
	err := New("test error 1")
	err2 := New("test error 2")
	err3 := New("test error 3")
	err4 := New("test error 4")
	err5 := New("test error 4")

	werr := Wrap(err, err2)
	werr = Wrap(werr, err3)
	werr2 := Wrap(err, err2)
	werr2 = Wrap(werr2, err3)
	werr3 := Wrap(err2, err3)
	werr4 := Wrap(err4, err5)
	testutils.Equal(t, werr.Error(), "test error 1: test error 2: test error 3")
	testutils.Equal(t, nativeErrors.Unwrap(werr).Error(), "test error 1: test error 2")
	testutils.Equal(t, nativeErrors.Is(werr, werr2), true)
	testutils.Equal(t, nativeErrors.Is(werr, werr3), false)
	testutils.Equal(t, nativeErrors.Is(werr3, werr4), false)

	testutils.Equal(t, nativeErrors.Is(werr, err), true)
	testutils.Equal(t, nativeErrors.Is(werr, err2), true)
	testutils.Equal(t, nativeErrors.Is(werr, err3), true)
	testutils.Equal(t, nativeErrors.Is(werr, err4), false)
	werr = nativeErrors.Unwrap(nativeErrors.Unwrap(werr))
	testutils.Equal(t, nativeErrors.Is(werr, err), true)
	testutils.Equal(t, nativeErrors.Is(werr, err2), false)
}

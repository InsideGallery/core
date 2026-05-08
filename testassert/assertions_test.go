package testassert

import (
	"errors"
	"testing"
)

func TestAssertions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "equal helpers compare values",
			run: func(t *testing.T) {
				t.Helper()

				if !IsEqual(1, 1) {
					t.Fatal("values should be equal")
				}

				if IsEqual(1, 2) {
					t.Fatal("values should not be equal")
				}
			},
		},
		{
			name: "error helper compares wrapped errors",
			run: func(t *testing.T) {
				t.Helper()

				expected := errors.New("expected")
				received := errors.Join(expected)

				if !EqualError(received, expected) {
					t.Fatal("errors should be equal")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

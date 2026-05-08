package testutils

import (
	"errors"
	"testing"
)

func TestComparisonHelpers(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "approximately equal",
			run: func(t *testing.T) {
				t.Helper()

				if !ApproximatelyEqual(1, 1) {
					t.Fatal("values should be approximately equal")
				}
			},
		},
		{
			name: "typed mismatches are not equal",
			run: func(t *testing.T) {
				t.Helper()

				if IsEqual(float64(1), "1") {
					t.Fatal("float64 should not equal string")
				}

				if IsEqual(float32(1), "1") {
					t.Fatal("float32 should not equal string")
				}

				if IsEqual(errors.New("err"), "err") {
					t.Fatal("error should not equal string")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

func TestJSONHelpers(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "equal empty json",
			run: func(t *testing.T) {
				t.Helper()

				EqualJSON(t, nil, nil)
			},
		},
		{
			name: "equal object json",
			run: func(t *testing.T) {
				t.Helper()

				EqualJSON(t, []byte(`{"b":2,"a":1}`), []byte(`{"a":1,"b":2}`))
			},
		},
		{
			name: "not equal empty expected",
			run: func(t *testing.T) {
				t.Helper()

				NotEqualJSON(t, []byte(`{"a":1}`), nil)
			},
		},
		{
			name: "not equal object json",
			run: func(t *testing.T) {
				t.Helper()

				NotEqualJSON(t, []byte(`{"a":1}`), []byte(`{"a":2}`))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

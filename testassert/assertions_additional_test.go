package testassert

import "testing"

func TestAssertionWrappers(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "approximately equal wrapper",
			run: func(t *testing.T) {
				t.Helper()

				if !ApproximatelyEqual(1, 1) {
					t.Fatal("values should be approximately equal")
				}
			},
		},
		{
			name: "equal and not equal wrappers",
			run: func(t *testing.T) {
				t.Helper()

				Equal(t, "same", "same")
				NotEqual(t, "left", "right")
			},
		},
		{
			name: "json wrappers",
			run: func(t *testing.T) {
				t.Helper()

				EqualJSON(t, []byte(`{"a":1}`), []byte(`{"a":1}`))
				NotEqualJSON(t, []byte(`{"a":1}`), []byte(`{"a":2}`))
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

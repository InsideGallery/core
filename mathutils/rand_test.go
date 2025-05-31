package mathutils

import "testing"

func TestRandomDigitString(t *testing.T) {
	testcases := map[string]struct {
		length int
	}{
		"simple_usage": {
			length: 15,
		},
		"zero_value": {
			length: 0,
		},
		"min_over_max": {
			length: 1,
		},
	}

	for name, test := range testcases {
		test := test

		t.Run(name, func(t *testing.T) {
			n := RandomDigitString(test.length)
			if len(n) != test.length {
				t.Fatalf("Unexpected len: %d != %d", test.length, len(n))
			}
		})
	}
}

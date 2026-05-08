package mathx

import (
	"math/big"
	"testing"
)

func TestMathX(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "cantor pair round trips",
			run: func(t *testing.T) {
				t.Helper()

				left, right := CantorUnpair(CantorPair(7, 11))
				if left != 7 || right != 11 {
					t.Fatalf("pair = (%d, %d), want (7, 11)", left, right)
				}
			},
		},
		{
			name: "big integer halves round trip",
			run: func(t *testing.T) {
				t.Helper()

				value := big.NewInt(123)
				high, low := BigIntToHighAndLow(value)
				got := HighAndLowToBigInt(high, low)

				if got.Cmp(value) != 0 {
					t.Fatalf("value = %s, want %s", got, value)
				}
			},
		},
		{
			name: "clamp constrains value",
			run: func(t *testing.T) {
				t.Helper()

				if got := Clamp(10, 0, 5); got != 5 {
					t.Fatalf("clamp = %f, want 5", got)
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

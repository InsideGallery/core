package mathx

import (
	"math/big"
	"testing"
	"unicode/utf8"
)

func TestRoundWithPrecisionCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		value     float64
		precision float64
		want      float64
	}{
		{
			name:      "zero precision uses default",
			value:     1.23456,
			precision: 0,
			want:      1.2346,
		},
		{
			name:      "rounds to tenths",
			value:     1.25,
			precision: 0.1,
			want:      1.3,
		},
		{
			name:      "rounds negative values",
			value:     -1.25,
			precision: 0.1,
			want:      -1.3,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := RoundWithPrecision(test.value, test.precision); got != test.want {
				t.Fatalf("RoundWithPrecision() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestApproximatelyEqualCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		a    float64
		b    float64
		want bool
	}{
		{
			name: "exactly equal",
			a:    42,
			b:    42,
			want: true,
		},
		{
			name: "different",
			a:    42,
			b:    42.00000000000001,
			want: false,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := ApproximatelyEqual(test.a, test.b); got != test.want {
				t.Fatalf("ApproximatelyEqual() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRoundCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		value     float64
		precision float64
		want      float64
	}{
		{
			name:      "zero precision returns zero",
			value:     9.99,
			precision: 0,
			want:      0,
		},
		{
			name:      "truncates to hundredths",
			value:     9.999,
			precision: 100,
			want:      9.99,
		},
		{
			name:      "truncates negative values",
			value:     -9.999,
			precision: 100,
			want:      -9.99,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := Round(test.value, test.precision); got != test.want {
				t.Fatalf("Round() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestIntStringToBigIntCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		value string
		want  *big.Int
	}{
		{
			name:  "zero",
			value: "0",
			want:  big.NewInt(0),
		},
		{
			name:  "large integer",
			value: "340282366920938463463374607431768211455",
			want:  mustBigInt(t, "340282366920938463463374607431768211455"),
		},
		{
			name:  "invalid integer",
			value: "12x",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := IntStringToBigInt(test.value)
			if got == nil || test.want == nil {
				if got != test.want {
					t.Fatalf("IntStringToBigInt() = %v, want %v", got, test.want)
				}

				return
			}

			if got.Cmp(test.want) != 0 {
				t.Fatalf("IntStringToBigInt() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestWeightIndexBoundaries(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		probabilities map[interface{}]uint64
		want          interface{}
	}{
		{
			name: "empty map",
		},
		{
			name: "single weighted item",
			probabilities: map[interface{}]uint64{
				"only": 1,
			},
			want: "only",
		},
		{
			name: "single zero weight item",
			probabilities: map[interface{}]uint64{
				"zero": 0,
			},
			want: "zero",
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			if got := WeightIndex(test.probabilities); got != test.want {
				t.Fatalf("WeightIndex() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestRandomDigitStringLength(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		length int
	}{
		{
			name: "empty",
		},
		{
			name:   "single byte",
			length: 1,
		},
		{
			name:   "multiple bytes",
			length: 16,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got := RandomDigitString(test.length)
			if len(got) != test.length {
				t.Fatalf("len(RandomDigitString()) = %d, want %d", len(got), test.length)
			}

			if utf8.RuneCountInString(got) > test.length {
				t.Fatalf("RandomDigitString() rune count exceeds byte length")
			}
		})
	}
}

func mustBigInt(t *testing.T, value string) *big.Int {
	t.Helper()

	got, ok := new(big.Int).SetString(value, 10)
	if !ok {
		t.Fatalf("parse big int %q", value)
	}

	return got
}

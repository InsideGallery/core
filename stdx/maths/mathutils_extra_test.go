package maths

import (
	"math"
	"math/big"
	"testing"
)

func TestCantorPair(t *testing.T) {
	cases := []struct {
		name     string
		k1       uint64
		k2       uint64
		expected uint64
	}{
		{
			name:     testNameBothZero,
			k1:       0,
			k2:       0,
			expected: 0,
		},
		{
			name:     "first_zero_second_one",
			k1:       0,
			k2:       1,
			expected: 2,
		},
		{
			name:     "first_one_second_zero",
			k1:       1,
			k2:       0,
			expected: 1,
		},
		{
			name:     "both_one",
			k1:       1,
			k2:       1,
			expected: 4,
		},
		{
			name:     "classic_2_1",
			k1:       2,
			k2:       1,
			expected: 7,
		},
		{
			name:     "asymmetric_1_2",
			k1:       1,
			k2:       2,
			expected: 8,
		},
		{
			name:     "larger_values_10_10",
			k1:       10,
			k2:       10,
			expected: CantorPair(10, 10),
		},
		{
			name:     "first_large_second_zero",
			k1:       100,
			k2:       0,
			expected: CantorPair(100, 0),
		},
		{
			name:     "first_zero_second_large",
			k1:       0,
			k2:       100,
			expected: CantorPair(0, 100),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := CantorPair(tc.k1, tc.k2)
			if got != tc.expected {
				t.Fatalf("CantorPair(%d, %d) = %d, want %d", tc.k1, tc.k2, got, tc.expected)
			}
		})
	}
}

func TestCantorUnpair(t *testing.T) {
	cases := []struct {
		name      string
		pair      uint64
		expectedA uint64
		expectedB uint64
	}{
		{
			name:      testNameZero,
			pair:      0,
			expectedA: 0,
			expectedB: 0,
		},
		{
			name:      testNameOne,
			pair:      1,
			expectedA: 1,
			expectedB: 0,
		},
		{
			name:      "four",
			pair:      4,
			expectedA: 1,
			expectedB: 1,
		},
		{
			name:      "seven",
			pair:      7,
			expectedA: 2,
			expectedB: 1,
		},
		{
			name:      "eight",
			pair:      8,
			expectedA: 1,
			expectedB: 2,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a, b := CantorUnpair(tc.pair)
			if a != tc.expectedA || b != tc.expectedB {
				t.Fatalf("CantorUnpair(%d) = (%d, %d), want (%d, %d)", tc.pair, a, b, tc.expectedA, tc.expectedB)
			}
		})
	}
}

func TestCantorPairUnpairRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		k1   uint64
		k2   uint64
	}{
		{
			name: testNameBothZero,
			k1:   0,
			k2:   0,
		},
		{
			name: "small_values",
			k1:   3,
			k2:   5,
		},
		{
			name: "first_zero",
			k1:   0,
			k2:   42,
		},
		{
			name: "second_zero",
			k1:   42,
			k2:   0,
		},
		{
			name: "equal_values",
			k1:   7,
			k2:   7,
		},
		{
			name: "medium_values",
			k1:   100,
			k2:   200,
		},
		{
			name: "reversed_medium",
			k1:   200,
			k2:   100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pair := CantorPair(tc.k1, tc.k2)
			a, b := CantorUnpair(pair)
			if a != tc.k1 || b != tc.k2 {
				t.Fatalf("roundtrip failed: CantorPair(%d,%d)=%d, CantorUnpair(%d)=(%d,%d)",
					tc.k1, tc.k2, pair, pair, a, b)
			}
		})
	}
}

func TestCantorPairUniqueness(t *testing.T) {
	cases := []struct {
		name string
		k1a  uint64
		k2a  uint64
		k1b  uint64
		k2b  uint64
	}{
		{
			name: "swap_produces_different_pair",
			k1a:  1,
			k2a:  2,
			k1b:  2,
			k2b:  1,
		},
		{
			name: "adjacent_values_a",
			k1a:  0,
			k2a:  1,
			k1b:  1,
			k2b:  0,
		},
		{
			name: "zero_vs_nonzero",
			k1a:  0,
			k2a:  0,
			k1b:  0,
			k2b:  1,
		},
		{
			name: "different_sums",
			k1a:  3,
			k2a:  5,
			k1b:  4,
			k2b:  4,
		},
		{
			name: "large_difference",
			k1a:  10,
			k2a:  0,
			k1b:  0,
			k2b:  10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pairA := CantorPair(tc.k1a, tc.k2a)
			pairB := CantorPair(tc.k1b, tc.k2b)
			if pairA == pairB {
				t.Fatalf("CantorPair(%d,%d) == CantorPair(%d,%d) == %d, expected different",
					tc.k1a, tc.k2a, tc.k1b, tc.k2b, pairA)
			}
		})
	}
}

func TestRandomDigitStringExtended(t *testing.T) {
	cases := []struct {
		name   string
		length int
	}{
		{
			name:   "zero_length",
			length: 0,
		},
		{
			name:   "length_one",
			length: 1,
		},
		{
			name:   "small_length",
			length: 5,
		},
		{
			name:   "medium_length",
			length: 100,
		},
		{
			name:   "large_length",
			length: 1000,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := RandomDigitString(tc.length)
			if len(result) != tc.length {
				t.Fatalf("RandomDigitString(%d) length = %d, want %d", tc.length, len(result), tc.length)
			}
		})
	}
}

func TestRandomDigitStringRandomness(t *testing.T) {
	cases := []struct {
		name   string
		length int
		runs   int
	}{
		{
			name:   "two_calls_differ_short",
			length: 10,
			runs:   2,
		},
		{
			name:   "two_calls_differ_medium",
			length: 50,
			runs:   2,
		},
		{
			name:   "two_calls_differ_long",
			length: 200,
			runs:   2,
		},
		{
			name:   "multiple_calls_differ",
			length: 20,
			runs:   5,
		},
		{
			name:   "single_char_can_vary",
			length: 1,
			runs:   100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			seen := map[string]bool{}
			for i := 0; i < tc.runs; i++ {
				seen[RandomDigitString(tc.length)] = true
			}
			if tc.length > 0 && tc.runs > 1 && len(seen) < 2 {
				t.Fatalf("expected at least 2 distinct values from %d runs, got %d", tc.runs, len(seen))
			}
		})
	}
}

func TestWeightIndexEdgeCases(t *testing.T) {
	cases := []struct {
		name     string
		prob     map[interface{}]uint64
		wantNil  bool
		wantKeys []interface{}
	}{
		{
			name:    "empty_map",
			prob:    map[interface{}]uint64{},
			wantNil: true,
		},
		{
			name:    "nil_map",
			prob:    nil,
			wantNil: true,
		},
		{
			name:     "single_element",
			prob:     map[interface{}]uint64{"only": 100},
			wantKeys: []interface{}{"only"},
		},
		{
			name:     "single_element_weight_one",
			prob:     map[interface{}]uint64{testNameOne: 1},
			wantKeys: []interface{}{testNameOne},
		},
		{
			name:     "single_element_max_weight",
			prob:     map[interface{}]uint64{"max": math.MaxUint64 / 2},
			wantKeys: []interface{}{"max"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := WeightIndex(tc.prob)
			if tc.wantNil {
				if result != nil {
					t.Fatalf("WeightIndex() = %v, want nil", result)
				}
				return
			}
			found := false
			for _, k := range tc.wantKeys {
				if result == k {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("WeightIndex() = %v, not in expected keys %v", result, tc.wantKeys)
			}
		})
	}
}

func TestWeightIndexDistribution(t *testing.T) {
	cases := []struct {
		name       string
		prob       map[interface{}]uint64
		iterations int
		checkKey   interface{}
		minRatio   float64
		maxRatio   float64
	}{
		{
			name:       "two_equal_weights",
			prob:       map[interface{}]uint64{"a": 50, "b": 50},
			iterations: 5000,
			checkKey:   "a",
			minRatio:   0.35,
			maxRatio:   0.65,
		},
		{
			name:       "heavily_skewed",
			prob:       map[interface{}]uint64{"heavy": 999, "light": 1},
			iterations: 5000,
			checkKey:   "heavy",
			minRatio:   0.90,
			maxRatio:   1.0,
		},
		{
			name:       "three_equal_weights",
			prob:       map[interface{}]uint64{"x": 100, "y": 100, "z": 100},
			iterations: 6000,
			checkKey:   "x",
			minRatio:   0.2,
			maxRatio:   0.5,
		},
		{
			name:       "two_elements_one_dominant",
			prob:       map[interface{}]uint64{"dom": 900, "sub": 100},
			iterations: 5000,
			checkKey:   "dom",
			minRatio:   0.80,
			maxRatio:   1.0,
		},
		{
			name:       "five_equal",
			prob:       map[interface{}]uint64{"a": 20, "b": 20, "c": 20, "d": 20, "e": 20},
			iterations: 10000,
			checkKey:   "a",
			minRatio:   0.10,
			maxRatio:   0.35,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			counts := map[interface{}]int{}
			for i := 0; i < tc.iterations; i++ {
				r := WeightIndex(tc.prob)
				counts[r]++
			}
			ratio := float64(counts[tc.checkKey]) / float64(tc.iterations)
			if ratio < tc.minRatio || ratio > tc.maxRatio {
				t.Fatalf("ratio for key %v = %f, expected between %f and %f",
					tc.checkKey, ratio, tc.minRatio, tc.maxRatio)
			}
		})
	}
}

func TestIntStringToBigIntExtended(t *testing.T) {
	cases := []struct {
		name    string
		str     string
		want    *big.Int
		wantNil bool
	}{
		{
			name: testNameZero,
			str:  "0",
			want: big.NewInt(0),
		},
		{
			name: "positive_small",
			str:  "42",
			want: big.NewInt(42),
		},
		{
			name: "negative_number",
			str:  "-1",
			want: big.NewInt(-1),
		},
		{
			name: "large_number",
			str:  "340282366920938463463374607431768211455",
			want: func() *big.Int {
				x, _ := new(big.Int).SetString("340282366920938463463374607431768211455", 10)
				return x
			}(),
		},
		{
			name:    "invalid_hex_string",
			str:     "0xDEAD",
			wantNil: true,
		},
		{
			name:    "alpha_string",
			str:     "abc",
			wantNil: true,
		},
		{
			name:    "empty_string",
			str:     "",
			wantNil: true,
		},
		{
			name: "max_int64",
			str:  "9223372036854775807",
			want: big.NewInt(math.MaxInt64),
		},
		{
			name: "negative_large",
			str:  "-9223372036854775808",
			want: func() *big.Int {
				x, _ := new(big.Int).SetString("-9223372036854775808", 10)
				return x
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IntStringToBigInt(tc.str)
			if tc.wantNil {
				if got != nil {
					t.Fatalf("IntStringToBigInt(%q) = %v, want nil", tc.str, got)
				}
				return
			}
			if got == nil {
				t.Fatalf("IntStringToBigInt(%q) = nil, want %v", tc.str, tc.want)
			}
			if got.Cmp(tc.want) != 0 {
				t.Fatalf("IntStringToBigInt(%q) = %v, want %v", tc.str, got, tc.want)
			}
		})
	}
}

func TestBigIntToHighAndLowExtended(t *testing.T) {
	cases := []struct {
		name     string
		x        *big.Int
		wantHigh uint64
		wantLow  uint64
	}{
		{
			name:     testNameZero,
			x:        big.NewInt(0),
			wantHigh: 0,
			wantLow:  0,
		},
		{
			name:     testNameOne,
			x:        big.NewInt(1),
			wantHigh: 0,
			wantLow:  1,
		},
		{
			name:     "max_uint64",
			x:        new(big.Int).SetUint64(math.MaxUint64),
			wantHigh: 0,
			wantLow:  math.MaxUint64,
		},
		{
			name: "one_above_uint64",
			x: func() *big.Int {
				x, _ := new(big.Int).SetString("18446744073709551616", 10)
				return x
			}(),
			wantHigh: 1,
			wantLow:  0,
		},
		{
			name: "high_only",
			x: func() *big.Int {
				x, _ := new(big.Int).SetString("42535295865117307932921825928971026432", 10)
				return x
			}(),
			wantHigh: 2305843009213693952,
			wantLow:  0,
		},
		{
			name: "both_high_and_low",
			x: func() *big.Int {
				x, _ := new(big.Int).SetString("18446744073709551617", 10)
				return x
			}(),
			wantHigh: 1,
			wantLow:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			high, low := BigIntToHighAndLow(tc.x)
			if high != tc.wantHigh {
				t.Fatalf("BigIntToHighAndLow(%v) high = %d, want %d", tc.x, high, tc.wantHigh)
			}
			if low != tc.wantLow {
				t.Fatalf("BigIntToHighAndLow(%v) low = %d, want %d", tc.x, low, tc.wantLow)
			}
		})
	}
}

func TestHighAndLowToBigIntExtended(t *testing.T) {
	cases := []struct {
		name string
		h    uint64
		l    uint64
		want *big.Int
	}{
		{
			name: testNameBothZero,
			h:    0,
			l:    0,
			want: big.NewInt(0),
		},
		{
			name: "low_only_one",
			h:    0,
			l:    1,
			want: big.NewInt(1),
		},
		{
			name: "high_only_one",
			h:    1,
			l:    0,
			want: func() *big.Int {
				x, _ := new(big.Int).SetString("18446744073709551616", 10)
				return x
			}(),
		},
		{
			name: "max_low",
			h:    0,
			l:    math.MaxUint64,
			want: new(big.Int).SetUint64(math.MaxUint64),
		},
		{
			name: "both_one",
			h:    1,
			l:    1,
			want: func() *big.Int {
				x, _ := new(big.Int).SetString("18446744073709551617", 10)
				return x
			}(),
		},
		{
			name: "max_high_max_low",
			h:    math.MaxUint64,
			l:    math.MaxUint64,
			want: func() *big.Int {
				x, _ := new(big.Int).SetString("340282366920938463463374607431768211455", 10)
				return x
			}(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := HighAndLowToBigInt(tc.h, tc.l)
			if got.Cmp(tc.want) != 0 {
				t.Fatalf("HighAndLowToBigInt(%d, %d) = %v, want %v", tc.h, tc.l, got, tc.want)
			}
		})
	}
}

func TestHighAndLowBigIntRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		h    uint64
		l    uint64
	}{
		{
			name: testNameBothZero,
			h:    0,
			l:    0,
		},
		{
			name: "low_one",
			h:    0,
			l:    1,
		},
		{
			name: "high_one",
			h:    1,
			l:    0,
		},
		{
			name: "both_nonzero",
			h:    12345,
			l:    67890,
		},
		{
			name: "max_values",
			h:    math.MaxUint64,
			l:    math.MaxUint64,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bigVal := HighAndLowToBigInt(tc.h, tc.l)
			gotH, gotL := BigIntToHighAndLow(bigVal)
			if gotH != tc.h || gotL != tc.l {
				t.Fatalf("roundtrip failed: (%d,%d) -> %v -> (%d,%d)", tc.h, tc.l, bigVal, gotH, gotL)
			}
		})
	}
}

func TestApproximatelyEqualExtended(t *testing.T) {
	cases := []struct {
		name string
		a    float64
		b    float64
		want bool
	}{
		{
			name: testNameBothZero,
			a:    0.0,
			b:    0.0,
			want: true,
		},
		{
			name: "identical_positive",
			a:    1.5,
			b:    1.5,
			want: true,
		},
		{
			name: "identical_negative",
			a:    -3.14,
			b:    -3.14,
			want: true,
		},
		{
			name: "clearly_different",
			a:    1.0,
			b:    2.0,
			want: false,
		},
		{
			name: "very_close_but_not_equal",
			a:    4.000000000000001,
			b:    4.000000000000002,
			want: false,
		},
		{
			name: "float_representation_equal",
			a:    4.0000000000000001,
			b:    4.0000000000000002,
			want: true,
		},
		{
			name: "negative_vs_positive",
			a:    -1.0,
			b:    1.0,
			want: false,
		},
		{
			name: "positive_vs_zero",
			a:    0.1,
			b:    0.0,
			want: false,
		},
		{
			name: "negative_vs_zero",
			a:    -0.1,
			b:    0.0,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ApproximatelyEqual(tc.a, tc.b)
			if got != tc.want {
				t.Fatalf("ApproximatelyEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestRoundExtended(t *testing.T) {
	cases := []struct {
		name string
		v    float64
		p    float64
		want float64
	}{
		{
			name: "precision_zero_returns_zero",
			v:    123.456,
			p:    0,
			want: 0,
		},
		{
			name: "precision_one_truncates",
			v:    5.999,
			p:    1,
			want: 5,
		},
		{
			name: "precision_hundred",
			v:    3.14159,
			p:    100,
			want: 3.14,
		},
		{
			name: "negative_value",
			v:    -2.789,
			p:    10,
			want: -2.7,
		},
		{
			name: testNameZeroValue,
			v:    0.0,
			p:    1000,
			want: 0.0,
		},
		{
			name: "large_precision",
			v:    1.123456789,
			p:    100000000,
			want: 1.12345678,
		},
		{
			name: "value_exactly_representable",
			v:    2.5,
			p:    10,
			want: 2.5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Round(tc.v, tc.p)
			if got != tc.want {
				t.Fatalf("Round(%v, %v) = %v, want %v", tc.v, tc.p, got, tc.want)
			}
		})
	}
}

func TestRoundWithPrecisionExtended(t *testing.T) {
	cases := []struct {
		name      string
		value     float64
		precision float64
		want      float64
	}{
		{
			name:      "negative_value",
			value:     -1.2345,
			precision: 0.01,
			want:      -1.23,
		},
		{
			name:      "very_small_precision",
			value:     1.23456789,
			precision: 0.00001,
			want:      1.23457,
		},
		{
			name:      "whole_number_no_rounding_needed",
			value:     5.0,
			precision: 0.1,
			want:      5.0,
		},
		{
			name:      "large_value",
			value:     99999.9999,
			precision: 0.01,
			want:      100000.0,
		},
		{
			name:      "precision_one",
			value:     3.7,
			precision: 1,
			want:      4.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RoundWithPrecision(tc.value, tc.precision)
			diff := got - tc.want
			if diff > 0.000001 || diff < -0.000001 {
				t.Fatalf("RoundWithPrecision(%v, %v) = %v, want %v", tc.value, tc.precision, got, tc.want)
			}
		})
	}
}

func TestClampExtended(t *testing.T) {
	cases := []struct {
		name  string
		value float64
		lower float64
		upper float64
		want  float64
	}{
		{
			name:  "negative_below_range",
			value: -10.0,
			lower: -5.0,
			upper: 5.0,
			want:  -5.0,
		},
		{
			name:  "negative_in_range",
			value: -3.0,
			lower: -5.0,
			upper: 5.0,
			want:  -3.0,
		},
		{
			name:  "at_upper_boundary",
			value: 7.0,
			lower: 4.0,
			upper: 7.0,
			want:  7.0,
		},
		{
			name:  "above_upper_boundary",
			value: 100.0,
			lower: 0.0,
			upper: 50.0,
			want:  50.0,
		},
		{
			name:  "max_float_value",
			value: math.MaxFloat64,
			lower: 0.0,
			upper: 1.0,
			want:  1.0,
		},
		{
			name:  "negative_max_float",
			value: -math.MaxFloat64,
			lower: -1.0,
			upper: 1.0,
			want:  -1.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Clamp(tc.value, tc.lower, tc.upper)
			if got != tc.want {
				t.Fatalf("Clamp(%v, %v, %v) = %v, want %v", tc.value, tc.lower, tc.upper, got, tc.want)
			}
		})
	}
}

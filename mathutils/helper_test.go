//go:build unit
// +build unit

package mathutils

import (
	"fmt"
	"math/big"
	"net"
	"testing"

	"github.com/InsideGallery/core/dataconv"
	"github.com/InsideGallery/core/testutils"
)

func TestRoundWithPrecision(t *testing.T) {
	testcases := map[string]struct {
		value     float64
		precision float64
		result    float64
	}{
		"zero_precision": {
			value:     1.55,
			precision: 0.0,
			result:    1.55,
		},
		"zero_value": {
			value:     0.0,
			precision: 0.1,
			result:    0.0,
		},
		"value:1.2": {
			value:     1.2345,
			precision: 0.1,
			result:    1.2,
		},
		"value:1.8": {
			value:     1.7654,
			precision: 0.1,
			result:    1.8,
		},
		"value:1.8555": {
			value:     1.8555,
			precision: 0.0001,
			result:    1.8555,
		},
		"value:1.9": {
			value:     1.8555,
			precision: 0.1,
			result:    1.9,
		},
	}

	for name, test := range testcases {
		test := test
		t.Run(name, func(t *testing.T) {
			testutils.Equal(t, RoundWithPrecision(test.value, test.precision), test.result)
		})
	}
}

func TestClamp(t *testing.T) {
	testcases := map[string]struct {
		value, lowerLimit, upperLimit, result float64
	}{
		"clamp:1.0": {
			value:      1.0,
			lowerLimit: 1.0,
			upperLimit: 1.0,
			result:     1.0,
		},
		"clamp:0.0": {
			value:      0.0,
			lowerLimit: 0.0,
			upperLimit: 0.0,
			result:     0.0,
		},
		"clamp:4.0": {
			value:      2.0,
			lowerLimit: 4.0,
			upperLimit: 7.0,
			result:     4.0,
		},
		"clamp:5.0": {
			value:      5.0,
			lowerLimit: 4.0,
			upperLimit: 7.0,
			result:     5.0,
		},
		"clamp:7.0": {
			value:      9.0,
			lowerLimit: 4.0,
			upperLimit: 7.0,
			result:     7.0,
		},
		"clamp_2:4.0": {
			value:      4.0,
			lowerLimit: 4.0,
			upperLimit: 7.0,
			result:     4.0,
		},
	}

	for name, test := range testcases {
		test := test
		t.Run(name, func(t *testing.T) {
			testutils.Equal(t, Clamp(test.value, test.lowerLimit, test.upperLimit), test.result)
		})
	}
}

func TestBigIntToHighAndLow2(t *testing.T) {
	rawip := "46.219.132.112"
	rip := net.ParseIP(rawip)
	intip := dataconv.IPv6ToBigInt(rip)

	high, low := BigIntToHighAndLow(intip)
	fmt.Println(high, low)
}

func TestApproximatelyEqual(t *testing.T) {
	testcases := map[string]struct {
		a, b float64
		r    bool
	}{
		"equal": {
			a: 4.0000000000000001,
			b: 4.0000000000000002,
			r: true,
		},
		"not_equal": {
			a: 4.000000000000001,
			b: 4.000000000000002,
			r: false,
		},
	}

	for name, test := range testcases {
		test := test
		t.Run(name, func(t *testing.T) {
			testutils.Equal(t, ApproximatelyEqual(test.a, test.b), test.r)
		})
	}
}

func TestRound(t *testing.T) {
	testcases := map[string]struct {
		v float64
		p float64
		r float64
	}{
		"8 digits": {
			v: 5.123456789123,
			p: 100000000,
			r: 5.12345678,
		},
		"0 digit": {
			v: 5.123456789123,
			p: 1,
			r: 5,
		},
		"zero": {
			v: 5.123456789123,
			p: 0,
			r: 0,
		},
	}

	for name, test := range testcases {
		test := test
		t.Run(name, func(t *testing.T) {
			testutils.Equal(t, Round(test.v, test.p), test.r)
		})
	}
}

func TestIntStringToBigInt(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "string 1",
			args: args{str: "1"},
			want: big.NewInt(1),
		},
		{
			name: "empty string",
			args: args{str: ""},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IntStringToBigInt(tt.args.str)
			if got == nil && tt.want == nil {
				return
			}
			if got.Cmp(tt.want) != 0 {
				t.Errorf("IntStringToBigInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBigIntToHighAndLow(t *testing.T) {
	type args struct {
		x *big.Int
	}
	tests := []struct {
		name  string
		args  args
		want  uint64
		want1 uint64
	}{
		{
			name:  "int 1",
			args:  args{x: IntStringToBigInt("1")},
			want:  0,
			want1: 1,
		},
		{
			name:  "int 42535295865117307932921825928971026432",
			args:  args{x: IntStringToBigInt("42535295865117307932921825928971026432")},
			want:  2305843009213693952,
			want1: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := BigIntToHighAndLow(tt.args.x)
			if got != tt.want {
				t.Errorf("BigIntToHighAndLow() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("BigIntToHighAndLow() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestHighAndLowToBigInt(t *testing.T) {
	type args struct {
		h uint64
		l uint64
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "18446744073709551616 to bigInt",
			args: args{
				h: 1,
				l: 0,
			},
			want: IntStringToBigInt("18446744073709551616"),
		},
		{
			name: "1 to bigInt",
			args: args{
				h: 0,
				l: 1,
			},
			want: IntStringToBigInt("1"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HighAndLowToBigInt(tt.args.h, tt.args.l)
			if got == nil && tt.want == nil {
				return
			}
			if got.Cmp(tt.want) != 0 {
				t.Errorf("IntStringToBigInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

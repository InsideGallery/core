package mathutils

import (
	"encoding/binary"
	"math"
	"math/big"
)

// mathematics constants
const (
	DefaultPrecision float64 = 0.0001
)

// RoundWithPrecision round precision
func RoundWithPrecision(value float64, precision float64) float64 {
	if ApproximatelyEqual(precision, 0) {
		precision = DefaultPrecision
	}

	precision = 1 / precision

	return math.Round(value*precision) / precision
}

// Clamp return the result of the "value" clamped by
func Clamp(value, lowerLimit, upperLimit float64) float64 {
	if value < lowerLimit {
		return lowerLimit
	} else if value > upperLimit {
		return upperLimit
	}

	return value
}

// ApproximatelyEqual function to test if two real numbers are (almost) equal
func ApproximatelyEqual(a, b float64) bool {
	epsilon := math.SmallestNonzeroFloat64
	difference := a - b

	return difference < epsilon && difference > -epsilon
}

// Round float to precision
func Round(v float64, p float64) float64 {
	if p == 0 {
		return 0
	}

	return float64(int64(v*p)) / p
}

func IntStringToBigInt(str string) *big.Int {
	x := big.NewInt(0)

	x, ok := x.SetString(str, 10) //nolint:mnd
	if !ok {
		return nil
	}

	return x
}

func BigIntToHighAndLow(x *big.Int) (uint64, uint64) {
	bytes := make([]byte, 16) //nolint:mnd
	x.FillBytes(bytes)
	v1 := binary.BigEndian.Uint64(bytes[0:8])
	v2 := binary.BigEndian.Uint64(bytes[8:16])

	return v1, v2
}

func HighAndLowToBigInt(h, l uint64) *big.Int {
	sh := make([]byte, 8) //nolint:mnd
	binary.BigEndian.PutUint64(sh, l)

	sl := make([]byte, 8) //nolint:mnd
	binary.BigEndian.PutUint64(sl, h)

	x := big.NewInt(0)
	x.SetBytes(append(sl, sh...))

	return x
}

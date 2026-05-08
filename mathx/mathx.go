// Package mathx provides focused math and probability helpers.
//
// New code should import this package instead of the legacy mathutils path:
//
//	import "github.com/InsideGallery/core/mathx"
//
// Compatibility: github.com/InsideGallery/core/mathutils remains available for
// existing consumers. Keep new math helpers in mathx so applications can migrate
// away from the legacy aggregate path without changing behavior.
package mathx

import (
	"math/big"

	"github.com/InsideGallery/core/mathutils"
)

// DefaultPrecision is the fallback precision used by RoundWithPrecision.
const DefaultPrecision = mathutils.DefaultPrecision

// CantorPair pairs two uint64 values with the Cantor pairing function.
func CantorPair(k1, k2 uint64) uint64 {
	return mathutils.CantorPair(k1, k2)
}

// CantorUnpair splits a Cantor-paired uint64 into two values.
func CantorUnpair(pair uint64) (uint64, uint64) {
	return mathutils.CantorUnpair(pair)
}

// RoundWithPrecision rounds a float using the requested precision.
func RoundWithPrecision(value, precision float64) float64 {
	return mathutils.RoundWithPrecision(value, precision)
}

// Clamp constrains value to the inclusive lower and upper limits.
func Clamp(value, lowerLimit, upperLimit float64) float64 {
	return mathutils.Clamp(value, lowerLimit, upperLimit)
}

// ApproximatelyEqual reports whether two floats are nearly equal.
func ApproximatelyEqual(a, b float64) bool {
	return mathutils.ApproximatelyEqual(a, b)
}

// Round truncates a float to a precision multiplier.
func Round(value, precision float64) float64 {
	return mathutils.Round(value, precision)
}

// IntStringToBigInt parses a base-10 integer string.
func IntStringToBigInt(str string) *big.Int {
	return mathutils.IntStringToBigInt(str)
}

// BigIntToHighAndLow splits a big integer into high and low uint64 halves.
func BigIntToHighAndLow(value *big.Int) (uint64, uint64) {
	return mathutils.BigIntToHighAndLow(value)
}

// HighAndLowToBigInt joins high and low uint64 halves into a big integer.
func HighAndLowToBigInt(high, low uint64) *big.Int {
	return mathutils.HighAndLowToBigInt(high, low)
}

// WeightIndex returns a weighted random key from the probability map.
func WeightIndex(probabilities map[interface{}]uint64) interface{} { //nolint:ireturn
	return mathutils.WeightIndex(probabilities)
}

// RandomDigitString returns a random byte string with the requested length.
func RandomDigitString(length int) string {
	return mathutils.RandomDigitString(length)
}

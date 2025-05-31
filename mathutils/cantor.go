package mathutils

import "math"

// CantorPair two uint64 numbers by cantor pairing function.
func CantorPair(k1, k2 uint64) uint64 {
	pair := k1 + k2
	pair *= pair + 1
	pair /= 2 // nolint mnd
	pair += k2

	return pair
}

// CantorUnpair one uint64 pair to two uint64 numbers by cantor pairing function.
func CantorUnpair(pair uint64) (uint64, uint64) {
	w := math.Floor((math.Sqrt(float64(8*pair+1)) - 1) / 2) // nolint:mnd
	t := (w*w + w) / 2                                      // nolint:mnd

	k2 := pair - uint64(t)
	k1 := uint64(w) - k2

	return k1, k2
}

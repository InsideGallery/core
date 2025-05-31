package mathutils

import "math/rand"

const maxSize = 255

// RandomDigitString return random digit string
func RandomDigitString(length int) (result string) {
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = uint8(rand.Intn(maxSize)) // nolint:gosec
	}

	return string(b)
}

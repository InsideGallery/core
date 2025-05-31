package rsa

import (
	"testing"

	"github.com/InsideGallery/core/testutils"
)

func TestRSACipher(t *testing.T) {
	a, err := NewRSA(DefaultBitsSize)
	testutils.Equal(t, err, nil)

	val := []byte("test string")
	res, err := a.Encrypt(val)
	testutils.Equal(t, err, nil)

	original, err := a.Decrypt(res)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, val, original)
}

func TestRSACipherRestore(t *testing.T) {
	a, err := NewRSA(DefaultBitsSize)
	testutils.Equal(t, err, nil)

	raw, err := a.ToBinary()
	testutils.Equal(t, err, nil)

	c, err := a.FromBinary(raw)
	testutils.Equal(t, err, nil)

	val := []byte("test string")
	res, err := c.Encrypt(val)
	testutils.Equal(t, err, nil)

	original, err := c.Decrypt(res)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, val, original)
}

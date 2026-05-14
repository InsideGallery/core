package rsa_test

import (
	"testing"

	"github.com/FrogoAI/testutils"

	corersa "github.com/InsideGallery/core/pki/rsa"
)

func TestRSACipher(t *testing.T) {
	a, err := corersa.NewRSA(corersa.DefaultBitsSize)
	testutils.Equal(t, err, nil)

	val := []byte("test strings")
	res, err := a.Encrypt(val)
	testutils.Equal(t, err, nil)

	original, err := a.Decrypt(res)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, val, original)
}

func TestRSACipherRestore(t *testing.T) {
	a, err := corersa.NewRSA(corersa.DefaultBitsSize)
	testutils.Equal(t, err, nil)

	raw, err := a.ToBinary()
	testutils.Equal(t, err, nil)

	c, err := a.FromBinary(raw)
	testutils.Equal(t, err, nil)

	val := []byte("test strings")
	res, err := c.Encrypt(val)
	testutils.Equal(t, err, nil)

	original, err := c.Decrypt(res)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, val, original)
}

package aes_test

import (
	"errors"
	"testing"

	coreaes "github.com/InsideGallery/core/pki/aes"
	"github.com/InsideGallery/core/testutils"
)

func TestAESCipher(t *testing.T) {
	a, err := coreaes.NewAES(coreaes.AES32)
	testutils.Equal(t, err, nil)

	val := []byte("test string")
	res, err := a.Encrypt(val)
	testutils.Equal(t, err, nil)

	original, err := a.Decrypt(res)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, val, original)
}

func TestNewAESStrict(t *testing.T) {
	cases := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{
			name: "supported size",
			size: coreaes.AES32,
		},
		{
			name:    "unsupported size",
			size:    1,
			wantErr: true,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			cipher, err := coreaes.NewAESStrict(test.size)
			if test.wantErr {
				if !errors.Is(err, coreaes.ErrInvalidAESSize) {
					t.Fatalf("err = %v, want %v", err, coreaes.ErrInvalidAESSize)
				}

				if cipher != nil {
					t.Fatal("cipher should be nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("new aes strict: %v", err)
			}

			raw, err := cipher.ToBinary()
			if err != nil {
				t.Fatalf("to binary: %v", err)
			}

			if len(raw) != test.size {
				t.Fatalf("key size = %d, want %d", len(raw), test.size)
			}
		})
	}
}

func TestAESCipherRestore(t *testing.T) {
	a, err := coreaes.NewAES(coreaes.AES32)
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

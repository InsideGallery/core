package aes_test

import (
	"errors"
	"testing"

	coreaes "github.com/InsideGallery/core/pki/aes"
)

func TestAESAdditional(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "supported sizes create keys",
			run: func(t *testing.T) {
				t.Helper()

				for _, size := range []int{coreaes.AES16, coreaes.AES24, coreaes.AES32} {
					cipher, err := coreaes.NewAES(size)
					if err != nil {
						t.Fatalf("new aes(%d): %v", size, err)
					}

					raw, err := cipher.ToBinary()
					if err != nil {
						t.Fatalf("to binary: %v", err)
					}

					if len(raw) != size {
						t.Fatalf("key size = %d, want %d", len(raw), size)
					}

					if cipher.Kind() != "aes" {
						t.Fatalf("kind = %q, want aes", cipher.Kind())
					}
				}
			},
		},
		{
			name: "unsupported size returns sentinel",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := coreaes.NewAES(1)
				if !errors.Is(err, coreaes.ErrInvalidAESSize) {
					t.Fatalf("err = %v, want %v", err, coreaes.ErrInvalidAESSize)
				}

				if cipher != nil {
					t.Fatal("cipher should be nil")
				}
			},
		},
		{
			name: "invalid restored key returns encrypt and decrypt errors",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := (&coreaes.AES{}).FromBinary([]byte("bad"))
				if err != nil {
					t.Fatalf("from binary: %v", err)
				}

				if _, err := cipher.Encrypt([]byte("data")); err == nil {
					t.Fatal("expected encrypt error")
				}

				if _, err := cipher.Decrypt([]byte("00")); err == nil {
					t.Fatal("expected decrypt error")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

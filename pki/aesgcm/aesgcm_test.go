package aesgcm

import (
	"errors"
	"testing"
)

func TestAESGCM(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new cipher encrypts and decrypts",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := New(KeySize32)
				if err != nil {
					t.Fatalf("new cipher: %v", err)
				}

				plaintext := []byte("message")
				encrypted, err := cipher.Encrypt(plaintext)
				if err != nil {
					t.Fatalf("encrypt: %v", err)
				}

				got, err := cipher.Decrypt(encrypted)
				if err != nil {
					t.Fatalf("decrypt: %v", err)
				}

				if string(got) != string(plaintext) {
					t.Fatalf("plaintext = %q, want %q", got, plaintext)
				}
			},
		},
		{
			name: "invalid key size returns sentinel",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := New(1)
				if !errors.Is(err, ErrInvalidKeySize) {
					t.Fatalf("err = %v, want %v", err, ErrInvalidKeySize)
				}

				if cipher != nil {
					t.Fatal("cipher should be nil")
				}
			},
		},
	}

	for _, test := range cases {
		t.Run(test.name, test.run)
	}
}

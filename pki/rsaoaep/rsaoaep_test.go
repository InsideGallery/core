package rsaoaep

import (
	"errors"
	"testing"
)

func TestRSAOAEP(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "new cipher restores from private key",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := New(DefaultKeyBits)
				if err != nil {
					t.Fatalf("new cipher: %v", err)
				}

				raw, err := cipher.ToBinary()
				if err != nil {
					t.Fatalf("to binary: %v", err)
				}

				restored, err := FromPrivateKey(raw)
				if err != nil {
					t.Fatalf("from private key: %v", err)
				}

				plaintext := []byte("message")
				encrypted, err := restored.Encrypt(plaintext)
				if err != nil {
					t.Fatalf("encrypt: %v", err)
				}

				got, err := restored.Decrypt(encrypted)
				if err != nil {
					t.Fatalf("decrypt: %v", err)
				}

				if string(got) != string(plaintext) {
					t.Fatalf("plaintext = %q, want %q", got, plaintext)
				}
			},
		},
		{
			name: "invalid private key returns sentinel",
			run: func(t *testing.T) {
				t.Helper()

				cipher, err := FromPrivateKey([]byte("bad"))
				if !errors.Is(err, ErrFailedToParsePEMBlock) {
					t.Fatalf("err = %v, want %v", err, ErrFailedToParsePEMBlock)
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

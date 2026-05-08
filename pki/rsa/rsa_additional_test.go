package rsa_test

import (
	"errors"
	"testing"

	corersa "github.com/InsideGallery/core/pki/rsa"
)

func TestRSAKind(t *testing.T) {
	cipher, err := corersa.NewRSA(1024)
	if err != nil {
		t.Fatalf("NewRSA(): %v", err)
	}

	if got := cipher.Kind(); got != "rsa" {
		t.Fatalf("Kind() = %q, want %q", got, "rsa")
	}
}

func TestFromPrivateKeyErrors(t *testing.T) {
	cases := []struct {
		name    string
		key     []byte
		wantErr error
	}{
		{
			name:    "invalid pem",
			key:     []byte("not pem"),
			wantErr: corersa.ErrFailedToParsePEMBlock,
		},
		{
			name: "invalid private key block",
			key: []byte(`-----BEGIN RSA PRIVATE KEY-----
bm90IGRlcg==
-----END RSA PRIVATE KEY-----`),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := corersa.FromPrivateKey(test.key)
			if err == nil {
				t.Fatalf("FromPrivateKey() err = nil, cipher = %v", got)
			}

			if test.wantErr != nil && !errors.Is(err, test.wantErr) {
				t.Fatalf("FromPrivateKey() err = %v, want %v", err, test.wantErr)
			}
		})
	}
}

func TestNewRSARejectsInvalidBits(t *testing.T) {
	got, err := corersa.NewRSA(0)
	if err == nil {
		t.Fatalf("NewRSA() err = nil, cipher = %v", got)
	}
}

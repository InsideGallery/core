package cryptor

import (
	"context"
	"errors"
	"testing"
)

func TestCryptor(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		operation func(context.Context, Cipher, []byte) (Result, error)
		cipher    Cipher
		input     []byte
		want      []byte
		wantErr   error
	}{
		{
			name:      "encrypt",
			operation: Encrypt,
			cipher:    stubCipher{},
			input:     []byte("plain"),
			want:      []byte("enc:plain"),
		},
		{
			name:      "decrypt",
			operation: Decrypt,
			cipher:    stubCipher{},
			input:     []byte("enc:plain"),
			want:      []byte("plain"),
		},
		{
			name:      "nil cipher",
			operation: Encrypt,
			input:     []byte("plain"),
			wantErr:   ErrCipherNotSet,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := test.operation(context.Background(), test.cipher, test.input)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("err = %v, want %v", err, test.wantErr)
				}

				return
			}

			if err != nil {
				t.Fatalf("operation: %v", err)
			}

			if got.Kind != "stub" {
				t.Fatalf("Kind = %q, want stub", got.Kind)
			}

			if string(got.Data) != string(test.want) {
				t.Fatalf("Data = %q, want %q", got.Data, test.want)
			}
		})
	}
}

type stubCipher struct{}

func (stubCipher) Encrypt(data []byte) ([]byte, error) {
	return append([]byte("enc:"), data...), nil
}

func (stubCipher) Decrypt(data []byte) ([]byte, error) {
	return []byte(string(data[4:])), nil
}

func (stubCipher) Kind() string {
	return "stub"
}

func (stubCipher) ToBinary() ([]byte, error) {
	return nil, nil
}

func (stubCipher) FromBinary([]byte) (Cipher, error) { //nolint:ireturn
	return stubCipher{}, nil
}

package diversify

import (
	"errors"
	"testing"
)

func TestKeyRejectsWrongLengths(t *testing.T) {
	cases := []struct {
		name string
		key  []byte
	}{
		{
			name: "nil key",
		},
		{
			name: "short key",
			key:  make([]byte, aesKeySize128-1),
		},
		{
			name: "unsupported key length",
			key:  make([]byte, aesKeySize192+1),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			got, err := Key(test.key, nil)
			if !errors.Is(err, ErrWrongKeyLen) {
				t.Fatalf("Key() err = %v, want %v", err, ErrWrongKeyLen)
			}

			if got != nil {
				t.Fatalf("Key() = %x, want nil", got)
			}
		})
	}
}

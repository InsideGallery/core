# pki/mock_cipher

Import path: `github.com/InsideGallery/core/pki/mock_cipher`

`mock_cipher` is a generated GoMock package for the legacy `pki.Cipher`
interface. It is generated from `pki/cipher.go` and is intended for tests that
need to set expectations on cipher behavior.

## Main API

- `NewMockCipher(ctrl)` creates a mock cipher.
- `MockCipher` implements the `pki.Cipher` methods: `Encrypt`, `Decrypt`,
  `Kind`, `ToBinary`, and `FromBinary`.
- `MockCipher.EXPECT()` returns the recorder used to define GoMock
  expectations.
- `MockCipherMockRecorder` provides expectation methods for each mocked method.

## Usage

```go
package example_test

import (
	"testing"

	"github.com/InsideGallery/core/pki/mock_cipher"
	"go.uber.org/mock/gomock"
)

func TestCipherKind(t *testing.T) {
	ctrl := gomock.NewController(t)
	cipher := mock_cipher.NewMockCipher(ctrl)

	cipher.EXPECT().Kind().Return("test")

	if got := cipher.Kind(); got != "test" {
		t.Fatalf("Kind() = %q, want test", got)
	}
}
```

## Compatibility Notes

Do not edit `cipher.go` in this package by hand. Regenerate it from
`pki/cipher.go` when the `pki.Cipher` interface changes.

# pki

Import path: `github.com/InsideGallery/core/pki`

`pki` is the legacy import path for the core cipher contract. New code should
prefer `github.com/InsideGallery/core/pki/cryptor`, which exposes the same
contract through a focused package name.

The Go package name is `cipher`, so callers commonly import this path with an
alias to avoid confusion with `crypto/cipher`.

## Main API

- `Cipher` is the shared interface implemented by the in-tree cipher packages.
  It requires `Encrypt`, `Decrypt`, `Kind`, `ToBinary`, and `FromBinary`.
- `Options` carries a requested cipher `Kind`.
- `Result` returns a cipher `Kind` and operation `Data` without exposing
  implementation-specific result types.
- `Encrypt(ctx, cipher, plaintext)` checks for a nil cipher, checks
  `ctx.Err()`, encrypts with the cipher, and wraps the result.
- `Decrypt(ctx, cipher, ciphertext)` performs the same boundary checks for
  decryption.
- `ErrCipherNotSet` is returned, wrapped with context, when the cipher argument
  is nil.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/pki/aesgcm"
	legacycipher "github.com/InsideGallery/core/pki"
)

func roundTrip(ctx context.Context, plaintext []byte) ([]byte, error) {
	cipher, err := aesgcm.New(aesgcm.KeySize32)
	if err != nil {
		return nil, err
	}

	encrypted, err := legacycipher.Encrypt(ctx, cipher, plaintext)
	if err != nil {
		return nil, err
	}

	decrypted, err := legacycipher.Decrypt(ctx, cipher, encrypted.Data)
	if err != nil {
		return nil, err
	}

	return decrypted.Data, nil
}
```

## Compatibility Notes

This path remains available for downstream consumers. Add new cipher boundary
helpers to `pki/cryptor` instead of this package so new call sites avoid a local
name collision with Go's standard `crypto/cipher` package.

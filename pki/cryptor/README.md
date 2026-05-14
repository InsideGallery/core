# pki/cryptor

Import path: `github.com/InsideGallery/core/pki/cryptor`

`cryptor` is the preferred package for the core cipher contract and operation
boundary helpers. It delegates to the legacy `pki` path for compatibility while
providing a clearer import path for new code.

## Main API

- `Cipher` is the shared interface implemented by core cipher packages.
- `Options` carries a requested cipher `Kind`.
- `Result` returns a cipher `Kind` and operation `Data`.
- `Encrypt(ctx, cipher, plaintext)` checks for nil cipher dependencies,
  respects a canceled context before work starts, and returns encrypted data.
- `Decrypt(ctx, cipher, ciphertext)` applies the same checks before decryption.
- `ErrCipherNotSet` reports a nil cipher argument.

## Usage

```go
package example

import (
	"context"

	"github.com/InsideGallery/core/pki/aesgcm"
	"github.com/InsideGallery/core/pki/cryptor"
)

func encrypt(ctx context.Context, plaintext []byte) (cryptor.Result, error) {
	cipher, err := aesgcm.New(aesgcm.KeySize32)
	if err != nil {
		return cryptor.Result{}, err
	}

	return cryptor.Encrypt(ctx, cipher, plaintext)
}
```

## Compatibility Notes

This package mirrors the legacy `github.com/InsideGallery/core/pki` API. The
legacy path remains available, but new call sites should use `pki/cryptor`.

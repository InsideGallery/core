# pki/aesgcm

Import path: `github.com/InsideGallery/core/pki/aesgcm`

`aesgcm` is the preferred AES-GCM import path. It re-exports the legacy
`pki/aes` implementation with names that avoid a local collision with Go's
standard `crypto/aes` package.

## Main API

- `Cipher` is an alias for the legacy AES-GCM cipher type.
- `New(size)` creates a cipher with a random key.
- `KeySize16`, `KeySize24`, and `KeySize32` are supported key sizes in bytes.
- `ErrInvalidKeySize` reports unsupported key sizes.
- `Cipher.Encrypt`, `Cipher.Decrypt`, `Cipher.Kind`, `Cipher.ToBinary`, and
  `Cipher.FromBinary` are inherited from `pki/aes`.

## Usage

```go
package example

import "github.com/InsideGallery/core/pki/aesgcm"

func roundTrip(plaintext []byte) ([]byte, error) {
	cipher, err := aesgcm.New(aesgcm.KeySize32)
	if err != nil {
		return nil, err
	}

	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return nil, err
	}

	return cipher.Decrypt(ciphertext)
}
```

## Security and Compatibility Notes

Encryption uses AES-GCM with a random nonce, prefixes that nonce to the
ciphertext, and hex-encodes the combined bytes. The implementation does not use
additional authenticated data. `github.com/InsideGallery/core/pki/aes` remains
available for existing consumers, but new code should import this package.

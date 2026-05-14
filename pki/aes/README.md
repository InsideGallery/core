# pki/aes

Import path: `github.com/InsideGallery/core/pki/aes`

`pki/aes` is the legacy AES-GCM cipher package. New code should prefer
`github.com/InsideGallery/core/pki/aesgcm`, which re-exports this behavior under
an algorithm-specific package name that does not collide with `crypto/aes`.

## Main API

- `AES` implements the `pki.Cipher` interface.
- `NewAES(size)` and `NewAESStrict(size)` create a cipher with a random key.
- `AES16`, `AES24`, and `AES32` are the supported key sizes in bytes.
- `ErrInvalidAESSize` reports unsupported key sizes.
- `Kind()` returns `"aes"`.
- `Encrypt(plaintext)` encrypts with AES-GCM, prefixes a random nonce to the
  ciphertext, and returns the nonce plus ciphertext as hex-encoded bytes.
- `Decrypt(ciphertext)` expects the hex-encoded format returned by `Encrypt`.
- `ToBinary()` returns the raw key bytes.
- `FromBinary(raw)` restores a cipher from raw key bytes.

## Usage

```go
package example

import coreaes "github.com/InsideGallery/core/pki/aes"

func encrypt(plaintext []byte) ([]byte, []byte, error) {
	cipher, err := coreaes.NewAES(coreaes.AES32)
	if err != nil {
		return nil, nil, err
	}

	key, err := cipher.ToBinary()
	if err != nil {
		return nil, nil, err
	}

	ciphertext, err := cipher.Encrypt(plaintext)
	if err != nil {
		return nil, nil, err
	}

	return ciphertext, key, nil
}
```

## Security and Compatibility Notes

AES-GCM encryption uses a random nonce and no additional authenticated data.
`ToBinary` exposes raw key material; store it only in a suitable secret store.
`FromBinary` accepts any byte slice and leaves invalid key sizes to fail during
encrypt or decrypt calls. This package remains for compatibility; prefer
`pki/aesgcm` for new code.

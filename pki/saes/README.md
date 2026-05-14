# pki/saes

Import path: `github.com/InsideGallery/core/pki/saes`

`saes` provides deterministic authenticated encryption with AES-SIV through
Tink's `subtle.NewAESSIV`. It implements the shared `pki.Cipher` contract with
a 64-byte AES-SIV key.

## Main API

- `SAES` implements the `pki.Cipher` interface.
- `NewSAES()` creates a cipher with a random 64-byte key.
- `AESSIV64` is the required key size in bytes.
- `Kind` is the string returned by `SAES.Kind()`.
- `Encrypt(origin)` calls deterministic AES-SIV encryption with nil associated
  data.
- `Decrypt(ciphertext)` decrypts deterministic AES-SIV ciphertext with nil
  associated data.
- `ToBinary()` returns raw key bytes.
- `FromBinary(raw)` restores an `SAES` cipher from raw key bytes.
- `ErrEncryptedDataIsEmpty` reports an empty ciphertext passed to `Decrypt`.
- `ErrEncryptedDataIsWrongLen` is exported for compatibility, but the current
  implementation does not return it.

## Usage

```go
package example

import "github.com/InsideGallery/core/pki/saes"

func seal(value []byte) ([]byte, []byte, error) {
	cipher, err := saes.NewSAES()
	if err != nil {
		return nil, nil, err
	}

	key, err := cipher.ToBinary()
	if err != nil {
		return nil, nil, err
	}

	ciphertext, err := cipher.Encrypt(value)
	if err != nil {
		return nil, nil, err
	}

	return ciphertext, key, nil
}
```

## Security Notes

AES-SIV encryption here is deterministic: the same key and plaintext produce
the same ciphertext. That is useful for stable encrypted identifiers, but it
does not hide equality of repeated plaintext values. `ToBinary` exposes raw key
material; store it only in a suitable secret store.

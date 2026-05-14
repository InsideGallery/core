# pki/rsa

Import path: `github.com/InsideGallery/core/pki/rsa`

`pki/rsa` is the legacy RSA-OAEP cipher package. New code should prefer
`github.com/InsideGallery/core/pki/rsaoaep`, which exposes the same behavior
without colliding with Go's standard `crypto/rsa` package name.

## Main API

- `RSA` implements the `pki.Cipher` interface.
- `NewRSA(bits)` generates a new RSA private key and matching public key.
- `DefaultBitsSize` is the default key size, currently 4096 bits.
- `FromPrivateKey(data)` restores a cipher from PKCS#1 private-key PEM data.
- `TypeRSAPrivateKey` is the PEM block type used by `ToBinary`.
- `ErrFailedToParsePEMBlock` reports data that is not a private-key PEM block.
- `Kind()` returns `"rsa"`.
- `Encrypt(data)` uses RSA-OAEP with SHA-256 and a nil label.
- `Decrypt(data)` decrypts the RSA-OAEP ciphertext with the private key.
- `ToBinary()` returns PKCS#1 private-key PEM bytes.
- `FromBinary(raw)` restores a cipher through `FromPrivateKey`.

## Usage

```go
package example

import corersa "github.com/InsideGallery/core/pki/rsa"

func roundTrip(plaintext []byte) ([]byte, error) {
	cipher, err := corersa.NewRSA(corersa.DefaultBitsSize)
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

RSA-OAEP can encrypt only short messages relative to the key size and hash
overhead. Use it for small payloads such as wrapped keys, not large data
streams. `ToBinary` exposes private-key material; store it only in a suitable
secret store. This legacy path remains available for existing consumers; prefer
`pki/rsaoaep` for new code.

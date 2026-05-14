# pki/rsaoaep

Import path: `github.com/InsideGallery/core/pki/rsaoaep`

`rsaoaep` is the preferred RSA-OAEP import path. It re-exports the legacy
`pki/rsa` implementation with names that avoid a local collision with Go's
standard `crypto/rsa` package.

## Main API

- `Cipher` is an alias for the legacy RSA-OAEP cipher type.
- `New(bits)` generates a new RSA private key and matching public key.
- `FromPrivateKey(data)` restores a cipher from PKCS#1 private-key PEM data.
- `DefaultKeyBits` is the default key size, currently 4096 bits.
- `PrivateKeyPEMBlockType` is the PEM block type used by `ToBinary`.
- `ErrFailedToParsePEMBlock` reports data that is not a private-key PEM block.
- `Cipher.Encrypt`, `Cipher.Decrypt`, `Cipher.Kind`, `Cipher.ToBinary`, and
  `Cipher.FromBinary` are inherited from `pki/rsa`.

## Usage

```go
package example

import "github.com/InsideGallery/core/pki/rsaoaep"

func restoreAndDecrypt(privateKeyPEM, ciphertext []byte) ([]byte, error) {
	cipher, err := rsaoaep.FromPrivateKey(privateKeyPEM)
	if err != nil {
		return nil, err
	}

	return cipher.Decrypt(ciphertext)
}
```

## Security and Compatibility Notes

Encryption uses RSA-OAEP with SHA-256 and a nil label. RSA-OAEP is suited to
short payloads such as wrapped keys. `github.com/InsideGallery/core/pki/rsa`
remains available for existing consumers, but new code should import this
package.

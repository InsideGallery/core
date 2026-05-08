// Package rsaoaep provides RSA-OAEP cipher helpers without colliding with crypto/rsa.
//
// New code should import this package instead of the legacy pki/rsa path:
//
//	import "github.com/InsideGallery/core/pki/rsaoaep"
//
// Compatibility: github.com/InsideGallery/core/pki/rsa remains available for
// existing consumers. Prefer New and FromPrivateKey from rsaoaep so RSA-OAEP
// usage has an algorithm-specific import path.
package rsaoaep

import legacy "github.com/InsideGallery/core/pki/rsa"

const (
	// PrivateKeyPEMBlockType is the PEM block type used for PKCS#1 private keys.
	PrivateKeyPEMBlockType = legacy.TypeRSAPrivateKey
	// DefaultKeyBits is the default RSA key size.
	DefaultKeyBits = legacy.DefaultBitsSize
)

// ErrFailedToParsePEMBlock reports invalid private-key PEM data.
var ErrFailedToParsePEMBlock = legacy.ErrFailedToParsePEMBlock

// Cipher encrypts and decrypts data with RSA-OAEP.
type Cipher = legacy.RSA

// New returns an RSA-OAEP cipher with a randomly generated key.
func New(bits int) (*Cipher, error) {
	return legacy.NewRSA(bits)
}

// FromPrivateKey restores an RSA-OAEP cipher from PKCS#1 private-key PEM data.
func FromPrivateKey(data []byte) (*Cipher, error) {
	return legacy.FromPrivateKey(data)
}

// Package aesgcm provides AES-GCM cipher helpers without colliding with crypto/aes.
//
// New code should import this package instead of the legacy pki/aes path:
//
//	import "github.com/InsideGallery/core/pki/aesgcm"
//
// Compatibility: github.com/InsideGallery/core/pki/aes remains available for
// existing consumers. Prefer New and KeySize constants from aesgcm so AES-GCM
// usage has an algorithm-specific import path.
package aesgcm

import legacy "github.com/InsideGallery/core/pki/aes"

const (
	// KeySize32 is the 256-bit AES key size in bytes.
	KeySize32 = legacy.AES32
	// KeySize24 is the 192-bit AES key size in bytes.
	KeySize24 = legacy.AES24
	// KeySize16 is the 128-bit AES key size in bytes.
	KeySize16 = legacy.AES16
)

// ErrInvalidKeySize reports an unsupported AES key size.
var ErrInvalidKeySize = legacy.ErrInvalidAESSize

// Cipher encrypts and decrypts data with AES-GCM.
type Cipher = legacy.AES

// New returns an AES-GCM cipher with a randomly generated key.
func New(size int) (*Cipher, error) {
	return legacy.NewAES(size)
}

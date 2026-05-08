// Package cryptor defines core-owned cipher contracts and boundary helpers.
//
// New code should import this package instead of the legacy pki path:
//
//	import "github.com/InsideGallery/core/pki/cryptor"
//
// Compatibility: github.com/InsideGallery/core/pki remains available for
// existing consumers. Prefer Options, Result, Cipher, Encrypt, and Decrypt from
// this package so cipher contracts stay behind a focused import path.
package cryptor

import (
	"context"

	legacy "github.com/InsideGallery/core/pki"
)

// ErrCipherNotSet reports a nil cipher dependency.
var ErrCipherNotSet = legacy.ErrCipherNotSet

// Options identifies the cipher behavior requested by a consumer.
type Options = legacy.Options

// Result is the core-owned result shape for cipher operations.
type Result = legacy.Result

// Cipher is the contract implemented by core-owned cipher implementations.
type Cipher = legacy.Cipher

// Encrypt encrypts plaintext through a Cipher.
func Encrypt(ctx context.Context, cipher Cipher, plaintext []byte) (Result, error) {
	return legacy.Encrypt(ctx, cipher, plaintext)
}

// Decrypt decrypts ciphertext through a Cipher.
func Decrypt(ctx context.Context, cipher Cipher, ciphertext []byte) (Result, error) {
	return legacy.Decrypt(ctx, cipher, ciphertext)
}

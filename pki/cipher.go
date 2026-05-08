//go:generate mockgen -source=cipher.go -destination=mock_cipher/cipher.go

// Package cipher is the legacy import path for core-owned cipher contracts.
//
// New code should import the focused replacement package:
//
//	import "github.com/InsideGallery/core/pki/cryptor"
//
// Compatibility: existing cipher contracts remain available for downstream
// consumers that still import pki. Do not add new cipher boundary helpers here;
// add them to pki/cryptor so call sites avoid a local name collision with
// crypto/cipher.
package cipher //nolint:revive

import (
	"context"
	"errors"
	"fmt"
)

// ErrCipherNotSet reports a nil cipher dependency.
var ErrCipherNotSet = errors.New("cipher is not set")

// Options identifies the cipher behavior requested by a consumer.
type Options struct {
	Kind string
}

// Result is the core-owned result shape for cipher operations.
type Result struct {
	Kind string
	Data []byte
}

type Cipher interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)

	Kind() string
	ToBinary() ([]byte, error)
	FromBinary([]byte) (Cipher, error)
}

// Encrypt encrypts plaintext through a Cipher without exposing implementation-specific result types.
func Encrypt(ctx context.Context, c Cipher, plaintext []byte) (Result, error) {
	if c == nil {
		return Result{}, fmt.Errorf("cipher encrypt: %w", ErrCipherNotSet)
	}

	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("cipher encrypt: %w", err)
	}

	data, err := c.Encrypt(plaintext)
	if err != nil {
		return Result{}, fmt.Errorf("cipher encrypt: %w", err)
	}

	return Result{Kind: c.Kind(), Data: data}, nil
}

// Decrypt decrypts ciphertext through a Cipher without exposing implementation-specific result types.
func Decrypt(ctx context.Context, c Cipher, ciphertext []byte) (Result, error) {
	if c == nil {
		return Result{}, fmt.Errorf("cipher decrypt: %w", ErrCipherNotSet)
	}

	if err := ctx.Err(); err != nil {
		return Result{}, fmt.Errorf("cipher decrypt: %w", err)
	}

	data, err := c.Decrypt(ciphertext)
	if err != nil {
		return Result{}, fmt.Errorf("cipher decrypt: %w", err)
	}

	return Result{Kind: c.Kind(), Data: data}, nil
}

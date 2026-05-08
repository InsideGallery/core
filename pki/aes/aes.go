// Package aes is the legacy AES-GCM import path.
//
// New code should import the focused replacement package:
//
//	import "github.com/InsideGallery/core/pki/aesgcm"
//
// Compatibility: existing AES-GCM exports remain available for downstream
// consumers that still import pki/aes. Do not add new helpers here; add AES-GCM
// behavior to pki/aesgcm so call sites avoid a local name collision with
// crypto/aes.
package aes //nolint:revive

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"

	localCipher "github.com/InsideGallery/core/pki"
)

const (
	AES32 = 32
	AES24 = 24
	AES16 = 16

	kind = "aes"
)

type AES struct {
	key []byte
}

var ErrInvalidAESSize = errors.New("invalid AES key size, must be 16, 24, or 32")

func NewAES(size int) (*AES, error) {
	return NewAESStrict(size)
}

func NewAESStrict(size int) (*AES, error) {
	switch size {
	case AES16, AES24, AES32:
	default:
		return nil, ErrInvalidAESSize
	}

	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return &AES{key: bytes}, nil
}

func (a *AES) Kind() string {
	return kind
}

func (a *AES) ToBinary() ([]byte, error) {
	return a.key, nil
}

func (a *AES) FromBinary(raw []byte) (localCipher.Cipher, error) { // nolint:ireturn
	return &AES{key: raw}, nil
}

func (a *AES) Encrypt(origin []byte) ([]byte, error) {
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	// Encrypt the data using aesGCM.Seal
	// Since we don't want to save the nonce somewhere else in this case,
	// we add it as a prefix to the encrypted data.
	// The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, origin, nil)

	enc := make([]byte, hex.EncodedLen(len(ciphertext)))

	hex.Encode(enc, ciphertext)

	return enc, nil
}

func (a *AES) Decrypt(encrypted []byte) ([]byte, error) {
	enc := make([]byte, hex.DecodedLen(len(encrypted)))

	_, err := hex.Decode(enc, encrypted)
	if err != nil {
		return nil, err
	}

	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return nil, err
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Get the nonce size
	nonceSize := aesGCM.NonceSize()

	// Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

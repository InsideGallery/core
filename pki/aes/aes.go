package aes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
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

func NewAES(size int) (*AES, error) {
	switch size {
	case AES16, AES24, AES32:
	default:
		size = AES32
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

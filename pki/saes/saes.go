package saes

import (
	"crypto/rand"

	"github.com/tink-crypto/tink-go/v2/daead/subtle"

	localCipher "github.com/InsideGallery/core/pki"
)

const (
	// AESSIV64 is byte size of SIV SAES Key which must be twice as long as AES32
	AESSIV64 = 64

	Kind = "saes"
)

type SAES struct {
	key []byte
}

func NewSAES() (*SAES, error) {
	bytes := make([]byte, AESSIV64)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}

	return &SAES{key: bytes}, nil
}

func (a *SAES) Kind() string {
	return Kind
}

func (a *SAES) ToBinary() ([]byte, error) {
	return a.key, nil
}

func (a *SAES) FromBinary(raw []byte) (localCipher.Cipher, error) {
	return &SAES{key: raw}, nil
}

func (a *SAES) Encrypt(origin []byte) ([]byte, error) {
	aessiv, err := subtle.NewAESSIV(a.key)
	if err != nil {
		return nil, err
	}

	ciphertext, err := aessiv.EncryptDeterministically(origin, nil)
	if err != nil {
		return nil, err
	}

	return ciphertext, err
}

func (a *SAES) Decrypt(ciphertext []byte) ([]byte, error) {
	if len(ciphertext) == 0 {
		return nil, ErrEncryptedDataIsEmpty
	}

	aessiv, err := subtle.NewAESSIV(a.key)
	if err != nil {
		return nil, err
	}

	plaintext, err := aessiv.DecryptDeterministically(ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

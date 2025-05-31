package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"

	localCipher "github.com/InsideGallery/core/pki"
)

const (
	TypeRSAPrivateKey = "RSA PRIVATE KEY"

	DefaultBitsSize = 4096

	kind = "rsa"
)

type RSA struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewRSA(bits int) (*RSA, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	return &RSA{
		privateKey: privateKey,
		publicKey:  &privateKey.PublicKey,
	}, nil
}

func FromPrivateKey(b []byte) (*RSA, error) {
	block, _ := pem.Decode(b)
	if block == nil {
		return nil, ErrFailedToParsePEMBlock
	}

	pkey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &RSA{
		privateKey: pkey,
		publicKey:  &pkey.PublicKey,
	}, nil
}

func (a *RSA) Kind() string {
	return kind
}

func (a *RSA) ToBinary() ([]byte, error) {
	raw := pem.EncodeToMemory(
		&pem.Block{
			Type:  TypeRSAPrivateKey,
			Bytes: x509.MarshalPKCS1PrivateKey(a.privateKey),
		},
	)

	return raw, nil
}

func (a *RSA) FromBinary(raw []byte) (localCipher.Cipher, error) { // nolint:ireturn
	return FromPrivateKey(raw)
}

func (a *RSA) Encrypt(data []byte) ([]byte, error) {
	rng := rand.Reader

	return rsa.EncryptOAEP(sha256.New(), rng, a.publicKey, data, nil)
}

func (a *RSA) Decrypt(data []byte) ([]byte, error) {
	rng := rand.Reader
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rng, a.privateKey, data, nil)

	return plaintext, err
}

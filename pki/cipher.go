//go:generate mockgen -source=cipher.go -destination=mock_cipher/cipher.go
package cipher

type Cipher interface {
	Encrypt([]byte) ([]byte, error)
	Decrypt([]byte) ([]byte, error)

	Kind() string
	ToBinary() ([]byte, error)
	FromBinary([]byte) (Cipher, error)
}

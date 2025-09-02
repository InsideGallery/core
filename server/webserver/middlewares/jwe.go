package middlewares

import (
	"crypto/sha256"
	"io"
	"net/http"

	"golang.org/x/crypto/hkdf"

	"github.com/go-jose/go-jose/v3"
	"github.com/gofiber/fiber/v2"
)

const (
	DecryptValueKey  = "decrypted_body"
	ResponseValueKey = "response_body"

	HeaderJOSE = "application/jose"
)

type JWE struct {
	decryptionKeyGetter func(c *fiber.Ctx) ([]byte, error)
}

func NewJWE(decryptionKeyGetter func(c *fiber.Ctx) ([]byte, error)) *JWE {
	return &JWE{
		decryptionKeyGetter: decryptionKeyGetter,
	}
}

func (j *JWE) DecryptMiddleware(c *fiber.Ctx) error {
	// 1. Отримуємо тіло запиту (очікуємо рядок JWE)
	jweString := string(c.Body())
	if jweString == "" {
		return c.Next()
	}

	decryptionKey, err := j.decryptionKeyGetter(c)
	if err != nil {
		return err
	}

	// 2. Парсимо JWE
	jweObject, err := jose.ParseEncrypted(jweString)
	if err != nil {
		return err
	}

	// 3. Розшифровуємо за допомогою приватного ключа сервера
	decryptedPayload, err := jweObject.Decrypt(decryptionKey)
	if err != nil {
		return err
	}

	c.Locals(DecryptValueKey, decryptedPayload)

	err = c.Next()
	if err != nil {
		return err
	}

	resp, ok := c.Locals(ResponseValueKey).([]byte)
	if ok && len(resp) != 0 {
		result, err := EncryptResponse(decryptionKey, resp)
		if err != nil {
			return err
		}

		c.Status(http.StatusOK)
		c.Set(fiber.HeaderContentType, HeaderJOSE)

		_, err = c.WriteString(result)
		return err
	}

	return nil
}

func EncryptResponse(sesKey, payload []byte) (string, error) {
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.DIRECT, Key: sesKey},
		nil,
	)
	if err != nil {
		return "", err
	}

	jweObject, err := encrypter.Encrypt(payload)
	if err != nil {
		return "", err
	}

	return jweObject.CompactSerialize()
}

func GetSessionKey(masterSecret []byte, nonce []byte) ([]byte, error) {
	kdf := hkdf.New(sha256.New, masterSecret, nonce, nil)

	sessionKey := make([]byte, 32)

	_, err := io.ReadFull(kdf, sessionKey)

	return sessionKey, err
}

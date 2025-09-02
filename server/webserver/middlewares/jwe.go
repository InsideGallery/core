package middlewares

import (
	"net/http"

	"github.com/go-jose/go-jose/v3"
	"github.com/gofiber/fiber/v2"
)

const (
	DecryptValueKey  = "decrypted_body"
	ResponseValueKey = "response_body"

	HeaderJOSE = "application/jose"
)

type JWE struct {
	privateKey []byte
	publicKey  []byte
}

func NewJWE(privateKey, publicKey []byte) *JWE {
	return &JWE{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

func (j *JWE) DecryptMiddleware(c *fiber.Ctx) error {
	// 1. Отримуємо тіло запиту (очікуємо рядок JWE)
	jweString := string(c.Body())
	if jweString == "" {
		return c.Next()
	}

	// 2. Парсимо JWE
	jweObject, err := jose.ParseEncrypted(jweString)
	if err != nil {
		return err
	}

	// 3. Розшифровуємо за допомогою приватного ключа сервера
	decryptedPayload, err := jweObject.Decrypt(j.privateKey)
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
		result, err := EncryptResponse(j.publicKey, resp)
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

func EncryptResponse(publicKey, payload []byte) (string, error) {
	encrypter, err := jose.NewEncrypter(
		jose.A256GCM,
		jose.Recipient{Algorithm: jose.RSA_OAEP, Key: publicKey},
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

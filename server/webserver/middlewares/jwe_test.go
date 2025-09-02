package middlewares

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/InsideGallery/core/pki/aes"
	"github.com/InsideGallery/core/testutils"
	"github.com/go-jose/go-jose/v3"
	"github.com/gofiber/fiber/v2"
)

func TestJWEAES(t *testing.T) {
	masterKey, err := aes.NewAES(32)
	testutils.Equal(t, err, nil)

	rawMasterKey, err := masterKey.ToBinary()
	testutils.Equal(t, err, nil)

	sharedSecretKey, err := GetSessionKey(rawMasterKey, []byte("key"))
	testutils.Equal(t, err, nil)

	requestStr := "hello world!"
	responseStr := "good, thanks!"

	raw, err := EncryptResponse(sharedSecretKey, []byte(requestStr))
	testutils.Equal(t, err, nil)

	j := NewJWE(func(_ *fiber.Ctx) ([]byte, error) {
		return GetSessionKey(rawMasterKey, []byte("key"))
	})

	app := fiber.New()
	app.Use(j.DecryptMiddleware)
	app.Post("/", func(ctx *fiber.Ctx) error {
		data := ctx.Locals(DecryptValueKey).([]byte)
		testutils.Equal(t, string(data), requestStr)

		ctx.Locals(ResponseValueKey, []byte("good, thanks!"))
		return nil
	})

	req, _ := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(raw)))
	res, err := app.Test(req, -1)
	testutils.Equal(t, err, nil)

	defer res.Body.Close()

	result, err := io.ReadAll(res.Body)
	testutils.Equal(t, err, nil)

	parsedJWE, err := jose.ParseEncrypted(string(result))
	testutils.Equal(t, err, nil)

	decryptedResponse, err := parsedJWE.Decrypt(sharedSecretKey)
	testutils.Equal(t, err, nil)

	testutils.Equal(t, string(decryptedResponse), responseStr)
}

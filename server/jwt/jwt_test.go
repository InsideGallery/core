//go:build unit
// +build unit

package jwt

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/InsideGallery/core/server/jwt/model"
	"github.com/InsideGallery/core/testutils"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	privateKey, err := os.ReadFile("test-data/test-jwt.key")
	testutils.Equal(t, err, nil)

	publicKey, err := os.ReadFile("test-data/test-jwt.pem")
	testutils.Equal(t, err, nil)

	jwtService, err := NewJWT(privateKey, publicKey)
	testutils.Equal(t, err, nil)

	scope, err := model.ScopeFrom("read:service:action")
	testutils.Equal(t, err, nil)

	accessString, refreshToken, err := jwtService.Generate(Payload{
		UserID: "F12E0A1B-8303-4A23-B4BF-CEC4797D95EC",
		Scopes: model.Scopes{scope},
	})
	testutils.Equal(t, err, nil)
	testutils.Equal(t, accessString != "", true)
	testutils.Equal(t, refreshToken != "", true)

	webServer := fiber.New()
	webServer.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtService.GetSigningKey(),
	}))
	webServer.Get("/test", func(c fiber.Ctx) error {
		jwtToken := jwtware.FromContext(c)
		assert.True(t, jwtToken.Valid)

		return nil
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Add("Authorization", "Bearer "+accessString)

	resp, err := webServer.Test(req, fiber.TestConfig{Timeout: 0})
	testutils.Equal(t, err, nil)
	testutils.Equal(t, resp.StatusCode, http.StatusOK)
}

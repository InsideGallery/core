package jwt

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestNewJWTErrors(t *testing.T) {
	privateKey, err := os.ReadFile("test-data/test-jwt.key")
	if err != nil {
		t.Fatalf("read private key: %v", err)
	}

	cases := []struct {
		name       string
		privateKey []byte
		publicKey  []byte
	}{
		{
			name:       "invalid private key",
			privateKey: []byte("bad"),
			publicKey:  []byte("bad"),
		},
		{
			name:       "invalid public key",
			privateKey: privateKey,
			publicKey:  []byte("bad"),
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			service, err := NewJWT(test.privateKey, test.publicKey)
			if err == nil {
				t.Fatal("expected error")
			}

			if service != nil {
				t.Fatal("service should be nil")
			}
		})
	}
}

func TestDecodeClaimsWithoutToken(t *testing.T) {
	cases := []struct {
		name string
	}{
		{name: "missing context token"},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/", func(c fiber.Ctx) error {
				payload, err := DecodeClaims(c)
				if !errors.Is(err, ErrJWTTokenNotFound) {
					t.Fatalf("err = %v, want %v", err, ErrJWTTokenNotFound)
				}

				if payload != nil {
					t.Fatal("payload should be nil")
				}

				return c.SendStatus(http.StatusOK)
			})

			resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/", nil))
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			defer resp.Body.Close()
		})
	}
}

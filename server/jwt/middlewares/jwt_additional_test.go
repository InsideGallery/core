package middlewares

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gofiber/fiber/v3"

	coreJWT "github.com/InsideGallery/core/server/jwt"
	jwtModel "github.com/InsideGallery/core/server/jwt/model"
)

func TestJWTMiddleware(t *testing.T) {
	jwtService := newTestJWTService(t)
	scope := jwtModel.Scope{
		AccessType: jwtModel.AccessTypeRead,
		Service:    "gallery",
		Action:     "view",
	}
	token, _, err := jwtService.Generate(coreJWT.Payload{
		UserID:   "user-1",
		OrgID:    "org-1",
		Role:     jwtModel.UserRoleManager,
		OrgSlug:  "org",
		OrgName:  "Org",
		UserName: "Ada",
		Scopes:   jwtModel.Scopes{scope},
	})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	cases := []struct {
		name       string
		auth       string
		useScope   bool
		wantStatus int
	}{
		{
			name:       "missing token returns bad request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid token returns unauthorized",
			auth:       "Bearer bad",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token passes jwt middleware",
			auth:       "Bearer " + token,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid token passes scope middleware",
			auth:       "Bearer " + token,
			useScope:   true,
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(NewJWT(jwtService))
			if test.useScope {
				app.Use(NewScopeMiddleware(context.Background()).CheckScope)
			}
			app.Get("/gallery/view", func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/gallery/view", nil)
			if test.auth != "" {
				req.Header.Set("Authorization", test.auth)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}
		})
	}
}

func TestScopeMiddlewareResponses(t *testing.T) {
	jwtService := newTestJWTService(t)

	deniedToken, _, err := jwtService.Generate(coreJWT.Payload{
		UserID:   "user-1",
		OrgID:    "org-1",
		Role:     jwtModel.UserRoleManager,
		OrgSlug:  "org",
		OrgName:  "Org",
		UserName: "Ada",
	})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	cases := []struct {
		name       string
		auth       string
		withJWT    bool
		wantStatus int
	}{
		{
			name:       "missing decoded claims returns bad request",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "denied scope returns forbidden",
			auth:       "Bearer " + deniedToken,
			withJWT:    true,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			if test.withJWT {
				app.Use(NewJWT(jwtService))
			}
			app.Use(NewScopeMiddleware(context.Background()).CheckScope)
			app.Get("/gallery/view", func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/gallery/view", nil)
			if test.auth != "" {
				req.Header.Set("Authorization", test.auth)
			}

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("app test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}
		})
	}
}

func newTestJWTService(t *testing.T) *coreJWT.Service {
	t.Helper()

	privateKey, err := os.ReadFile("../test-data/test-jwt.key")
	if err != nil {
		t.Fatalf("read private key: %v", err)
	}

	publicKey, err := os.ReadFile("../test-data/test-jwt.pem")
	if err != nil {
		t.Fatalf("read public key: %v", err)
	}

	jwtService, err := coreJWT.NewJWT(privateKey, publicKey)
	if err != nil {
		t.Fatalf("new jwt service: %v", err)
	}

	return jwtService
}

package middlewares

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/FrogoAI/testutils"
	"github.com/gofiber/fiber/v3"
	jwtlib "github.com/golang-jwt/jwt/v5"

	coreJWT "github.com/InsideGallery/core/server/jwt"
	jwtModel "github.com/InsideGallery/core/server/jwt/model"
)

const trackedJWTMiddlewareRSAKeyBits = 2048

func Test_parseScope(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		target         string
		requestURI     string
		want           string
		wantStatusCode int
	}{
		{
			name:           "read GET",
			method:         http.MethodGet,
			target:         "/organizations",
			want:           "read:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET",
			method:         http.MethodGet,
			target:         "/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET",
			method:         http.MethodGet,
			target:         "/v1/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read OPTION",
			method:         http.MethodOptions,
			target:         "/organizations",
			want:           "read:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read OPTION",
			method:         http.MethodOptions,
			target:         "/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read HEAD",
			method:         http.MethodHead,
			target:         "/organizations",
			want:           "read:organizations:",
			wantStatusCode: http.StatusOK,
		},
		// todo: потрібно прописати ендпоінти для ролбеку обʼєктів
		{
			name:           "read HEAD",
			method:         http.MethodHead,
			target:         "/organizations/some-id",
			want:           "read:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write POST",
			method:         http.MethodPost,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write POST",
			method:         http.MethodPost,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT",
			method:         http.MethodPut,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT",
			method:         http.MethodPut,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PATCH",
			method:         http.MethodPatch,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PATCH",
			method:         http.MethodPatch,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write DELETE",
			method:         http.MethodDelete,
			target:         "/organizations",
			want:           "write:organizations:",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write DELETE",
			method:         http.MethodDelete,
			target:         "/organizations/some-id",
			want:           "write:organizations:some-id",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET with query params",
			method:         http.MethodGet,
			target:         "/v1/policies/history",
			requestURI:     "/v1/policies/history?page=1&per-page=100",
			want:           "read:policies:history",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "read GET entity by ID",
			method:         http.MethodGet,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "read:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT entity by ID",
			method:         http.MethodPut,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "write:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PATCH entity by ID",
			method:         http.MethodPatch,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "write:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
		{
			name:           "write PUT entity by ID",
			method:         http.MethodDelete,
			target:         "/v1/policies/history/:id",
			requestURI:     "/v1/policies/history/64ac106b7cb02d8d34948708",
			want:           "write:policies:history/64ac106b7cb02d8d34948708",
			wantStatusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var triggered bool

			webServer := fiber.New()
			webServer.All(tt.target, func(c fiber.Ctx) error {
				triggered = true

				if got := parseScope(c); got != tt.want {
					t.Errorf("parseScope() = %v, want %v", got, tt.want)
				}

				return nil
			})

			var req *http.Request

			if tt.requestURI == "" {
				req = httptest.NewRequest(tt.method, tt.target, nil)
			} else {
				req = httptest.NewRequest(tt.method, tt.requestURI, nil)
			}

			resp, err := webServer.Test(req, fiber.TestConfig{Timeout: 0})
			testutils.Equal(t, err, nil)
			testutils.Equal(t, triggered, true)
			testutils.Equal(t, resp.StatusCode, tt.wantStatusCode)
		})
	}
}

func TestJWTMiddlewareBoundaryResponses(t *testing.T) {
	jwtService, privateKey := newTrackedJWTMiddlewareService(t)
	validToken, _, err := jwtService.Generate(coreJWT.Payload{
		UserID: "user-1",
		Role:   jwtModel.UserRoleManager,
	})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	expiredToken := signTrackedJWTMiddlewarePayload(t, privateKey, time.Now().Add(-time.Minute))
	unsignedToken := signUnsignedTrackedJWTMiddlewarePayload(t, time.Now().Add(time.Hour))

	cases := []struct {
		name       string
		auth       string
		wantStatus int
	}{
		{
			name:       "missing header",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed token",
			auth:       "Bearer not-a-jwt",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token",
			auth:       "Bearer " + expiredToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "unsigned token",
			auth:       "Bearer " + unsignedToken,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token",
			auth:       "Bearer " + validToken,
			wantStatus: http.StatusOK,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			webServer := fiber.New()
			webServer.Use(NewJWT(jwtService))
			webServer.Get("/gallery/view", func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/gallery/view", nil)
			if test.auth != "" {
				req.Header.Set("Authorization", test.auth)
			}

			resp, err := webServer.Test(req, fiber.TestConfig{Timeout: 0})
			if err != nil {
				t.Fatalf("web server test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}
		})
	}
}

func TestScopeMiddlewareBoundaryResponses(t *testing.T) {
	jwtService, _ := newTrackedJWTMiddlewareService(t)
	allowedToken, _, err := jwtService.Generate(coreJWT.Payload{
		UserID: "user-1",
		Role:   jwtModel.UserRoleManager,
		Scopes: jwtModel.Scopes{
			{
				AccessType: jwtModel.AccessTypeRead,
				Service:    "gallery",
				Action:     "view",
			},
		},
	})
	if err != nil {
		t.Fatalf("generate allowed token: %v", err)
	}

	deniedToken, _, err := jwtService.Generate(coreJWT.Payload{
		UserID: "user-1",
		Role:   jwtModel.UserRoleManager,
	})
	if err != nil {
		t.Fatalf("generate denied token: %v", err)
	}

	cases := []struct {
		name       string
		auth       string
		withJWT    bool
		wantStatus int
	}{
		{
			name:       "missing decoded claims",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "allowed scope",
			auth:       "Bearer " + allowedToken,
			withJWT:    true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "denied scope",
			auth:       "Bearer " + deniedToken,
			withJWT:    true,
			wantStatus: http.StatusForbidden,
		},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			webServer := fiber.New()
			if test.withJWT {
				webServer.Use(NewJWT(jwtService))
			}
			webServer.Use(NewScopeMiddleware(context.Background()).CheckScope)
			webServer.Get("/gallery/view", func(c fiber.Ctx) error {
				return c.SendStatus(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/gallery/view", nil)
			if test.auth != "" {
				req.Header.Set("Authorization", test.auth)
			}

			resp, err := webServer.Test(req, fiber.TestConfig{Timeout: 0})
			if err != nil {
				t.Fatalf("web server test: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != test.wantStatus {
				t.Fatalf("status = %d, want %d", resp.StatusCode, test.wantStatus)
			}
		})
	}
}

func newTrackedJWTMiddlewareService(t *testing.T) (*coreJWT.Service, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, trackedJWTMiddlewareRSAKeyBits)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	})

	jwtService, err := coreJWT.NewJWT(privateKeyPEM, publicKeyPEM)
	if err != nil {
		t.Fatalf("new jwt service: %v", err)
	}

	return jwtService, privateKey
}

func signTrackedJWTMiddlewarePayload(t *testing.T, privateKey *rsa.PrivateKey, expiresAt time.Time) string {
	t.Helper()

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS512, jwtlib.MapClaims{
		"exp": jwtlib.NewNumericDate(expiresAt).Unix(),
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return tokenString
}

func signUnsignedTrackedJWTMiddlewarePayload(t *testing.T, expiresAt time.Time) string {
	t.Helper()

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, jwtlib.MapClaims{
		"exp": jwtlib.NewNumericDate(expiresAt).Unix(),
	})

	tokenString, err := token.SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign unsigned token: %v", err)
	}

	return tokenString
}

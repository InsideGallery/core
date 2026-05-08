package jwt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	jwtlib "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"github.com/InsideGallery/core/server/jwt/model"
	"github.com/InsideGallery/core/testutils"
)

const generatedTestRSAKeyBits = 2048

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

func TestServiceGenerateAndParsePayload(t *testing.T) {
	t.Parallel()

	jwtService, _ := newGeneratedTestJWTService(t)
	scope := model.Scope{
		AccessType: model.AccessTypeRead,
		Service:    "gallery",
		Action:     "view",
	}
	payload := Payload{
		UserID:         "user-1",
		OrgID:          "org-1",
		Role:           model.UserRoleManager,
		OrgSlug:        "inside-gallery",
		OrgName:        "Inside Gallery",
		UserName:       "Ada",
		Scopes:         model.Scopes{scope},
		ChangePassword: true,
	}

	accessToken, refreshToken, err := jwtService.Generate(payload)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	cases := []struct {
		name  string
		token string
		want  Payload
	}{
		{
			name:  "access token keeps full payload",
			token: accessToken,
			want:  payload,
		},
		{
			name:  "refresh token keeps refresh payload",
			token: refreshToken,
			want: Payload{
				UserID:  payload.UserID,
				OrgSlug: payload.OrgSlug,
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := jwtService.ParsePayload(test.token)
			if err != nil {
				t.Fatalf("ParsePayload() error: %v", err)
			}

			if !reflect.DeepEqual(*got, test.want) {
				t.Fatalf("ParsePayload() = %#v, want %#v", *got, test.want)
			}
		})
	}
}

func TestServiceParsePayloadRejectsInvalidTokens(t *testing.T) {
	t.Parallel()

	jwtService, privateKey := newGeneratedTestJWTService(t)
	payload := Payload{
		UserID:  "user-1",
		OrgSlug: "inside-gallery",
	}
	expiredToken := signGeneratedTestPayload(t, privateKey, payload, time.Now().Add(-time.Minute))
	unsignedToken := signUnsignedGeneratedTestPayload(t, payload, time.Now().Add(time.Hour))

	cases := []struct {
		name  string
		token string
	}{
		{
			name:  "malformed token",
			token: "not-a-jwt",
		},
		{
			name:  "expired token",
			token: expiredToken,
		},
		{
			name:  "unsigned token",
			token: unsignedToken,
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := jwtService.ParsePayload(test.token)
			if err == nil {
				t.Fatalf("ParsePayload() payload = %#v, want error", got)
			}
		})
	}
}

func TestPublicSigningKey(t *testing.T) {
	t.Parallel()

	jwtService, _ := newGeneratedTestJWTService(t)

	signingKey, err := jwtService.PublicSigningKey()
	if err != nil {
		t.Fatalf("PublicSigningKey() error: %v", err)
	}

	if signingKey.Algorithm != jwtlib.SigningMethodRS512.Name {
		t.Fatalf("Algorithm = %q, want %q", signingKey.Algorithm, jwtlib.SigningMethodRS512.Name)
	}

	block, rest := pem.Decode(signingKey.PublicKey)
	if block == nil {
		t.Fatal("PublicKey should be PEM encoded")
	}

	if len(rest) != 0 {
		t.Fatalf("PublicKey has trailing PEM data: %q", string(rest))
	}

	if block.Type != "PUBLIC KEY" {
		t.Fatalf("PEM type = %q, want PUBLIC KEY", block.Type)
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		t.Fatalf("parse public key: %v", err)
	}

	if _, ok := publicKey.(*rsa.PublicKey); !ok {
		t.Fatalf("public key type = %T, want *rsa.PublicKey", publicKey)
	}
}

func TestDecodeClaimsWithScopes(t *testing.T) {
	jwtService, _ := newGeneratedTestJWTService(t)
	scope := model.Scope{
		AccessType: model.AccessTypeWrite,
		Service:    "gallery",
		Action:     "edit",
	}
	payload := Payload{
		UserID:   "user-1",
		OrgID:    "org-1",
		Role:     model.UserRoleManager,
		OrgSlug:  "inside-gallery",
		OrgName:  "Inside Gallery",
		UserName: "Ada",
		Scopes:   model.Scopes{scope},
	}
	accessToken, _, err := jwtService.Generate(payload)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}

	var got *Payload
	var decodeErr error

	webServer := fiber.New()
	webServer.Use(jwtware.New(jwtware.Config{
		SigningKey: jwtService.GetSigningKey(),
	}))
	webServer.Get("/claims", func(c fiber.Ctx) error {
		got, decodeErr = DecodeClaims(c)

		return c.SendStatus(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/claims", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := webServer.Test(req, fiber.TestConfig{Timeout: 0})
	if err != nil {
		t.Fatalf("web server test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	if decodeErr != nil {
		t.Fatalf("DecodeClaims() error: %v", decodeErr)
	}

	if got == nil {
		t.Fatal("DecodeClaims() payload should not be nil")
	}

	if !reflect.DeepEqual(got.Scopes, payload.Scopes) {
		t.Fatalf("Scopes = %#v, want %#v", got.Scopes, payload.Scopes)
	}

	if got.UserID != payload.UserID || got.OrgID != payload.OrgID || got.Role != payload.Role {
		t.Fatalf("DecodeClaims() = %#v, want user/org/role from %#v", got, payload)
	}
}

func newGeneratedTestJWTService(t *testing.T) (*Service, *rsa.PrivateKey) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, generatedTestRSAKeyBits)
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

	jwtService, err := NewJWT(privateKeyPEM, publicKeyPEM)
	if err != nil {
		t.Fatalf("NewJWT() error: %v", err)
	}

	return jwtService, privateKey
}

func signGeneratedTestPayload(
	t *testing.T,
	privateKey *rsa.PrivateKey,
	payload Payload,
	expiresAt time.Time,
) string {
	t.Helper()

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodRS512, payloadClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		},
		Payload: payload,
	})

	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	return tokenString
}

func signUnsignedGeneratedTestPayload(t *testing.T, payload Payload, expiresAt time.Time) string {
	t.Helper()

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodNone, payloadClaims{
		RegisteredClaims: jwtlib.RegisteredClaims{
			ExpiresAt: jwtlib.NewNumericDate(expiresAt),
		},
		Payload: payload,
	})

	tokenString, err := token.SignedString(jwtlib.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign unsigned token: %v", err)
	}

	return tokenString
}

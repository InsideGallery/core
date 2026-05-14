package jwt

import (
	"os"
	"testing"

	"github.com/InsideGallery/core/server/jwt/model"
)

func TestTokenPairBoundary(t *testing.T) {
	t.Parallel()

	privateKey, err := os.ReadFile("test-data/test-jwt.key")
	if err != nil {
		t.Fatalf("read private key: %v", err)
	}

	publicKey, err := os.ReadFile("test-data/test-jwt.pem")
	if err != nil {
		t.Fatalf("read public key: %v", err)
	}

	service, err := NewJWT(privateKey, publicKey)
	if err != nil {
		t.Fatalf("NewJWT() error: %v", err)
	}

	scope, err := model.ScopeFrom("read:service:action")
	if err != nil {
		t.Fatalf("ScopeFrom() error: %v", err)
	}

	cases := []struct {
		name    string
		payload Payload
	}{
		{
			name: "generate and parse access token",
			payload: Payload{
				UserID: "user-1",
				OrgID:  "org-1",
				Role:   model.UserRoleRoot,
				Scopes: model.Scopes{scope},
			},
		},
	}

	for _, test := range cases {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			tokens, err := service.GenerateTokenPair(test.payload)
			if err != nil {
				t.Fatalf("GenerateTokenPair() error: %v", err)
			}

			if tokens.AccessToken == "" || tokens.RefreshToken == "" {
				t.Fatal("generated tokens should not be empty")
			}

			got, err := service.ParsePayload(tokens.AccessToken)
			if err != nil {
				t.Fatalf("ParsePayload() error: %v", err)
			}

			if got.UserID != test.payload.UserID {
				t.Fatalf("UserID = %q, want %q", got.UserID, test.payload.UserID)
			}
		})
	}
}

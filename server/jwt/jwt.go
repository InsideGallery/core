package jwt

import (
	"crypto/rsa"
	"time"

	"github.com/InsideGallery/core/server/jwt/model"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v4"
	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenExpire  = 4 * time.Hour
	refreshTokenExpire = 24 * time.Hour

	ContextJWTKey = "user"
)

type Payload struct {
	UserID         string         `json:"user_id"`
	OrgID          string         `json:"org_id"`
	Role           model.UserRole `json:"role"`
	OrgSlug        string         `json:"org_slug"`
	OrgName        string         `json:"org_name"`
	UserName       string         `json:"user_name"`
	Scopes         model.Scopes   `json:"scopes"`
	ChangePassword bool           `json:"change_password"`
}

type Service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewJWT(privateKey, publicKey []byte) (*Service, error) {
	privKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKey)
	if err != nil {
		return nil, err
	}

	return &Service{
		privateKey: privKey,
		publicKey:  pubKey,
	}, nil
}

func (j *Service) Generate(payload Payload) (accessToken, refreshToken string, err error) {
	type cl struct {
		jwt.RegisteredClaims
		Payload
	}

	// access token
	token := jwt.New(jwt.GetSigningMethod(jwt.SigningMethodRS512.Name))
	token.Claims = cl{
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpire)),
		},
	}

	accessToken, err = token.SignedString(j.privateKey)
	if err != nil {
		return
	}

	// refresh token
	token = jwt.New(jwt.GetSigningMethod(jwt.SigningMethodRS512.Name))
	token.Claims = cl{
		Payload: Payload{
			UserID:  payload.UserID,
			OrgSlug: payload.OrgSlug,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpire)),
		},
	}

	refreshToken, err = token.SignedString(j.privateKey)
	if err != nil {
		return
	}

	return
}

func (j *Service) GetSigningKey() jwtware.SigningKey {
	return jwtware.SigningKey{
		JWTAlg: jwt.SigningMethodRS512.Name,
		Key:    j.publicKey,
	}
}

func DecodeClaims(c *fiber.Ctx) (*Payload, error) {
	jwtToken, ok := c.Locals(ContextJWTKey).(*jwt.Token)
	if !ok {
		return nil, ErrJWTTokenNotFound
	}

	claims, ok := jwtToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrJWTTokenNotFound
	}

	payload := &Payload{
		UserID:   claims["user_id"].(string),
		OrgID:    claims["org_id"].(string),
		Role:     model.UserRole(claims["role"].(string)),
		OrgSlug:  claims["org_slug"].(string),
		OrgName:  claims["org_name"].(string),
		UserName: claims["user_name"].(string),
	}

	scopes, ok := claims["scopes"].([]interface{})
	if ok {
		for _, scopeItem := range scopes {
			if scopeItem == nil {
				continue
			}

			scopeInfo := scopeItem.(map[string]interface{})
			payload.Scopes = append(payload.Scopes, model.Scope{
				AccessType: model.AccessType(scopeInfo["access_type"].(string)),
				Service:    scopeInfo["service"].(string),
				Action:     scopeInfo["action"].(string),
			})
		}
	}

	return payload, nil
}

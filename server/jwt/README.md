# server/jwt

Import path: `github.com/InsideGallery/core/server/jwt`

`jwt` signs, parses, and configures RSA JWT credentials for server
authentication. Tokens are signed with RS512 using `github.com/golang-jwt/jwt/v5`.

## Main APIs

- `Config`: JWT key configuration loaded from `JWT_PRIVATE_KEY` and
  `JWT_PUBLIC_KEY`. Values may be PEM strings or file paths.
- `GetConfigFromEnv()`: parses `JWT_*` environment variables with
  `github.com/caarlos0/env/v10`.
- `NewJWT(privateKey, publicKey []byte)`: builds a `Service` from PEM encoded RSA
  keys.
- `Service.Generate(payload)`: returns access and refresh token strings.
- `Service.GenerateTokenPair(payload)`: returns a `TokenPair`.
- `Service.ParsePayload(tokenString)`: validates RS512 tokens and returns a
  `Payload`.
- `Service.PublicSigningKey()`: returns a core-owned `SigningKey`.
- `ErrJWTTokenNotFound`: sentinel returned when Fiber claim decoding has no JWT.

`GetSigningKey` and `DecodeClaims` are deprecated compatibility shims for
`github.com/gofiber/contrib/v3/jwt` and Fiber middleware.

## Usage

```go
cfg, err := jwt.GetConfigFromEnv()
if err != nil {
	return err
}

privateKey, err := cfg.GetPrivateKey()
if err != nil {
	return err
}

publicKey, err := cfg.GetPublicKey()
if err != nil {
	return err
}

service, err := jwt.NewJWT(privateKey, publicKey)
if err != nil {
	return err
}

tokens, err := service.GenerateTokenPair(jwt.Payload{UserID: "user-1"})
```

Access tokens keep the full `Payload` and expire after four hours. Refresh tokens
keep `UserID` and `OrgSlug` and expire after 24 hours.

## Operational Notes

Applications own key storage and rotation. Do not hardcode private keys in
source; provide key paths or PEM values through environment configuration.

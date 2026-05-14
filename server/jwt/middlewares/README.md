# server/jwt/middlewares

Import path: `github.com/InsideGallery/core/server/jwt/middlewares`

This package provides Fiber middleware for JWT validation and scope-based route
authorization.

## Main APIs

- `NewJWT(jwtService) fiber.Handler`: installs `gofiber/contrib/v3/jwt`
  validation using the signing key from `server/jwt.Service`.
- `NewScopeMiddleware(ctx)`: creates a `ScopeMiddleware`.
- `ScopeMiddleware.CheckScope(c fiber.Ctx) error`: decodes JWT claims and checks
  whether the request method and path are allowed by the user's role and scopes.

## Usage

```go
app := fiber.New()
app.Use(jwtmiddlewares.NewJWT(jwtService))
app.Use(jwtmiddlewares.NewScopeMiddleware(context.Background()).CheckScope)

app.Get("/gallery/view", func(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusOK)
})
```

`CheckScope` maps HTTP methods through `model.MethodMap`, drops a leading version
segment such as `/v1`, and builds scopes like `read:gallery:view`.

## Operational Notes

Missing or malformed JWT headers return HTTP 400. Invalid, expired, or unsigned
tokens return HTTP 401. Missing decoded claims return HTTP 400, and denied scopes
return HTTP 403. Error bodies use `server/webserver.Response`.

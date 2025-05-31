package middlewares

import (
	"errors"

	jwtCore "github.com/InsideGallery/core/server/jwt"
	"github.com/InsideGallery/core/server/webserver"
	"github.com/gofiber/fiber/v2"
	jwtware "github.com/gofiber/jwt/v4"
)

func NewJWT(jwtService *jwtCore.Service) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtService.GetSigningKey(),
		ContextKey: jwtCore.ContextJWTKey,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err.Error() == "Missing or malformed JWT" {
				return c.Status(fiber.StatusBadRequest).
					JSON(webserver.GetResponseWithError(errors.New("missing or malformed JWT"), 0))
			}

			return c.Status(fiber.StatusUnauthorized).
				JSON(webserver.GetResponseWithError(errors.New("invalid or expired JWT"), 0))
		},
	})
}

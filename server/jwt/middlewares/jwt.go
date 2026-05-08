package middlewares

import (
	"errors"

	jwtware "github.com/gofiber/contrib/v3/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/extractors"

	jwtCore "github.com/InsideGallery/core/server/jwt"
	"github.com/InsideGallery/core/server/webserver"
)

func NewJWT(jwtService *jwtCore.Service) fiber.Handler {
	return jwtware.New(jwtware.Config{
		SigningKey: jwtService.GetSigningKey(), //nolint:staticcheck // Fiber middleware needs the legacy signing key shim
		ErrorHandler: func(c fiber.Ctx, err error) error {
			if errors.Is(err, extractors.ErrNotFound) {
				return c.Status(fiber.StatusBadRequest).
					JSON(webserver.GetResponseWithError(errors.New("missing or malformed JWT"), 0))
			}

			return c.Status(fiber.StatusUnauthorized).
				JSON(webserver.GetResponseWithError(errors.New("invalid or expired JWT"), 0))
		},
	})
}

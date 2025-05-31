package middlewares

import (
	"log/slog"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rval := recover(); rval != nil {
				slog.Default().Error("Recovered request panic", "rval", rval)
				http.Error(w, "Panic during the request", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func RecoverFiber(next fiber.Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if rval := recover(); rval != nil {
				slog.Default().Error("Recovered request panic", "rval", rval)

				c.Status(http.StatusInternalServerError)

				_, err := c.WriteString("Panic during the request")
				if err != nil {
					slog.Default().Error("Error write response", "rval", rval)
				}
			}
		}()

		return next(c)
	}
}

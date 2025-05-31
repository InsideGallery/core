package webserver

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

var ErrorHandler = func(c *fiber.Ctx, err error) error {
	return c.Status(http.StatusInternalServerError).JSON(Response{
		Ok: false,
		Error: &ErrorResponse{
			Message: err.Error(),
			Code:    0,
		},
	})
}

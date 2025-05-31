package webserver

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

const ReadBufferSize = 8192

func NewFiberApp(name string) *fiber.App {
	return fiber.New(fiber.Config{
		ReadBufferSize:    ReadBufferSize,
		EnablePrintRoutes: os.Getenv("DEPLOYMENT_ENVIRONMENT") == "",
		ErrorHandler:      ErrorHandler,
		AppName:           name,
	})
}

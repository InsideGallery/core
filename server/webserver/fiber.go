package webserver

import "github.com/gofiber/fiber/v3"

const ReadBufferSize = 16384

func NewFiberApp(name string) *fiber.App {
	return fiber.New(fiber.Config{
		ReadBufferSize: ReadBufferSize,
		ServerHeader:   name,
		ErrorHandler:   ErrorHandler,
		AppName:        name,
	})
}

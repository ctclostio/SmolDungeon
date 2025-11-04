package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// ErrorHandler returns a custom error handler for Fiber
func ErrorHandler(c *fiber.Ctx, err error) error {
	log.Printf("Error: %v", err)
	return c.Status(500).JSON(fiber.Map{
		"error": "Internal server error",
	})
}

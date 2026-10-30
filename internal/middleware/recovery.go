package middleware

import (
	"errors"
	"github.com/Group-project-bridgeon/purchasing/internal/logger"
	"github.com/gofiber/fiber/v2"
)

func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) (err error) {
		defer func() {
			if recover() != nil {
				logger.Failure(c.UserContext(), "http", "panic_recovery", errors.New("panic recovered"))
				err = c.Status(500).JSON(fiber.Map{"error": "internal server error", "code": 500, "request_id": GetRequestID(c)})
			}
		}()
		return c.Next()
	}
}

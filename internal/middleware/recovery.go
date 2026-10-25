package middleware

import (
	"fmt"
	"log/slog"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

// Recovery intercepts unexpected panics and returns an HTTP 500 error cleanly.
func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				reqID := GetRequestID(c)
				stack := string(debug.Stack())

				slog.ErrorContext(
					c.UserContext(),
					"panic_recovered",
					slog.String("request_id", reqID),
					slog.Any("error", r),
					slog.String("stack", stack),
				)

				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error":      fmt.Sprintf("internal server error: %v", r),
					"code":       fiber.StatusInternalServerError,
					"request_id": reqID,
				})
			}
		}()

		return c.Next()
	}
}

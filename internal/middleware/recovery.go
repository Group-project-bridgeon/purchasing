package middleware

import (
	"fmt"
	"log"
	"runtime/debug"

	"github.com/gofiber/fiber/v2"
)

// Recovery intercepts unexpected panics and returns an HTTP 500 error cleanly.
func Recovery() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				reqID := GetRequestID(c)
				log.Printf(
					"[PANIC RECOVERED] req_id=%s error=%v\nstack:\n%s",
					reqID,
					r,
					debug.Stack(),
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

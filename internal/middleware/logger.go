package middleware

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Logger logs incoming HTTP requests with latency, status, path, and request ID.
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.OriginalURL()
		clientIP := c.IP()
		reqID := GetRequestID(c)

		log.Printf(
			"[HTTP] %3d | %10v | %-15s | %-7s %s | req_id=%s",
			status,
			latency,
			clientIP,
			method,
			path,
			reqID,
		)

		return err
	}
}

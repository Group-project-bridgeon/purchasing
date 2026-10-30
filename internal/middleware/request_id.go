package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"regexp"
)

const (
	HeaderXRequestID = "X-Request-ID"
	LocalsRequestID  = "requestId"
)

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

// RequestID attaches or propagates a unique correlation ID across requests.
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		reqID := c.Get(HeaderXRequestID)
		if !validRequestID.MatchString(reqID) || len(c.Request().Header.PeekAll(HeaderXRequestID)) != 1 {
			reqID = uuid.New().String()
		}

		c.Locals(LocalsRequestID, reqID)
		c.Set(HeaderXRequestID, reqID)

		return c.Next()
	}
}

// GetRequestID retrieves the request ID from the Fiber context.
func GetRequestID(c *fiber.Ctx) string {
	if val, ok := c.Locals(LocalsRequestID).(string); ok && val != "" {
		return val
	}
	return c.Get(HeaderXRequestID)
}

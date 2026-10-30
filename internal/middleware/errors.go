package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/logger"
	"github.com/gofiber/fiber/v2"
)

func SafeError(c *fiber.Ctx, err error) string {
	status := c.Response().StatusCode()
	reason, _ := logger.ErrorReason(err)
	switch reason {
	case "forbidden":
		status = 403
	case "conflict", "document_state_conflict":
		status = 409
	case "timeout":
		status = 504
	case "request_canceled":
		status = 408
	}
	if status < 400 {
		status = 500
	}
	c.Status(status)
	operation := "request"
	if v, ok := c.Locals(operationKey).(string); ok {
		operation = v
	}
	logger.Failure(c.UserContext(), "http", operation, err)
	switch status {
	case 400:
		return "invalid request"
	case 401:
		return "invalid or missing bearer token"
	case 403:
		return "permission denied"
	case 404:
		return "resource not found"
	case 408:
		return "request canceled"
	case 409:
		return "operation conflicts with current document state"
	case 413:
		return "request body too large"
	case 415:
		return "unsupported media type"
	case 503:
		return "service unavailable"
	case 504:
		return "operation timed out"
	default:
		return "internal server error"
	}
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	var httpErr *fiber.Error
	if errors.As(err, &httpErr) {
		status = httpErr.Code
	}
	c.Status(status)
	message := SafeError(c, err)
	return c.JSON(fiber.Map{"error": message, "request_id": GetRequestID(c)})
}

func RequestBounds(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()
		c.SetUserContext(ctx)
		return c.Next()
	}
}

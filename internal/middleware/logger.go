package middleware

import (
	"log/slog"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/logger"
	"github.com/gofiber/fiber/v2"
)

// Logger logs incoming HTTP requests with structured fields and appropriate log levels.
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		reqID := GetRequestID(c)

		// Attach request ID into context
		ctx := logger.WithRequestID(c.UserContext(), reqID)
		c.SetUserContext(ctx)

		err := c.Next()

		latency := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		clientIP := c.IP()
		bytesSent := len(c.Response().Body())

		attrs := []any{
			slog.Int("status", status),
			slog.Int64("latency_ms", latency.Milliseconds()),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("ip", clientIP),
			slog.String("request_id", reqID),
			slog.Int("bytes_sent", bytesSent),
		}

		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
		}

		logMsg := "http_request"
		if status >= 500 || err != nil {
			slog.ErrorContext(ctx, logMsg, attrs...)
		} else if status >= 400 {
			slog.WarnContext(ctx, logMsg, attrs...)
		} else {
			slog.InfoContext(ctx, logMsg, attrs...)
		}

		return err
	}
}

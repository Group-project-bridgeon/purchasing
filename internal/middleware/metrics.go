package middleware

import (
	"strings"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/metrics"
	"github.com/gofiber/fiber/v2"
)

// Metrics records Prometheus metrics for every incoming HTTP request.
func Metrics() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Increment in-flight requests gauge
		metrics.HTTPRequestsInFlight.Inc()
		defer metrics.HTTPRequestsInFlight.Dec()

		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		statusCode := c.Response().StatusCode()

		method := c.Method()
		routePath := "unmatched"

		if r := c.Route(); r != nil {
			if r.Method != "" {
				method = r.Method
			}
			if r.Path != "" {
				routePath = r.Path
			}
		}

		metrics.RecordHTTPRequest(strings.Clone(method), strings.Clone(routePath), statusCode, duration)

		return err
	}
}

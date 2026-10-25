package middleware

import (
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

		// Use route path template (e.g., "/api/v1/purchase-orders/:id") to avoid high-cardinality explosions.
		routePath := ""
		if r := c.Route(); r != nil && r.Path != "" {
			routePath = r.Path
		} else {
			routePath = "unmatched"
		}

		metrics.RecordHTTPRequest(method, routePath, statusCode, duration)

		return err
	}
}

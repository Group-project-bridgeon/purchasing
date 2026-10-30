package api

import (
	"context"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/logger"
	"github.com/Group-project-bridgeon/purchasing/internal/security"
	"github.com/gofiber/fiber/v2"
)

type ReadinessProbe func(context.Context) error

func readiness(probe ReadinessProbe, timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.UserContext(), timeout)
		defer cancel()
		if probe == nil {
			return c.Status(503).JSON(fiber.Map{"status": "unavailable"})
		}
		if err := probe(ctx); err != nil {
			logger.Failure(c.UserContext(), "readiness", "database_check", err)
			return c.Status(503).JSON(fiber.Map{"status": "unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	}
}

type RouterOptions struct {
	Verifier                         *security.Verifier
	Readiness                        ReadinessProbe
	ReadinessTimeout, RequestTimeout time.Duration
}

func defaultRouterOptions(options []RouterOptions) RouterOptions {
	var v RouterOptions
	if len(options) > 0 {
		v = options[0]
	}
	if v.ReadinessTimeout <= 0 {
		v.ReadinessTimeout = time.Second
	}
	if v.RequestTimeout <= 0 {
		v.RequestTimeout = 5 * time.Second
	}
	return v
}

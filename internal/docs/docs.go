package docs

import (
	_ "embed"

	"github.com/gofiber/fiber/v2"
)

//go:embed swagger.html
var swaggerHTML []byte

//go:embed openapi.json
var openAPISpec []byte

// RegisterRoutes registers Swagger UI, raw OpenAPI 3.0 specification, and health endpoints.
func RegisterRoutes(app *fiber.App) {
	// Swagger UI HTML endpoint
	renderSwaggerUI := func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Send(swaggerHTML)
	}

	app.Get("/docs", renderSwaggerUI)
	app.Get("/docs/", renderSwaggerUI)
	app.Get("/swagger", func(c *fiber.Ctx) error {
		return c.Redirect("/docs", fiber.StatusMovedPermanently)
	})
	app.Get("/swagger/", func(c *fiber.Ctx) error {
		return c.Redirect("/docs", fiber.StatusMovedPermanently)
	})

	// Raw OpenAPI 3.0 JSON specification endpoint
	app.Get("/docs/openapi.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json; charset=utf-8")
		return c.Send(openAPISpec)
	})

	// Service healthcheck endpoints
	healthHandler := func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "ok",
			"service": "purchasing-service",
			"version": "1.0.0",
		})
	}

	app.Get("/health", healthHandler)
	app.Get("/api/v1/health", healthHandler)
}

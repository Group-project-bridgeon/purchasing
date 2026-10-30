package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterOutboxRoutes registers all routes for querying and retrying Transactional Outbox domain events.
func RegisterOutboxRoutes(
	r fiber.Router,
	h *handler.OutboxHandler,
	cfg config.OutboxRoutes,
) {
	// Disabled by approved policy; no administration endpoint is registered.
	if h == nil {
		return
	}
}

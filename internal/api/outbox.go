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
	if h == nil {
		return
	}

	r.Get(cfg.Base, h.List)    // GET  /api/v1/outbox/events
	r.Get(cfg.ByID, h.GetByID) // GET  /api/v1/outbox/events/:id
	r.Post(cfg.Retry, h.Retry) // POST /api/v1/outbox/events/:id/retry
}

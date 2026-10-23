package v1

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterOutboxRoutes registers all routes for inspecting and retrying Outbox domain events.
func RegisterOutboxRoutes(
	r fiber.Router,
	h *handler.OutboxHandler,
	cfg config.OutboxRoutes,
) {
	if h == nil {
		return
	}

	r.Get(cfg.Base, h.List)
	r.Get(cfg.ByID, h.GetByID)
	r.Post(cfg.Retry, h.Retry)
}

package v1

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterPurchaseReturnRoutes registers all routes for Purchase Returns (RTV).
func RegisterPurchaseReturnRoutes(
	r fiber.Router,
	h *handler.PurchaseReturnHandler,
	cfg config.PurchaseReturnRoutes,
) {
	if h == nil {
		return
	}
	r.Post(cfg.Base, h.Create)
	r.Get(cfg.Base, h.List)
	r.Get(cfg.ByID, h.GetByID)
	r.Get(cfg.ByNumber, h.GetByNumber)
	r.Put(cfg.ByID, h.Update)
	r.Post(cfg.Approve, h.Approve)
	r.Post(cfg.Ship, h.Ship)
	r.Post(cfg.Complete, h.Complete)
	r.Post(cfg.Cancel, h.Cancel)
}

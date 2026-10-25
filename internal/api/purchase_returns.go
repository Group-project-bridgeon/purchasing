package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterPurchaseReturnRoutes registers all routes for Purchase Returns (RTV) and Debit Notes.
func RegisterPurchaseReturnRoutes(
	r fiber.Router,
	h *handler.PurchaseReturnHandler,
	cfg config.PurchaseReturnRoutes,
) {
	if h == nil {
		return
	}

	r.Post(cfg.Base, h.Create)         // POST /api/v1/purchase-returns
	r.Get(cfg.Base, h.List)            // GET  /api/v1/purchase-returns
	r.Get(cfg.ByID, h.GetByID)         // GET  /api/v1/purchase-returns/:id
	r.Get(cfg.ByNumber, h.GetByNumber) // GET  /api/v1/purchase-returns/number/:returnNumber
	r.Put(cfg.ByID, h.Update)          // PUT  /api/v1/purchase-returns/:id
	r.Post(cfg.Approve, h.Approve)     // POST /api/v1/purchase-returns/:id/approve
	r.Post(cfg.Ship, h.Ship)           // POST /api/v1/purchase-returns/:id/ship
	r.Post(cfg.Complete, h.Complete)   // POST /api/v1/purchase-returns/:id/complete
	r.Post(cfg.Cancel, h.Cancel)       // POST /api/v1/purchase-returns/:id/cancel
}

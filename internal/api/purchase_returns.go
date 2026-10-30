package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
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

	r.Post(cfg.Base, middleware.Require("purchasing.purchase-returns:create"), h.Create)         // POST /api/v1/purchase-returns
	r.Get(cfg.Base, middleware.Require("purchasing.purchase-returns:read"), h.List)              // GET  /api/v1/purchase-returns
	r.Get(cfg.ByID, middleware.Require("purchasing.purchase-returns:read"), h.GetByID)           // GET  /api/v1/purchase-returns/:id
	r.Get(cfg.ByNumber, middleware.Require("purchasing.purchase-returns:read"), h.GetByNumber)   // GET  /api/v1/purchase-returns/number/:returnNumber
	r.Put(cfg.ByID, middleware.Require("purchasing.purchase-returns:update"), h.Update)          // PUT  /api/v1/purchase-returns/:id
	r.Post(cfg.Approve, middleware.Require("purchasing.purchase-returns:approve"), h.Approve)    // POST /api/v1/purchase-returns/:id/approve
	r.Post(cfg.Ship, middleware.Require("purchasing.purchase-returns:ship"), h.Ship)             // POST /api/v1/purchase-returns/:id/ship
	r.Post(cfg.Complete, middleware.Require("purchasing.purchase-returns:complete"), h.Complete) // POST /api/v1/purchase-returns/:id/complete
	r.Post(cfg.Cancel, middleware.Require("purchasing.purchase-returns:cancel"), h.Cancel)       // POST /api/v1/purchase-returns/:id/cancel
}

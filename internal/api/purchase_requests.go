package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// RegisterPurchaseRequestRoutes registers all routes for Purchase Requests and PR line items.
func RegisterPurchaseRequestRoutes(
	r fiber.Router,
	h *handler.PurchaseRequestHandler,
	cfg config.PurchaseRequestRoutes,
) {
	if h == nil {
		return
	}

	// Requisition header operations
	r.Post(cfg.Base, middleware.Require("purchasing.purchase-requests:create"), h.Create)      // POST   /api/v1/purchase-requests
	r.Get(cfg.Base, middleware.Require("purchasing.purchase-requests:read"), h.List)           // GET    /api/v1/purchase-requests
	r.Get(cfg.ByID, middleware.Require("purchasing.purchase-requests:read"), h.GetByID)        // GET    /api/v1/purchase-requests/:id
	r.Put(cfg.ByID, middleware.Require("purchasing.purchase-requests:update"), h.Update)       // PUT    /api/v1/purchase-requests/:id
	r.Post(cfg.Submit, middleware.Require("purchasing.purchase-requests:submit"), h.Submit)    // POST   /api/v1/purchase-requests/:id/submit
	r.Post(cfg.Approve, middleware.Require("purchasing.purchase-requests:approve"), h.Approve) // POST   /api/v1/purchase-requests/:id/approve
	r.Post(cfg.Reject, middleware.Require("purchasing.purchase-requests:reject"), h.Reject)    // POST   /api/v1/purchase-requests/:id/reject

	// Requisition line items CRUD
	r.Get(cfg.Items, middleware.Require("purchasing.purchase-requests:read"), h.GetItems)           // GET    /api/v1/purchase-requests/:id/items
	r.Post(cfg.Items, middleware.Require("purchasing.purchase-requests:update"), h.AddItem)         // POST   /api/v1/purchase-requests/:id/items
	r.Put(cfg.ItemByID, middleware.Require("purchasing.purchase-requests:update"), h.UpdateItem)    // PUT    /api/v1/purchase-requests/:id/items/:itemId
	r.Delete(cfg.ItemByID, middleware.Require("purchasing.purchase-requests:delete"), h.DeleteItem) // DELETE /api/v1/purchase-requests/:id/items/:itemId
}

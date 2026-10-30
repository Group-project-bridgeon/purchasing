package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// RegisterPurchaseOrderRoutes registers all routes for Purchase Orders and PO items.
func RegisterPurchaseOrderRoutes(
	r fiber.Router,
	h *handler.PurchaseOrderHandler,
	cfg config.PurchaseOrderRoutes,
) {
	if h == nil {
		return
	}

	// Order header operations & lifecycle
	r.Post(cfg.Base, middleware.Require("purchasing.purchase-orders:create"), h.Create)                                                   // POST   /api/v1/purchase-orders
	r.Post(cfg.FromRequest, middleware.Require("purchasing.purchase-orders:create", "purchasing.purchase-requests:read"), h.CreateFromPR) // POST   /api/v1/purchase-orders/from-request/:prId
	r.Get(cfg.Base, middleware.Require("purchasing.purchase-orders:read"), h.List)                                                        // GET    /api/v1/purchase-orders
	r.Get(cfg.ByID, middleware.Require("purchasing.purchase-orders:read"), h.GetByID)                                                     // GET    /api/v1/purchase-orders/:id
	r.Put(cfg.ByID, middleware.Require("purchasing.purchase-orders:update"), h.Update)                                                    // PUT    /api/v1/purchase-orders/:id
	r.Post(cfg.Approve, middleware.Require("purchasing.purchase-orders:approve"), h.Approve)                                              // POST   /api/v1/purchase-orders/:id/approve
	r.Post(cfg.Issue, middleware.Require("purchasing.purchase-orders:issue"), h.Issue)                                                    // POST   /api/v1/purchase-orders/:id/issue
	r.Post(cfg.Cancel, middleware.Require("purchasing.purchase-orders:cancel"), h.Cancel)                                                 // POST   /api/v1/purchase-orders/:id/cancel

	// Order line items CRUD
	r.Get(cfg.Items, middleware.Require("purchasing.purchase-orders:read"), h.GetItems)           // GET    /api/v1/purchase-orders/:id/items
	r.Post(cfg.Items, middleware.Require("purchasing.purchase-orders:update"), h.AddItem)         // POST   /api/v1/purchase-orders/:id/items
	r.Put(cfg.ItemByID, middleware.Require("purchasing.purchase-orders:update"), h.UpdateItem)    // PUT    /api/v1/purchase-orders/:id/items/:itemId
	r.Delete(cfg.ItemByID, middleware.Require("purchasing.purchase-orders:delete"), h.DeleteItem) // DELETE /api/v1/purchase-orders/:id/items/:itemId
}

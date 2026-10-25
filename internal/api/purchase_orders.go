package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
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
	r.Post(cfg.Base, h.Create)              // POST   /api/v1/purchase-orders
	r.Post(cfg.FromRequest, h.CreateFromPR) // POST   /api/v1/purchase-orders/from-request/:prId
	r.Get(cfg.Base, h.List)                 // GET    /api/v1/purchase-orders
	r.Get(cfg.ByID, h.GetByID)              // GET    /api/v1/purchase-orders/:id
	r.Put(cfg.ByID, h.Update)               // PUT    /api/v1/purchase-orders/:id
	r.Post(cfg.Approve, h.Approve)          // POST   /api/v1/purchase-orders/:id/approve
	r.Post(cfg.Issue, h.Issue)              // POST   /api/v1/purchase-orders/:id/issue
	r.Post(cfg.Cancel, h.Cancel)            // POST   /api/v1/purchase-orders/:id/cancel

	// Order line items CRUD
	r.Get(cfg.Items, h.GetItems)         // GET    /api/v1/purchase-orders/:id/items
	r.Post(cfg.Items, h.AddItem)         // POST   /api/v1/purchase-orders/:id/items
	r.Put(cfg.ItemByID, h.UpdateItem)    // PUT    /api/v1/purchase-orders/:id/items/:itemId
	r.Delete(cfg.ItemByID, h.DeleteItem) // DELETE /api/v1/purchase-orders/:id/items/:itemId
}

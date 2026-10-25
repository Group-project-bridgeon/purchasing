package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
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
	r.Post(cfg.Base, h.Create)     // POST   /api/v1/purchase-requests
	r.Get(cfg.Base, h.List)        // GET    /api/v1/purchase-requests
	r.Get(cfg.ByID, h.GetByID)     // GET    /api/v1/purchase-requests/:id
	r.Put(cfg.ByID, h.Update)      // PUT    /api/v1/purchase-requests/:id
	r.Post(cfg.Submit, h.Submit)   // POST   /api/v1/purchase-requests/:id/submit
	r.Post(cfg.Approve, h.Approve) // POST   /api/v1/purchase-requests/:id/approve
	r.Post(cfg.Reject, h.Reject)   // POST   /api/v1/purchase-requests/:id/reject

	// Requisition line items CRUD
	r.Get(cfg.Items, h.GetItems)         // GET    /api/v1/purchase-requests/:id/items
	r.Post(cfg.Items, h.AddItem)         // POST   /api/v1/purchase-requests/:id/items
	r.Put(cfg.ItemByID, h.UpdateItem)    // PUT    /api/v1/purchase-requests/:id/items/:itemId
	r.Delete(cfg.ItemByID, h.DeleteItem) // DELETE /api/v1/purchase-requests/:id/items/:itemId
}

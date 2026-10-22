package v1

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterPurchaseRequestRoutes registers all routes for Purchase Requests and PR items.
func RegisterPurchaseRequestRoutes(
	r fiber.Router,
	h *handler.PurchaseRequestHandler,
	cfg config.PurchaseRequestRoutes,
) {
	if h == nil {
		return
	}
	// Requisition header operations
	r.Post(cfg.Base, h.Create)
	r.Get(cfg.Base, h.List)
	r.Get(cfg.ByID, h.GetByID)
	r.Put(cfg.ByID, h.Update)
	r.Post(cfg.Submit, h.Submit)
	r.Post(cfg.Approve, h.Approve)
	r.Post(cfg.Reject, h.Reject)

	// Requisition line items CRUD
	r.Get(cfg.Items, h.GetItems)
	r.Post(cfg.Items, h.AddItem)
	r.Put(cfg.ItemByID, h.UpdateItem)
	r.Delete(cfg.ItemByID, h.DeleteItem)
}

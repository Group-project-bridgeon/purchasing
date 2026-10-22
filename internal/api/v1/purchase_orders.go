package v1

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
	r.Post(cfg.Base, h.Create)
	r.Post(cfg.FromRequest, h.CreateFromPR)
	r.Get(cfg.Base, h.List)
	r.Get(cfg.ByID, h.GetByID)
	r.Put(cfg.ByID, h.Update)
	r.Post(cfg.Approve, h.Approve)
	r.Post(cfg.Issue, h.Issue)
	r.Post(cfg.Cancel, h.Cancel)

	// Order line items CRUD
	r.Get(cfg.Items, h.GetItems)
	r.Post(cfg.Items, h.AddItem)
	r.Put(cfg.ItemByID, h.UpdateItem)
	r.Delete(cfg.ItemByID, h.DeleteItem)
}

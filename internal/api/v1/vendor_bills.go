package v1

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterVendorBillRoutes registers all routes for Vendor Bills and 3-Way Matching.
func RegisterVendorBillRoutes(
	r fiber.Router,
	h *handler.VendorBillHandler,
	cfg config.VendorBillRoutes,
	poCfg config.PurchaseOrderRoutes,
) {
	if h == nil {
		return
	}
	r.Post(cfg.Base, h.Create)
	r.Get(cfg.Base, h.List)
	r.Get(cfg.ByID, h.GetByID)
	r.Post(cfg.Match, h.Match)
	r.Post(cfg.Approve, h.Approve)
	r.Post(cfg.Cancel, h.Cancel)

	// Sub-resource route under PO
	r.Get(poCfg.VendorBills, h.ListByPO)
}

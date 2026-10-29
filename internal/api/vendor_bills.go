package api

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

	r.Post(cfg.Base, h.Create)     // POST /api/v1/vendor-bills
	r.Get(cfg.Base, h.List)        // GET  /api/v1/vendor-bills
	r.Get(cfg.ByID, h.GetByID)     // GET  /api/v1/vendor-bills/:id
	r.Post(cfg.Match, h.Match)     // POST /api/v1/vendor-bills/:id/match
	r.Post(cfg.Approve, h.Approve) // POST /api/v1/vendor-bills/:id/approve
	r.Post(cfg.Cancel, h.Cancel)   // POST /api/v1/vendor-bills/:id/cancel

	// Sub-resource route under PO
	r.Get(poCfg.VendorBills, h.ListByPO) // GET /api/v1/purchase-orders/:id/vendor-bills
}

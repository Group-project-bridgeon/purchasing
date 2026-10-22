package v1

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterGoodsReceiptRoutes registers all routes for Goods Receipts (GRN).
func RegisterGoodsReceiptRoutes(
	r fiber.Router,
	h *handler.GoodsReceiptHandler,
	cfg config.GoodsReceiptRoutes,
	poCfg config.PurchaseOrderRoutes,
) {
	if h == nil {
		return
	}
	r.Post(cfg.Base, h.Create)
	r.Get(cfg.Base, h.List)
	r.Get(cfg.ByID, h.GetByID)
	r.Post(cfg.Cancel, h.Cancel)

	// Sub-resource route under PO
	r.Get(poCfg.GoodsReceipts, h.ListByPO)
}

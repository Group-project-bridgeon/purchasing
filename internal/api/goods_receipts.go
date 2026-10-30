package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
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

	r.Post(cfg.Base, middleware.Require("purchasing.goods-receipts:create"), h.Create)   // POST /api/v1/goods-receipts
	r.Get(cfg.Base, middleware.Require("purchasing.goods-receipts:read"), h.List)        // GET  /api/v1/goods-receipts
	r.Get(cfg.ByID, middleware.Require("purchasing.goods-receipts:read"), h.GetByID)     // GET  /api/v1/goods-receipts/:id
	r.Post(cfg.Cancel, middleware.Require("purchasing.goods-receipts:cancel"), h.Cancel) // POST /api/v1/goods-receipts/:id/cancel

	// Sub-resource route under PO
	r.Get(poCfg.GoodsReceipts, middleware.Require("purchasing.goods-receipts:read"), h.ListByPO) // GET /api/v1/purchase-orders/:id/goods-receipts
}

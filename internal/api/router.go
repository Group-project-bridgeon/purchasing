package api

import (
	v1 "github.com/Group-project-bridgeon/purchasing/internal/api/v1"
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/docs"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

// Handlers bundles all domain handlers for clean dependency injection into the router.
type Handlers struct {
	PurchaseRequest *handler.PurchaseRequestHandler
	PurchaseOrder   *handler.PurchaseOrderHandler
	GoodsReceipt    *handler.GoodsReceiptHandler
	VendorBill      *handler.VendorBillHandler
	PurchaseReturn  *handler.PurchaseReturnHandler
	RFQ             *handler.RFQHandler
	Analytics       *handler.AnalyticsHandler
}

// SetupRouter configures the HTTP middleware pipeline, system endpoints, and all domain routes.
func SetupRouter(
	app *fiber.App,
	handlers Handlers,
	routesCfg *config.RoutesConfig,
) {
	// 1. Mount Global Middleware Pipeline
	app.Use(middleware.RequestID())
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())

	// 2. Mount Documentation & System Health
	docs.RegisterRoutes(app)

	// 3. Mount API V1 Routes Group
	v1Group := app.Group(routesCfg.APIPrefix)

	v1.RegisterPurchaseRequestRoutes(v1Group, handlers.PurchaseRequest, routesCfg.PurchaseRequests)
	v1.RegisterPurchaseOrderRoutes(v1Group, handlers.PurchaseOrder, routesCfg.PurchaseOrders)
	v1.RegisterGoodsReceiptRoutes(v1Group, handlers.GoodsReceipt, routesCfg.GoodsReceipts, routesCfg.PurchaseOrders)
	v1.RegisterVendorBillRoutes(v1Group, handlers.VendorBill, routesCfg.VendorBills, routesCfg.PurchaseOrders)
	v1.RegisterPurchaseReturnRoutes(v1Group, handlers.PurchaseReturn, routesCfg.PurchaseReturns)
	v1.RegisterRFQRoutes(v1Group, handlers.RFQ, routesCfg.RFQs)
	v1.RegisterAnalyticsRoutes(v1Group, handlers.Analytics, routesCfg.Analytics)
}

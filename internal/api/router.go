package api

import (
	"strings"

	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/docs"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	Outbox          *handler.OutboxHandler
}

// SetupRouter configures the HTTP middleware pipeline, system endpoints, and all domain routes.
func SetupRouter(
	app *fiber.App,
	handlers Handlers,
	routesCfg *config.RoutesConfig,
	options ...RouterOptions,
) {
	settings := defaultRouterOptions(options)
	auth := middleware.HumanIdentity(settings.Verifier)
	// 0. Normalize duplicate slashes (e.g. "//api/v1" -> "/api/v1")
	app.Use(func(c *fiber.Ctx) error {
		p := c.Path()
		if strings.Contains(p, "//") {
			for strings.Contains(p, "//") {
				p = strings.ReplaceAll(p, "//", "/")
			}
			c.Path(p)
		}
		return c.Next()
	})

	// 1. Mount Global Middleware Pipeline
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestBounds(settings.RequestTimeout))
	app.Use(middleware.Metrics())
	app.Use(middleware.Logger())
	app.Use(middleware.Recovery())
	app.Use(middleware.CORS())

	// 2. Mount Documentation, Health & Prometheus Metrics
	app.Use("/docs", auth, middleware.Require("purchasing.documentation:read"))
	app.Use("/swagger", auth, middleware.Require("purchasing.documentation:read"))
	docs.RegisterRoutes(app)
	app.Get("/ready", readiness(settings.Readiness, settings.ReadinessTimeout))
	app.Get("/metrics", auth, middleware.Require("purchasing.metrics:read"), adaptor.HTTPHandler(promhttp.Handler()))

	// 3. Mount API V1 Routes Group
	v1Group := app.Group(routesCfg.APIPrefix, auth)

	RegisterPurchaseRequestRoutes(v1Group, handlers.PurchaseRequest, routesCfg.PurchaseRequests)
	RegisterPurchaseOrderRoutes(v1Group, handlers.PurchaseOrder, routesCfg.PurchaseOrders)
	RegisterGoodsReceiptRoutes(v1Group, handlers.GoodsReceipt, routesCfg.GoodsReceipts, routesCfg.PurchaseOrders)
	RegisterVendorBillRoutes(v1Group, handlers.VendorBill, routesCfg.VendorBills, routesCfg.PurchaseOrders)
	RegisterPurchaseReturnRoutes(v1Group, handlers.PurchaseReturn, routesCfg.PurchaseReturns)
	RegisterRFQRoutes(v1Group, handlers.RFQ, routesCfg.RFQs)
	RegisterAnalyticsRoutes(v1Group, handlers.Analytics, routesCfg.Analytics)
	RegisterOutboxRoutes(v1Group, handlers.Outbox, routesCfg.Outbox)
}

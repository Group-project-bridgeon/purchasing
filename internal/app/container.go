package app

import (
	"context"
	"errors"
	"github.com/Group-project-bridgeon/purchasing/internal/api"
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/database"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/logger"
	"github.com/Group-project-bridgeon/purchasing/internal/metrics"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/Group-project-bridgeon/purchasing/internal/security"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Container manages the dependency injection lifecycle of the purchasing microservice.
type Container struct {
	Config        *config.Config
	Routes        *config.RoutesConfig
	DB            *pgxpool.Pool
	FiberApp      *fiber.App
	Handlers      api.Handlers
	OutboxService service.OutboxService
}

// NewContainer initializes repositories, services, handlers, and the Fiber application.
func NewContainer(
	db *pgxpool.Pool,
	cfg *config.Config,
	routesCfg *config.RoutesConfig,
) (*Container, error) {
	verifier, err := security.Load(cfg.JWTPublicKeyPath, cfg.JWTIssuer, cfg.JWTAudience)
	if err != nil {
		return nil, errors.New("invalid human JWT configuration")
	}
	if cfg.OutboxPublisherEnabled {
		return nil, service.ErrDispatcherNotConfigured
	}
	// -------------------------------------------------------------------------
	// 1. Repositories & Transaction Managers
	// -------------------------------------------------------------------------
	purchaseRequestRepo := repository.NewPostgresPurchaseRequestRepository(db)
	purchaseRequestItemRepo := repository.NewPostgresPurchaseRequestItemRepository(db)
	prTxManager := repository.NewPostgresTransactionManager(db)

	purchaseOrderRepo := repository.NewPostgresPurchaseOrderRepository(db)
	purchaseOrderItemRepo := repository.NewPostgresPurchaseOrderItemRepository(db)
	documentNumberRepo := repository.NewPostgresDocumentNumberRepository(db)
	poTxManager := repository.NewPostgresPurchaseOrderTransactionManager(db)

	goodsReceiptRepo := repository.NewPostgresGoodsReceiptRepository(db)
	goodsReceiptItemRepo := repository.NewPostgresGoodsReceiptItemRepository(db)
	grnTxManager := repository.NewPostgresGoodsReceiptTransactionManager(db)

	vendorBillRepo := repository.NewPostgresVendorBillRepository(db)
	vendorBillItemRepo := repository.NewPostgresVendorBillItemRepository(db)
	billTxManager := repository.NewPostgresVendorBillTransactionManager(db)

	purchaseReturnRepo := repository.NewPostgresPurchaseReturnRepository(db)
	purchaseReturnItemRepo := repository.NewPostgresPurchaseReturnItemRepository(db)
	purchaseReturnTxManager := repository.NewPostgresPurchaseReturnTransactionManager(db)

	rfqRepo := repository.NewPostgresRFQRepository(db)
	rfqItemRepo := repository.NewPostgresRFQItemRepository(db)
	vendorQuotationRepo := repository.NewPostgresVendorQuotationRepository(db)
	vendorQuotationItemRepo := repository.NewPostgresVendorQuotationItemRepository(db)
	rfqTxManager := repository.NewPostgresRFQTransactionManager(db)

	analyticsRepo := repository.NewPostgresAnalyticsRepository(db)
	outboxRepo := repository.NewPostgresOutboxRepository(db)

	// -------------------------------------------------------------------------
	// 2. Business Services
	// -------------------------------------------------------------------------
	outboxService := service.NewOutboxService(outboxRepo, nil)

	prService := service.NewPurchaseRequestService(
		purchaseRequestRepo,
		purchaseRequestItemRepo,
		prTxManager,
	)

	poService := service.NewPurchaseOrderService(
		purchaseOrderRepo,
		purchaseOrderItemRepo,
		poTxManager,
	)

	grnService := service.NewGoodsReceiptService(
		goodsReceiptRepo,
		goodsReceiptItemRepo,
		grnTxManager,
	)

	billService := service.NewVendorBillService(
		vendorBillRepo,
		vendorBillItemRepo,
		billTxManager,
	)

	returnService := service.NewPurchaseReturnService(
		purchaseReturnRepo,
		purchaseReturnItemRepo,
		purchaseOrderRepo,
		purchaseOrderItemRepo,
		documentNumberRepo,
		purchaseReturnTxManager,
	)

	rfqService := service.NewRFQService(
		rfqRepo,
		rfqItemRepo,
		vendorQuotationRepo,
		vendorQuotationItemRepo,
		purchaseOrderRepo,
		purchaseOrderItemRepo,
		purchaseRequestRepo,
		purchaseRequestItemRepo,
		documentNumberRepo,
		rfqTxManager,
	)

	analyticsService := service.NewAnalyticsService(
		analyticsRepo,
	)

	// -------------------------------------------------------------------------
	// 3. HTTP Handlers
	// -------------------------------------------------------------------------
	handlers := api.Handlers{
		PurchaseRequest: handler.NewPurchaseRequestHandler(prService),
		PurchaseOrder:   handler.NewPurchaseOrderHandler(poService),
		GoodsReceipt:    handler.NewGoodsReceiptHandler(grnService),
		VendorBill:      handler.NewVendorBillHandler(billService),
		PurchaseReturn:  handler.NewPurchaseReturnHandler(returnService),
		RFQ:             handler.NewRFQHandler(rfqService),
		Analytics:       handler.NewAnalyticsHandler(analyticsService),
		Outbox:          handler.NewOutboxHandler(outboxService),
	}

	// -------------------------------------------------------------------------
	// 4. Fiber Engine & Master Router Setup
	// -------------------------------------------------------------------------
	if db != nil {
		metrics.RegisterDBPoolCollector(db)
	}

	app := fiber.New(fiber.Config{
		AppName:      "ERP Purchasing Microservice",
		ServerHeader: "Purchasing",
		ReadTimeout:  cfg.HTTPReadTimeout, WriteTimeout: cfg.HTTPWriteTimeout, IdleTimeout: cfg.HTTPIdleTimeout, BodyLimit: cfg.HTTPMaxBodyBytes, ErrorHandler: middleware.ErrorHandler, DisableStartupMessage: true,
	})

	api.SetupRouter(app, handlers, routesCfg, api.RouterOptions{Verifier: verifier, ReadinessTimeout: cfg.ReadinessTimeout, RequestTimeout: cfg.RequestTimeout, Readiness: func(ctx context.Context) error {
		if db == nil {
			return database.ErrSchemaNotReady
		}
		if err := db.Ping(ctx); err != nil {
			return err
		}
		if err := database.CheckSchema(ctx, db); err != nil {
			return &logger.StageError{Component: "database", Operation: "readiness_schema", Reason: "schema_not_ready", Err: err}
		}
		return nil
	}})

	return &Container{
		Config:        cfg,
		Routes:        routesCfg,
		DB:            db,
		FiberApp:      app,
		Handlers:      handlers,
		OutboxService: outboxService,
	}, nil
}

// App returns the configured Fiber HTTP engine ready for listening or testing.
func (c *Container) App() *fiber.App {
	return c.FiberApp
}

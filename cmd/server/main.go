package main

import (
	"context"
	"log"

	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/database"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/gofiber/fiber/v2"
)

func main() {
	// Load application configuration.
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	// Create PostgreSQL connection pool.
	ctx := context.Background()

	db, err := database.NewPostgresPool(
		ctx,
		cfg.DatabaseURL,
	)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create repositories.
	purchaseRequestRepository :=
		repository.NewPostgresPurchaseRequestRepository(db)

	purchaseRequestItemRepository :=
		repository.NewPostgresPurchaseRequestItemRepository(db)

	transactionManager :=
		repository.NewPostgresTransactionManager(db)

	// Create service.
	purchaseRequestService :=
		service.NewPurchaseRequestService(
			purchaseRequestRepository,
			purchaseRequestItemRepository,
			transactionManager,
		)

	// Create handler.
	purchaseRequestHandler :=
		handler.NewPurchaseRequestHandler(
			purchaseRequestService,
		)

	// Create Purchase Order repositories.
	purchaseOrderRepository :=
		repository.NewPostgresPurchaseOrderRepository(db)

	purchaseOrderItemRepository :=
		repository.NewPostgresPurchaseOrderItemRepository(db)

	documentNumberRepository :=
		repository.NewPostgresDocumentNumberRepository(db)

	poTransactionManager :=
		repository.NewPostgresPurchaseOrderTransactionManager(db)

	// Create Purchase Order service.
	purchaseOrderService :=
		service.NewPurchaseOrderService(
			purchaseOrderRepository,
			purchaseOrderItemRepository,
			poTransactionManager,
		)

	// Create Purchase Order handler.
	purchaseOrderHandler :=
		handler.NewPurchaseOrderHandler(
			purchaseOrderService,
		)

	// Create Goods Receipt repositories.
	goodsReceiptRepository :=
		repository.NewPostgresGoodsReceiptRepository(db)

	goodsReceiptItemRepository :=
		repository.NewPostgresGoodsReceiptItemRepository(db)

	grnTransactionManager :=
		repository.NewPostgresGoodsReceiptTransactionManager(db)

	// Create Goods Receipt service.
	goodsReceiptService :=
		service.NewGoodsReceiptService(
			goodsReceiptRepository,
			goodsReceiptItemRepository,
			grnTransactionManager,
		)

	// Create Goods Receipt handler.
	goodsReceiptHandler :=
		handler.NewGoodsReceiptHandler(
			goodsReceiptService,
		)

	// Create Vendor Bill repositories.
	vendorBillRepository :=
		repository.NewPostgresVendorBillRepository(db)

	vendorBillItemRepository :=
		repository.NewPostgresVendorBillItemRepository(db)

	billTransactionManager :=
		repository.NewPostgresVendorBillTransactionManager(db)

	// Create Vendor Bill service.
	vendorBillService :=
		service.NewVendorBillService(
			vendorBillRepository,
			vendorBillItemRepository,
			billTransactionManager,
		)

	// Create Vendor Bill handler.
	vendorBillHandler :=
		handler.NewVendorBillHandler(
			vendorBillService,
		)

	// Create Purchase Return repositories.
	purchaseReturnRepository :=
		repository.NewPostgresPurchaseReturnRepository(db)

	purchaseReturnItemRepository :=
		repository.NewPostgresPurchaseReturnItemRepository(db)

	purchaseReturnTxManager :=
		repository.NewPostgresPurchaseReturnTransactionManager(db)

	// Create Purchase Return service.
	purchaseReturnService :=
		service.NewPurchaseReturnService(
			purchaseReturnRepository,
			purchaseReturnItemRepository,
			purchaseOrderRepository,
			purchaseOrderItemRepository,
			documentNumberRepository,
			purchaseReturnTxManager,
		)

	// Create Purchase Return handler.
	purchaseReturnHandler :=
		handler.NewPurchaseReturnHandler(
			purchaseReturnService,
		)

	// Create RFQ repositories.
	rfqRepository :=
		repository.NewPostgresRFQRepository(db)

	rfqItemRepository :=
		repository.NewPostgresRFQItemRepository(db)

	vendorQuotationRepository :=
		repository.NewPostgresVendorQuotationRepository(db)

	vendorQuotationItemRepository :=
		repository.NewPostgresVendorQuotationItemRepository(db)

	rfqTxManager :=
		repository.NewPostgresRFQTransactionManager(db)

	// Create RFQ service.
	rfqService :=
		service.NewRFQService(
			rfqRepository,
			rfqItemRepository,
			vendorQuotationRepository,
			vendorQuotationItemRepository,
			purchaseOrderRepository,
			purchaseOrderItemRepository,
			purchaseRequestRepository,
			purchaseRequestItemRepository,
			documentNumberRepository,
			rfqTxManager,
		)

	// Create RFQ handler.
	rfqHandler :=
		handler.NewRFQHandler(
			rfqService,
		)

	// Create Analytics repository.
	analyticsRepository :=
		repository.NewPostgresAnalyticsRepository(db)

	// Create Analytics service.
	analyticsService :=
		service.NewAnalyticsService(
			analyticsRepository,
		)

	// Create Analytics handler.
	analyticsHandler :=
		handler.NewAnalyticsHandler(
			analyticsService,
		)

	// Create Fiber application.
	app := fiber.New()

	// Purchase Request routes.
	app.Post(
		"/api/v1/purchase-requests",
		purchaseRequestHandler.Create,
	)

	app.Get(
		"/api/v1/purchase-requests",
		purchaseRequestHandler.List,
	)

	app.Get(
		"/api/v1/purchase-requests/:id",
		purchaseRequestHandler.GetByID,
	)

	app.Put(
		"/api/v1/purchase-requests/:id",
		purchaseRequestHandler.Update,
	)

	app.Post(
		"/api/v1/purchase-requests/:id/submit",
		purchaseRequestHandler.Submit,
	)

	app.Post(
		"/api/v1/purchase-requests/:id/approve",
		purchaseRequestHandler.Approve,
	)

	app.Post(
		"/api/v1/purchase-requests/:id/reject",
		purchaseRequestHandler.Reject,
	)

	// Purchase Request Item routes.
	app.Get(
		"/api/v1/purchase-requests/:id/items",
		purchaseRequestHandler.GetItems,
	)

	app.Post(
		"/api/v1/purchase-requests/:id/items",
		purchaseRequestHandler.AddItem,
	)

	app.Put(
		"/api/v1/purchase-requests/:id/items/:itemId",
		purchaseRequestHandler.UpdateItem,
	)

	app.Delete(
		"/api/v1/purchase-requests/:id/items/:itemId",
		purchaseRequestHandler.DeleteItem,
	)

	// Purchase Order routes.
	app.Post(
		"/api/v1/purchase-orders",
		purchaseOrderHandler.Create,
	)

	app.Post(
		"/api/v1/purchase-orders/from-request/:prId",
		purchaseOrderHandler.CreateFromPR,
	)

	app.Get(
		"/api/v1/purchase-orders",
		purchaseOrderHandler.List,
	)

	app.Get(
		"/api/v1/purchase-orders/:id",
		purchaseOrderHandler.GetByID,
	)

	app.Put(
		"/api/v1/purchase-orders/:id",
		purchaseOrderHandler.Update,
	)

	app.Post(
		"/api/v1/purchase-orders/:id/approve",
		purchaseOrderHandler.Approve,
	)

	app.Post(
		"/api/v1/purchase-orders/:id/issue",
		purchaseOrderHandler.Issue,
	)

	app.Post(
		"/api/v1/purchase-orders/:id/cancel",
		purchaseOrderHandler.Cancel,
	)

	// Purchase Order Item routes.
	app.Get(
		"/api/v1/purchase-orders/:id/items",
		purchaseOrderHandler.GetItems,
	)

	app.Post(
		"/api/v1/purchase-orders/:id/items",
		purchaseOrderHandler.AddItem,
	)

	app.Put(
		"/api/v1/purchase-orders/:id/items/:itemId",
		purchaseOrderHandler.UpdateItem,
	)

	app.Delete(
		"/api/v1/purchase-orders/:id/items/:itemId",
		purchaseOrderHandler.DeleteItem,
	)

	// Goods Receipt routes.
	app.Post(
		"/api/v1/goods-receipts",
		goodsReceiptHandler.Create,
	)

	app.Get(
		"/api/v1/goods-receipts",
		goodsReceiptHandler.List,
	)

	app.Get(
		"/api/v1/goods-receipts/:id",
		goodsReceiptHandler.GetByID,
	)

	app.Post(
		"/api/v1/goods-receipts/:id/cancel",
		goodsReceiptHandler.Cancel,
	)

	app.Get(
		"/api/v1/purchase-orders/:id/goods-receipts",
		goodsReceiptHandler.ListByPO,
	)

	// Vendor Bill routes.
	app.Post(
		"/api/v1/vendor-bills",
		vendorBillHandler.Create,
	)

	app.Get(
		"/api/v1/vendor-bills",
		vendorBillHandler.List,
	)

	app.Get(
		"/api/v1/vendor-bills/:id",
		vendorBillHandler.GetByID,
	)

	app.Post(
		"/api/v1/vendor-bills/:id/match",
		vendorBillHandler.Match,
	)

	app.Post(
		"/api/v1/vendor-bills/:id/approve",
		vendorBillHandler.Approve,
	)

	app.Post(
		"/api/v1/vendor-bills/:id/cancel",
		vendorBillHandler.Cancel,
	)

	app.Get(
		"/api/v1/purchase-orders/:id/vendor-bills",
		vendorBillHandler.ListByPO,
	)

	// Purchase Return routes.
	app.Post(
		"/api/v1/purchase-returns",
		purchaseReturnHandler.Create,
	)

	app.Get(
		"/api/v1/purchase-returns",
		purchaseReturnHandler.List,
	)

	app.Get(
		"/api/v1/purchase-returns/:id",
		purchaseReturnHandler.GetByID,
	)

	app.Get(
		"/api/v1/purchase-returns/number/:returnNumber",
		purchaseReturnHandler.GetByNumber,
	)

	app.Put(
		"/api/v1/purchase-returns/:id",
		purchaseReturnHandler.Update,
	)

	app.Post(
		"/api/v1/purchase-returns/:id/approve",
		purchaseReturnHandler.Approve,
	)

	app.Post(
		"/api/v1/purchase-returns/:id/ship",
		purchaseReturnHandler.Ship,
	)

	app.Post(
		"/api/v1/purchase-returns/:id/complete",
		purchaseReturnHandler.Complete,
	)

	app.Post(
		"/api/v1/purchase-returns/:id/cancel",
		purchaseReturnHandler.Cancel,
	)

	// RFQ routes.
	app.Post(
		"/api/v1/rfqs",
		rfqHandler.Create,
	)

	app.Post(
		"/api/v1/rfqs/from-request/:prId",
		rfqHandler.CreateFromPR,
	)

	app.Get(
		"/api/v1/rfqs",
		rfqHandler.List,
	)

	app.Get(
		"/api/v1/rfqs/:id",
		rfqHandler.GetByID,
	)

	app.Get(
		"/api/v1/rfqs/number/:rfqNumber",
		rfqHandler.GetByNumber,
	)

	app.Put(
		"/api/v1/rfqs/:id",
		rfqHandler.Update,
	)

	app.Post(
		"/api/v1/rfqs/:id/publish",
		rfqHandler.Publish,
	)

	app.Post(
		"/api/v1/rfqs/:id/close",
		rfqHandler.Close,
	)

	app.Post(
		"/api/v1/rfqs/:id/cancel",
		rfqHandler.Cancel,
	)

	app.Post(
		"/api/v1/rfqs/:id/quotations",
		rfqHandler.SubmitQuotation,
	)

	app.Get(
		"/api/v1/rfqs/:id/quotations",
		rfqHandler.ListQuotations,
	)

	app.Get(
		"/api/v1/rfqs/:id/quotations/:quotationId",
		rfqHandler.GetQuotationByID,
	)

	app.Get(
		"/api/v1/rfqs/:id/compare",
		rfqHandler.Compare,
	)

	app.Post(
		"/api/v1/rfqs/:id/award/:quotationId",
		rfqHandler.Award,
	)

	// Analytics routes.
	app.Get(
		"/api/v1/analytics/vendors/:vendorId/scorecard",
		analyticsHandler.GetVendorScorecard,
	)

	app.Get(
		"/api/v1/analytics/vendors/performance",
		analyticsHandler.GetVendorPerformanceRankings,
	)

	app.Get(
		"/api/v1/analytics/dashboard/kpis",
		analyticsHandler.GetDashboardKPIs,
	)

	app.Get(
		"/api/v1/analytics/spend/by-vendor",
		analyticsHandler.GetSpendByVendor,
	)

	// Start server.
	log.Printf(
		"purchasing service running on :%s",
		cfg.ServerPort,
	)

	log.Fatal(
		app.Listen(":" + cfg.ServerPort),
	)
}

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

	// Start server.
	log.Printf(
		"purchasing service running on :%s",
		cfg.ServerPort,
	)

	log.Fatal(
		app.Listen(":" + cfg.ServerPort),
	)
}

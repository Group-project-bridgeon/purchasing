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

	// Create Fiber application.
	app := fiber.New()

	// Purchase Request routes.
	app.Post(
		"/api/v1/purchase-requests",
		purchaseRequestHandler.Create,
	)
	app.Post(
		"/api/v1/purchase-requests",
		purchaseRequestHandler.Create,
	)

	app.Get(
		"/api/v1/purchase-requests/:id",
		purchaseRequestHandler.GetByID,
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

	// Start server.
	log.Printf(
		"purchasing service running on :%s",
		cfg.ServerPort,
	)

	log.Fatal(
		app.Listen(":" + cfg.ServerPort),
	)
}

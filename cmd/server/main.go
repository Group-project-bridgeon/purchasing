package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/app"
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/database"
)

func main() {
	// 1. Load application & route configurations.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load environment config: %v", err)
	}

	routesCfg, err := config.LoadRoutes()
	if err != nil {
		log.Fatalf("failed to load routes config: %v", err)
	}

	// 2. Initialize PostgreSQL connection pool.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Automatically run embedded database migrations.
	if err := database.RunMigrations(ctx, db); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

	// 4. Initialize Dependency Injection Container & Fiber Engine.
	container := app.NewContainer(db, cfg, routesCfg)
	server := container.App()

	// 5. Start background outbox domain event publisher.
	publisherCtx, stopPublisher := context.WithCancel(context.Background())
	defer stopPublisher()
	container.OutboxService.StartBackgroundPublisher(publisherCtx, 3*time.Second)

	// 6. Start HTTP server in a non-blocking goroutine.
	go func() {
		log.Printf("purchasing service running on :%s (api prefix: %s)", cfg.ServerPort, routesCfg.APIPrefix)
		if err := server.Listen(":" + cfg.ServerPort); err != nil {
			log.Printf("server listener stopped: %v", err)
		}
	}()

	// 7. Graceful shutdown on OS interrupt / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("received shutdown signal, gracefully closing server...")

	// Stop background outbox publisher
	stopPublisher()
	container.OutboxService.Stop()

	if err := server.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("error during graceful shutdown: %v", err)
	} else {
		log.Println("server shut down cleanly")
	}
}

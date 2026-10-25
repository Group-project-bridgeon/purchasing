package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/app"
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/database"
	"github.com/Group-project-bridgeon/purchasing/internal/logger"
)

func main() {
	// 1. Load application & route configurations.
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load environment config", slog.Any("error", err))
		os.Exit(1)
	}

	// Initialize production-grade structured logging.
	logger.Init(logger.Config{
		Environment: cfg.Environment,
		Level:       cfg.LogLevel,
		Format:      cfg.LogFormat,
	})

	routesCfg, err := config.LoadRoutes()
	if err != nil {
		slog.Error("failed to load routes config", slog.Any("error", err))
		os.Exit(1)
	}

	// 2. Initialize PostgreSQL connection pool.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	// 3. Automatically run embedded database migrations.
	if err := database.RunMigrations(ctx, db); err != nil {
		slog.Error("failed to run database migrations", slog.Any("error", err))
		os.Exit(1)
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
		slog.Info(
			"purchasing service running",
			slog.String("port", cfg.ServerPort),
			slog.String("api_prefix", routesCfg.APIPrefix),
			slog.String("metrics", "/metrics"),
			slog.String("docs", "/docs"),
		)
		if err := server.Listen(":" + cfg.ServerPort); err != nil {
			slog.Info("server listener stopped", slog.Any("error", err))
		}
	}()

	// 7. Graceful shutdown on OS interrupt / SIGTERM.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("received shutdown signal, gracefully closing server...")

	// Stop background outbox publisher
	stopPublisher()
	container.OutboxService.Stop()

	if err := server.ShutdownWithTimeout(10 * time.Second); err != nil {
		slog.Error("error during graceful shutdown", slog.Any("error", err))
	} else {
		slog.Info("server shut down cleanly")
	}
}

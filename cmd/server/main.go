package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Group-project-bridgeon/purchasing/internal/app"
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/database"
	"github.com/Group-project-bridgeon/purchasing/internal/logger"
)

func main() {
	logger.Init(logger.Config{Format: "json", Level: "info"})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		logger.Failure(context.Background(), "application", "startup_or_serve", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return &logger.StageError{Component: "configuration", Operation: "load", Reason: "configuration_invalid", SafeHint: err.Error(), Err: err}
	}
	logger.Init(logger.Config{Environment: cfg.Environment, Level: cfg.LogLevel, Format: cfg.LogFormat})
	routes, err := config.LoadRoutes()
	if err != nil {
		return &logger.StageError{Component: "configuration", Operation: "routes", Reason: "routes_invalid", Err: err}
	}
	startup, cancel := context.WithTimeout(ctx, cfg.StartupTimeout)
	defer cancel()
	db, err := database.NewPostgresPool(startup, cfg.DatabaseURL, cfg)
	if err != nil {
		return &logger.StageError{Component: "database", Operation: "startup_connect", Reason: "database_connection_failed", Err: err}
	}
	defer db.Close()
	if err = database.CheckSchema(startup, db); err != nil {
		return &logger.StageError{Component: "database", Operation: "startup_schema", Reason: "schema_not_ready", SafeHint: "requires reviewed clean golang-migrate schema 9", Err: err}
	}
	container, err := app.NewContainer(db, cfg, routes)
	if err != nil {
		return &logger.StageError{Component: "authentication", Operation: "startup_key", Reason: "jwt_configuration_invalid", Err: err}
	}
	// No production dispatcher exists. Config rejects enabled publishing; the
	// disabled service never starts a worker or acknowledges pending events.
	defer container.OutboxService.Stop()
	if err := app.Serve(ctx, container.App(), cfg.ListenAddress(), cfg.ShutdownTimeout); err != nil {
		return &logger.StageError{Component: "http", Operation: "serve", Reason: "listener_failed", Err: err}
	}
	return nil
}

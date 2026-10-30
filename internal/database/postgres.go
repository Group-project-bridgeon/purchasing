package database

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, databaseURL string, settings ...*config.Config) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}
	parsed, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	parsed.MaxConns = 10
	parsed.MinConns = 0
	parsed.ConnConfig.ConnectTimeout = 5 * time.Second
	operation := 5 * time.Second
	if len(settings) > 0 && settings[0] != nil {
		c := settings[0]
		parsed.MaxConns = c.DBMaxConns
		parsed.MinConns = c.DBMinConns
		parsed.MaxConnLifetime = c.DBMaxConnLifetime
		parsed.MaxConnIdleTime = c.DBMaxConnIdleTime
		parsed.HealthCheckPeriod = c.DBHealthCheckPeriod
		parsed.ConnConfig.ConnectTimeout = c.StartupTimeout
		operation = c.DBOperationTimeout
	}
	parsed.ConnConfig.RuntimeParams["statement_timeout"] = strconv.FormatInt(operation.Milliseconds(), 10)
	parsed.ConnConfig.RuntimeParams["lock_timeout"] = strconv.FormatInt(operation.Milliseconds(), 10)
	parsed.ConnConfig.RuntimeParams["idle_in_transaction_session_timeout"] = strconv.FormatInt((2 * operation).Milliseconds(), 10)
	pool, err := pgxpool.NewWithConfig(ctx, parsed)
	if err != nil {
		return nil, fmt.Errorf("create database pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

package database

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"strings"

	"github.com/Group-project-bridgeon/purchasing/db/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RunMigrations applies all unapplied embedded .up.sql database migrations in sequential order.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	// 1. Ensure the schema_migrations tracking table exists.
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(100) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("create schema_migrations table: %w", err)
	}

	// 2. Read all embedded migration files.
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return fmt.Errorf("read embedded migrations directory: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".up.sql") {
			upFiles = append(upFiles, entry.Name())
		}
	}

	// 3. Sort migrations in lexical ascending order (000001, 000002, ...).
	sort.Strings(upFiles)

	// 4. Apply each migration within an atomic transaction.
	for _, file := range upFiles {
		version := strings.TrimSuffix(file, ".up.sql")

		var exists bool
		err := pool.QueryRow(
			ctx,
			"SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)",
			version,
		).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration status for %s: %w", version, err)
		}

		if exists {
			continue
		}

		content, err := fs.ReadFile(migrations.FS, file)
		if err != nil {
			return fmt.Errorf("read migration file %s: %w", file, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin transaction for migration %s: %w", version, err)
		}

		if _, err := tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute migration %s: %w", version, err)
		}

		if _, err := tx.Exec(
			ctx,
			"INSERT INTO schema_migrations (version) VALUES ($1)",
			version,
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record applied migration %s: %w", version, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}

		log.Printf("Successfully applied database migration: %s", version)
	}

	return nil
}

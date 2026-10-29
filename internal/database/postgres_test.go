package database

import (
	"context"
	"testing"
)

const testDatabaseURL = "postgres://postgres:postgres@localhost:5433/purchasing?sslmode=disable"

func TestNewPostgresPool(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("NewPostgresPool() returned error: %v", err)
	}

	t.Cleanup(func() {
		pool.Close()
	})

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("database ping failed: %v", err)
	}
}

func TestNewPostgresPoolRequiresURL(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool(ctx, "")
	if err == nil {
		if pool != nil {
			pool.Close()
		}

		t.Fatal("expected error when database URL is empty")
	}
}

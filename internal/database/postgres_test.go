package database

import (
	"context"
	"testing"

	"github.com/Group-project-bridgeon/purchasing/internal/testfixture"
)

func TestNewPostgresPool(t *testing.T) {
	ctx := context.Background()

	pool, err := NewPostgresPool(ctx, testfixture.URL(t))
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

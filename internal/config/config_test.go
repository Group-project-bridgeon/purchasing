package config

import (
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY_PATH", "synthetic-public-key.pem")
	t.Setenv(
		"DATABASE_URL",
		"postgres://postgres:postgres@localhost:5433/purchasing?sslmode=disable",
	)
	t.Setenv("SERVER_PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.DatabaseURL == "" {
		t.Fatal("DatabaseURL should not be empty")
	}

	if cfg.ServerPort != "9090" {
		t.Fatalf("expected port 9090, got %s", cfg.ServerPort)
	}

	if cfg.Environment != "development" {
		t.Fatalf("expected default environment development, got %s", cfg.Environment)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected default log level info, got %s", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Fatalf("expected default log format json, got %s", cfg.LogFormat)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("SERVER_PORT", "")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DATABASE_URL is missing")
	}
}

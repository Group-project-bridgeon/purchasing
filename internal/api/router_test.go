package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/gofiber/fiber/v2"
)

func TestSetupRouter_SystemEndpoints(t *testing.T) {
	app := fiber.New()
	routesCfg, err := config.LoadRoutes()
	if err != nil {
		t.Fatalf("failed to load routes config: %v", err)
	}

	// For system endpoint testing, handlers can be empty or mocked
	SetupRouter(app, Handlers{}, routesCfg)

	// Test 1: Healthcheck
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error on /health: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /health, got %d", resp.StatusCode)
	}

	// Test 2: Swagger UI
	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error on /docs: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /docs, got %d", resp.StatusCode)
	}

	// Test 3: OpenAPI spec
	req = httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error on /docs/openapi.json: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from /docs/openapi.json, got %d", resp.StatusCode)
	}
}

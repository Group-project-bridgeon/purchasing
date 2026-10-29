package docs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestDocsRoutes(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app)

	// Test 1: GET /health
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to test /health: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /health, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var healthResp map[string]string
	if err := json.Unmarshal(body, &healthResp); err != nil {
		t.Fatalf("invalid json from /health: %v", err)
	}
	if healthResp["status"] != "ok" {
		t.Fatalf("expected status 'ok', got %q", healthResp["status"])
	}

	// Test 2: GET /docs (Swagger UI)
	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to test /docs: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /docs, got %d", resp.StatusCode)
	}
	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Fatalf("expected text/html, got %q", contentType)
	}

	// Test 3: GET /swagger (Redirect to /docs)
	req = httptest.NewRequest(http.MethodGet, "/swagger", nil)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to test /swagger: %v", err)
	}
	if resp.StatusCode != http.StatusMovedPermanently {
		t.Fatalf("expected 301 from /swagger, got %d", resp.StatusCode)
	}

	// Test 4: GET /docs/openapi.json
	req = httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil)
	resp, err = app.Test(req, -1)
	if err != nil {
		t.Fatalf("failed to test /docs/openapi.json: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK from /docs/openapi.json, got %d", resp.StatusCode)
	}

	body, _ = io.ReadAll(resp.Body)
	var openAPIMap map[string]interface{}
	if err := json.Unmarshal(body, &openAPIMap); err != nil {
		t.Fatalf("invalid json from /docs/openapi.json: %v", err)
	}
	if openAPIMap["openapi"] != "3.0.3" {
		t.Fatalf("expected openapi 3.0.3, got %v", openAPIMap["openapi"])
	}
}

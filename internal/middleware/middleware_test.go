package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestMiddleware_RequestID(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString(GetRequestID(c))
	})

	// 1. Without header -> should generate UUID
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	headerID := resp.Header.Get(HeaderXRequestID)
	if headerID == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != headerID {
		t.Fatalf("expected body %q to match header %q", string(body), headerID)
	}

	// 2. With existing header -> should preserve it
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set(HeaderXRequestID, "custom-trace-12345")
	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2.Header.Get(HeaderXRequestID) != "custom-trace-12345" {
		t.Fatalf("expected preserved X-Request-ID, got %q", resp2.Header.Get(HeaderXRequestID))
	}
}

func TestMiddleware_CORS(t *testing.T) {
	app := fiber.New()
	app.Use(CORS())

	app.Get("/data", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Preflight OPTIONS request
	req := httptest.NewRequest(http.MethodOptions, "/data", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content for OPTIONS, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("expected '*' CORS origin")
	}
}

func TestMiddleware_Recovery(t *testing.T) {
	app := fiber.New()
	app.Use(RequestID())
	app.Use(Recovery())

	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("simulated critical crash")
	})

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected 500 Internal Server Error, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var errMap map[string]interface{}
	if err := json.Unmarshal(body, &errMap); err != nil {
		t.Fatalf("expected valid JSON error response, got %s", string(body))
	}
	if errMap["code"] != float64(500) {
		t.Errorf("expected code 500, got %v", errMap["code"])
	}
}

package middleware

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestErrorsAndLogsNeverExposeRawInfrastructure(t *testing.T) {
	var capture bytes.Buffer
	old := slog.Default()
	t.Cleanup(func() { slog.SetDefault(old) })
	logger.Init(logger.Config{Format: "json", Output: &capture})
	const canary = "SYNTHETIC_SENSITIVE_VALUE_MUST_NOT_APPEAR"
	app := fiber.New(fiber.Config{ErrorHandler: ErrorHandler, BodyLimit: 1024})
	app.Use(RequestID(), RequestBounds(20*time.Millisecond), Recovery())
	app.Get("/db", func(c *fiber.Ctx) error {
		return c.Status(500).JSON(fiber.Map{"error": SafeError(c, &pgconn.PgError{Code: "28P01", Message: canary, Detail: canary})})
	})
	app.Get("/panic", func(c *fiber.Ctx) error { panic(canary) })
	app.Post("/body", func(c *fiber.Ctx) error { return c.SendStatus(204) })
	for _, path := range []string{"/db", "/panic"} {
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("X-Request-ID", strings.Repeat("x", 1000))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal("safe error request failed")
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 500 || strings.Contains(string(body), canary) || !validRequestID.MatchString(resp.Header.Get("X-Request-ID")) {
			t.Fatal("unsafe response or request ID")
		}
	}
	if strings.Contains(capture.String(), canary) || !strings.Contains(capture.String(), `"sqlstate":"28P01"`) || !strings.Contains(capture.String(), `"component":"http"`) {
		t.Fatal("unsafe or unhelpful captured diagnostics")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go func() { _ = app.Listener(listener) }()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = app.ShutdownWithContext(ctx)
		_ = listener.Close()
	})
	client := http.Client{Timeout: time.Second}
	resp, err := client.Post("http://"+listener.Addr().String()+"/body", "text/plain", strings.NewReader(strings.Repeat("x", 2048)))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 413 {
		t.Fatal("body bound not enforced")
	}
}

func TestRequestContextHasBoundedDeadline(t *testing.T) {
	app := fiber.New()
	app.Use(RequestBounds(5 * time.Millisecond))
	app.Get("/", func(c *fiber.Ctx) error {
		<-c.UserContext().Done()
		if c.UserContext().Err() != context.DeadlineExceeded {
			t.Fatal("wrong cancellation")
		}
		return c.SendStatus(204)
	})
	resp, err := app.Test(httptest.NewRequest("GET", "/", nil))
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Fatal("request budget failed")
	}
}

package api

import (
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestReadinessDependencyFailureIsBoundedAndSafe(t *testing.T) {
	for _, test := range []struct {
		name   string
		probe  ReadinessProbe
		status int
	}{
		{"ready", func(context.Context) error { return nil }, 200},
		{"unavailable", func(context.Context) error { return errors.New("SYNTHETIC_RAW_DATABASE_ERROR") }, 503},
		{"timeout", func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }, 503},
	} {
		t.Run(test.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/ready", readiness(test.probe, 5*time.Millisecond))
			start := time.Now()
			resp, err := app.Test(httptest.NewRequest("GET", "/ready", nil))
			if err != nil {
				t.Fatal(err)
			}
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode != test.status || strings.Contains(string(body), "RAW_DATABASE") || time.Since(start) > time.Second {
				t.Fatal("unsafe or unbounded readiness")
			}
		})
	}
}

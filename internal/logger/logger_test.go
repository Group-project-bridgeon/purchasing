package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestInit_JSON(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := Config{
		Environment: "production",
		Level:       "info",
		Format:      "json",
		Output:      buf,
	}

	l := Init(cfg)
	l.Info("server starting", "port", "8080")

	out := buf.String()
	if !strings.Contains(out, `"service":"purchasing-service"`) {
		t.Errorf("expected output to contain service attribute, got: %s", out)
	}
	if !strings.Contains(out, `"port":"8080"`) {
		t.Errorf("expected output to contain port, got: %s", out)
	}

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("expected valid JSON, got err: %v", err)
	}
}

func TestInit_Text(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := Config{
		Environment: "development",
		Level:       "debug",
		Format:      "text",
		Output:      buf,
	}

	l := Init(cfg)
	l.Debug("debug trace message", "key", "val")

	out := buf.String()
	if !strings.Contains(out, "debug trace message") {
		t.Errorf("expected text output to contain message, got: %s", out)
	}
	if !strings.Contains(out, "key=val") {
		t.Errorf("expected text output to contain key=val, got: %s", out)
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-test-123")

	if got := GetRequestID(ctx); got != "req-test-123" {
		t.Errorf("expected req-test-123, got: %s", got)
	}

	buf := &bytes.Buffer{}
	Init(Config{
		Environment: "test",
		Level:       "info",
		Format:      "json",
		Output:      buf,
	})

	l := FromContext(ctx)
	l.Info("context test")

	out := buf.String()
	if !strings.Contains(out, `"request_id":"req-test-123"`) {
		t.Errorf("expected log to contain request_id, got: %s", out)
	}
}

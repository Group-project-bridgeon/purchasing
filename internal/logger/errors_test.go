package logger

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
	"syscall"
	"testing"
)

func TestStageFailureKeepsUsefulCategoriesWithoutRawCause(t *testing.T) {
	previous := slog.Default()
	t.Cleanup(func() { slog.SetDefault(previous) })
	var output bytes.Buffer
	Init(Config{Environment: "test", Format: "json", Output: &output})
	Failure(WithRequestID(context.Background(), "safe-id"), "startup", "run", &StageError{Component: "configuration", Operation: "load", Reason: "configuration_invalid", SafeHint: "DATABASE_URL is required", Err: errors.New("synthetic-secret-canary")})
	for _, field := range []string{`"service":"purchasing-service"`, `"component":"configuration"`, `"operation":"load"`, `"reason":"configuration_invalid"`, `"request_id":"safe-id"`, `"configuration_issue":"DATABASE_URL is required"`} {
		if !strings.Contains(output.String(), field) {
			t.Fatal("stable startup diagnostic missing")
		}
	}
	if strings.Contains(output.String(), "synthetic-secret-canary") {
		t.Fatal("raw cause leaked")
	}
	if reason, _ := ErrorReason(syscall.EADDRINUSE); reason != "address_in_use" {
		t.Fatal("listener category missing")
	}
}

package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	UserIDKey    contextKey = "user_id"
)

// Config defines the configuration options for the structured logger.
type Config struct {
	Environment string // "production" | "development" | "test"
	Level       string // "debug" | "info" | "warn" | "error"
	Format      string // "json" | "text"
	Output      io.Writer
}

// Init initializes the global slog default logger according to the configuration.
func Init(cfg Config) *slog.Logger {
	var level slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.Environment == "development" || strings.ToLower(cfg.Level) == "debug",
	}

	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}

	var handler slog.Handler
	if strings.ToLower(cfg.Format) == "text" || (cfg.Format == "" && cfg.Environment == "development") {
		handler = slog.NewTextHandler(out, opts)
	} else {
		handler = slog.NewJSONHandler(out, opts)
	}

	// Attach service identity attributes
	handler = handler.WithAttrs([]slog.Attr{
		slog.String("service", "purchasing-service"),
		slog.String("env", cfg.Environment),
	})

	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}

// WithRequestID returns a copy of parent context with the request_id attached.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request_id from context if available.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}

// FromContext returns an *slog.Logger decorated with attributes found in the context.
func FromContext(ctx context.Context) *slog.Logger {
	l := slog.Default()
	if ctx == nil {
		return l
	}

	if reqID := GetRequestID(ctx); reqID != "" {
		l = l.With("request_id", reqID)
	}
	if uID, ok := ctx.Value(UserIDKey).(string); ok && uID != "" {
		l = l.With("user_id", uID)
	}

	return l
}

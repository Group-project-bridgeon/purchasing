package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is environment-only: application/test startup never reads local .env files.
type Config struct {
	DatabaseURL, ServerPort, ListenHost, Environment, LogLevel, LogFormat         string
	JWTPublicKeyPath, JWTIssuer, JWTAudience                                      string
	HTTPReadTimeout, HTTPWriteTimeout, HTTPIdleTimeout, RequestTimeout            time.Duration
	StartupTimeout, ReadinessTimeout, ShutdownTimeout                             time.Duration
	DBOperationTimeout, DBMaxConnLifetime, DBMaxConnIdleTime, DBHealthCheckPeriod time.Duration
	HTTPMaxBodyBytes                                                              int
	DBMaxConns, DBMinConns                                                        int32
	OutboxPublisherEnabled                                                        bool
}

func Load() (*Config, error) {
	c := &Config{DatabaseURL: strings.TrimSpace(os.Getenv("DATABASE_URL")), ServerPort: value("SERVER_PORT", "8080"), ListenHost: value("LISTEN_HOST", "0.0.0.0"), Environment: value("APP_ENV", "development"), LogLevel: value("LOG_LEVEL", "info"), LogFormat: value("LOG_FORMAT", "json")}
	if c.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	port, err := strconv.Atoi(c.ServerPort)
	if err != nil || port < 1 || port > 65535 || strconv.Itoa(port) != c.ServerPort {
		return nil, fmt.Errorf("SERVER_PORT must be a canonical integer between 1 and 65535")
	}
	if net.ParseIP(c.ListenHost) == nil {
		return nil, fmt.Errorf("LISTEN_HOST must be an IP address")
	}
	if c.Environment != "development" && c.Environment != "test" && c.Environment != "production" {
		return nil, fmt.Errorf("APP_ENV must be development, test or production")
	}
	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		return nil, fmt.Errorf("LOG_LEVEL is invalid")
	}
	if c.LogFormat != "json" && c.LogFormat != "text" {
		return nil, fmt.Errorf("LOG_FORMAT is invalid")
	}
	for _, setting := range []struct {
		name, fallback string
		target         *time.Duration
		max            time.Duration
	}{
		{"HTTP_READ_TIMEOUT", "5s", &c.HTTPReadTimeout, time.Minute}, {"HTTP_WRITE_TIMEOUT", "15s", &c.HTTPWriteTimeout, time.Minute}, {"HTTP_IDLE_TIMEOUT", "60s", &c.HTTPIdleTimeout, 5 * time.Minute},
		{"REQUEST_TIMEOUT", "5s", &c.RequestTimeout, 30 * time.Second}, {"STARTUP_TIMEOUT", "15s", &c.StartupTimeout, 2 * time.Minute}, {"READINESS_TIMEOUT", "1s", &c.ReadinessTimeout, 5 * time.Second}, {"SHUTDOWN_TIMEOUT", "10s", &c.ShutdownTimeout, time.Minute},
		{"DB_OPERATION_TIMEOUT", "5s", &c.DBOperationTimeout, 30 * time.Second}, {"DB_MAX_CONN_LIFETIME", "30m", &c.DBMaxConnLifetime, 24 * time.Hour}, {"DB_MAX_CONN_IDLE_TIME", "5m", &c.DBMaxConnIdleTime, time.Hour}, {"DB_HEALTH_CHECK_PERIOD", "1m", &c.DBHealthCheckPeriod, 5 * time.Minute},
	} {
		v, e := time.ParseDuration(value(setting.name, setting.fallback))
		if e != nil || v < time.Millisecond || v > setting.max {
			return nil, fmt.Errorf("%s duration is outside the supported bounds", setting.name)
		}
		*setting.target = v
	}
	if c.HTTPWriteTimeout <= c.RequestTimeout || c.RequestTimeout < c.DBOperationTimeout || c.DBMaxConnIdleTime > c.DBMaxConnLifetime {
		return nil, fmt.Errorf("HTTP/database timeout ordering is invalid")
	}
	for _, setting := range []struct {
		name               string
		fallback, min, max int
		target             *int
	}{{"HTTP_MAX_BODY_BYTES", 1048576, 1024, 16777216, &c.HTTPMaxBodyBytes}} {
		v, e := strconv.Atoi(value(setting.name, strconv.Itoa(setting.fallback)))
		if e != nil || v < setting.min || v > setting.max {
			return nil, fmt.Errorf("%s is outside the supported bounds", setting.name)
		}
		*setting.target = v
	}
	max, e := strconv.Atoi(value("DB_MAX_CONNS", "10"))
	if e != nil || max < 1 || max > 100 {
		return nil, fmt.Errorf("DB_MAX_CONNS must be 1-100")
	}
	min, e := strconv.Atoi(value("DB_MIN_CONNS", "0"))
	if e != nil || min < 0 || min > max {
		return nil, fmt.Errorf("DB_MIN_CONNS must be 0-DB_MAX_CONNS")
	}
	c.DBMaxConns = int32(max)
	c.DBMinConns = int32(min)
	switch value("OUTBOX_PUBLISHER_ENABLED", "false") {
	case "false":
	case "true":
		return nil, fmt.Errorf("OUTBOX_PUBLISHER_ENABLED cannot be enabled without an approved production dispatcher")
	default:
		return nil, fmt.Errorf("OUTBOX_PUBLISHER_ENABLED must be false")
	}
	c.JWTPublicKeyPath = strings.TrimSpace(os.Getenv("JWT_PUBLIC_KEY_PATH"))
	c.JWTIssuer = value("JWT_ISSUER", "erp-auth")
	c.JWTAudience = value("JWT_AUDIENCE", "erp-api")
	if c.JWTPublicKeyPath == "" {
		return nil, fmt.Errorf("JWT_PUBLIC_KEY_PATH is required")
	}
	if c.JWTAudience != "erp-api" {
		return nil, fmt.Errorf("JWT_AUDIENCE must be erp-api")
	}
	for _, name := range []string{"JWT_ISSUER", "JWT_AUDIENCE"} {
		raw := os.Getenv(name)
		if raw != "" && raw != strings.TrimSpace(raw) {
			return nil, fmt.Errorf("%s must not have surrounding whitespace", name)
		}
	}
	return c, nil
}
func value(name, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(name)); v != "" {
		return v
	}
	return fallback
}
func (c *Config) ListenAddress() string { return net.JoinHostPort(c.ListenHost, c.ServerPort) }

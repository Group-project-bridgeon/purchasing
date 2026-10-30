package config

import (
	"testing"
)

func TestOperationalBounds(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY_PATH", "synthetic-public-key.pem")
	t.Setenv("DATABASE_URL", "postgres://fixture@127.0.0.1:15432/purchasing_fixture?sslmode=disable")
	for _, tc := range []struct{ key, value string }{
		{"SERVER_PORT", "0"}, {"SERVER_PORT", "65536"}, {"SERVER_PORT", "08080"},
		{"LISTEN_HOST", "untrusted-host"}, {"HTTP_READ_TIMEOUT", "0s"},
		{"HTTP_WRITE_TIMEOUT", "5s"}, {"REQUEST_TIMEOUT", "1ms"},
		{"HTTP_MAX_BODY_BYTES", "999999999"}, {"DB_MAX_CONNS", "0"},
		{"DB_MIN_CONNS", "101"}, {"DB_OPERATION_TIMEOUT", "31s"},
		{"OUTBOX_PUBLISHER_ENABLED", "true"}, {"OUTBOX_PUBLISHER_ENABLED", "TRUE"},
		{"SHUTDOWN_TIMEOUT", "0s"}, {"LOG_FORMAT", "invalid"},
	} {
		t.Run(tc.key+"/"+tc.value, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			if _, err := Load(); err == nil {
				t.Fatal("invalid bounds accepted")
			}
		})
	}
	c, err := Load()
	if err != nil || c.OutboxPublisherEnabled || c.ListenAddress() != "0.0.0.0:8080" {
		t.Fatal("invalid safe defaults", err)
	}
}

func TestConfigurationDoesNotLoadDotenv(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("configuration unexpectedly read local dotenv or accepted missing database URL")
	}
}

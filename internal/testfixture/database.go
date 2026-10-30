// Package testfixture refuses database tests without an explicit run-owned target.
package testfixture

import (
	"context"
	"errors"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

var runPattern = regexp.MustCompile(`^[a-f0-9]{32}$`)
var passwordPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

func targetConfig(raw, run string) (*pgx.ConnConfig, error) {
	u, err := url.Parse(raw)
	if err != nil || !runPattern.MatchString(run) || u == nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") || u.User == nil {
		return nil, errors.New("invalid fixture target")
	}
	password, explicit := u.User.Password()
	if !explicit || !passwordPattern.MatchString(password) || u.User.Username() != "purchasing_fixture" || u.Hostname() != "127.0.0.1" || u.Path != "/purchasing_test_"+run || u.Query().Get("sslmode") != "disable" || u.Fragment != "" {
		return nil, errors.New("invalid fixture target")
	}
	for key, values := range u.Query() {
		if len(values) != 1 {
			return nil, errors.New("ambiguous fixture connection option")
		}
		switch key {
		case "sslmode", "statement_timeout", "lock_timeout", "idle_in_transaction_session_timeout":
		default:
			return nil, errors.New("unapproved fixture connection option")
		}
	}
	cfg, err := pgx.ParseConfig(raw)
	if err != nil || cfg.Host != "127.0.0.1" || cfg.Port == 5432 || cfg.Port == 5433 || cfg.Port == 0 || len(cfg.Fallbacks) != 0 || cfg.TLSConfig != nil {
		return nil, errors.New("invalid fixture target")
	}
	cfg.ConnectTimeout = 3 * time.Second
	return cfg, nil
}

// URL never reads DATABASE_URL or local dotenv files. The ownership marker is
// created only by the disposable harness, before any test can delete fixture rows.
func URL(t *testing.T) string {
	t.Helper()
	run, raw := os.Getenv("PURCHASING_TEST_RUN_ID"), os.Getenv("PURCHASING_TEST_DATABASE_URL")
	if run == "" && raw == "" {
		t.Skip("requires the run-owned Purchasing PostgreSQL fixture")
	}
	cfg, err := targetConfig(raw, run)
	if err != nil {
		t.Fatal("invalid or unowned Purchasing test target; connection details withheld")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := pgx.ConnectConfig(ctx, cfg)
	if err != nil {
		t.Fatal("Purchasing test ownership connection failed; details withheld")
	}
	defer func() {
		closeCtx, stop := context.WithTimeout(context.Background(), time.Second)
		defer stop()
		_ = conn.Close(closeCtx)
	}()
	var owner string
	err = conn.QueryRow(ctx, "SELECT run_id FROM public.purchasing_fixture_ownership WHERE run_id=$1", run).Scan(&owner)
	if err != nil || owner != run {
		t.Fatal("Purchasing test ownership marker mismatch; refusing database tests")
	}
	return raw
}

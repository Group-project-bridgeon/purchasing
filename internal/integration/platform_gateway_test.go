package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/api"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type gatewayState struct{ RunID, BaseURL, AuthDSN, PurchasingDSN, Password, CAFile, PrivateKey string }
type purchasingGateway struct {
	t        *testing.T
	ctx      context.Context
	state    gatewayState
	client   *http.Client
	auth, db *pgxpool.Pool
}

func (g *purchasingGateway) check(ok bool, reason string) {
	g.t.Helper()
	if !ok {
		g.t.Fatal(reason)
	}
}
func (g *purchasingGateway) sql(pool *pgxpool.Pool, query string, args ...any) {
	g.t.Helper()
	_, err := pool.Exec(g.ctx, query, args...)
	if err != nil {
		state := "none"
		var pg *pgconn.PgError
		if errors.As(err, &pg) {
			state = pg.Code
		}
		g.t.Fatalf("owned fixture SQL failed (SQLSTATE %s); raw diagnostics withheld", state)
	}
}
func (g *purchasingGateway) call(method, path, token string, body any, want int) map[string]any {
	g.t.Helper()
	var data []byte
	var err error
	if body != nil {
		data, err = json.Marshal(body)
		g.check(err == nil, "fixture JSON failed")
	}
	req, err := http.NewRequestWithContext(g.ctx, method, g.state.BaseURL+path, bytes.NewReader(data))
	g.check(err == nil, "fixture request invalid")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := g.client.Do(req)
	g.check(err == nil, "verified HTTPS request failed; URL/token/error withheld")
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, (1<<20)+1))
	g.check(err == nil && len(b) <= 1<<20, "fixture response read failed")
	g.check(strings.Contains(strings.ToLower(res.Header.Get("Via")), "kong") || strings.Contains(strings.ToLower(res.Header.Get("Server")), "kong") || res.Header.Get("X-Kong-Request-ID") != "", "response did not traverse Kong")
	g.check(res.StatusCode == want, fmt.Sprintf("%s %s got HTTP %d want %d; body withheld", method, path, res.StatusCode, want))
	g.t.Logf("%s %s HTTP %d", method, path, want)
	var out map[string]any
	_ = json.Unmarshal(b, &out)
	return out
}
func (g *purchasingGateway) login(permissions []string, scope string) (uuid.UUID, string) {
	g.t.Helper()
	random := make([]byte, 32)
	_, err := rand.Read(random)
	g.check(err == nil, "random generation failed")
	password := hex.EncodeToString(random)
	email := "purchasing-" + uuid.NewString() + "@example.invalid"
	response := g.call("POST", "/api/v1/auth/register", "", map[string]string{"name": "Synthetic Purchasing User", "email": email, "password": password}, 201)
	user, ok := response["user"].(map[string]any)
	g.check(ok, "register shape mismatch")
	id, err := uuid.Parse(fmt.Sprint(user["id"]))
	g.check(err == nil, "register identity invalid")
	// Explicit synthetic email verification and grants, never normal provisioning.
	g.sql(g.auth, `UPDATE identity_users SET email_verified_at=now() WHERE id=$1`, id)
	if len(permissions) > 0 {
		role := uuid.New()
		g.sql(g.auth, `INSERT INTO identity_roles(id,name) VALUES($1,$2)`, role, "pgw-"+role.String())
		unique := map[string]bool{}
		for _, p := range permissions {
			if unique[p] {
				continue
			}
			unique[p] = true
			parts := strings.Split(p, ":")
			g.check(len(parts) == 2, "invalid fixture grant")
			tag, err := g.auth.Exec(g.ctx, `INSERT INTO identity_role_permissions(role_id,permission_id) SELECT $1,id FROM identity_permissions WHERE resource=$2 AND action=$3`, role, parts[0], parts[1])
			g.check(err == nil && tag.RowsAffected() == 1, "exact approved permission missing")
		}
		var scopeID any
		if scope == "WAREHOUSE" {
			scopeID = uuid.New()
		}
		g.sql(g.auth, `INSERT INTO identity_user_roles(user_id,role_id,scope_type,scope_id) VALUES($1,$2,$3,$4)`, id, role, scope, scopeID)
	}
	response = g.call("POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": password}, 200)
	token, ok := response["access_token"].(string)
	g.check(ok && token != "" && response["token_type"] == "Bearer", "real Auth token missing")
	me := g.call("GET", "/api/v1/auth/me", token, nil, 200)
	g.check(fmt.Sprint(me["id"]) == id.String(), "actual Auth identity mismatch")
	return id, token
}
func openPurchasingGateway(t *testing.T) *purchasingGateway {
	t.Helper()
	isolated, path := os.Getenv("PURCHASING_GATEWAY_ISOLATED"), os.Getenv("PURCHASING_GATEWAY_STATE")
	if isolated == "" && path == "" {
		t.Skip("requires owned Purchasing HTTPS/Kong fixture")
	}
	if isolated != "1" || !filepath.IsAbs(path) {
		t.Fatal("explicit owned gateway fixture required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal("fixture state missing")
	}
	var state gatewayState
	if json.Unmarshal(raw, &state) != nil {
		t.Fatal("fixture state invalid")
	}
	directory := filepath.Dir(path)
	if !regexp.MustCompile(`^purchasing-gateway-[0-9a-f]{32}$`).MatchString(state.RunID) || filepath.Base(directory) != state.RunID || !strings.EqualFold(filepath.Clean(filepath.Dir(directory)), filepath.Clean(os.TempDir())) {
		t.Fatal("fixture directory/run identity invalid")
	}
	owner, err := os.ReadFile(filepath.Join(directory, "owner.txt"))
	if err != nil || string(owner) != state.RunID {
		t.Fatal("fixture ownership marker invalid")
	}
	for _, p := range []string{state.CAFile, state.PrivateKey} {
		relative, err := filepath.Rel(directory, p)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) || filepath.IsAbs(relative) {
			t.Fatal("fixture key path escaped owned directory")
		}
	}
	base, err := url.Parse(state.BaseURL)
	if err != nil || base.Scheme != "https" || base.Hostname() != "localhost" || base.Port() == "" || base.User != nil || base.Path != "" || base.RawQuery != "" {
		t.Fatal("fixture gateway must use explicit loopback HTTPS")
	}
	ca, err := os.ReadFile(state.CAFile)
	if err != nil {
		t.Fatal("fixture public CA missing")
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(ca) {
		t.Fatal("fixture public CA invalid")
	}
	transport := &http.Transport{TLSClientConfig: &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}, DialContext: (&net.Dialer{Timeout: 3 * time.Second}).DialContext, TLSHandshakeTimeout: 3 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)
	t.Cleanup(transport.CloseIdleConnections)
	g := &purchasingGateway{t: t, ctx: ctx, state: state, client: &http.Client{Timeout: 8 * time.Second, Transport: transport, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
	for i, d := range []struct{ raw, user, database string }{{state.AuthDSN, "auth_fixture", "auth_purchasing_gateway"}, {state.PurchasingDSN, "purchasing_service", "purchasing_db"}} {
		u, err := url.Parse(d.raw)
		g.check(err == nil && u.Scheme == "postgres" && u.Hostname() == "127.0.0.1" && u.Port() != "" && u.Port() != "5432" && u.Port() != "5433" && u.User != nil && u.User.Username() == d.user && u.Path == "/"+d.database, "unsafe fixture DB target")
		_, hasPassword := u.User.Password()
		g.check(!hasPassword, "fixture URI must not embed password")
		cfg, err := pgxpool.ParseConfig(d.raw)
		g.check(err == nil, "fixture DB config invalid")
		cfg.ConnConfig.Password = state.Password
		cfg.ConnConfig.ConnectTimeout = 3 * time.Second
		cfg.ConnConfig.RuntimeParams["statement_timeout"] = "5000"
		cfg.ConnConfig.RuntimeParams["lock_timeout"] = "2000"
		cfg.MaxConns = 4
		pool, err := pgxpool.NewWithConfig(ctx, cfg)
		g.check(err == nil, "fixture pool invalid")
		t.Cleanup(pool.Close)
		var marker string
		err = pool.QueryRow(ctx, `SELECT run_id FROM purchasing_gateway_fixture_ownership`).Scan(&marker)
		g.check(err == nil && marker == state.RunID, "fixture DB ownership mismatch")
		if i == 0 {
			g.auth = pool
		} else {
			g.db = pool
		}
	}
	return g
}

func TestPurchasingGateway(t *testing.T) {
	g := openPurchasingGateway(t)
	const prefix = "/api/v1/purchasing"
	ctx := g.ctx
	var count int
	err := g.auth.QueryRow(ctx, `SELECT count(*) FROM identity_role_permissions rp JOIN identity_permissions p ON p.id=rp.permission_id WHERE p.resource LIKE 'purchasing.%'`).Scan(&count)
	g.check(err == nil && count == 0, "definitions unexpectedly granted Purchasing authority")
	var version int
	var dirty bool
	err = g.auth.QueryRow(ctx, `SELECT version,dirty FROM schema_migrations`).Scan(&version, &dirty)
	g.check(err == nil && version == 13 && !dirty, "clean Auth schema 13 missing")
	err = g.auth.QueryRow(ctx, `SELECT count(*) FROM identity_permissions WHERE resource LIKE 'purchasing.%'`).Scan(&count)
	g.check(err == nil && count == 41, "exactly 41 Purchasing definitions required")
	err = g.db.QueryRow(ctx, `SELECT version,dirty FROM schema_migrations`).Scan(&version, &dirty)
	g.check(err == nil && version == 9 && !dirty, "clean Purchasing schema 9 missing")
	t.Log("PASS: clean Auth 13/Purchasing 9, 41 Purchasing definitions and no automatic grants")
	event := uuid.New()
	g.sql(g.db, `INSERT INTO outbox_events(id,aggregate_type,aggregate_id,event_type,payload) VALUES($1,'fixture','fixture','fixture.pending','{}')`, event)
	g.call("GET", prefix+"/purchase-orders", "", nil, 401)
	g.call("GET", prefix+"/purchase-orders", "invalid", nil, 401)
	for _, path := range []string{"/else/api/v1/purchasing/purchase-orders", "/api/v1/purchasing-extra/purchase-orders", "/api/v1/purchase-orders", "/api/v1/party/namespace-probe"} {
		g.call("GET", path, "", nil, 404)
	}
	_, noGrant := g.login(nil, "GLOBAL")
	g.call("GET", prefix+"/purchase-orders", noGrant, nil, 403)
	all := []string{"catalog.units:read", "inventory.warehouses:read"}
	for _, grants := range api.PermissionMatrix() {
		all = append(all, grants...)
	}
	creator, creatorToken := g.login(all, "GLOBAL")
	approver, approverToken := g.login(all, "GLOBAL")
	g.call("GET", "/api/v1/catalog/units", creatorToken, nil, 200)
	g.call("GET", "/api/v1/inventory/warehouses", creatorToken, nil, 200)
	_, warehouseToken := g.login([]string{"purchasing.purchase-orders:read"}, "WAREHOUSE")
	g.call("GET", prefix+"/purchase-orders", warehouseToken, nil, 403)
	// Only negative cases are locally signed; all positive authority comes from Auth.
	private, err := os.ReadFile(g.state.PrivateKey)
	g.check(err == nil, "negative fixture key missing")
	key, err := jwt.ParseRSAPrivateKeyFromPEM(private)
	g.check(err == nil, "negative fixture key invalid")
	for _, kind := range []string{"expired", "machine"} {
		claims := jwt.MapClaims{}
		_, _, err = jwt.NewParser().ParseUnverified(creatorToken, claims)
		g.check(err == nil, "real token shape invalid")
		if kind == "expired" {
			claims["iat"] = time.Now().Add(-20 * time.Minute).Unix()
			claims["nbf"] = claims["iat"]
			claims["exp"] = time.Now().Add(-10 * time.Minute).Unix()
		} else {
			claims["token_kind"] = "machine"
		}
		negative, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
		g.check(err == nil, "negative fixture signing failed")
		g.call("GET", prefix+"/purchase-orders", negative, nil, 401)
	}
	id := uuid.NewString()
	for _, tc := range []struct {
		permissions []string
		paths       []string
	}{
		{[]string{"purchasing.purchase-orders:create"}, []string{"/purchase-orders/from-request/" + id, "/rfqs/" + id + "/award/" + id}},
		{[]string{"purchasing.purchase-requests:read"}, []string{"/purchase-orders/from-request/" + id, "/rfqs/from-request/" + id}},
		{[]string{"purchasing.rfqs:create"}, []string{"/rfqs/from-request/" + id}},
		{[]string{"purchasing.rfqs:award"}, []string{"/rfqs/" + id + "/award/" + id}},
	} {
		_, token := g.login(tc.permissions, "GLOBAL")
		for _, path := range tc.paths {
			g.call("POST", prefix+path, token, map[string]any{}, 403)
		}
	}
	g.call("POST", prefix+"/purchase-orders", creatorToken, map[string]any{"created_by": approver.String()}, 403)
	prResult := g.call("POST", prefix+"/purchase-requests", creatorToken, map[string]any{"purpose": "synthetic gateway request", "items": []map[string]any{{"item_id": uuid.NewString(), "description": "fixture", "quantity": "2"}}}, 201)
	pr := prResult["purchase_request"].(map[string]any)
	prID := pr["id"].(string)
	g.check(pr["requested_by"] == creator.String(), "request creator not trusted")
	g.call("POST", prefix+"/purchase-requests/"+prID+"/submit", creatorToken, nil, 200)
	g.call("POST", prefix+"/purchase-requests/"+prID+"/approve", creatorToken, nil, 403)
	g.call("POST", prefix+"/purchase-requests/"+prID+"/approve", approverToken, nil, 200)
	g.call("PUT", prefix+"/purchase-requests/"+prID, approverToken, map[string]any{"purpose": "blocked"}, 409)
	g.call("POST", prefix+"/purchase-requests/"+prID+"/items", approverToken, map[string]any{"item_id": uuid.NewString(), "quantity": "1"}, 409)
	vendor := uuid.NewString()
	result := g.call("POST", prefix+"/purchase-orders", creatorToken, map[string]any{"vendor_id": vendor, "items": []map[string]any{{"item_id": uuid.NewString(), "description": "fixture", "quantity": "2", "unit_price": "10"}}}, 201)
	po := result["purchase_order"].(map[string]any)
	poID := po["id"].(string)
	g.check(po["created_by"] == creator.String(), "order creator not trusted")
	// A different newer vendor catches lost filtering, not just lost pagination.
	g.call("POST", prefix+"/purchase-orders", creatorToken, map[string]any{"vendor_id": uuid.NewString(), "items": []map[string]any{{"item_id": uuid.NewString(), "quantity": "1", "unit_price": "1"}}}, 201)
	query := g.call("GET", prefix+"/purchase-orders?vendor_id="+vendor+"&limit=1&offset=0", creatorToken, nil, 200)
	g.check(query["limit"] == float64(1) && query["offset"] == float64(0), "query limit/offset not forwarded")
	orders := query["purchase_orders"].([]any)
	g.check(len(orders) == 1 && orders[0].(map[string]any)["id"] == poID, "vendor filter not forwarded")
	g.call("GET", prefix+"/purchase-orders?limit=invalid", creatorToken, nil, 400)
	g.call("POST", prefix+"/purchase-orders/"+poID+"/approve", creatorToken, nil, 403)
	g.call("POST", prefix+"/purchase-orders/"+poID+"/approve", approverToken, nil, 200)
	g.call("POST", prefix+"/purchase-orders/"+poID+"/issue", approverToken, nil, 200)
	g.call("PUT", prefix+"/purchase-orders/"+poID, approverToken, map[string]any{"currency": "EUR"}, 409)
	g.call("POST", prefix+"/purchase-orders/"+poID+"/items", approverToken, map[string]any{"item_id": uuid.NewString(), "quantity": "1", "unit_price": "1"}, 409)
	var by uuid.UUID
	err = g.db.QueryRow(ctx, `SELECT approved_by FROM purchase_orders WHERE id=$1`, poID).Scan(&by)
	g.check(err == nil && by == approver, "actual approver not persisted")
	for _, path := range []string{"/outbox/events", "/outbox/events/" + id} {
		g.call("GET", prefix+path, creatorToken, nil, 404)
	}
	g.call("POST", prefix+"/outbox/events/"+id+"/retry", creatorToken, map[string]any{}, 404)
	for _, path := range []string{"/health", "/ready", "/docs", "/metrics"} {
		g.call("GET", path, "", nil, 404)
	}
	for _, path := range []string{"/ready", "/docs", "/metrics"} {
		g.call("GET", prefix+path, creatorToken, nil, 404)
	}
	// The existing public /api/v1/health is intentionally observed, not reclassified.
	g.call("GET", prefix+"/health", "", nil, 200)
	deadline := time.NewTimer(4 * time.Second)
	defer deadline.Stop()
	select {
	case <-deadline.C:
	case <-ctx.Done():
		t.Fatal("fixture timed out")
	}
	var status string
	var published *time.Time
	var retries int
	err = g.db.QueryRow(ctx, `SELECT status,published_at,retry_count FROM outbox_events WHERE id=$1`, event).Scan(&status, &published, &retries)
	g.check(err == nil && status == "pending" && published == nil && retries == 0, "disabled Outbox acknowledged pending fixture event")
	t.Log("PASS: real Auth HTTPS issuance, exact grants, compound rejection, trusted actors, immutable approval/lifecycle, Kong translation/query forwarding and Auth/Catalog/Inventory regressions; Party absent from inspected source")
}

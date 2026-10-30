package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/api"
	"github.com/Group-project-bridgeon/purchasing/internal/app"
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/database"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/Group-project-bridgeon/purchasing/internal/testfixture"
	"github.com/google/uuid"
)

// This is real PostgreSQL plus the actual in-process HTTP stack, using a
// synthetic RS256 issuer. It does not claim real Auth issuance or Kong coverage.
func TestPrerequisiteHTTPAndDisabledOutbox(t *testing.T) {
	raw := testfixture.URL(t)
	signer := testfixture.NewSigner(t)
	t.Setenv("DATABASE_URL", raw)
	t.Setenv("JWT_PUBLIC_KEY_PATH", signer.PublicPath)
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_FORMAT", "json")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal("owned configuration failed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := database.NewPostgresPool(ctx, raw, cfg)
	if err != nil {
		t.Fatal("owned pool failed")
	}
	defer pool.Close()
	if err = database.CheckSchema(ctx, pool); err != nil {
		t.Fatal("owned schema not ready")
	}
	routes, err := config.LoadRoutes()
	if err != nil {
		t.Fatal(err)
	}
	container, err := app.NewContainer(pool, cfg, routes)
	if err != nil {
		t.Fatal("owned app wiring failed")
	}
	defer container.OutboxService.Stop()
	var all []string
	for _, p := range api.PermissionMatrix() {
		all = append(all, p...)
	}
	creatorToken := signer.Token(t, all, nil)
	creator := signer.UserID
	signer.UserID = uuid.New()
	approver := signer.UserID
	approverToken := signer.Token(t, all, nil)
	call := func(method, path, token string, payload any, want int) map[string]any {
		t.Helper()
		var data []byte
		if payload != nil {
			data, _ = json.Marshal(payload)
		}
		req := httptest.NewRequest(method, path, bytes.NewReader(data))
		if payload != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		resp, e := container.App().Test(req, 10000)
		if e != nil {
			t.Fatal("owned HTTP request failed")
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != want {
			t.Fatalf("%s %s: got %d want %d (response withheld)", method, path, resp.StatusCode, want)
		}
		var out map[string]any
		_ = json.Unmarshal(body, &out)
		return out
	}
	call("GET", "/ready", "", nil, 200)
	request := call("POST", "/api/v1/purchase-requests", creatorToken, map[string]any{"purpose": "synthetic prerequisite", "items": []map[string]any{{"item_id": uuid.NewString(), "description": "synthetic", "quantity": "2"}}}, 201)
	pr := request["purchase_request"].(map[string]any)
	id := pr["id"].(string)
	if pr["requested_by"] != creator.String() {
		t.Fatal("request creator not from JWT")
	}
	call("POST", "/api/v1/purchase-requests/"+id+"/submit", creatorToken, nil, 200)
	call("POST", "/api/v1/purchase-requests/"+id+"/approve", creatorToken, nil, 403)
	call("POST", "/api/v1/purchase-requests/"+id+"/approve", approverToken, nil, 200)
	var recorded uuid.UUID
	if err = pool.QueryRow(ctx, `SELECT approved_by FROM purchase_requests WHERE id=$1`, id).Scan(&recorded); err != nil || recorded != approver {
		t.Fatal("actual approver not persisted")
	}
	call("PUT", "/api/v1/purchase-requests/"+id, approverToken, map[string]any{"purpose": "blocked edit"}, 409)
	call("POST", "/api/v1/purchase-requests/"+id+"/items", approverToken, map[string]any{"item_id": uuid.NewString(), "quantity": "1"}, 409)
	order := call("POST", "/api/v1/purchase-orders", creatorToken, map[string]any{"vendor_id": uuid.NewString(), "items": []map[string]any{{"item_id": uuid.NewString(), "description": "synthetic", "quantity": "2", "unit_price": "10"}}}, 201)
	po := order["purchase_order"].(map[string]any)
	poID := po["id"].(string)
	if po["created_by"] != creator.String() {
		t.Fatal("order creator not trusted")
	}
	call("POST", "/api/v1/purchase-orders/"+poID+"/approve", creatorToken, nil, 403)
	call("POST", "/api/v1/purchase-orders/"+poID+"/approve", approverToken, nil, 200)
	call("POST", "/api/v1/purchase-orders/"+poID+"/issue", approverToken, nil, 200)
	call("PUT", "/api/v1/purchase-orders/"+poID, approverToken, map[string]any{"currency": "EUR"}, 409)
	call("POST", "/api/v1/purchase-orders/"+poID+"/items", approverToken, map[string]any{"item_id": uuid.NewString(), "quantity": "1", "unit_price": "1"}, 409)
	items := order["items"].([]any)
	poItem := items[0].(map[string]any)["id"].(string)
	call("POST", "/api/v1/goods-receipts", approverToken, map[string]any{"purchase_order_id": poID, "items": []map[string]any{{"purchase_order_item_id": poItem, "received_quantity": "1"}}}, 201)
	var receiver uuid.UUID
	if err = pool.QueryRow(ctx, `SELECT received_by FROM goods_receipts WHERE purchase_order_id=$1`, poID).Scan(&receiver); err != nil || receiver != approver {
		t.Fatal("receipt actor not from verified claims")
	}
	// Nil/unconfigured publishing must preserve an actual pending row unchanged.
	svc := service.NewOutboxService(repository.NewPostgresOutboxRepository(pool), nil)
	event, err := svc.RecordEvent(ctx, "fixture", uuid.NewString(), "fixture.pending", map[string]string{"synthetic": "yes"})
	if err != nil {
		t.Fatal("owned pending event failed")
	}
	svc.StartBackgroundPublisher(ctx, time.Millisecond)
	_, err = svc.ProcessBatch(ctx, 10)
	if !errors.Is(err, service.ErrDispatcherNotConfigured) {
		t.Fatal("unconfigured delivery accepted")
	}
	svc.Stop()
	var status string
	var published *time.Time
	var retry int
	if err = pool.QueryRow(ctx, `SELECT status,published_at,retry_count FROM outbox_events WHERE id=$1`, event.ID).Scan(&status, &published, &retry); err != nil || status != "pending" || published != nil || retry != 0 {
		t.Fatal("disabled worker changed durable pending event")
	}
	pool.Close()
	call("GET", "/ready", "", nil, 503)
	call("GET", "/health", "", nil, 200)
}

package api

import (
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"github.com/Group-project-bridgeon/purchasing/internal/security"
	"github.com/Group-project-bridgeon/purchasing/internal/testfixture"
	"github.com/gofiber/fiber/v2"
)

func TestEveryApprovedRouteRequiresExactAuthority(t *testing.T) {
	s := testfixture.NewSigner(t)
	verifier, err := security.Load(s.PublicPath, "erp-auth", "erp-api")
	if err != nil {
		t.Fatal(err)
	}
	routes, err := config.LoadRoutes()
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	SetupRouter(app, Handlers{PurchaseRequest: handler.NewPurchaseRequestHandler(nil), PurchaseOrder: handler.NewPurchaseOrderHandler(nil), GoodsReceipt: handler.NewGoodsReceiptHandler(nil), VendorBill: handler.NewVendorBillHandler(nil), PurchaseReturn: handler.NewPurchaseReturnHandler(nil), RFQ: handler.NewRFQHandler(nil), Analytics: handler.NewAnalyticsHandler(nil), Outbox: handler.NewOutboxHandler(nil)}, routes, RouterOptions{Verifier: verifier})
	// Catch new registered endpoints that were omitted from the approved matrix.
	for _, route := range app.GetRoutes(true) {
		if route.Method == "HEAD" {
			continue
		}
		key := route.Method + " " + route.Path
		if key == "GET /health" || key == "GET /api/v1/health" || key == "GET /ready" {
			continue
		}
		if _, ok := PermissionMatrix()[key]; !ok {
			t.Fatalf("registered route missing explicit authority: %s", key)
		}
	}
	noGrants := s.Token(t, nil, nil)
	for key := range PermissionMatrix() {
		parts := strings.SplitN(key, " ", 2)
		path := regexp.MustCompile(`:[A-Za-z0-9_]+`).ReplaceAllString(parts[1], "30000000-0000-4000-8000-000000000001")
		for _, authority := range []struct {
			token  string
			status int
		}{{"", 401}, {noGrants, 403}} {
			req := httptest.NewRequest(parts[0], path, nil)
			if authority.token != "" {
				req.Header.Set("Authorization", "Bearer "+authority.token)
			}
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(key, "request failed")
			}
			resp.Body.Close()
			if resp.StatusCode != authority.status {
				t.Fatalf("%s authority boundary: got %d want %d", key, resp.StatusCode, authority.status)
			}
		}
	}
	for _, path := range []string{"/api/v1/outbox/events", "/api/v1/outbox/events/id", "/api/v1/outbox/events/id/retry"} {
		method := "GET"
		if strings.HasSuffix(path, "retry") {
			method = "POST"
		}
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+s.Token(t, []string{"purchasing.outbox:read", "purchasing.outbox:retry"}, nil))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 404 {
			t.Fatal("outbox administration registered")
		}
	}
	for _, tc := range []struct{ path, field string }{{"/api/v1/purchase-orders", "created_by"}, {"/api/v1/vendor-bills", "created_by"}, {"/api/v1/purchase-requests", "requested_by"}, {"/api/v1/goods-receipts", "received_by"}, {"/api/v1/purchase-returns", "created_by"}, {"/api/v1/rfqs", "created_by"}, {"/api/v1/rfqs/from-request/30000000-0000-4000-8000-000000000001", "created_by"}} {
		var grants []string
		for _, p := range PermissionMatrix() {
			grants = append(grants, p...)
		}
		req := httptest.NewRequest("POST", tc.path, strings.NewReader(`{"`+tc.field+`":"30000000-0000-4000-8000-000000000099"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.Token(t, grants, nil))
		resp, err := app.Test(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != 403 {
			t.Fatalf("actor spoofing on %s: %d", tc.path, resp.StatusCode)
		}
	}
}

func TestCompoundAndIndependentPermissions(t *testing.T) {
	s := testfixture.NewSigner(t)
	v, err := security.Load(s.PublicPath, "erp-auth", "erp-api")
	if err != nil {
		t.Fatal(err)
	}
	for _, needed := range PermissionMatrix() {
		app := fiber.New()
		app.Get("/test", middleware.HumanIdentity(v), middleware.Require(needed...), func(c *fiber.Ctx) error { return c.SendStatus(204) })
		for _, grants := range [][]string{needed, {"purchasing.admin:all"}, {"purchasing.*:read"}} {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+s.Token(t, grants, nil))
			resp, err := app.Test(req)
			if err != nil {
				t.Fatal(err)
			}
			resp.Body.Close()
			want := 403
			if len(grants) > 0 && grants[0] == needed[0] {
				want = 204
			}
			if resp.StatusCode != want {
				t.Fatal("exact permission mismatch")
			}
		}
		if len(needed) > 1 {
			for _, one := range needed {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer "+s.Token(t, []string{one}, nil))
				resp, err := app.Test(req)
				if err != nil {
					t.Fatal(err)
				}
				resp.Body.Close()
				if resp.StatusCode != 403 {
					t.Fatal("missing compound grant accepted")
				}
			}
		}
	}
	if p := PermissionMatrix()["POST /api/v1/purchase-orders/:id/items"]; len(p) != 1 || p[0] != "purchasing.purchase-orders:update" {
		t.Fatal("add order item must require update")
	}
	if p := PermissionMatrix()["POST /api/v1/purchase-requests/:id/items"]; len(p) != 1 || p[0] != "purchasing.purchase-requests:update" {
		t.Fatal("add request item must require update")
	}
}

package config

import (
	"testing"
)

func TestLoadRoutes_EmbeddedFallback(t *testing.T) {
	routes, err := LoadRoutes("/non-existent/path/routes.yaml")
	if err != nil {
		t.Fatalf("expected LoadRoutes to succeed with embedded fallback, got error: %v", err)
	}

	if routes.APIPrefix != "/api/v1" {
		t.Errorf("expected APIPrefix '/api/v1', got %q", routes.APIPrefix)
	}
	if routes.PurchaseRequests.Base != "/purchase-requests" {
		t.Errorf("expected PR Base '/purchase-requests', got %q", routes.PurchaseRequests.Base)
	}
	if routes.PurchaseOrders.Base != "/purchase-orders" {
		t.Errorf("expected PO Base '/purchase-orders', got %q", routes.PurchaseOrders.Base)
	}
	if routes.GoodsReceipts.Base != "/goods-receipts" {
		t.Errorf("expected GRN Base '/goods-receipts', got %q", routes.GoodsReceipts.Base)
	}
	if routes.VendorBills.Match != "/vendor-bills/:id/match" {
		t.Errorf("expected Bill Match '/vendor-bills/:id/match', got %q", routes.VendorBills.Match)
	}
	if routes.PurchaseReturns.Complete != "/purchase-returns/:id/complete" {
		t.Errorf("expected Return Complete '/purchase-returns/:id/complete', got %q", routes.PurchaseReturns.Complete)
	}
	if routes.RFQs.Award != "/rfqs/:id/award/:quotationId" {
		t.Errorf("expected RFQ Award '/rfqs/:id/award/:quotationId', got %q", routes.RFQs.Award)
	}
	if routes.Analytics.DashboardKPIs != "/analytics/dashboard/kpis" {
		t.Errorf("expected Analytics DashboardKPIs '/analytics/dashboard/kpis', got %q", routes.Analytics.DashboardKPIs)
	}
	if routes.Outbox.Base != "/outbox/events" {
		t.Errorf("expected Outbox Base '/outbox/events', got %q", routes.Outbox.Base)
	}
	if routes.Outbox.Retry != "/outbox/events/:id/retry" {
		t.Errorf("expected Outbox Retry '/outbox/events/:id/retry', got %q", routes.Outbox.Retry)
	}
}

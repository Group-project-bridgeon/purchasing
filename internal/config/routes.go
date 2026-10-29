package config

import (
	_ "embed"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed routes.yaml
var defaultRoutesYAML []byte

type SystemRoutes struct {
	Health    string `yaml:"health"`
	APIHealth string `yaml:"api_health"`
	Docs      string `yaml:"docs"`
	Swagger   string `yaml:"swagger"`
	OpenAPI   string `yaml:"openapi"`
}

type PurchaseRequestRoutes struct {
	Base     string `yaml:"base"`
	ByID     string `yaml:"by_id"`
	Submit   string `yaml:"submit"`
	Approve  string `yaml:"approve"`
	Reject   string `yaml:"reject"`
	Items    string `yaml:"items"`
	ItemByID string `yaml:"item_by_id"`
}

type PurchaseOrderRoutes struct {
	Base          string `yaml:"base"`
	FromRequest   string `yaml:"from_request"`
	ByID          string `yaml:"by_id"`
	Approve       string `yaml:"approve"`
	Issue         string `yaml:"issue"`
	Cancel        string `yaml:"cancel"`
	Items         string `yaml:"items"`
	ItemByID      string `yaml:"item_by_id"`
	GoodsReceipts string `yaml:"goods_receipts"`
	VendorBills   string `yaml:"vendor_bills"`
}

type GoodsReceiptRoutes struct {
	Base   string `yaml:"base"`
	ByID   string `yaml:"by_id"`
	Cancel string `yaml:"cancel"`
}

type VendorBillRoutes struct {
	Base    string `yaml:"base"`
	ByID    string `yaml:"by_id"`
	Match   string `yaml:"match"`
	Approve string `yaml:"approve"`
	Cancel  string `yaml:"cancel"`
}

type PurchaseReturnRoutes struct {
	Base     string `yaml:"base"`
	ByID     string `yaml:"by_id"`
	ByNumber string `yaml:"by_number"`
	Approve  string `yaml:"approve"`
	Ship     string `yaml:"ship"`
	Complete string `yaml:"complete"`
	Cancel   string `yaml:"cancel"`
}

type RFQRoutes struct {
	Base          string `yaml:"base"`
	FromRequest   string `yaml:"from_request"`
	ByID          string `yaml:"by_id"`
	ByNumber      string `yaml:"by_number"`
	Publish       string `yaml:"publish"`
	Close         string `yaml:"close"`
	Cancel        string `yaml:"cancel"`
	Quotations    string `yaml:"quotations"`
	QuotationByID string `yaml:"quotation_by_id"`
	Compare       string `yaml:"compare"`
	Award         string `yaml:"award"`
}

type AnalyticsRoutes struct {
	VendorScorecard   string `yaml:"vendor_scorecard"`
	VendorPerformance string `yaml:"vendor_performance"`
	DashboardKPIs     string `yaml:"dashboard_kpis"`
	SpendByVendor     string `yaml:"spend_by_vendor"`
}

type OutboxRoutes struct {
	Base  string `yaml:"base"`
	ByID  string `yaml:"by_id"`
	Retry string `yaml:"retry"`
}

type RoutesConfig struct {
	Version          string                `yaml:"version"`
	APIPrefix        string                `yaml:"api_prefix"`
	System           SystemRoutes          `yaml:"system"`
	PurchaseRequests PurchaseRequestRoutes `yaml:"purchase_requests"`
	PurchaseOrders   PurchaseOrderRoutes   `yaml:"purchase_orders"`
	GoodsReceipts    GoodsReceiptRoutes    `yaml:"goods_receipts"`
	VendorBills      VendorBillRoutes      `yaml:"vendor_bills"`
	PurchaseReturns  PurchaseReturnRoutes  `yaml:"purchase_returns"`
	RFQs             RFQRoutes             `yaml:"rfqs"`
	Analytics        AnalyticsRoutes       `yaml:"analytics"`
	Outbox           OutboxRoutes          `yaml:"outbox"`
}

// LoadRoutes loads the routes configuration from disk if available, otherwise from embedded defaults.
func LoadRoutes(customPaths ...string) (*RoutesConfig, error) {
	paths := append(customPaths, "config/routes.yaml", "routes.yaml")

	for _, p := range paths {
		if data, err := os.ReadFile(p); err == nil {
			var cfg RoutesConfig
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				return &cfg, nil
			}
		}
	}

	// Fallback to embedded default YAML
	var cfg RoutesConfig
	if err := yaml.Unmarshal(defaultRoutesYAML, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse embedded routes configuration: %w", err)
	}
	return &cfg, nil
}

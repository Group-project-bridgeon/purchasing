package service

import (
	"context"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

type mockAnalyticsRepository struct {
	deliveryMetrics *repository.VendorDeliveryRawMetrics
	returnMetrics   *repository.VendorReturnRawMetrics
	billingMetrics  *repository.VendorBillingRawMetrics
	orderSummary    *repository.VendorOrderRawSummary
	dashboardSum    *repository.DashboardRawSummary
	liabilities     *repository.DashboardRawLiabilities
	spendByVendor   []*repository.VendorSpendRaw
	activeVendors   []string

	deliveryMetricsMap map[string]*repository.VendorDeliveryRawMetrics
	orderSummaryMap    map[string]*repository.VendorOrderRawSummary
	billingMetricsMap  map[string]*repository.VendorBillingRawMetrics
}

func (m *mockAnalyticsRepository) GetVendorDeliveryMetrics(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*repository.VendorDeliveryRawMetrics, error) {
	if m.deliveryMetricsMap != nil {
		if val, ok := m.deliveryMetricsMap[vendorID]; ok {
			return val, nil
		}
	}
	if m.deliveryMetrics != nil {
		return m.deliveryMetrics, nil
	}
	return &repository.VendorDeliveryRawMetrics{}, nil
}

func (m *mockAnalyticsRepository) GetVendorReturnMetrics(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*repository.VendorReturnRawMetrics, error) {
	if m.returnMetrics != nil {
		return m.returnMetrics, nil
	}
	return &repository.VendorReturnRawMetrics{}, nil
}

func (m *mockAnalyticsRepository) GetVendorBillingMetrics(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*repository.VendorBillingRawMetrics, error) {
	if m.billingMetricsMap != nil {
		if val, ok := m.billingMetricsMap[vendorID]; ok {
			return val, nil
		}
	}
	if m.billingMetrics != nil {
		return m.billingMetrics, nil
	}
	return &repository.VendorBillingRawMetrics{}, nil
}

func (m *mockAnalyticsRepository) GetVendorOrderSummary(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*repository.VendorOrderRawSummary, error) {
	if m.orderSummaryMap != nil {
		if val, ok := m.orderSummaryMap[vendorID]; ok {
			return val, nil
		}
	}
	if m.orderSummary != nil {
		return m.orderSummary, nil
	}
	return &repository.VendorOrderRawSummary{}, nil
}

func (m *mockAnalyticsRepository) GetDashboardSummary(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
) (*repository.DashboardRawSummary, error) {
	if m.dashboardSum != nil {
		return m.dashboardSum, nil
	}
	return &repository.DashboardRawSummary{OrdersByStatus: map[string]int64{}}, nil
}

func (m *mockAnalyticsRepository) GetDashboardLiabilities(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
) (*repository.DashboardRawLiabilities, error) {
	if m.liabilities != nil {
		return m.liabilities, nil
	}
	return &repository.DashboardRawLiabilities{}, nil
}

func (m *mockAnalyticsRepository) GetSpendByVendor(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
	limit int,
) ([]*repository.VendorSpendRaw, error) {
	return m.spendByVendor, nil
}

func (m *mockAnalyticsRepository) ListActiveVendors(
	ctx context.Context,
) ([]string, error) {
	return m.activeVendors, nil
}

func TestGetVendorScorecard_Calculations(t *testing.T) {
	repo := &mockAnalyticsRepository{
		orderSummary: &repository.VendorOrderRawSummary{
			TotalOrders:          10,
			TotalCommittedAmount: decimal.NewFromFloat(50000.00),
		},
		deliveryMetrics: &repository.VendorDeliveryRawMetrics{
			TotalReceipts:     10,
			OnTimeReceipts:    8,  // 80% on-time
			TotalOrderedItems: 20,
			FulfilledItems:    18, // 90% in-full
			TotalReceivedQty:  decimal.NewFromInt(1000),
			TotalRejectedQty:  decimal.NewFromInt(50), // 95% quality acceptance
		},
		returnMetrics: &repository.VendorReturnRawMetrics{
			TotalReturns:        1,
			TotalReturnedQty:    decimal.NewFromInt(20), // 2% return rate
			TotalReturnedAmount: decimal.NewFromFloat(1000.00),
		},
		billingMetrics: &repository.VendorBillingRawMetrics{
			TotalBills:          5,
			MatchedBills:        4, // 80% billing accuracy
			DiscrepantBills:     1,
			TotalInvoicedAmount: decimal.NewFromFloat(20000.00),
		},
	}

	service := NewAnalyticsService(repo)

	scorecard, err := service.GetVendorScorecard(context.Background(), "vendor-1", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error getting scorecard: %v", err)
	}

	// OnTimeRate: 8/10 = 80.00%
	if !scorecard.OnTimeDeliveryRate.Equal(decimal.NewFromFloat(80.00)) {
		t.Errorf("expected OnTimeDeliveryRate 80.00, got %s", scorecard.OnTimeDeliveryRate)
	}

	// InFullRate: 18/20 = 90.00%
	if !scorecard.InFullDeliveryRate.Equal(decimal.NewFromFloat(90.00)) {
		t.Errorf("expected InFullDeliveryRate 90.00, got %s", scorecard.InFullDeliveryRate)
	}

	// OTIFRate: (80 + 90) / 2 = 85.00%
	if !scorecard.OTIFRate.Equal(decimal.NewFromFloat(85.00)) {
		t.Errorf("expected OTIFRate 85.00, got %s", scorecard.OTIFRate)
	}

	// QualityAcceptanceRate: (1000 - 50) / 1000 = 95.00%
	if !scorecard.QualityAcceptanceRate.Equal(decimal.NewFromFloat(95.00)) {
		t.Errorf("expected QualityAcceptanceRate 95.00, got %s", scorecard.QualityAcceptanceRate)
	}

	// ReturnRate: 20 / 1000 = 2.00%
	if !scorecard.ReturnRate.Equal(decimal.NewFromFloat(2.00)) {
		t.Errorf("expected ReturnRate 2.00, got %s", scorecard.ReturnRate)
	}

	// BillingAccuracyRate: 4/5 = 80.00%
	if !scorecard.BillingAccuracyRate.Equal(decimal.NewFromFloat(80.00)) {
		t.Errorf("expected BillingAccuracyRate 80.00, got %s", scorecard.BillingAccuracyRate)
	}

	// CompositeScore calculation:
	// 0.40 * 85 = 34.00
	// 0.35 * 95 = 33.25
	// 0.15 * 80 = 12.00
	// 0.10 * (100 - 2) = 9.80
	// Total = 34.00 + 33.25 + 12.00 + 9.80 = 89.05
	expectedComposite := decimal.NewFromFloat(89.05)
	if !scorecard.CompositeScore.Equal(expectedComposite) {
		t.Errorf("expected CompositeScore %s, got %s", expectedComposite, scorecard.CompositeScore)
	}
}

func TestGetVendorPerformanceRankings(t *testing.T) {
	repo := &mockAnalyticsRepository{
		activeVendors: []string{"vendor-a", "vendor-b"},
		orderSummaryMap: map[string]*repository.VendorOrderRawSummary{
			"vendor-a": {TotalOrders: 5, TotalCommittedAmount: decimal.NewFromFloat(10000.00)},
			"vendor-b": {TotalOrders: 8, TotalCommittedAmount: decimal.NewFromFloat(20000.00)},
		},
		deliveryMetricsMap: map[string]*repository.VendorDeliveryRawMetrics{
			"vendor-a": {
				TotalReceipts: 5, OnTimeReceipts: 5, TotalOrderedItems: 5, FulfilledItems: 5, // 100% OTIF
				TotalReceivedQty: decimal.NewFromInt(100), TotalRejectedQty: decimal.Zero,
			},
			"vendor-b": {
				TotalReceipts: 8, OnTimeReceipts: 4, TotalOrderedItems: 8, FulfilledItems: 4, // 50% OTIF
				TotalReceivedQty: decimal.NewFromInt(100), TotalRejectedQty: decimal.NewFromInt(10),
			},
		},
		billingMetricsMap: map[string]*repository.VendorBillingRawMetrics{
			"vendor-a": {TotalBills: 2, MatchedBills: 2},
			"vendor-b": {TotalBills: 4, MatchedBills: 2},
		},
	}

	service := NewAnalyticsService(repo)

	rankings, total, err := service.GetVendorPerformanceRankings(context.Background(), nil, nil, 10, 0)
	if err != nil {
		t.Fatalf("unexpected error getting rankings: %v", err)
	}

	if total != 2 {
		t.Fatalf("expected 2 vendors, got %d", total)
	}

	// Vendor A has 100% OTIF, so should be Rank 1
	if rankings[0].VendorID != "vendor-a" || rankings[0].Rank != 1 {
		t.Errorf("expected vendor-a to be rank 1, got %s rank %d", rankings[0].VendorID, rankings[0].Rank)
	}

	// Vendor B has 50% OTIF, so should be Rank 2
	if rankings[1].VendorID != "vendor-b" || rankings[1].Rank != 2 {
		t.Errorf("expected vendor-b to be rank 2, got %s rank %d", rankings[1].VendorID, rankings[1].Rank)
	}
}

func TestGetDashboardKPIs(t *testing.T) {
	repo := &mockAnalyticsRepository{
		dashboardSum: &repository.DashboardRawSummary{
			TotalOrders:         25,
			TotalCommittedSpend: decimal.NewFromFloat(150000.00),
			OrdersByStatus: map[string]int64{
				"issued":   10,
				"received": 15,
			},
		},
		liabilities: &repository.DashboardRawLiabilities{
			UnbilledReceivingLiability: decimal.NewFromFloat(12500.00),
			PendingDeliveryExposure:    decimal.NewFromFloat(35000.00),
		},
	}

	service := NewAnalyticsService(repo)

	kpis, err := service.GetDashboardKPIs(context.Background(), nil, nil)
	if err != nil {
		t.Fatalf("unexpected error getting dashboard kpis: %v", err)
	}

	if kpis.TotalOrders != 25 {
		t.Errorf("expected total orders 25, got %d", kpis.TotalOrders)
	}
	if !kpis.UnbilledReceivingLiability.Equal(decimal.NewFromFloat(12500.00)) {
		t.Errorf("expected unbilled liability 12500.00, got %s", kpis.UnbilledReceivingLiability)
	}
	if !kpis.PendingDeliveryExposure.Equal(decimal.NewFromFloat(35000.00)) {
		t.Errorf("expected pending delivery exposure 35000.00, got %s", kpis.PendingDeliveryExposure)
	}
}

func TestGetSpendByVendor(t *testing.T) {
	repo := &mockAnalyticsRepository{
		spendByVendor: []*repository.VendorSpendRaw{
			{VendorID: "ven-1", TotalOrders: 10, CommittedSpend: decimal.NewFromFloat(75000.00)},
			{VendorID: "ven-2", TotalOrders: 5, CommittedSpend: decimal.NewFromFloat(25000.00)},
		},
	}

	service := NewAnalyticsService(repo)

	spend, total, err := service.GetSpendByVendor(context.Background(), nil, nil, 10)
	if err != nil {
		t.Fatalf("unexpected error getting spend by vendor: %v", err)
	}

	// Total: 75k + 25k = 100k
	if !total.Equal(decimal.NewFromFloat(100000.00)) {
		t.Errorf("expected total spend 100000.00, got %s", total)
	}

	if len(spend) != 2 {
		t.Fatalf("expected 2 vendors, got %d", len(spend))
	}

	// Vendor 1: 75k / 100k = 75.00%
	if !spend[0].SpendPercentage.Equal(decimal.NewFromFloat(75.00)) {
		t.Errorf("expected ven-1 spend percentage 75.00, got %s", spend[0].SpendPercentage)
	}
	// Vendor 2: 25k / 100k = 25.00%
	if !spend[1].SpendPercentage.Equal(decimal.NewFromFloat(25.00)) {
		t.Errorf("expected ven-2 spend percentage 25.00, got %s", spend[1].SpendPercentage)
	}
}

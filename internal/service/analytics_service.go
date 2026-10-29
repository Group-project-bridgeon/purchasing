package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

var (
	hundred       = decimal.NewFromInt(100)
	weightOTIF    = decimal.NewFromFloat(0.40)
	weightQuality = decimal.NewFromFloat(0.35)
	weightBilling = decimal.NewFromFloat(0.15)
	weightReturn  = decimal.NewFromFloat(0.10)
)

type AnalyticsService interface {
	GetVendorScorecard(
		ctx context.Context,
		vendorID string,
		fromDate *time.Time,
		toDate *time.Time,
	) (*domain.VendorScorecard, error)

	GetVendorPerformanceRankings(
		ctx context.Context,
		fromDate *time.Time,
		toDate *time.Time,
		limit int,
		offset int,
	) ([]*domain.VendorPerformanceRanking, int64, error)

	GetDashboardKPIs(
		ctx context.Context,
		fromDate *time.Time,
		toDate *time.Time,
	) (*domain.DashboardKPIs, error)

	GetSpendByVendor(
		ctx context.Context,
		fromDate *time.Time,
		toDate *time.Time,
		limit int,
	) ([]*domain.VendorSpendSummary, decimal.Decimal, error)
}

type analyticsService struct {
	repo repository.AnalyticsRepository
}

var _ AnalyticsService = (*analyticsService)(nil)

func NewAnalyticsService(
	repo repository.AnalyticsRepository,
) AnalyticsService {
	return &analyticsService{
		repo: repo,
	}
}

func (s *analyticsService) GetVendorScorecard(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*domain.VendorScorecard, error) {
	if strings.TrimSpace(vendorID) == "" {
		return nil, fmt.Errorf("vendor id is required")
	}

	orderSummary, err := s.repo.GetVendorOrderSummary(ctx, vendorID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get order summary: %w", err)
	}

	deliveryMetrics, err := s.repo.GetVendorDeliveryMetrics(ctx, vendorID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get delivery metrics: %w", err)
	}

	returnMetrics, err := s.repo.GetVendorReturnMetrics(ctx, vendorID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get return metrics: %w", err)
	}

	billingMetrics, err := s.repo.GetVendorBillingMetrics(ctx, vendorID, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get billing metrics: %w", err)
	}

	// 1. On-Time Delivery Rate
	onTimeRate := hundred
	if deliveryMetrics.TotalReceipts > 0 {
		onTimeRate = decimal.NewFromInt(deliveryMetrics.OnTimeReceipts).
			Div(decimal.NewFromInt(deliveryMetrics.TotalReceipts)).
			Mul(hundred)
	}

	// 2. In-Full Delivery Rate
	inFullRate := hundred
	if deliveryMetrics.TotalOrderedItems > 0 {
		inFullRate = decimal.NewFromInt(deliveryMetrics.FulfilledItems).
			Div(decimal.NewFromInt(deliveryMetrics.TotalOrderedItems)).
			Mul(hundred)
	}

	// 3. Combined OTIF Rate
	otifRate := onTimeRate.Add(inFullRate).Div(decimal.NewFromInt(2))

	// 4. Quality Acceptance Rate
	qualityRate := hundred
	if deliveryMetrics.TotalReceivedQty.GreaterThan(decimal.Zero) {
		acceptedQty := deliveryMetrics.TotalReceivedQty.Sub(deliveryMetrics.TotalRejectedQty)
		if acceptedQty.LessThan(decimal.Zero) {
			acceptedQty = decimal.Zero
		}
		qualityRate = acceptedQty.Div(deliveryMetrics.TotalReceivedQty).Mul(hundred)
	}

	// 5. Return Rate
	returnRate := decimal.Zero
	if deliveryMetrics.TotalReceivedQty.GreaterThan(decimal.Zero) {
		returnRate = returnMetrics.TotalReturnedQty.Div(deliveryMetrics.TotalReceivedQty).Mul(hundred)
	}

	// 6. Billing Accuracy Rate
	billingAccuracy := hundred
	if billingMetrics.TotalBills > 0 {
		billingAccuracy = decimal.NewFromInt(billingMetrics.MatchedBills).
			Div(decimal.NewFromInt(billingMetrics.TotalBills)).
			Mul(hundred)
	}

	// 7. Composite Score Calculation
	// score = (0.40 * OTIF) + (0.35 * Quality) + (0.15 * Billing) + (0.10 * max(0, 100 - ReturnRate))
	returnSatisfaction := hundred.Sub(returnRate)
	if returnSatisfaction.LessThan(decimal.Zero) {
		returnSatisfaction = decimal.Zero
	}

	compositeScore := otifRate.Mul(weightOTIF).
		Add(qualityRate.Mul(weightQuality)).
		Add(billingAccuracy.Mul(weightBilling)).
		Add(returnSatisfaction.Mul(weightReturn))

	if compositeScore.GreaterThan(hundred) {
		compositeScore = hundred
	}
	if compositeScore.LessThan(decimal.Zero) {
		compositeScore = decimal.Zero
	}

	return &domain.VendorScorecard{
		VendorID:              vendorID,
		TotalOrders:           orderSummary.TotalOrders,
		TotalCommittedSpend:   orderSummary.TotalCommittedAmount,
		TotalInvoicedSpend:    billingMetrics.TotalInvoicedAmount,
		TotalReceipts:         deliveryMetrics.TotalReceipts,
		OnTimeReceipts:        deliveryMetrics.OnTimeReceipts,
		OnTimeDeliveryRate:    onTimeRate.Round(2),
		TotalOrderedItems:     deliveryMetrics.TotalOrderedItems,
		FulfilledItems:        deliveryMetrics.FulfilledItems,
		InFullDeliveryRate:    inFullRate.Round(2),
		OTIFRate:              otifRate.Round(2),
		TotalReceivedQuantity: deliveryMetrics.TotalReceivedQty,
		TotalRejectedQuantity: deliveryMetrics.TotalRejectedQty,
		QualityAcceptanceRate: qualityRate.Round(2),
		TotalReturns:          returnMetrics.TotalReturns,
		TotalReturnedQuantity: returnMetrics.TotalReturnedQty,
		TotalReturnedAmount:   returnMetrics.TotalReturnedAmount,
		ReturnRate:            returnRate.Round(2),
		TotalBills:            billingMetrics.TotalBills,
		MatchedBills:          billingMetrics.MatchedBills,
		DiscrepantBills:       billingMetrics.DiscrepantBills,
		BillingAccuracyRate:   billingAccuracy.Round(2),
		CompositeScore:        compositeScore.Round(2),
	}, nil
}

func (s *analyticsService) GetVendorPerformanceRankings(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
	limit int,
	offset int,
) ([]*domain.VendorPerformanceRanking, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	vendorIDs, err := s.repo.ListActiveVendors(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list active vendors: %w", err)
	}

	rankings := make([]*domain.VendorPerformanceRanking, 0, len(vendorIDs))

	for _, vID := range vendorIDs {
		scorecard, err := s.GetVendorScorecard(ctx, vID, fromDate, toDate)
		if err != nil {
			continue
		}

		rankings = append(rankings, &domain.VendorPerformanceRanking{
			VendorID:              vID,
			TotalOrders:           scorecard.TotalOrders,
			CommittedSpend:        scorecard.TotalCommittedSpend,
			OTIFRate:              scorecard.OTIFRate,
			QualityAcceptanceRate: scorecard.QualityAcceptanceRate,
			CompositeScore:        scorecard.CompositeScore,
		})
	}

	sort.Slice(rankings, func(i, j int) bool {
		if !rankings[i].CompositeScore.Equal(rankings[j].CompositeScore) {
			return rankings[i].CompositeScore.GreaterThan(rankings[j].CompositeScore)
		}
		return rankings[i].CommittedSpend.GreaterThan(rankings[j].CommittedSpend)
	})

	for i := range rankings {
		rankings[i].Rank = i + 1
	}

	total := int64(len(rankings))

	start := offset
	if start > len(rankings) {
		start = len(rankings)
	}
	end := start + limit
	if end > len(rankings) {
		end = len(rankings)
	}

	return rankings[start:end], total, nil
}

func (s *analyticsService) GetDashboardKPIs(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
) (*domain.DashboardKPIs, error) {
	summary, err := s.repo.GetDashboardSummary(ctx, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get dashboard summary: %w", err)
	}

	liabilities, err := s.repo.GetDashboardLiabilities(ctx, fromDate, toDate)
	if err != nil {
		return nil, fmt.Errorf("get dashboard liabilities: %w", err)
	}

	return &domain.DashboardKPIs{
		TotalOrders:                summary.TotalOrders,
		TotalCommittedSpend:        summary.TotalCommittedSpend,
		UnbilledReceivingLiability: liabilities.UnbilledReceivingLiability,
		PendingDeliveryExposure:    liabilities.PendingDeliveryExposure,
		OrdersByStatus:             summary.OrdersByStatus,
	}, nil
}

func (s *analyticsService) GetSpendByVendor(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
	limit int,
) ([]*domain.VendorSpendSummary, decimal.Decimal, error) {
	rawSpend, err := s.repo.GetSpendByVendor(ctx, fromDate, toDate, limit)
	if err != nil {
		return nil, decimal.Zero, fmt.Errorf("get spend by vendor: %w", err)
	}

	totalSpend := decimal.Zero
	for _, row := range rawSpend {
		totalSpend = totalSpend.Add(row.CommittedSpend)
	}

	results := make([]*domain.VendorSpendSummary, 0, len(rawSpend))
	for _, row := range rawSpend {
		pct := decimal.Zero
		if totalSpend.GreaterThan(decimal.Zero) {
			pct = row.CommittedSpend.Div(totalSpend).Mul(hundred).Round(2)
		}

		results = append(results, &domain.VendorSpendSummary{
			VendorID:        row.VendorID,
			TotalOrders:     row.TotalOrders,
			CommittedSpend:  row.CommittedSpend,
			SpendPercentage: pct,
		})
	}

	return results, totalSpend, nil
}

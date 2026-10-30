package helper

import (
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/gofiber/fiber/v2"
	"github.com/shopspring/decimal"
)

func ParseDateRange(c *fiber.Ctx) (*time.Time, *time.Time, error) {
	var fromDate *time.Time
	if from := c.Query("from_date"); from != "" {
		t, err := time.Parse("2006-01-02", from)
		if err != nil {
			t, err = time.Parse(time.RFC3339, from)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid from_date (expected YYYY-MM-DD): %w", err)
			}
		}
		fromDate = &t
	}

	var toDate *time.Time
	if to := c.Query("to_date"); to != "" {
		t, err := time.Parse("2006-01-02", to)
		if err != nil {
			t, err = time.Parse(time.RFC3339, to)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid to_date (expected YYYY-MM-DD): %w", err)
			}
		}
		toDate = &t
	}

	return fromDate, toDate, nil
}

func MapScorecardToResponse(sc *domain.VendorScorecard) dto.VendorScorecardResponse {
	return dto.VendorScorecardResponse{
		VendorID:              sc.VendorID,
		TotalOrders:           sc.TotalOrders,
		TotalCommittedSpend:   sc.TotalCommittedSpend.StringFixed(2),
		TotalInvoicedSpend:    sc.TotalInvoicedSpend.StringFixed(2),
		TotalReceipts:         sc.TotalReceipts,
		OnTimeReceipts:        sc.OnTimeReceipts,
		OnTimeDeliveryRate:    sc.OnTimeDeliveryRate.StringFixed(2),
		TotalOrderedItems:     sc.TotalOrderedItems,
		FulfilledItems:        sc.FulfilledItems,
		InFullDeliveryRate:    sc.InFullDeliveryRate.StringFixed(2),
		OTIFRate:              sc.OTIFRate.StringFixed(2),
		TotalReceivedQuantity: sc.TotalReceivedQuantity.String(),
		TotalRejectedQuantity: sc.TotalRejectedQuantity.String(),
		QualityAcceptanceRate: sc.QualityAcceptanceRate.StringFixed(2),
		TotalReturns:          sc.TotalReturns,
		TotalReturnedQuantity: sc.TotalReturnedQuantity.String(),
		TotalReturnedAmount:   sc.TotalReturnedAmount.StringFixed(2),
		ReturnRate:            sc.ReturnRate.StringFixed(2),
		TotalBills:            sc.TotalBills,
		MatchedBills:          sc.MatchedBills,
		DiscrepantBills:       sc.DiscrepantBills,
		BillingAccuracyRate:   sc.BillingAccuracyRate.StringFixed(2),
		CompositeScore:        sc.CompositeScore.StringFixed(2),
	}
}

func MapRankingsToResponse(rankings []*domain.VendorPerformanceRanking, total int64) dto.VendorPerformanceListResponse {
	list := make([]dto.VendorPerformanceRankingResponse, 0, len(rankings))
	for _, r := range rankings {
		list = append(list, dto.VendorPerformanceRankingResponse{
			VendorID:              r.VendorID,
			TotalOrders:           r.TotalOrders,
			CommittedSpend:        r.CommittedSpend.StringFixed(2),
			OTIFRate:              r.OTIFRate.StringFixed(2),
			QualityAcceptanceRate: r.QualityAcceptanceRate.StringFixed(2),
			CompositeScore:        r.CompositeScore.StringFixed(2),
			Rank:                  r.Rank,
		})
	}
	return dto.VendorPerformanceListResponse{
		Vendors: list,
		Total:   total,
	}
}

func MapDashboardKPIsToResponse(kpis *domain.DashboardKPIs) dto.DashboardKPIsResponse {
	return dto.DashboardKPIsResponse{
		TotalOrders:                kpis.TotalOrders,
		TotalCommittedSpend:        kpis.TotalCommittedSpend.StringFixed(2),
		UnbilledReceivingLiability: kpis.UnbilledReceivingLiability.StringFixed(2),
		PendingDeliveryExposure:    kpis.PendingDeliveryExposure.StringFixed(2),
		OrdersByStatus:             kpis.OrdersByStatus,
	}
}

func MapSpendListToResponse(spend []*domain.VendorSpendSummary, total decimal.Decimal) dto.VendorSpendListResponse {
	list := make([]dto.VendorSpendSummaryResponse, 0, len(spend))
	for _, s := range spend {
		list = append(list, dto.VendorSpendSummaryResponse{
			VendorID:        s.VendorID,
			TotalOrders:     s.TotalOrders,
			CommittedSpend:  s.CommittedSpend.StringFixed(2),
			SpendPercentage: s.SpendPercentage.StringFixed(2),
		})
	}
	return dto.VendorSpendListResponse{
		SpendByVendor: list,
		TotalSpend:    total.StringFixed(2),
	}
}

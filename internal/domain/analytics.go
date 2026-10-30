package domain

import (
	"github.com/shopspring/decimal"
)

type VendorScorecard struct {
	VendorID              string
	TotalOrders           int64
	TotalCommittedSpend   decimal.Decimal
	TotalInvoicedSpend    decimal.Decimal
	TotalReceipts         int64
	OnTimeReceipts        int64
	OnTimeDeliveryRate    decimal.Decimal
	TotalOrderedItems     int64
	FulfilledItems        int64
	InFullDeliveryRate    decimal.Decimal
	OTIFRate              decimal.Decimal
	TotalReceivedQuantity decimal.Decimal
	TotalRejectedQuantity decimal.Decimal
	QualityAcceptanceRate decimal.Decimal
	TotalReturns          int64
	TotalReturnedQuantity decimal.Decimal
	TotalReturnedAmount   decimal.Decimal
	ReturnRate            decimal.Decimal
	TotalBills            int64
	MatchedBills          int64
	DiscrepantBills       int64
	BillingAccuracyRate   decimal.Decimal
	CompositeScore        decimal.Decimal
}

type VendorPerformanceRanking struct {
	VendorID              string
	TotalOrders           int64
	CommittedSpend        decimal.Decimal
	OTIFRate              decimal.Decimal
	QualityAcceptanceRate decimal.Decimal
	CompositeScore        decimal.Decimal
	Rank                  int
}

type DashboardKPIs struct {
	TotalOrders                int64
	TotalCommittedSpend        decimal.Decimal
	UnbilledReceivingLiability decimal.Decimal
	PendingDeliveryExposure    decimal.Decimal
	OrdersByStatus             map[string]int64
}

type VendorSpendSummary struct {
	VendorID        string
	TotalOrders     int64
	CommittedSpend  decimal.Decimal
	SpendPercentage decimal.Decimal
}

package repository

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

type VendorDeliveryRawMetrics struct {
	TotalReceipts     int64
	OnTimeReceipts    int64
	TotalOrderedItems int64
	FulfilledItems    int64
	TotalReceivedQty  decimal.Decimal
	TotalRejectedQty  decimal.Decimal
}

type VendorReturnRawMetrics struct {
	TotalReturns        int64
	TotalReturnedQty    decimal.Decimal
	TotalReturnedAmount decimal.Decimal
}

type VendorBillingRawMetrics struct {
	TotalBills          int64
	MatchedBills        int64
	DiscrepantBills     int64
	TotalInvoicedAmount decimal.Decimal
}

type VendorOrderRawSummary struct {
	TotalOrders          int64
	TotalCommittedAmount decimal.Decimal
}

type DashboardRawSummary struct {
	TotalOrders         int64
	TotalCommittedSpend decimal.Decimal
	OrdersByStatus      map[string]int64
}

type DashboardRawLiabilities struct {
	UnbilledReceivingLiability decimal.Decimal
	PendingDeliveryExposure    decimal.Decimal
}

type VendorSpendRaw struct {
	VendorID       string
	TotalOrders    int64
	CommittedSpend decimal.Decimal
}

type AnalyticsRepository interface {
	GetVendorDeliveryMetrics(
		ctx context.Context,
		vendorID string,
		fromDate *time.Time,
		toDate *time.Time,
	) (*VendorDeliveryRawMetrics, error)

	GetVendorReturnMetrics(
		ctx context.Context,
		vendorID string,
		fromDate *time.Time,
		toDate *time.Time,
	) (*VendorReturnRawMetrics, error)

	GetVendorBillingMetrics(
		ctx context.Context,
		vendorID string,
		fromDate *time.Time,
		toDate *time.Time,
	) (*VendorBillingRawMetrics, error)

	GetVendorOrderSummary(
		ctx context.Context,
		vendorID string,
		fromDate *time.Time,
		toDate *time.Time,
	) (*VendorOrderRawSummary, error)

	GetDashboardSummary(
		ctx context.Context,
		fromDate *time.Time,
		toDate *time.Time,
	) (*DashboardRawSummary, error)

	GetDashboardLiabilities(
		ctx context.Context,
		fromDate *time.Time,
		toDate *time.Time,
	) (*DashboardRawLiabilities, error)

	GetSpendByVendor(
		ctx context.Context,
		fromDate *time.Time,
		toDate *time.Time,
		limit int,
	) ([]*VendorSpendRaw, error)

	ListActiveVendors(
		ctx context.Context,
	) ([]string, error)
}

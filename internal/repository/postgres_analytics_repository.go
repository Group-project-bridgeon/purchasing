package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type postgresAnalyticsRepository struct {
	queries *generated.Queries
}

var _ AnalyticsRepository = (*postgresAnalyticsRepository)(nil)

func NewPostgresAnalyticsRepository(
	db *pgxpool.Pool,
) AnalyticsRepository {
	return &postgresAnalyticsRepository{
		queries: generated.New(db),
	}
}

func (r *postgresAnalyticsRepository) GetVendorDeliveryMetrics(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*VendorDeliveryRawMetrics, error) {
	vUUID := pgtype.UUID{}
	if err := vUUID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	var fromTS pgtype.Timestamptz
	if fromDate != nil {
		fromTS = pgtype.Timestamptz{Time: *fromDate, Valid: true}
	}

	var toTS pgtype.Timestamptz
	if toDate != nil {
		toTS = pgtype.Timestamptz{Time: *toDate, Valid: true}
	}

	row, err := r.queries.GetVendorDeliveryMetrics(
		ctx,
		generated.GetVendorDeliveryMetricsParams{
			VendorID: vUUID,
			Column2:  fromTS,
			Column3:  toTS,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get vendor delivery metrics: %w", err)
	}

	receivedQty, err := numericToDecimal(row.TotalReceivedQuantity)
	if err != nil {
		receivedQty = decimal.Zero
	}

	rejectedQty, err := numericToDecimal(row.TotalRejectedQuantity)
	if err != nil {
		rejectedQty = decimal.Zero
	}

	return &VendorDeliveryRawMetrics{
		TotalReceipts:     row.TotalReceipts,
		OnTimeReceipts:    row.OnTimeReceipts,
		TotalOrderedItems: row.TotalOrderedItems,
		FulfilledItems:    row.FulfilledItems,
		TotalReceivedQty:  receivedQty,
		TotalRejectedQty:  rejectedQty,
	}, nil
}

func (r *postgresAnalyticsRepository) GetVendorReturnMetrics(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*VendorReturnRawMetrics, error) {
	vUUID := pgtype.UUID{}
	if err := vUUID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	var fromDateParam pgtype.Date
	if fromDate != nil {
		fromDateParam = pgtype.Date{Time: *fromDate, Valid: true}
	}

	var toDateParam pgtype.Date
	if toDate != nil {
		toDateParam = pgtype.Date{Time: *toDate, Valid: true}
	}

	row, err := r.queries.GetVendorReturnMetrics(
		ctx,
		generated.GetVendorReturnMetricsParams{
			VendorID: vUUID,
			Column2:  fromDateParam,
			Column3:  toDateParam,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get vendor return metrics: %w", err)
	}

	returnedQty, err := numericToDecimal(row.TotalReturnedQuantity)
	if err != nil {
		returnedQty = decimal.Zero
	}

	returnedAmount, err := numericToDecimal(row.TotalReturnedAmount)
	if err != nil {
		returnedAmount = decimal.Zero
	}

	return &VendorReturnRawMetrics{
		TotalReturns:        row.TotalReturns,
		TotalReturnedQty:    returnedQty,
		TotalReturnedAmount: returnedAmount,
	}, nil
}

func (r *postgresAnalyticsRepository) GetVendorBillingMetrics(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*VendorBillingRawMetrics, error) {
	vUUID := pgtype.UUID{}
	if err := vUUID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	var fromDateParam pgtype.Date
	if fromDate != nil {
		fromDateParam = pgtype.Date{Time: *fromDate, Valid: true}
	}

	var toDateParam pgtype.Date
	if toDate != nil {
		toDateParam = pgtype.Date{Time: *toDate, Valid: true}
	}

	row, err := r.queries.GetVendorBillingMetrics(
		ctx,
		generated.GetVendorBillingMetricsParams{
			VendorID: vUUID,
			Column2:  fromDateParam,
			Column3:  toDateParam,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get vendor billing metrics: %w", err)
	}

	invoicedAmount, err := numericToDecimal(row.TotalInvoicedAmount)
	if err != nil {
		invoicedAmount = decimal.Zero
	}

	return &VendorBillingRawMetrics{
		TotalBills:          row.TotalBills,
		MatchedBills:        row.MatchedBills,
		DiscrepantBills:     row.DiscrepantBills,
		TotalInvoicedAmount: invoicedAmount,
	}, nil
}

func (r *postgresAnalyticsRepository) GetVendorOrderSummary(
	ctx context.Context,
	vendorID string,
	fromDate *time.Time,
	toDate *time.Time,
) (*VendorOrderRawSummary, error) {
	vUUID := pgtype.UUID{}
	if err := vUUID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	var fromTS pgtype.Timestamptz
	if fromDate != nil {
		fromTS = pgtype.Timestamptz{Time: *fromDate, Valid: true}
	}

	var toTS pgtype.Timestamptz
	if toDate != nil {
		toTS = pgtype.Timestamptz{Time: *toDate, Valid: true}
	}

	row, err := r.queries.GetVendorOrderSummary(
		ctx,
		generated.GetVendorOrderSummaryParams{
			VendorID: vUUID,
			Column2:  fromTS,
			Column3:  toTS,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get vendor order summary: %w", err)
	}

	committedAmount, err := numericToDecimal(row.TotalCommittedAmount)
	if err != nil {
		committedAmount = decimal.Zero
	}

	return &VendorOrderRawSummary{
		TotalOrders:          row.TotalOrders,
		TotalCommittedAmount: committedAmount,
	}, nil
}

func (r *postgresAnalyticsRepository) GetDashboardSummary(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
) (*DashboardRawSummary, error) {
	var fromTS pgtype.Timestamptz
	if fromDate != nil {
		fromTS = pgtype.Timestamptz{Time: *fromDate, Valid: true}
	}

	var toTS pgtype.Timestamptz
	if toDate != nil {
		toTS = pgtype.Timestamptz{Time: *toDate, Valid: true}
	}

	row, err := r.queries.GetDashboardSummary(
		ctx,
		generated.GetDashboardSummaryParams{
			Column1: fromTS,
			Column2: toTS,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard summary: %w", err)
	}

	committedSpend, err := numericToDecimal(row.TotalCommittedSpend)
	if err != nil {
		committedSpend = decimal.Zero
	}

	statusMap := map[string]int64{
		"draft":              row.DraftOrders,
		"pending_approval":   row.PendingApprovalOrders,
		"approved":           row.ApprovedOrders,
		"issued":             row.IssuedOrders,
		"partially_received": row.PartiallyReceivedOrders,
		"received":           row.ReceivedOrders,
		"billed":             row.BilledOrders,
		"closed":             row.ClosedOrders,
		"cancelled":          row.CancelledOrders,
	}

	return &DashboardRawSummary{
		TotalOrders:         row.TotalOrders,
		TotalCommittedSpend: committedSpend,
		OrdersByStatus:      statusMap,
	}, nil
}

func (r *postgresAnalyticsRepository) GetDashboardLiabilities(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
) (*DashboardRawLiabilities, error) {
	var fromTS pgtype.Timestamptz
	if fromDate != nil {
		fromTS = pgtype.Timestamptz{Time: *fromDate, Valid: true}
	}

	var toTS pgtype.Timestamptz
	if toDate != nil {
		toTS = pgtype.Timestamptz{Time: *toDate, Valid: true}
	}

	row, err := r.queries.GetDashboardLiabilities(
		ctx,
		generated.GetDashboardLiabilitiesParams{
			Column1: fromTS,
			Column2: toTS,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get dashboard liabilities: %w", err)
	}

	unbilled, err := numericToDecimal(row.UnbilledReceivingLiability)
	if err != nil {
		unbilled = decimal.Zero
	}

	pending, err := numericToDecimal(row.PendingDeliveryExposure)
	if err != nil {
		pending = decimal.Zero
	}

	return &DashboardRawLiabilities{
		UnbilledReceivingLiability: unbilled,
		PendingDeliveryExposure:    pending,
	}, nil
}

func (r *postgresAnalyticsRepository) GetSpendByVendor(
	ctx context.Context,
	fromDate *time.Time,
	toDate *time.Time,
	limit int,
) ([]*VendorSpendRaw, error) {
	if limit <= 0 {
		limit = 10
	}

	var fromTS pgtype.Timestamptz
	if fromDate != nil {
		fromTS = pgtype.Timestamptz{Time: *fromDate, Valid: true}
	}

	var toTS pgtype.Timestamptz
	if toDate != nil {
		toTS = pgtype.Timestamptz{Time: *toDate, Valid: true}
	}

	rows, err := r.queries.GetSpendByVendor(
		ctx,
		generated.GetSpendByVendorParams{
			Column1: fromTS,
			Column2: toTS,
			Limit:   int32(limit),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("get spend by vendor: %w", err)
	}

	result := make([]*VendorSpendRaw, 0, len(rows))
	for _, row := range rows {
		spend, err := numericToDecimal(row.CommittedSpend)
		if err != nil {
			spend = decimal.Zero
		}

		result = append(result, &VendorSpendRaw{
			VendorID:       row.VendorID.String(),
			TotalOrders:    row.TotalOrders,
			CommittedSpend: spend,
		})
	}

	return result, nil
}

func (r *postgresAnalyticsRepository) ListActiveVendors(
	ctx context.Context,
) ([]string, error) {
	rows, err := r.queries.ListActiveVendors(ctx)
	if err != nil {
		return nil, fmt.Errorf("list active vendors: %w", err)
	}

	result := make([]string, 0, len(rows))
	for _, id := range rows {
		result = append(result, id.String())
	}

	return result, nil
}

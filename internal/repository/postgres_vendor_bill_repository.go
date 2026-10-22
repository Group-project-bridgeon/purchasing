package repository

import (
	"context"
	"fmt"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVendorBillRepository struct {
	queries *generated.Queries
}

var _ VendorBillRepository = (*postgresVendorBillRepository)(nil)

func NewPostgresVendorBillRepository(
	db *pgxpool.Pool,
) VendorBillRepository {
	return &postgresVendorBillRepository{
		queries: generated.New(db),
	}
}

func NewPostgresVendorBillRepositoryWithTx(
	tx pgx.Tx,
) VendorBillRepository {
	return &postgresVendorBillRepository{
		queries: generated.New(tx),
	}
}

func mapVendorBillToDomain(
	vb generated.VendorBill,
) (*domain.VendorBill, error) {
	subtotal, err := numericToDecimal(vb.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("convert subtotal: %w", err)
	}

	taxAmount, err := numericToDecimal(vb.TaxAmount)
	if err != nil {
		return nil, fmt.Errorf("convert tax amount: %w", err)
	}

	totalAmount, err := numericToDecimal(vb.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("convert total amount: %w", err)
	}

	result := &domain.VendorBill{
		ID:                 vb.ID.String(),
		InternalBillNumber: vb.InternalBillNumber,
		BillNumber:         vb.BillNumber,
		PurchaseOrderID:    vb.PurchaseOrderID.String(),
		VendorID:           vb.VendorID.String(),
		BillDate:           vb.BillDate.Time,
		Currency:           vb.Currency,
		Subtotal:           subtotal,
		TaxAmount:          taxAmount,
		TotalAmount:        totalAmount,
		MatchStatus:        domain.BillMatchStatus(vb.MatchStatus),
		Status:             domain.BillStatus(vb.Status),
		CreatedAt:          vb.CreatedAt.Time,
		UpdatedAt:          vb.UpdatedAt.Time,
	}

	if vb.DueDate.Valid {
		d := vb.DueDate.Time
		result.DueDate = &d
	}

	if vb.DiscrepancyReason.Valid {
		dr := vb.DiscrepancyReason.String
		result.DiscrepancyReason = &dr
	}

	if vb.Notes.Valid {
		n := vb.Notes.String
		result.Notes = &n
	}

	return result, nil
}

func (r *postgresVendorBillRepository) CreateVendorBill(
	ctx context.Context,
	bill *domain.VendorBill,
) (*domain.VendorBill, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(bill.PurchaseOrderID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	vendorID := pgtype.UUID{}
	if err := vendorID.Scan(bill.VendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	var dueDate pgtype.Date
	if bill.DueDate != nil {
		dueDate = pgtype.Date{Time: *bill.DueDate, Valid: true}
	}

	subtotal, err := decimalToNumeric(bill.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("invalid subtotal: %w", err)
	}

	taxAmount, err := decimalToNumeric(bill.TaxAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid tax amount: %w", err)
	}

	totalAmount, err := decimalToNumeric(bill.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid total amount: %w", err)
	}

	var notes pgtype.Text
	if bill.Notes != nil {
		notes = pgtype.Text{String: *bill.Notes, Valid: true}
	}

	var discrepancyReason pgtype.Text
	if bill.DiscrepancyReason != nil {
		discrepancyReason = pgtype.Text{String: *bill.DiscrepancyReason, Valid: true}
	}

	matchStatus := string(bill.MatchStatus)
	if matchStatus == "" {
		matchStatus = string(domain.BillMatchStatusUnmatched)
	}

	status := string(bill.Status)
	if status == "" {
		status = string(domain.BillStatusDraft)
	}

	currency := bill.Currency
	if currency == "" {
		currency = "USD"
	}

	created, err := r.queries.CreateVendorBill(
		ctx,
		generated.CreateVendorBillParams{
			InternalBillNumber: bill.InternalBillNumber,
			BillNumber:         bill.BillNumber,
			PurchaseOrderID:    poID,
			VendorID:           vendorID,
			BillDate:           pgtype.Date{Time: bill.BillDate, Valid: true},
			DueDate:            dueDate,
			Currency:           currency,
			Subtotal:           subtotal,
			TaxAmount:          taxAmount,
			TotalAmount:        totalAmount,
			MatchStatus:        matchStatus,
			Status:             status,
			DiscrepancyReason:  discrepancyReason,
			Notes:              notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create vendor bill: %w", err)
	}

	return mapVendorBillToDomain(created)
}

func (r *postgresVendorBillRepository) GetVendorBillByID(
	ctx context.Context,
	id string,
) (*domain.VendorBill, error) {
	billID := pgtype.UUID{}
	if err := billID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid vendor bill id: %w", err)
	}

	vb, err := r.queries.GetVendorBillByID(ctx, billID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor bill not found: %w", err)
		}
		return nil, fmt.Errorf("get vendor bill by id: %w", err)
	}

	return mapVendorBillToDomain(vb)
}

func (r *postgresVendorBillRepository) GetVendorBillByInternalNumber(
	ctx context.Context,
	internalBillNumber string,
) (*domain.VendorBill, error) {
	vb, err := r.queries.GetVendorBillByInternalNumber(ctx, internalBillNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor bill not found: %w", err)
		}
		return nil, fmt.Errorf("get vendor bill by internal number: %w", err)
	}

	return mapVendorBillToDomain(vb)
}

func (r *postgresVendorBillRepository) GetVendorBillByVendorInvoice(
	ctx context.Context,
	vendorID string,
	billNumber string,
) (*domain.VendorBill, error) {
	vID := pgtype.UUID{}
	if err := vID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	vb, err := r.queries.GetVendorBillByVendorInvoice(
		ctx,
		generated.GetVendorBillByVendorInvoiceParams{
			VendorID:   vID,
			BillNumber: billNumber,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not an error if checking existence
		}
		return nil, fmt.Errorf("get vendor bill by invoice: %w", err)
	}

	return mapVendorBillToDomain(vb)
}

func (r *postgresVendorBillRepository) ListVendorBills(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	bills, err := r.queries.ListVendorBills(
		ctx,
		generated.ListVendorBillsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list vendor bills: %w", err)
	}

	result := make([]*domain.VendorBill, 0, len(bills))
	for _, vb := range bills {
		mapped, err := mapVendorBillToDomain(vb)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresVendorBillRepository) ListVendorBillsByPO(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.VendorBill, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(purchaseOrderID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	bills, err := r.queries.ListVendorBillsByPO(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("list vendor bills by po: %w", err)
	}

	result := make([]*domain.VendorBill, 0, len(bills))
	for _, vb := range bills {
		mapped, err := mapVendorBillToDomain(vb)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresVendorBillRepository) ListVendorBillsByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	vID := pgtype.UUID{}
	if err := vID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	bills, err := r.queries.ListVendorBillsByVendor(
		ctx,
		generated.ListVendorBillsByVendorParams{
			VendorID: vID,
			Limit:    int32(limit),
			Offset:   int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list vendor bills by vendor: %w", err)
	}

	result := make([]*domain.VendorBill, 0, len(bills))
	for _, vb := range bills {
		mapped, err := mapVendorBillToDomain(vb)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresVendorBillRepository) ListVendorBillsByStatus(
	ctx context.Context,
	status domain.BillStatus,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	bills, err := r.queries.ListVendorBillsByStatus(
		ctx,
		generated.ListVendorBillsByStatusParams{
			Status: string(status),
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list vendor bills by status: %w", err)
	}

	result := make([]*domain.VendorBill, 0, len(bills))
	for _, vb := range bills {
		mapped, err := mapVendorBillToDomain(vb)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresVendorBillRepository) ListVendorBillsByMatchStatus(
	ctx context.Context,
	matchStatus domain.BillMatchStatus,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	bills, err := r.queries.ListVendorBillsByMatchStatus(
		ctx,
		generated.ListVendorBillsByMatchStatusParams{
			MatchStatus: string(matchStatus),
			Limit:       int32(limit),
			Offset:      int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list vendor bills by match status: %w", err)
	}

	result := make([]*domain.VendorBill, 0, len(bills))
	for _, vb := range bills {
		mapped, err := mapVendorBillToDomain(vb)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresVendorBillRepository) UpdateVendorBill(
	ctx context.Context,
	bill *domain.VendorBill,
) (*domain.VendorBill, error) {
	billID := pgtype.UUID{}
	if err := billID.Scan(bill.ID); err != nil {
		return nil, fmt.Errorf("invalid vendor bill id: %w", err)
	}

	var dueDate pgtype.Date
	if bill.DueDate != nil {
		dueDate = pgtype.Date{Time: *bill.DueDate, Valid: true}
	}

	subtotal, err := decimalToNumeric(bill.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("invalid subtotal: %w", err)
	}

	taxAmount, err := decimalToNumeric(bill.TaxAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid tax amount: %w", err)
	}

	totalAmount, err := decimalToNumeric(bill.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid total amount: %w", err)
	}

	var notes pgtype.Text
	if bill.Notes != nil {
		notes = pgtype.Text{String: *bill.Notes, Valid: true}
	}

	updated, err := r.queries.UpdateVendorBill(
		ctx,
		generated.UpdateVendorBillParams{
			ID:          billID,
			BillDate:    pgtype.Date{Time: bill.BillDate, Valid: true},
			DueDate:     dueDate,
			Currency:    bill.Currency,
			Subtotal:    subtotal,
			TaxAmount:   taxAmount,
			TotalAmount: totalAmount,
			Notes:       notes,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor bill not found or not editable: %w", err)
		}
		return nil, fmt.Errorf("update vendor bill: %w", err)
	}

	return mapVendorBillToDomain(updated)
}

func (r *postgresVendorBillRepository) UpdateVendorBillMatchStatus(
	ctx context.Context,
	id string,
	matchStatus domain.BillMatchStatus,
	discrepancyReason *string,
) (*domain.VendorBill, error) {
	billID := pgtype.UUID{}
	if err := billID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid vendor bill id: %w", err)
	}

	var reason pgtype.Text
	if discrepancyReason != nil {
		reason = pgtype.Text{String: *discrepancyReason, Valid: true}
	}

	updated, err := r.queries.UpdateVendorBillMatchStatus(
		ctx,
		generated.UpdateVendorBillMatchStatusParams{
			ID:                billID,
			MatchStatus:       string(matchStatus),
			DiscrepancyReason: reason,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor bill not found: %w", err)
		}
		return nil, fmt.Errorf("update vendor bill match status: %w", err)
	}

	return mapVendorBillToDomain(updated)
}

func (r *postgresVendorBillRepository) UpdateVendorBillStatus(
	ctx context.Context,
	id string,
	status domain.BillStatus,
) (*domain.VendorBill, error) {
	billID := pgtype.UUID{}
	if err := billID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid vendor bill id: %w", err)
	}

	updated, err := r.queries.UpdateVendorBillStatus(
		ctx,
		generated.UpdateVendorBillStatusParams{
			ID:     billID,
			Status: string(status),
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor bill not found: %w", err)
		}
		return nil, fmt.Errorf("update vendor bill status: %w", err)
	}

	return mapVendorBillToDomain(updated)
}

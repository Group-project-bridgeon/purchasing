package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresVendorQuotationRepository struct {
	queries *generated.Queries
}

var _ VendorQuotationRepository = (*postgresVendorQuotationRepository)(nil)

func NewPostgresVendorQuotationRepository(
	db *pgxpool.Pool,
) VendorQuotationRepository {
	return &postgresVendorQuotationRepository{
		queries: generated.New(db),
	}
}

func NewPostgresVendorQuotationRepositoryWithTx(
	tx pgx.Tx,
) VendorQuotationRepository {
	return &postgresVendorQuotationRepository{
		queries: generated.New(tx),
	}
}

func mapVendorQuotationToDomain(
	q generated.VendorQuotation,
) (*domain.VendorQuotation, error) {
	subtotal, err := numericToDecimal(q.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("convert subtotal: %w", err)
	}

	taxAmount, err := numericToDecimal(q.TaxAmount)
	if err != nil {
		return nil, fmt.Errorf("convert tax amount: %w", err)
	}

	totalAmount, err := numericToDecimal(q.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("convert total amount: %w", err)
	}

	result := &domain.VendorQuotation{
		ID:             q.ID.String(),
		RFQID:          q.RfqID.String(),
		VendorID:       q.VendorID.String(),
		QuoteReference: q.QuoteReference,
		QuotationDate:  q.QuotationDate.Time,
		ValidUntil:     q.ValidUntil.Time,
		Currency:       q.Currency,
		Subtotal:       subtotal,
		TaxAmount:      taxAmount,
		TotalAmount:    totalAmount,
		LeadTimeDays:   int(q.LeadTimeDays),
		Status:         domain.QuotationStatus(q.Status),
		CreatedAt:      q.CreatedAt.Time,
		UpdatedAt:      q.UpdatedAt.Time,
	}

	if q.PaymentTerms.Valid {
		pt := q.PaymentTerms.String
		result.PaymentTerms = &pt
	}

	if q.Notes.Valid {
		n := q.Notes.String
		result.Notes = &n
	}

	return result, nil
}

func (r *postgresVendorQuotationRepository) CreateVendorQuotation(
	ctx context.Context,
	q *domain.VendorQuotation,
) (*domain.VendorQuotation, error) {
	rfqID := pgtype.UUID{}
	if err := rfqID.Scan(q.RFQID); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	vendorID := pgtype.UUID{}
	if err := vendorID.Scan(q.VendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	subtotalNumeric := pgtype.Numeric{}
	if err := subtotalNumeric.Scan(q.Subtotal.String()); err != nil {
		return nil, fmt.Errorf("invalid subtotal: %w", err)
	}

	taxAmountNumeric := pgtype.Numeric{}
	if err := taxAmountNumeric.Scan(q.TaxAmount.String()); err != nil {
		return nil, fmt.Errorf("invalid tax amount: %w", err)
	}

	totalAmountNumeric := pgtype.Numeric{}
	if err := totalAmountNumeric.Scan(q.TotalAmount.String()); err != nil {
		return nil, fmt.Errorf("invalid total amount: %w", err)
	}

	var paymentTerms pgtype.Text
	if q.PaymentTerms != nil {
		paymentTerms = pgtype.Text{String: *q.PaymentTerms, Valid: true}
	}

	var notes pgtype.Text
	if q.Notes != nil {
		notes = pgtype.Text{String: *q.Notes, Valid: true}
	}

	quoteDate := pgtype.Date{Time: q.QuotationDate, Valid: true}
	if q.QuotationDate.IsZero() {
		quoteDate = pgtype.Date{Time: time.Now(), Valid: true}
	}

	validUntil := pgtype.Date{Time: q.ValidUntil, Valid: true}

	currency := q.Currency
	if currency == "" {
		currency = "USD"
	}

	created, err := r.queries.CreateVendorQuotation(
		ctx,
		generated.CreateVendorQuotationParams{
			RfqID:          rfqID,
			VendorID:       vendorID,
			QuoteReference: q.QuoteReference,
			Column4:        quoteDate,
			ValidUntil:     validUntil,
			Currency:       currency,
			Subtotal:       subtotalNumeric,
			TaxAmount:      taxAmountNumeric,
			TotalAmount:    totalAmountNumeric,
			PaymentTerms:   paymentTerms,
			LeadTimeDays:   int32(q.LeadTimeDays),
			Status:         string(q.Status),
			Notes:          notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create vendor quotation: %w", err)
	}

	return mapVendorQuotationToDomain(created)
}

func (r *postgresVendorQuotationRepository) GetVendorQuotationByID(
	ctx context.Context,
	id string,
) (*domain.VendorQuotation, error) {
	quotationID := pgtype.UUID{}
	if err := quotationID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid quotation id: %w", err)
	}

	row, err := r.queries.GetVendorQuotationByID(ctx, quotationID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrQuotationNotFound
		}
		return nil, fmt.Errorf("get vendor quotation by id: %w", err)
	}

	return mapVendorQuotationToDomain(row)
}

func (r *postgresVendorQuotationRepository) GetVendorQuotationByRef(
	ctx context.Context,
	rfqID string,
	vendorID string,
	ref string,
) (*domain.VendorQuotation, error) {
	rUUID := pgtype.UUID{}
	if err := rUUID.Scan(rfqID); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	vUUID := pgtype.UUID{}
	if err := vUUID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	row, err := r.queries.GetVendorQuotationByRef(
		ctx,
		generated.GetVendorQuotationByRefParams{
			RfqID:          rUUID,
			VendorID:       vUUID,
			QuoteReference: ref,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get vendor quotation by ref: %w", err)
	}

	return mapVendorQuotationToDomain(row)
}

func (r *postgresVendorQuotationRepository) ListVendorQuotationsByRFQ(
	ctx context.Context,
	rfqID string,
) ([]*domain.VendorQuotation, error) {
	rUUID := pgtype.UUID{}
	if err := rUUID.Scan(rfqID); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	rows, err := r.queries.ListVendorQuotationsByRFQ(ctx, rUUID)
	if err != nil {
		return nil, fmt.Errorf("list vendor quotations: %w", err)
	}

	result := make([]*domain.VendorQuotation, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapVendorQuotationToDomain(row)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresVendorQuotationRepository) UpdateVendorQuotationStatus(
	ctx context.Context,
	id string,
	status domain.QuotationStatus,
) (*domain.VendorQuotation, error) {
	qUUID := pgtype.UUID{}
	if err := qUUID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid quotation id: %w", err)
	}

	updated, err := r.queries.UpdateVendorQuotationStatus(
		ctx,
		generated.UpdateVendorQuotationStatusParams{
			ID:     qUUID,
			Status: string(status),
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrQuotationNotFound
		}
		return nil, fmt.Errorf("update vendor quotation status: %w", err)
	}

	return mapVendorQuotationToDomain(updated)
}

func (r *postgresVendorQuotationRepository) RejectOtherVendorQuotations(
	ctx context.Context,
	rfqID string,
	awardedQuotationID string,
) error {
	rUUID := pgtype.UUID{}
	if err := rUUID.Scan(rfqID); err != nil {
		return fmt.Errorf("invalid rfq id: %w", err)
	}

	qUUID := pgtype.UUID{}
	if err := qUUID.Scan(awardedQuotationID); err != nil {
		return fmt.Errorf("invalid quotation id: %w", err)
	}

	return r.queries.RejectOtherVendorQuotations(
		ctx,
		generated.RejectOtherVendorQuotationsParams{
			RfqID: rUUID,
			ID:    qUUID,
		},
	)
}

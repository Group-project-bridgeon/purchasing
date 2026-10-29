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

type postgresGoodsReceiptRepository struct {
	queries *generated.Queries
}

var _ GoodsReceiptRepository = (*postgresGoodsReceiptRepository)(nil)

func NewPostgresGoodsReceiptRepository(
	db *pgxpool.Pool,
) GoodsReceiptRepository {
	return &postgresGoodsReceiptRepository{
		queries: generated.New(db),
	}
}

func NewPostgresGoodsReceiptRepositoryWithTx(
	tx pgx.Tx,
) GoodsReceiptRepository {
	return &postgresGoodsReceiptRepository{
		queries: generated.New(tx),
	}
}

func mapGoodsReceiptToDomain(
	gr generated.GoodsReceipt,
) *domain.GoodsReceipt {
	result := &domain.GoodsReceipt{
		ID:              gr.ID.String(),
		GRNNumber:       gr.GrnNumber,
		PurchaseOrderID: gr.PurchaseOrderID.String(),
		VendorID:        gr.VendorID.String(),
		ReceivedBy:      gr.ReceivedBy.String(),
		ReceivedAt:      gr.ReceivedAt.Time,
		Status:          domain.GoodsReceiptStatus(gr.Status),
		CreatedAt:       gr.CreatedAt.Time,
		UpdatedAt:       gr.UpdatedAt.Time,
	}

	if gr.Notes.Valid {
		n := gr.Notes.String
		result.Notes = &n
	}

	return result
}

func (r *postgresGoodsReceiptRepository) CreateGoodsReceipt(
	ctx context.Context,
	receipt *domain.GoodsReceipt,
) (*domain.GoodsReceipt, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(receipt.PurchaseOrderID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	vendorID := pgtype.UUID{}
	if err := vendorID.Scan(receipt.VendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	receivedBy := pgtype.UUID{}
	if err := receivedBy.Scan(receipt.ReceivedBy); err != nil {
		return nil, fmt.Errorf("invalid received_by user id: %w", err)
	}

	var notes pgtype.Text
	if receipt.Notes != nil {
		notes = pgtype.Text{String: *receipt.Notes, Valid: true}
	}

	status := string(receipt.Status)
	if status == "" {
		status = string(domain.GoodsReceiptStatusCompleted)
	}

	receivedAt := receipt.ReceivedAt
	if receivedAt.IsZero() {
		receivedAt = time.Now()
	}

	created, err := r.queries.CreateGoodsReceipt(
		ctx,
		generated.CreateGoodsReceiptParams{
			GrnNumber:       receipt.GRNNumber,
			PurchaseOrderID: poID,
			VendorID:        vendorID,
			ReceivedBy:      receivedBy,
			Column5:         pgtype.Timestamptz{Time: receivedAt, Valid: true},
			Status:          status,
			Notes:           notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create goods receipt: %w", err)
	}

	return mapGoodsReceiptToDomain(created), nil
}

func (r *postgresGoodsReceiptRepository) GetGoodsReceiptByID(
	ctx context.Context,
	id string,
) (*domain.GoodsReceipt, error) {
	grnID := pgtype.UUID{}
	if err := grnID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid goods receipt id: %w", err)
	}

	grn, err := r.queries.GetGoodsReceiptByID(ctx, grnID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("goods receipt not found: %w", err)
		}
		return nil, fmt.Errorf("get goods receipt by id: %w", err)
	}

	return mapGoodsReceiptToDomain(grn), nil
}

func (r *postgresGoodsReceiptRepository) GetGoodsReceiptByNumber(
	ctx context.Context,
	grnNumber string,
) (*domain.GoodsReceipt, error) {
	grn, err := r.queries.GetGoodsReceiptByNumber(ctx, grnNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("goods receipt not found: %w", err)
		}
		return nil, fmt.Errorf("get goods receipt by number: %w", err)
	}

	return mapGoodsReceiptToDomain(grn), nil
}

func (r *postgresGoodsReceiptRepository) ListGoodsReceipts(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.GoodsReceipt, error) {
	receipts, err := r.queries.ListGoodsReceipts(
		ctx,
		generated.ListGoodsReceiptsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list goods receipts: %w", err)
	}

	result := make([]*domain.GoodsReceipt, 0, len(receipts))
	for _, gr := range receipts {
		result = append(result, mapGoodsReceiptToDomain(gr))
	}

	return result, nil
}

func (r *postgresGoodsReceiptRepository) ListGoodsReceiptsByPO(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.GoodsReceipt, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(purchaseOrderID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	receipts, err := r.queries.ListGoodsReceiptsByPO(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("list goods receipts by po: %w", err)
	}

	result := make([]*domain.GoodsReceipt, 0, len(receipts))
	for _, gr := range receipts {
		result = append(result, mapGoodsReceiptToDomain(gr))
	}

	return result, nil
}

func (r *postgresGoodsReceiptRepository) ListGoodsReceiptsByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.GoodsReceipt, error) {
	vID := pgtype.UUID{}
	if err := vID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	receipts, err := r.queries.ListGoodsReceiptsByVendor(
		ctx,
		generated.ListGoodsReceiptsByVendorParams{
			VendorID: vID,
			Limit:    int32(limit),
			Offset:   int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list goods receipts by vendor: %w", err)
	}

	result := make([]*domain.GoodsReceipt, 0, len(receipts))
	for _, gr := range receipts {
		result = append(result, mapGoodsReceiptToDomain(gr))
	}

	return result, nil
}

func (r *postgresGoodsReceiptRepository) UpdateGoodsReceiptStatus(
	ctx context.Context,
	id string,
	status domain.GoodsReceiptStatus,
) (*domain.GoodsReceipt, error) {
	grnID := pgtype.UUID{}
	if err := grnID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid goods receipt id: %w", err)
	}

	updated, err := r.queries.UpdateGoodsReceiptStatus(
		ctx,
		generated.UpdateGoodsReceiptStatusParams{
			ID:     grnID,
			Status: string(status),
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("goods receipt not found: %w", err)
		}
		return nil, fmt.Errorf("update goods receipt status: %w", err)
	}

	return mapGoodsReceiptToDomain(updated), nil
}

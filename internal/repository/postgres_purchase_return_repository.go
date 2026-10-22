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

type postgresPurchaseReturnRepository struct {
	queries *generated.Queries
}

var _ PurchaseReturnRepository = (*postgresPurchaseReturnRepository)(nil)

func NewPostgresPurchaseReturnRepository(
	db *pgxpool.Pool,
) PurchaseReturnRepository {
	return &postgresPurchaseReturnRepository{
		queries: generated.New(db),
	}
}

func NewPostgresPurchaseReturnRepositoryWithTx(
	tx pgx.Tx,
) PurchaseReturnRepository {
	return &postgresPurchaseReturnRepository{
		queries: generated.New(tx),
	}
}

func mapPurchaseReturnToDomain(
	pr generated.PurchaseReturn,
) (*domain.PurchaseReturn, error) {
	totalAmount, err := numericToDecimal(pr.TotalReturnAmount)
	if err != nil {
		return nil, fmt.Errorf("convert total return amount: %w", err)
	}

	result := &domain.PurchaseReturn{
		ID:                pr.ID.String(),
		ReturnNumber:      pr.ReturnNumber,
		PurchaseOrderID:   pr.PurchaseOrderID.String(),
		VendorID:          pr.VendorID.String(),
		Status:            domain.PurchaseReturnStatus(pr.Status),
		ReturnDate:        pr.ReturnDate.Time,
		TotalReturnAmount: totalAmount,
		Reason:            pr.Reason,
		CreatedBy:         pr.CreatedBy.String(),
		CreatedAt:         pr.CreatedAt.Time,
		UpdatedAt:         pr.UpdatedAt.Time,
	}

	if pr.GoodsReceiptID.Valid {
		grnID := pr.GoodsReceiptID.String()
		result.GoodsReceiptID = &grnID
	}

	if pr.Carrier.Valid {
		c := pr.Carrier.String
		result.Carrier = &c
	}

	if pr.TrackingNumber.Valid {
		tn := pr.TrackingNumber.String
		result.TrackingNumber = &tn
	}

	if pr.Notes.Valid {
		n := pr.Notes.String
		result.Notes = &n
	}

	return result, nil
}

func (r *postgresPurchaseReturnRepository) CreatePurchaseReturn(
	ctx context.Context,
	ret *domain.PurchaseReturn,
) (*domain.PurchaseReturn, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(ret.PurchaseOrderID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	var grnID pgtype.UUID
	if ret.GoodsReceiptID != nil && *ret.GoodsReceiptID != "" {
		if err := grnID.Scan(*ret.GoodsReceiptID); err != nil {
			return nil, fmt.Errorf("invalid goods receipt id: %w", err)
		}
	}

	vendorID := pgtype.UUID{}
	if err := vendorID.Scan(ret.VendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	createdBy := pgtype.UUID{}
	if err := createdBy.Scan(ret.CreatedBy); err != nil {
		return nil, fmt.Errorf("invalid created_by id: %w", err)
	}

	totalAmountNumeric := pgtype.Numeric{}
	if err := totalAmountNumeric.Scan(ret.TotalReturnAmount.String()); err != nil {
		return nil, fmt.Errorf("invalid total return amount: %w", err)
	}

	var carrier pgtype.Text
	if ret.Carrier != nil {
		carrier = pgtype.Text{String: *ret.Carrier, Valid: true}
	}

	var trackingNumber pgtype.Text
	if ret.TrackingNumber != nil {
		trackingNumber = pgtype.Text{String: *ret.TrackingNumber, Valid: true}
	}

	var notes pgtype.Text
	if ret.Notes != nil {
		notes = pgtype.Text{String: *ret.Notes, Valid: true}
	}

	var returnDate pgtype.Date
	if !ret.ReturnDate.IsZero() {
		returnDate = pgtype.Date{Time: ret.ReturnDate, Valid: true}
	} else {
		returnDate = pgtype.Date{Time: time.Now(), Valid: true}
	}

	created, err := r.queries.CreatePurchaseReturn(
		ctx,
		generated.CreatePurchaseReturnParams{
			ReturnNumber:      ret.ReturnNumber,
			PurchaseOrderID:   poID,
			GoodsReceiptID:    grnID,
			VendorID:          vendorID,
			Status:            string(ret.Status),
			Column6:           returnDate,
			TotalReturnAmount: totalAmountNumeric,
			Reason:            ret.Reason,
			Carrier:           carrier,
			TrackingNumber:    trackingNumber,
			CreatedBy:         createdBy,
			Notes:             notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create purchase return: %w", err)
	}

	return mapPurchaseReturnToDomain(created)
}

func (r *postgresPurchaseReturnRepository) GetPurchaseReturnByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseReturn, error) {
	returnID := pgtype.UUID{}
	if err := returnID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase return id: %w", err)
	}

	ret, err := r.queries.GetPurchaseReturnByID(ctx, returnID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPurchaseReturnNotFound
		}
		return nil, fmt.Errorf("get purchase return by id: %w", err)
	}

	return mapPurchaseReturnToDomain(ret)
}

func (r *postgresPurchaseReturnRepository) GetPurchaseReturnByNumber(
	ctx context.Context,
	returnNumber string,
) (*domain.PurchaseReturn, error) {
	ret, err := r.queries.GetPurchaseReturnByNumber(ctx, returnNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPurchaseReturnNotFound
		}
		return nil, fmt.Errorf("get purchase return by number: %w", err)
	}

	return mapPurchaseReturnToDomain(ret)
}

func (r *postgresPurchaseReturnRepository) ListPurchaseReturns(
	ctx context.Context,
	poID *string,
	vendorID *string,
	status *domain.PurchaseReturnStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseReturn, error) {
	var poUUID pgtype.UUID
	if poID != nil && *poID != "" {
		if err := poUUID.Scan(*poID); err != nil {
			return nil, fmt.Errorf("invalid purchase order id: %w", err)
		}
	}

	var vendorUUID pgtype.UUID
	if vendorID != nil && *vendorID != "" {
		if err := vendorUUID.Scan(*vendorID); err != nil {
			return nil, fmt.Errorf("invalid vendor id: %w", err)
		}
	}

	var statusStr string
	if status != nil {
		statusStr = string(*status)
	}

	rows, err := r.queries.ListPurchaseReturns(
		ctx,
		generated.ListPurchaseReturnsParams{
			Column1: poUUID,
			Column2: vendorUUID,
			Column3: statusStr,
			Limit:   int32(limit),
			Offset:  int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list purchase returns: %w", err)
	}

	results := make([]*domain.PurchaseReturn, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapPurchaseReturnToDomain(row)
		if err != nil {
			return nil, err
		}
		results = append(results, mapped)
	}

	return results, nil
}

func (r *postgresPurchaseReturnRepository) CountPurchaseReturns(
	ctx context.Context,
	poID *string,
	vendorID *string,
	status *domain.PurchaseReturnStatus,
) (int64, error) {
	var poUUID pgtype.UUID
	if poID != nil && *poID != "" {
		if err := poUUID.Scan(*poID); err != nil {
			return 0, fmt.Errorf("invalid purchase order id: %w", err)
		}
	}

	var vendorUUID pgtype.UUID
	if vendorID != nil && *vendorID != "" {
		if err := vendorUUID.Scan(*vendorID); err != nil {
			return 0, fmt.Errorf("invalid vendor id: %w", err)
		}
	}

	var statusStr string
	if status != nil {
		statusStr = string(*status)
	}

	count, err := r.queries.CountPurchaseReturns(
		ctx,
		generated.CountPurchaseReturnsParams{
			Column1: poUUID,
			Column2: vendorUUID,
			Column3: statusStr,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("count purchase returns: %w", err)
	}

	return count, nil
}

func (r *postgresPurchaseReturnRepository) UpdatePurchaseReturn(
	ctx context.Context,
	ret *domain.PurchaseReturn,
) (*domain.PurchaseReturn, error) {
	returnID := pgtype.UUID{}
	if err := returnID.Scan(ret.ID); err != nil {
		return nil, fmt.Errorf("invalid purchase return id: %w", err)
	}

	var carrier pgtype.Text
	if ret.Carrier != nil {
		carrier = pgtype.Text{String: *ret.Carrier, Valid: true}
	}

	var trackingNumber pgtype.Text
	if ret.TrackingNumber != nil {
		trackingNumber = pgtype.Text{String: *ret.TrackingNumber, Valid: true}
	}

	var notes pgtype.Text
	if ret.Notes != nil {
		notes = pgtype.Text{String: *ret.Notes, Valid: true}
	}

	updated, err := r.queries.UpdatePurchaseReturn(
		ctx,
		generated.UpdatePurchaseReturnParams{
			ID:             returnID,
			Reason:         ret.Reason,
			Carrier:        carrier,
			TrackingNumber: trackingNumber,
			Notes:          notes,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPurchaseReturnNotFound
		}
		return nil, fmt.Errorf("update purchase return: %w", err)
	}

	return mapPurchaseReturnToDomain(updated)
}

func (r *postgresPurchaseReturnRepository) UpdatePurchaseReturnStatus(
	ctx context.Context,
	id string,
	status domain.PurchaseReturnStatus,
) (*domain.PurchaseReturn, error) {
	returnID := pgtype.UUID{}
	if err := returnID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase return id: %w", err)
	}

	updated, err := r.queries.UpdatePurchaseReturnStatus(
		ctx,
		generated.UpdatePurchaseReturnStatusParams{
			ID:     returnID,
			Status: string(status),
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPurchaseReturnNotFound
		}
		return nil, fmt.Errorf("update purchase return status: %w", err)
	}

	return mapPurchaseReturnToDomain(updated)
}

func (r *postgresPurchaseReturnRepository) ShipPurchaseReturn(
	ctx context.Context,
	id string,
	carrier string,
	trackingNumber string,
) (*domain.PurchaseReturn, error) {
	returnID := pgtype.UUID{}
	if err := returnID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase return id: %w", err)
	}

	carrierText := pgtype.Text{String: carrier, Valid: true}
	trackingText := pgtype.Text{String: trackingNumber, Valid: true}

	updated, err := r.queries.ShipPurchaseReturn(
		ctx,
		generated.ShipPurchaseReturnParams{
			ID:             returnID,
			Carrier:        carrierText,
			TrackingNumber: trackingText,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrPurchaseReturnNotFound
		}
		return nil, fmt.Errorf("ship purchase return: %w", err)
	}

	return mapPurchaseReturnToDomain(updated)
}

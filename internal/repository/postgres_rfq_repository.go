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

type postgresRFQRepository struct {
	queries *generated.Queries
}

var _ RFQRepository = (*postgresRFQRepository)(nil)

func NewPostgresRFQRepository(
	db *pgxpool.Pool,
) RFQRepository {
	return &postgresRFQRepository{
		queries: generated.New(db),
	}
}

func NewPostgresRFQRepositoryWithTx(
	tx pgx.Tx,
) RFQRepository {
	return &postgresRFQRepository{
		queries: generated.New(tx),
	}
}

func mapRFQToDomain(r generated.Rfq) *domain.RFQ {
	result := &domain.RFQ{
		ID:                 r.ID.String(),
		RFQNumber:          r.RfqNumber,
		Title:              r.Title,
		Status:             domain.RFQStatus(r.Status),
		SubmissionDeadline: r.SubmissionDeadline.Time,
		CreatedBy:          r.CreatedBy.String(),
		CreatedAt:          r.CreatedAt.Time,
		UpdatedAt:          r.UpdatedAt.Time,
	}

	if r.PurchaseRequestID.Valid {
		prID := r.PurchaseRequestID.String()
		result.PurchaseRequestID = &prID
	}

	if r.Notes.Valid {
		notes := r.Notes.String
		result.Notes = &notes
	}

	return result
}

func (r *postgresRFQRepository) CreateRFQ(
	ctx context.Context,
	rfq *domain.RFQ,
) (*domain.RFQ, error) {
	var prUUID pgtype.UUID
	if rfq.PurchaseRequestID != nil && *rfq.PurchaseRequestID != "" {
		if err := prUUID.Scan(*rfq.PurchaseRequestID); err != nil {
			return nil, fmt.Errorf("invalid purchase request id: %w", err)
		}
	}

	createdBy := pgtype.UUID{}
	if err := createdBy.Scan(rfq.CreatedBy); err != nil {
		return nil, fmt.Errorf("invalid created_by user id: %w", err)
	}

	deadline := pgtype.Timestamptz{Time: rfq.SubmissionDeadline, Valid: true}

	var notes pgtype.Text
	if rfq.Notes != nil {
		notes = pgtype.Text{String: *rfq.Notes, Valid: true}
	}

	created, err := r.queries.CreateRFQ(
		ctx,
		generated.CreateRFQParams{
			RfqNumber:          rfq.RFQNumber,
			PurchaseRequestID:  prUUID,
			Title:              rfq.Title,
			Status:             string(rfq.Status),
			SubmissionDeadline: deadline,
			CreatedBy:          createdBy,
			Notes:              notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create rfq: %w", err)
	}

	return mapRFQToDomain(created), nil
}

func (r *postgresRFQRepository) GetRFQByID(
	ctx context.Context,
	id string,
) (*domain.RFQ, error) {
	rfqID := pgtype.UUID{}
	if err := rfqID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	row, err := r.queries.GetRFQByID(ctx, rfqID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRFQNotFound
		}
		return nil, fmt.Errorf("get rfq by id: %w", err)
	}

	return mapRFQToDomain(row), nil
}

func (r *postgresRFQRepository) GetRFQByNumber(
	ctx context.Context,
	rfqNumber string,
) (*domain.RFQ, error) {
	row, err := r.queries.GetRFQByNumber(ctx, rfqNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRFQNotFound
		}
		return nil, fmt.Errorf("get rfq by number: %w", err)
	}

	return mapRFQToDomain(row), nil
}

func (r *postgresRFQRepository) ListRFQs(
	ctx context.Context,
	status *domain.RFQStatus,
	purchaseRequestID *string,
	limit int,
	offset int,
) ([]*domain.RFQ, error) {
	var statusStr string
	if status != nil {
		statusStr = string(*status)
	}

	var prUUID pgtype.UUID
	if purchaseRequestID != nil && *purchaseRequestID != "" {
		if err := prUUID.Scan(*purchaseRequestID); err != nil {
			return nil, fmt.Errorf("invalid purchase request id: %w", err)
		}
	}

	rows, err := r.queries.ListRFQs(
		ctx,
		generated.ListRFQsParams{
			Column1: statusStr,
			Column2: prUUID,
			Limit:   int32(limit),
			Offset:  int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list rfqs: %w", err)
	}

	result := make([]*domain.RFQ, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapRFQToDomain(row))
	}

	return result, nil
}

func (r *postgresRFQRepository) CountRFQs(
	ctx context.Context,
	status *domain.RFQStatus,
	purchaseRequestID *string,
) (int64, error) {
	var statusStr string
	if status != nil {
		statusStr = string(*status)
	}

	var prUUID pgtype.UUID
	if purchaseRequestID != nil && *purchaseRequestID != "" {
		if err := prUUID.Scan(*purchaseRequestID); err != nil {
			return 0, fmt.Errorf("invalid purchase request id: %w", err)
		}
	}

	count, err := r.queries.CountRFQs(
		ctx,
		generated.CountRFQsParams{
			Column1: statusStr,
			Column2: prUUID,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("count rfqs: %w", err)
	}

	return count, nil
}

func (r *postgresRFQRepository) UpdateRFQ(
	ctx context.Context,
	rfq *domain.RFQ,
) (*domain.RFQ, error) {
	rfqID := pgtype.UUID{}
	if err := rfqID.Scan(rfq.ID); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	deadline := pgtype.Timestamptz{Time: rfq.SubmissionDeadline, Valid: true}

	var notes pgtype.Text
	if rfq.Notes != nil {
		notes = pgtype.Text{String: *rfq.Notes, Valid: true}
	}

	updated, err := r.queries.UpdateRFQ(
		ctx,
		generated.UpdateRFQParams{
			ID:                 rfqID,
			Title:              rfq.Title,
			SubmissionDeadline: deadline,
			Notes:              notes,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRFQNotFound
		}
		return nil, fmt.Errorf("update rfq: %w", err)
	}

	return mapRFQToDomain(updated), nil
}

func (r *postgresRFQRepository) UpdateRFQStatus(
	ctx context.Context,
	id string,
	status domain.RFQStatus,
) (*domain.RFQ, error) {
	rfqID := pgtype.UUID{}
	if err := rfqID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	updated, err := r.queries.UpdateRFQStatus(
		ctx,
		generated.UpdateRFQStatusParams{
			ID:     rfqID,
			Status: string(status),
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrRFQNotFound
		}
		return nil, fmt.Errorf("update rfq status: %w", err)
	}

	return mapRFQToDomain(updated), nil
}

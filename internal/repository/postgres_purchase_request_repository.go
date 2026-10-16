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

type postgresPurchaseRequestRepository struct {
	db      *pgxpool.Pool
	queries *generated.Queries
}

func NewPostgresPurchaseRequestRepository(
    db *pgxpool.Pool,
) PurchaseRequestRepository {
    return &postgresPurchaseRequestRepository{
        db:      db,
        queries: generated.New(db),
    }
}

func (r *postgresPurchaseRequestRepository) GetPurchaseRequestByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	var requestID pgtype.UUID

	if err := requestID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase request id: %w", err)
	}

	request, err := r.queries.GetPurchaseRequestByID(ctx, requestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase request not found: %w", err)
		}

		return nil, fmt.Errorf("get purchase request by id: %w", err)
	}

	return mapPurchaseRequestToDomain(request), nil
}

func mapPurchaseRequestToDomain(
	request generated.PurchaseRequest,
) *domain.PurchaseRequest {
	result := &domain.PurchaseRequest{
		ID:            request.ID.String(),
		RequestNumber: request.RequestNumber,
		RequestedBy:   request.RequestedBy.String(),
		Status:        domain.PurchaseRequestStatus(request.Status),
		Purpose:       request.Purpose,
		CreatedAt:     request.CreatedAt.Time,
		UpdatedAt:     request.UpdatedAt.Time,
		RequestedAt:   request.RequestedAt.Time,
	}

	if request.DepartmentID.Valid {
		departmentID := request.DepartmentID.String()
		result.DepartmentID = &departmentID
	}

	if request.RequiredBy.Valid {
		requiredBy := request.RequiredBy.Time
		result.RequiredBy = &requiredBy
	}

	if request.Notes.Valid {
		notes := request.Notes.String
		result.Notes = &notes
	}

	return result
}

func (r *postgresPurchaseRequestRepository) GetPurchaseRequestByNumber(
	ctx context.Context,
	requestNumber string,
) (*domain.PurchaseRequest, error) {
	request, err := r.queries.GetPurchaseRequestByNumber(ctx, requestNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase request not found: %w", err)
		}

		return nil, fmt.Errorf("get purchase request by number: %w", err)
	}

	return mapPurchaseRequestToDomain(request), nil
}

func (r *postgresPurchaseRequestRepository) ListPurchaseRequests(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.PurchaseRequest, error) {
	requests, err := r.queries.ListPurchaseRequests(
		ctx,
		generated.ListPurchaseRequestsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list purchase requests: %w", err)
	}

	result := make([]*domain.PurchaseRequest, 0, len(requests))

	for _, request := range requests {
		result = append(result, mapPurchaseRequestToDomain(request))
	}

	return result, nil
}

func (r *postgresPurchaseRequestRepository) ListPurchaseRequestsByStatus(
	ctx context.Context,
	status domain.PurchaseRequestStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseRequest, error) {
	requests, err := r.queries.ListPurchaseRequestsByStatus(
		ctx,
		generated.ListPurchaseRequestsByStatusParams{
			Status: string(status),
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list purchase requests by status: %w", err)
	}

	result := make([]*domain.PurchaseRequest, 0, len(requests))

	for _, request := range requests {
		result = append(result, mapPurchaseRequestToDomain(request))
	}

	return result, nil
}

func (r *postgresPurchaseRequestRepository) CreatePurchaseRequest(
	ctx context.Context,
	request *domain.PurchaseRequest,
) (*domain.PurchaseRequest, error) {
	requestedBy := pgtype.UUID{}
	if err := requestedBy.Scan(request.RequestedBy); err != nil {
		return nil, fmt.Errorf("invalid requested_by: %w", err)
	}

	departmentID := pgtype.UUID{}
	if request.DepartmentID != nil {
		if err := departmentID.Scan(*request.DepartmentID); err != nil {
			return nil, fmt.Errorf("invalid department_id: %w", err)
		}
	}

	var requestedAt pgtype.Timestamptz
	if !request.RequestedAt.IsZero() {
		requestedAt = pgtype.Timestamptz{
			Time:  request.RequestedAt,
			Valid: true,
		}
	}

	var requiredBy pgtype.Date
	if request.RequiredBy != nil {
		requiredBy = pgtype.Date{
			Time:  *request.RequiredBy,
			Valid: true,
		}
	}

	var notes pgtype.Text
	if request.Notes != nil {
		notes = pgtype.Text{
			String: *request.Notes,
			Valid:  true,
		}
	}

	created, err := r.queries.CreatePurchaseRequest(
		ctx,
		generated.CreatePurchaseRequestParams{
			RequestNumber: request.RequestNumber,
			RequestedBy:   requestedBy,
			DepartmentID:  departmentID,
			Column4:       requestedAt,
			RequiredBy:    requiredBy,
			Purpose:       request.Purpose,
			Notes:         notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create purchase request: %w", err)
	}

	return mapPurchaseRequestToDomain(created), nil
}

func (r *postgresPurchaseRequestRepository) UpdatePurchaseRequest(
	ctx context.Context,
	request *domain.PurchaseRequest,
) (*domain.PurchaseRequest, error) {
	requestID := pgtype.UUID{}

	if err := requestID.Scan(request.ID); err != nil {
		return nil, fmt.Errorf("invalid purchase request id: %w", err)
	}

	var requiredBy pgtype.Date
	if request.RequiredBy != nil {
		requiredBy = pgtype.Date{
			Time:  *request.RequiredBy,
			Valid: true,
		}
	}

	var notes pgtype.Text
	if request.Notes != nil {
		notes = pgtype.Text{
			String: *request.Notes,
			Valid:  true,
		}
	}

	updated, err := r.queries.UpdatePurchaseRequest(
		ctx,
		generated.UpdatePurchaseRequestParams{
			ID:         requestID,
			RequiredBy: requiredBy,
			Purpose:    request.Purpose,
			Notes:      notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("update purchase request: %w", err)
	}

	return mapPurchaseRequestToDomain(updated), nil
}

func (r *postgresPurchaseRequestRepository) SubmitPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	requestID := pgtype.UUID{}

	if err := requestID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase request id: %w", err)
	}

	request, err := r.queries.SubmitPurchaseRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("submit purchase request: %w", err)
	}

	return mapPurchaseRequestToDomain(request), nil
}

func (r *postgresPurchaseRequestRepository) ApprovePurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	requestID := pgtype.UUID{}

	if err := requestID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase request id: %w", err)
	}

	request, err := r.queries.ApprovePurchaseRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("approve purchase request: %w", err)
	}

	return mapPurchaseRequestToDomain(request), nil
}

func (r *postgresPurchaseRequestRepository) RejectPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	requestID := pgtype.UUID{}

	if err := requestID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase request id: %w", err)
	}

	request, err := r.queries.RejectPurchaseRequest(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("reject purchase request: %w", err)
	}

	return mapPurchaseRequestToDomain(request), nil
}


var _ PurchaseRequestRepository = (*postgresPurchaseRequestRepository)(nil)
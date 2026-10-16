package service

import (
	"context"
	"fmt"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
)

type PurchaseRequestService interface {
	CreatePurchaseRequest(
		ctx context.Context,
		request *domain.PurchaseRequest,
	) (*domain.PurchaseRequest, error)

	GetPurchaseRequestByID(
		ctx context.Context,
		id string,
	) (*domain.PurchaseRequest, error)

	GetPurchaseRequestByNumber(
		ctx context.Context,
		requestNumber string,
	) (*domain.PurchaseRequest, error)

	ListPurchaseRequests(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.PurchaseRequest, error)

	ListPurchaseRequestsByStatus(
		ctx context.Context,
		status domain.PurchaseRequestStatus,
		limit int,
		offset int,
	) ([]*domain.PurchaseRequest, error)

	UpdatePurchaseRequest(
		ctx context.Context,
		request *domain.PurchaseRequest,
	) (*domain.PurchaseRequest, error)

	SubmitPurchaseRequest(
		ctx context.Context,
		id string,
	) (*domain.PurchaseRequest, error)

	ApprovePurchaseRequest(
		ctx context.Context,
		id string,
	) (*domain.PurchaseRequest, error)

	RejectPurchaseRequest(
		ctx context.Context,
		id string,
	) (*domain.PurchaseRequest, error)
}

type purchaseRequestService struct {
	repository repository.PurchaseRequestRepository
}

func NewPurchaseRequestService(
	repository repository.PurchaseRequestRepository,
) PurchaseRequestService {
	return &purchaseRequestService{
		repository: repository,
	}
}

func (s *purchaseRequestService) CreatePurchaseRequest(
	ctx context.Context,
	request *domain.PurchaseRequest,
) (*domain.PurchaseRequest, error) {
	return s.repository.CreatePurchaseRequest(ctx, request)
}

func (s *purchaseRequestService) GetPurchaseRequestByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	return s.repository.GetPurchaseRequestByID(ctx, id)
}

func (s *purchaseRequestService) GetPurchaseRequestByNumber(
	ctx context.Context,
	requestNumber string,
) (*domain.PurchaseRequest, error) {
	return s.repository.GetPurchaseRequestByNumber(ctx, requestNumber)
}

func (s *purchaseRequestService) ListPurchaseRequests(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.PurchaseRequest, error) {
	return s.repository.ListPurchaseRequests(ctx, limit, offset)
}

func (s *purchaseRequestService) ListPurchaseRequestsByStatus(
	ctx context.Context,
	status domain.PurchaseRequestStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseRequest, error) {
	return s.repository.ListPurchaseRequestsByStatus(
		ctx,
		status,
		limit,
		offset,
	)
}

func (s *purchaseRequestService) UpdatePurchaseRequest(
	ctx context.Context,
	request *domain.PurchaseRequest,
) (*domain.PurchaseRequest, error) {
	if request == nil {
		return nil, fmt.Errorf("purchase request cannot be nil")
	}

	current, err := s.repository.GetPurchaseRequestByID(ctx, request.ID)
	if err != nil {
		return nil, fmt.Errorf("get purchase request: %w", err)
	}

	if current.Status != domain.PurchaseRequestStatusDraft {
		return nil, fmt.Errorf(
			"purchase request cannot be updated in status %q",
			current.Status,
		)
	}

	return s.repository.UpdatePurchaseRequest(ctx, request)
}

func (s *purchaseRequestService) SubmitPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	return s.repository.SubmitPurchaseRequest(ctx, id)
}

func (s *purchaseRequestService) ApprovePurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	return s.repository.ApprovePurchaseRequest(ctx, id)
}

func (s *purchaseRequestService) RejectPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	return s.repository.RejectPurchaseRequest(ctx, id)
}

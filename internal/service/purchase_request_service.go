package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/jackc/pgx/v5"
)

const purchaseRequestDocumentType = "purchase_request"

type CreatePurchaseRequestInput struct {
	Request *domain.PurchaseRequest
	Items   []*domain.PurchaseRequestItem
}

type PurchaseRequestService interface {
	CreatePurchaseRequest(
		ctx context.Context,
		input *CreatePurchaseRequestInput,
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

	GetPurchaseRequestItems(
		ctx context.Context,
		purchaseRequestID string,
	) ([]*domain.PurchaseRequestItem, error)

	AddPurchaseRequestItem(
		ctx context.Context,
		purchaseRequestID string,
		item *domain.PurchaseRequestItem,
	) (*domain.PurchaseRequestItem, error)

	UpdatePurchaseRequestItem(
		ctx context.Context,
		item *domain.PurchaseRequestItem,
	) (*domain.PurchaseRequestItem, error)

	DeletePurchaseRequestItem(
		ctx context.Context,
		id string,
	) error
}

type purchaseRequestService struct {
	repository     repository.PurchaseRequestRepository
	itemRepository repository.PurchaseRequestItemRepository
	transaction    repository.TransactionManager
}

func NewPurchaseRequestService(
	repository repository.PurchaseRequestRepository,
	itemRepository repository.PurchaseRequestItemRepository,
	transaction repository.TransactionManager,
) PurchaseRequestService {
	return &purchaseRequestService{
		repository:     repository,
		itemRepository: itemRepository,
		transaction:    transaction,
	}
}

func (s *purchaseRequestService) CreatePurchaseRequest(
	ctx context.Context,
	input *CreatePurchaseRequestInput,
) (*domain.PurchaseRequest, error) {
	if input == nil {
		return nil, fmt.Errorf("create purchase request input cannot be nil")
	}

	if input.Request == nil {
		return nil, fmt.Errorf("purchase request cannot be nil")
	}

	if len(input.Items) == 0 {
		return nil, fmt.Errorf(
			"purchase request must contain at least one item",
		)
	}

	for i, item := range input.Items {
		if item == nil {
			return nil, fmt.Errorf(
				"purchase request item %d cannot be nil",
				i+1,
			)
		}

		if item.Quantity.IsZero() || item.Quantity.IsNegative() {
			return nil, fmt.Errorf(
				"purchase request item %d must have a positive quantity",
				i+1,
			)
		}

		if item.EstimatedUnitPrice != nil &&
			item.EstimatedUnitPrice.IsNegative() {
			return nil, fmt.Errorf(
				"purchase request item %d cannot have a negative estimated unit price",
				i+1,
			)
		}
	}

	var createdRequest *domain.PurchaseRequest

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			requestRepository repository.PurchaseRequestRepository,
			itemRepository repository.PurchaseRequestItemRepository,
			documentNumberRepository repository.DocumentNumberRepository,
		) error {
			/*
				Generate the purchase request number inside
				the same transaction as the request creation.

				Example:
				PR-2026-000001
				PR-2026-000002
			*/
			year := time.Now().Year()

			nextNumber, err := documentNumberRepository.NextDocumentNumber(
				ctx,
				purchaseRequestDocumentType,
				year,
			)
			if err != nil {
				return fmt.Errorf(
					"generate purchase request number: %w",
					err,
				)
			}

			requestNumber := fmt.Sprintf(
				"PR-%d-%06d",
				year,
				nextNumber,
			)

			/*
				Create a copy so we don't mutate the caller's
				original request object.
			*/
			request := *input.Request
			request.RequestNumber = requestNumber

			/*
				The service controls the initial state.
				A newly created purchase request must always
				start as draft.
			*/
			request.Status = domain.PurchaseRequestStatusDraft

			createdRequest, err = requestRepository.CreatePurchaseRequest(
				ctx,
				&request,
			)
			if err != nil {
				return fmt.Errorf(
					"create purchase request: %w",
					err,
				)
			}

			for i, item := range input.Items {
				/*
					Create a copy so we don't mutate the
					caller's item object.
				*/
				itemCopy := *item
				itemCopy.PurchaseRequestID = createdRequest.ID

				if _, err := itemRepository.CreatePurchaseRequestItem(
					ctx,
					&itemCopy,
				); err != nil {
					return fmt.Errorf(
						"create purchase request item %d: %w",
						i+1,
						err,
					)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return createdRequest, nil
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
	return s.repository.GetPurchaseRequestByNumber(
		ctx,
		requestNumber,
	)
}

func (s *purchaseRequestService) ListPurchaseRequests(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.PurchaseRequest, error) {
	return s.repository.ListPurchaseRequests(
		ctx,
		limit,
		offset,
	)
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
		return nil, fmt.Errorf(
			"purchase request cannot be nil",
		)
	}

	current, err := s.repository.GetPurchaseRequestByID(
		ctx,
		request.ID,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get purchase request: %w",
			err,
		)
	}

	if current.Status != domain.PurchaseRequestStatusDraft {
		return nil, domain.ErrPurchaseRequestNotEditable
	}

	return s.repository.UpdatePurchaseRequest(
		ctx,
		request,
	)
}

func (s *purchaseRequestService) SubmitPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	request, err := s.repository.SubmitPurchaseRequest(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPurchaseRequestCannotSubmit
		}

		return nil, fmt.Errorf("submit purchase request: %w", err)
	}

	return request, nil
}

func (s *purchaseRequestService) ApprovePurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	request, err := s.repository.ApprovePurchaseRequest(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPurchaseRequestCannotApprove
		}

		return nil, fmt.Errorf("approve purchase request: %w", err)
	}

	return request, nil
}

func (s *purchaseRequestService) RejectPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	request, err := s.repository.RejectPurchaseRequest(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPurchaseRequestCannotReject
		}

		return nil, fmt.Errorf("reject purchase request: %w", err)
	}

	return request, nil
}

func (s *purchaseRequestService) GetPurchaseRequestItems(
	ctx context.Context,
	purchaseRequestID string,
) ([]*domain.PurchaseRequestItem, error) {
	return s.itemRepository.GetPurchaseRequestItems(
		ctx,
		purchaseRequestID,
	)
}

func (s *purchaseRequestService) AddPurchaseRequestItem(
	ctx context.Context,
	purchaseRequestID string,
	item *domain.PurchaseRequestItem,
) (*domain.PurchaseRequestItem, error) {
	if purchaseRequestID == "" {
		return nil, fmt.Errorf("purchase request id cannot be empty")
	}

	if item == nil {
		return nil, fmt.Errorf("purchase request item cannot be nil")
	}

	if item.Quantity.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	if item.EstimatedUnitPrice != nil && item.EstimatedUnitPrice.IsNegative() {
		return nil, fmt.Errorf("estimated unit price cannot be negative")
	}

	current, err := s.repository.GetPurchaseRequestByID(ctx, purchaseRequestID)
	if err != nil {
		return nil, fmt.Errorf("get purchase request: %w", err)
	}

	if current.Status != domain.PurchaseRequestStatusDraft {
		return nil, domain.ErrPurchaseRequestNotEditable
	}

	itemCopy := *item
	itemCopy.PurchaseRequestID = purchaseRequestID

	created, err := s.itemRepository.CreatePurchaseRequestItem(ctx, &itemCopy)
	if err != nil {
		return nil, fmt.Errorf("create purchase request item: %w", err)
	}

	return created, nil
}

func (s *purchaseRequestService) UpdatePurchaseRequestItem(
	ctx context.Context,
	item *domain.PurchaseRequestItem,
) (*domain.PurchaseRequestItem, error) {
	if item == nil {
		return nil, fmt.Errorf("purchase request item cannot be nil")
	}

	if item.Quantity.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	if item.EstimatedUnitPrice != nil &&
		item.EstimatedUnitPrice.IsNegative() {
		return nil, fmt.Errorf(
			"estimated unit price cannot be negative",
		)
	}

	updatedItem, err := s.itemRepository.UpdatePurchaseRequestItem(
		ctx,
		item,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"update purchase request item: %w",
			err,
		)
	}

	return updatedItem, nil
}

func (s *purchaseRequestService) DeletePurchaseRequestItem(
	ctx context.Context,
	id string,
) error {
	if id == "" {
		return fmt.Errorf("purchase request item id cannot be empty")
	}

	if err := s.itemRepository.DeletePurchaseRequestItem(ctx, id); err != nil {
		return fmt.Errorf(
			"delete purchase request item: %w",
			err,
		)
	}

	return nil
}

package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type PurchaseRequestRepository interface {
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

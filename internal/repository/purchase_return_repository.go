package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type PurchaseReturnRepository interface {
	CreatePurchaseReturn(
		ctx context.Context,
		ret *domain.PurchaseReturn,
	) (*domain.PurchaseReturn, error)

	GetPurchaseReturnByID(
		ctx context.Context,
		id string,
	) (*domain.PurchaseReturn, error)

	GetPurchaseReturnByNumber(
		ctx context.Context,
		returnNumber string,
	) (*domain.PurchaseReturn, error)

	ListPurchaseReturns(
		ctx context.Context,
		poID *string,
		vendorID *string,
		status *domain.PurchaseReturnStatus,
		limit int,
		offset int,
	) ([]*domain.PurchaseReturn, error)

	CountPurchaseReturns(
		ctx context.Context,
		poID *string,
		vendorID *string,
		status *domain.PurchaseReturnStatus,
	) (int64, error)

	UpdatePurchaseReturn(
		ctx context.Context,
		ret *domain.PurchaseReturn,
	) (*domain.PurchaseReturn, error)

	UpdatePurchaseReturnStatus(
		ctx context.Context,
		id string,
		status domain.PurchaseReturnStatus,
	) (*domain.PurchaseReturn, error)

	ShipPurchaseReturn(
		ctx context.Context,
		id string,
		carrier string,
		trackingNumber string,
	) (*domain.PurchaseReturn, error)
}

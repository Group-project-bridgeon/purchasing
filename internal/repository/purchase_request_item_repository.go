package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type PurchaseRequestItemRepository interface {
	CreatePurchaseRequestItem(
		ctx context.Context,
		item *domain.PurchaseRequestItem,
	) (*domain.PurchaseRequestItem, error)

	GetPurchaseRequestItems(
		ctx context.Context,
		purchaseRequestID string,
	) ([]*domain.PurchaseRequestItem, error)

	GetPurchaseRequestItemByID(
		ctx context.Context,
		id string,
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
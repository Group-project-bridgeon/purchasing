package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/shopspring/decimal"
)

type PurchaseReturnItemRepository interface {
	CreatePurchaseReturnItem(
		ctx context.Context,
		item *domain.PurchaseReturnItem,
	) (*domain.PurchaseReturnItem, error)

	GetPurchaseReturnItems(
		ctx context.Context,
		purchaseReturnID string,
	) ([]*domain.PurchaseReturnItem, error)

	GetPurchaseReturnItemByID(
		ctx context.Context,
		id string,
	) (*domain.PurchaseReturnItem, error)

	GetTotalReturnedQuantityForPOItem(
		ctx context.Context,
		purchaseOrderItemID string,
	) (decimal.Decimal, error)
}

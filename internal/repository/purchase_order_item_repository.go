package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/shopspring/decimal"
)

type PurchaseOrderItemRepository interface {
	CreatePurchaseOrderItem(
		ctx context.Context,
		item *domain.PurchaseOrderItem,
	) (*domain.PurchaseOrderItem, error)

	GetPurchaseOrderItems(
		ctx context.Context,
		purchaseOrderID string,
	) ([]*domain.PurchaseOrderItem, error)

	GetPurchaseOrderItemByID(
		ctx context.Context,
		id string,
	) (*domain.PurchaseOrderItem, error)

	UpdatePurchaseOrderItem(
		ctx context.Context,
		item *domain.PurchaseOrderItem,
	) (*domain.PurchaseOrderItem, error)

	DeletePurchaseOrderItem(
		ctx context.Context,
		id string,
	) error

	UpdatePurchaseOrderItemFulfillment(
		ctx context.Context,
		id string,
		receivedQuantity decimal.Decimal,
		billedQuantity decimal.Decimal,
	) (*domain.PurchaseOrderItem, error)
}

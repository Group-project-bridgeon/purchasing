package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type GoodsReceiptItemRepository interface {
	CreateGoodsReceiptItem(
		ctx context.Context,
		item *domain.GoodsReceiptItem,
	) (*domain.GoodsReceiptItem, error)

	GetGoodsReceiptItems(
		ctx context.Context,
		goodsReceiptID string,
	) ([]*domain.GoodsReceiptItem, error)

	GetGoodsReceiptItemsByPOItemID(
		ctx context.Context,
		purchaseOrderItemID string,
	) ([]*domain.GoodsReceiptItem, error)
}

package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type GoodsReceiptRepository interface {
	CreateGoodsReceipt(
		ctx context.Context,
		receipt *domain.GoodsReceipt,
	) (*domain.GoodsReceipt, error)

	GetGoodsReceiptByID(
		ctx context.Context,
		id string,
	) (*domain.GoodsReceipt, error)

	GetGoodsReceiptByNumber(
		ctx context.Context,
		grnNumber string,
	) (*domain.GoodsReceipt, error)

	ListGoodsReceipts(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.GoodsReceipt, error)

	ListGoodsReceiptsByPO(
		ctx context.Context,
		purchaseOrderID string,
	) ([]*domain.GoodsReceipt, error)

	ListGoodsReceiptsByVendor(
		ctx context.Context,
		vendorID string,
		limit int,
		offset int,
	) ([]*domain.GoodsReceipt, error)

	UpdateGoodsReceiptStatus(
		ctx context.Context,
		id string,
		status domain.GoodsReceiptStatus,
	) (*domain.GoodsReceipt, error)
}

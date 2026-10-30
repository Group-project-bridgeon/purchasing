package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type PurchaseOrderRepository interface {
	CreatePurchaseOrder(
		ctx context.Context,
		order *domain.PurchaseOrder,
	) (*domain.PurchaseOrder, error)

	GetPurchaseOrderByID(
		ctx context.Context,
		id string,
	) (*domain.PurchaseOrder, error)

	GetPurchaseOrderByNumber(
		ctx context.Context,
		poNumber string,
	) (*domain.PurchaseOrder, error)

	ListPurchaseOrders(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.PurchaseOrder, error)

	ListPurchaseOrdersByStatus(
		ctx context.Context,
		status domain.PurchaseOrderStatus,
		limit int,
		offset int,
	) ([]*domain.PurchaseOrder, error)

	ListPurchaseOrdersByVendor(
		ctx context.Context,
		vendorID string,
		limit int,
		offset int,
	) ([]*domain.PurchaseOrder, error)

	UpdatePurchaseOrder(
		ctx context.Context,
		order *domain.PurchaseOrder,
	) (*domain.PurchaseOrder, error)

	UpdatePurchaseOrderStatus(
		ctx context.Context,
		id string,
		status domain.PurchaseOrderStatus,
	) (*domain.PurchaseOrder, error)
}

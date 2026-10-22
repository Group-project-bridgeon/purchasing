package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/shopspring/decimal"
)

type VendorBillItemRepository interface {
	CreateVendorBillItem(
		ctx context.Context,
		item *domain.VendorBillItem,
	) (*domain.VendorBillItem, error)

	GetVendorBillItems(
		ctx context.Context,
		vendorBillID string,
	) ([]*domain.VendorBillItem, error)

	GetVendorBillItemByID(
		ctx context.Context,
		id string,
	) (*domain.VendorBillItem, error)

	UpdateVendorBillItemMatch(
		ctx context.Context,
		id string,
		matchedQuantity decimal.Decimal,
		matchStatus domain.BillItemMatchStatus,
	) (*domain.VendorBillItem, error)
}

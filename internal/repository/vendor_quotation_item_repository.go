package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type VendorQuotationItemRepository interface {
	CreateVendorQuotationItem(
		ctx context.Context,
		item *domain.VendorQuotationItem,
	) (*domain.VendorQuotationItem, error)

	GetVendorQuotationItems(
		ctx context.Context,
		vendorQuotationID string,
	) ([]*domain.VendorQuotationItem, error)
}

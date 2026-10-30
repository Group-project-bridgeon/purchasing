package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type VendorQuotationRepository interface {
	CreateVendorQuotation(
		ctx context.Context,
		q *domain.VendorQuotation,
	) (*domain.VendorQuotation, error)

	GetVendorQuotationByID(
		ctx context.Context,
		id string,
	) (*domain.VendorQuotation, error)

	GetVendorQuotationByRef(
		ctx context.Context,
		rfqID string,
		vendorID string,
		ref string,
	) (*domain.VendorQuotation, error)

	ListVendorQuotationsByRFQ(
		ctx context.Context,
		rfqID string,
	) ([]*domain.VendorQuotation, error)

	UpdateVendorQuotationStatus(
		ctx context.Context,
		id string,
		status domain.QuotationStatus,
	) (*domain.VendorQuotation, error)

	RejectOtherVendorQuotations(
		ctx context.Context,
		rfqID string,
		awardedQuotationID string,
	) error
}

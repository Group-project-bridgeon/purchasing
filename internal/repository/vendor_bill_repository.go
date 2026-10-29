package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type VendorBillRepository interface {
	CreateVendorBill(
		ctx context.Context,
		bill *domain.VendorBill,
	) (*domain.VendorBill, error)

	GetVendorBillByID(
		ctx context.Context,
		id string,
	) (*domain.VendorBill, error)

	GetVendorBillByInternalNumber(
		ctx context.Context,
		internalBillNumber string,
	) (*domain.VendorBill, error)

	GetVendorBillByVendorInvoice(
		ctx context.Context,
		vendorID string,
		billNumber string,
	) (*domain.VendorBill, error)

	ListVendorBills(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByPO(
		ctx context.Context,
		purchaseOrderID string,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByVendor(
		ctx context.Context,
		vendorID string,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByStatus(
		ctx context.Context,
		status domain.BillStatus,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByMatchStatus(
		ctx context.Context,
		matchStatus domain.BillMatchStatus,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	UpdateVendorBill(
		ctx context.Context,
		bill *domain.VendorBill,
	) (*domain.VendorBill, error)

	UpdateVendorBillMatchStatus(
		ctx context.Context,
		id string,
		matchStatus domain.BillMatchStatus,
		discrepancyReason *string,
	) (*domain.VendorBill, error)

	UpdateVendorBillStatus(
		ctx context.Context,
		id string,
		status domain.BillStatus,
	) (*domain.VendorBill, error)
}

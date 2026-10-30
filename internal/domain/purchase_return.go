package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type PurchaseReturnStatus string

const (
	PurchaseReturnStatusDraft     PurchaseReturnStatus = "draft"
	PurchaseReturnStatusApproved  PurchaseReturnStatus = "approved"
	PurchaseReturnStatusShipped   PurchaseReturnStatus = "shipped"
	PurchaseReturnStatusCompleted PurchaseReturnStatus = "completed"
	PurchaseReturnStatusCancelled PurchaseReturnStatus = "cancelled"
)

type PurchaseReturn struct {
	ApprovedBy *string
	ApprovedAt *time.Time

	ID                string
	ReturnNumber      string
	PurchaseOrderID   string
	GoodsReceiptID    *string
	VendorID          string
	Status            PurchaseReturnStatus
	ReturnDate        time.Time
	TotalReturnAmount decimal.Decimal
	Reason            string
	Carrier           *string
	TrackingNumber    *string
	CreatedBy         string
	Notes             *string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type PurchaseReturnItem struct {
	ID                  string
	PurchaseReturnID    string
	PurchaseOrderItemID string
	ItemID              string
	Description         string
	ReturnQuantity      decimal.Decimal
	UnitPrice           decimal.Decimal
	TotalPrice          decimal.Decimal
	Reason              *string
	CreatedAt           time.Time
}

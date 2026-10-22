package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type GoodsReceiptStatus string

const (
	GoodsReceiptStatusDraft     GoodsReceiptStatus = "draft"
	GoodsReceiptStatusCompleted GoodsReceiptStatus = "completed"
	GoodsReceiptStatusCancelled GoodsReceiptStatus = "cancelled"
)

type GoodsReceipt struct {
	ID              string
	GRNNumber       string
	PurchaseOrderID string
	VendorID        string
	ReceivedBy      string
	ReceivedAt      time.Time
	Status          GoodsReceiptStatus
	Notes           *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type GoodsReceiptItem struct {
	ID                  string
	GoodsReceiptID      string
	PurchaseOrderItemID string
	ItemID              string
	ReceivedQuantity    decimal.Decimal
	RejectedQuantity    decimal.Decimal
	RejectionReason     *string
	CreatedAt           time.Time
}

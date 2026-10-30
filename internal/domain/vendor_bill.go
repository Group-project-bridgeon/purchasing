package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type BillStatus string

const (
	BillStatusDraft     BillStatus = "draft"
	BillStatusApproved  BillStatus = "approved"
	BillStatusPaid      BillStatus = "paid"
	BillStatusCancelled BillStatus = "cancelled"
)

type BillMatchStatus string

const (
	BillMatchStatusUnmatched  BillMatchStatus = "unmatched"
	BillMatchStatusMatched    BillMatchStatus = "matched"
	BillMatchStatusDiscrepant BillMatchStatus = "discrepant"
)

type BillItemMatchStatus string

const (
	BillItemMatchStatusUnmatched        BillItemMatchStatus = "unmatched"
	BillItemMatchStatusMatched          BillItemMatchStatus = "matched"
	BillItemMatchStatusPriceVariance    BillItemMatchStatus = "price_variance"
	BillItemMatchStatusQuantityVariance BillItemMatchStatus = "quantity_variance"
)

type VendorBill struct {
	ApprovedBy *string
	ApprovedAt *time.Time
	CreatedBy  *string

	ID                 string
	InternalBillNumber string
	BillNumber         string // Seller's invoice number
	PurchaseOrderID    string
	VendorID           string
	BillDate           time.Time
	DueDate            *time.Time
	Currency           string
	Subtotal           decimal.Decimal
	TaxAmount          decimal.Decimal
	TotalAmount        decimal.Decimal
	MatchStatus        BillMatchStatus
	Status             BillStatus
	DiscrepancyReason  *string
	Notes              *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type VendorBillItem struct {
	ID                  string
	VendorBillID        string
	PurchaseOrderItemID string
	ItemID              string
	Description         string
	Quantity            decimal.Decimal
	UnitPrice           decimal.Decimal
	TaxRate             decimal.Decimal
	TotalPrice          decimal.Decimal
	MatchedQuantity     decimal.Decimal
	MatchStatus         BillItemMatchStatus
	CreatedAt           time.Time
}

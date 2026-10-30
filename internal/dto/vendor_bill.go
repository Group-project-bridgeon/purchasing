package dto

import (
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type CreateVendorBillBody struct {
	CreatedBy       string                     `json:"created_by,omitempty"`
	BillNumber      string                     `json:"bill_number"`
	PurchaseOrderID string                     `json:"purchase_order_id"`
	BillDate        string                     `json:"bill_date"`
	DueDate         *string                    `json:"due_date,omitempty"`
	Currency        string                     `json:"currency,omitempty"`
	Notes           *string                    `json:"notes,omitempty"`
	Items           []CreateVendorBillItemBody `json:"items"`
}

type CreateVendorBillItemBody struct {
	PurchaseOrderItemID string  `json:"purchase_order_item_id"`
	Description         string  `json:"description"`
	Quantity            string  `json:"quantity"`
	UnitPrice           string  `json:"unit_price"`
	TaxRate             *string `json:"tax_rate,omitempty"`
}

type UpdateVendorBillBody struct {
	BillDate *string `json:"bill_date,omitempty"`
	DueDate  *string `json:"due_date,omitempty"`
	Currency string  `json:"currency,omitempty"`
	Notes    *string `json:"notes,omitempty"`
}

type VendorBillResponse struct {
	ApprovedBy *string    `json:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at"`
	CreatedBy  *string    `json:"created_by"`

	ID                 string                 `json:"id"`
	InternalBillNumber string                 `json:"internal_bill_number"`
	BillNumber         string                 `json:"bill_number"`
	PurchaseOrderID    string                 `json:"purchase_order_id"`
	VendorID           string                 `json:"vendor_id"`
	BillDate           time.Time              `json:"bill_date"`
	DueDate            *time.Time             `json:"due_date,omitempty"`
	Currency           string                 `json:"currency"`
	Subtotal           string                 `json:"subtotal"`
	TaxAmount          string                 `json:"tax_amount"`
	TotalAmount        string                 `json:"total_amount"`
	MatchStatus        domain.BillMatchStatus `json:"match_status"`
	Status             domain.BillStatus      `json:"status"`
	DiscrepancyReason  *string                `json:"discrepancy_reason,omitempty"`
	Notes              *string                `json:"notes,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

type VendorBillItemResponse struct {
	ID                  string                     `json:"id"`
	VendorBillID        string                     `json:"vendor_bill_id"`
	PurchaseOrderItemID string                     `json:"purchase_order_item_id"`
	ItemID              string                     `json:"item_id"`
	Description         string                     `json:"description"`
	Quantity            string                     `json:"quantity"`
	UnitPrice           string                     `json:"unit_price"`
	TaxRate             string                     `json:"tax_rate"`
	TotalPrice          string                     `json:"total_price"`
	MatchedQuantity     string                     `json:"matched_quantity"`
	MatchStatus         domain.BillItemMatchStatus `json:"match_status"`
	CreatedAt           time.Time                  `json:"created_at"`
}

type CreateVendorBillResponse struct {
	VendorBill VendorBillResponse       `json:"vendor_bill"`
	Items      []VendorBillItemResponse `json:"items"`
}

type ListVendorBillsResponse struct {
	VendorBills []VendorBillResponse `json:"vendor_bills"`
	Limit       int                  `json:"limit"`
	Offset      int                  `json:"offset"`
}

type MatchBillResponse struct {
	VendorBill    VendorBillResponse       `json:"vendor_bill"`
	Items         []VendorBillItemResponse `json:"items"`
	IsMatched     bool                     `json:"is_matched"`
	Discrepancies []string                 `json:"discrepancies,omitempty"`
}

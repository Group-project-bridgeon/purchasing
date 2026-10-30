package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type PurchaseOrderStatus string

const (
	PurchaseOrderStatusDraft             PurchaseOrderStatus = "draft"
	PurchaseOrderStatusPendingApproval   PurchaseOrderStatus = "pending_approval"
	PurchaseOrderStatusApproved          PurchaseOrderStatus = "approved"
	PurchaseOrderStatusIssued            PurchaseOrderStatus = "issued"
	PurchaseOrderStatusPartiallyReceived PurchaseOrderStatus = "partially_received"
	PurchaseOrderStatusReceived          PurchaseOrderStatus = "received"
	PurchaseOrderStatusBilled            PurchaseOrderStatus = "billed"
	PurchaseOrderStatusClosed            PurchaseOrderStatus = "closed"
	PurchaseOrderStatusCancelled         PurchaseOrderStatus = "cancelled"
)

type PurchaseOrder struct {
	ApprovedBy *string
	ApprovedAt *time.Time
	CreatedBy  *string

	ID                   string
	PONumber             string
	PurchaseRequestID    *string
	VendorID             string
	Status               PurchaseOrderStatus
	OrderDate            time.Time
	ExpectedDeliveryDate *time.Time
	Currency             string
	Subtotal             decimal.Decimal
	TaxAmount            decimal.Decimal
	TotalAmount          decimal.Decimal
	PaymentTerms         *string
	ShippingAddress      *string
	Notes                *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type PurchaseOrderItem struct {
	ID               string
	PurchaseOrderID  string
	ItemID           string
	Description      string
	Quantity         decimal.Decimal
	UnitPrice        decimal.Decimal
	TaxRate          decimal.Decimal
	TotalPrice       decimal.Decimal
	ReceivedQuantity decimal.Decimal
	BilledQuantity   decimal.Decimal
	CreatedAt        time.Time
}

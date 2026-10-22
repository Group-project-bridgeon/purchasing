package dto

import (
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type CreatePurchaseOrderBody struct {
	PurchaseRequestID    *string                       `json:"purchase_request_id,omitempty"`
	VendorID             string                        `json:"vendor_id"`
	ExpectedDeliveryDate *string                       `json:"expected_delivery_date,omitempty"`
	Currency             string                        `json:"currency,omitempty"`
	PaymentTerms         *string                       `json:"payment_terms,omitempty"`
	ShippingAddress      *string                       `json:"shipping_address,omitempty"`
	Notes                *string                       `json:"notes,omitempty"`
	Items                []CreatePurchaseOrderItemBody `json:"items"`
}

type CreatePurchaseOrderItemBody struct {
	ItemID      string  `json:"item_id"`
	Description string  `json:"description"`
	Quantity    string  `json:"quantity"`
	UnitPrice   string  `json:"unit_price"`
	TaxRate     *string `json:"tax_rate,omitempty"`
}

type UpdatePurchaseOrderBody struct {
	ExpectedDeliveryDate *string `json:"expected_delivery_date,omitempty"`
	Currency             string  `json:"currency,omitempty"`
	PaymentTerms         *string `json:"payment_terms,omitempty"`
	ShippingAddress      *string `json:"shipping_address,omitempty"`
	Notes                *string `json:"notes,omitempty"`
}

type UpdatePurchaseOrderItemBody struct {
	ItemID      string  `json:"item_id"`
	Description string  `json:"description"`
	Quantity    string  `json:"quantity"`
	UnitPrice   string  `json:"unit_price"`
	TaxRate     *string `json:"tax_rate,omitempty"`
}

type PurchaseOrderResponse struct {
	ID                   string                     `json:"id"`
	PONumber             string                     `json:"po_number"`
	PurchaseRequestID    *string                    `json:"purchase_request_id,omitempty"`
	VendorID             string                     `json:"vendor_id"`
	Status               domain.PurchaseOrderStatus `json:"status"`
	OrderDate            time.Time                  `json:"order_date"`
	ExpectedDeliveryDate *time.Time                 `json:"expected_delivery_date,omitempty"`
	Currency             string                     `json:"currency"`
	Subtotal             string                     `json:"subtotal"`
	TaxAmount            string                     `json:"tax_amount"`
	TotalAmount          string                     `json:"total_amount"`
	PaymentTerms         *string                    `json:"payment_terms,omitempty"`
	ShippingAddress      *string                    `json:"shipping_address,omitempty"`
	Notes                *string                    `json:"notes,omitempty"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
}

type PurchaseOrderItemResponse struct {
	ID               string `json:"id"`
	PurchaseOrderID  string `json:"purchase_order_id"`
	ItemID           string `json:"item_id"`
	Description      string `json:"description"`
	Quantity         string `json:"quantity"`
	UnitPrice        string `json:"unit_price"`
	TaxRate          string `json:"tax_rate"`
	TotalPrice       string `json:"total_price"`
	ReceivedQuantity string `json:"received_quantity"`
	BilledQuantity   string `json:"billed_quantity"`
}

type CreatePurchaseOrderResponse struct {
	PurchaseOrder PurchaseOrderResponse       `json:"purchase_order"`
	Items         []PurchaseOrderItemResponse `json:"items"`
}

type ListPurchaseOrdersResponse struct {
	PurchaseOrders []PurchaseOrderResponse `json:"purchase_orders"`
	Limit          int                     `json:"limit"`
	Offset         int                     `json:"offset"`
}

type ConvertPRToPOBody struct {
	VendorID             string  `json:"vendor_id"`
	ExpectedDeliveryDate *string `json:"expected_delivery_date,omitempty"`
	Currency             string  `json:"currency,omitempty"`
	PaymentTerms         *string `json:"payment_terms,omitempty"`
	ShippingAddress      *string `json:"shipping_address,omitempty"`
	Notes                *string `json:"notes,omitempty"`
}

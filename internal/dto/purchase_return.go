package dto

import (
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type CreatePurchaseReturnBody struct {
	PurchaseOrderID string                         `json:"purchase_order_id"`
	GoodsReceiptID  *string                        `json:"goods_receipt_id,omitempty"`
	ReturnDate      *string                        `json:"return_date,omitempty"`
	Reason          string                         `json:"reason"`
	Carrier         *string                        `json:"carrier,omitempty"`
	TrackingNumber  *string                        `json:"tracking_number,omitempty"`
	CreatedBy       string                         `json:"created_by"`
	Notes           *string                        `json:"notes,omitempty"`
	Items           []CreatePurchaseReturnItemBody `json:"items"`
}

type CreatePurchaseReturnItemBody struct {
	PurchaseOrderItemID string  `json:"purchase_order_item_id"`
	ReturnQuantity      string  `json:"return_quantity"`
	Reason              *string `json:"reason,omitempty"`
}

type ShipPurchaseReturnBody struct {
	Carrier        string `json:"carrier"`
	TrackingNumber string `json:"tracking_number"`
}

type UpdatePurchaseReturnBody struct {
	Reason         *string `json:"reason,omitempty"`
	Carrier        *string `json:"carrier,omitempty"`
	TrackingNumber *string `json:"tracking_number,omitempty"`
	Notes          *string `json:"notes,omitempty"`
}

type PurchaseReturnResponse struct {
	ID                string                      `json:"id"`
	ReturnNumber      string                      `json:"return_number"`
	PurchaseOrderID   string                      `json:"purchase_order_id"`
	GoodsReceiptID    *string                     `json:"goods_receipt_id,omitempty"`
	VendorID          string                      `json:"vendor_id"`
	Status            domain.PurchaseReturnStatus `json:"status"`
	ReturnDate        time.Time                   `json:"return_date"`
	TotalReturnAmount string                      `json:"total_return_amount"`
	Reason            string                      `json:"reason"`
	Carrier           *string                     `json:"carrier,omitempty"`
	TrackingNumber    *string                     `json:"tracking_number,omitempty"`
	CreatedBy         string                      `json:"created_by"`
	Notes             *string                     `json:"notes,omitempty"`
	CreatedAt         time.Time                   `json:"created_at"`
	UpdatedAt         time.Time                   `json:"updated_at"`
}

type PurchaseReturnItemResponse struct {
	ID                  string    `json:"id"`
	PurchaseReturnID    string    `json:"purchase_return_id"`
	PurchaseOrderItemID string    `json:"purchase_order_item_id"`
	ItemID              string    `json:"item_id"`
	Description         string    `json:"description"`
	ReturnQuantity      string    `json:"return_quantity"`
	UnitPrice           string    `json:"unit_price"`
	TotalPrice          string    `json:"total_price"`
	Reason              *string   `json:"reason,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type CreatePurchaseReturnResponse struct {
	PurchaseReturn PurchaseReturnResponse       `json:"purchase_return"`
	Items          []PurchaseReturnItemResponse `json:"items"`
}

type PurchaseReturnDetailResponse struct {
	PurchaseReturn PurchaseReturnResponse       `json:"purchase_return"`
	Items          []PurchaseReturnItemResponse `json:"items"`
}

type ListPurchaseReturnsResponse struct {
	PurchaseReturns []PurchaseReturnResponse `json:"purchase_returns"`
	Total           int64                    `json:"total"`
	Limit           int                      `json:"limit"`
	Offset          int                      `json:"offset"`
}

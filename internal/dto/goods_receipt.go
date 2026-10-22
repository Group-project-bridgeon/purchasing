package dto

import (
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type CreateGoodsReceiptBody struct {
	PurchaseOrderID string                       `json:"purchase_order_id"`
	ReceivedBy      string                       `json:"received_by"`
	ReceivedAt      *string                      `json:"received_at,omitempty"`
	Notes           *string                      `json:"notes,omitempty"`
	Items           []CreateGoodsReceiptItemBody `json:"items"`
}

type CreateGoodsReceiptItemBody struct {
	PurchaseOrderItemID string  `json:"purchase_order_item_id"`
	ReceivedQuantity    string  `json:"received_quantity"`
	RejectedQuantity    *string `json:"rejected_quantity,omitempty"`
	RejectionReason     *string `json:"rejection_reason,omitempty"`
}

type GoodsReceiptResponse struct {
	ID              string                    `json:"id"`
	GRNNumber       string                    `json:"grn_number"`
	PurchaseOrderID string                    `json:"purchase_order_id"`
	VendorID        string                    `json:"vendor_id"`
	ReceivedBy      string                    `json:"received_by"`
	ReceivedAt      time.Time                 `json:"received_at"`
	Status          domain.GoodsReceiptStatus `json:"status"`
	Notes           *string                   `json:"notes,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

type GoodsReceiptItemResponse struct {
	ID                  string  `json:"id"`
	GoodsReceiptID      string  `json:"goods_receipt_id"`
	PurchaseOrderItemID string  `json:"purchase_order_item_id"`
	ItemID              string  `json:"item_id"`
	ReceivedQuantity    string  `json:"received_quantity"`
	RejectedQuantity    string  `json:"rejected_quantity"`
	RejectionReason     *string `json:"rejection_reason,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

type CreateGoodsReceiptResponse struct {
	GoodsReceipt GoodsReceiptResponse       `json:"goods_receipt"`
	Items        []GoodsReceiptItemResponse `json:"items"`
}

type ListGoodsReceiptsResponse struct {
	GoodsReceipts []GoodsReceiptResponse `json:"goods_receipts"`
	Limit         int                    `json:"limit"`
	Offset        int                    `json:"offset"`
}

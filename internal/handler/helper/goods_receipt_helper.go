package helper

import (
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/shopspring/decimal"
)

func ParseCreateGoodsReceipt(
	body *dto.CreateGoodsReceiptBody,
) (*service.CreateGoodsReceiptInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	var receivedAt *time.Time
	if body.ReceivedAt != nil && *body.ReceivedAt != "" {
		parsedDate, err := time.Parse(time.RFC3339, *body.ReceivedAt)
		if err != nil {
			parsedDate, err = time.Parse("2006-01-02", *body.ReceivedAt)
			if err != nil {
				return nil, fmt.Errorf("invalid received_at date: %w", err)
			}
		}
		receivedAt = &parsedDate
	}

	items := make([]*domain.GoodsReceiptItem, 0, len(body.Items))
	for i, itemBody := range body.Items {
		receivedQty, err := decimal.NewFromString(itemBody.ReceivedQuantity)
		if err != nil {
			return nil, fmt.Errorf("invalid received_quantity for item %d: %w", i+1, err)
		}

		rejectedQty := decimal.Zero
		if itemBody.RejectedQuantity != nil && *itemBody.RejectedQuantity != "" {
			parsedRejected, err := decimal.NewFromString(*itemBody.RejectedQuantity)
			if err != nil {
				return nil, fmt.Errorf("invalid rejected_quantity for item %d: %w", i+1, err)
			}
			rejectedQty = parsedRejected
		}

		items = append(items, &domain.GoodsReceiptItem{
			PurchaseOrderItemID: itemBody.PurchaseOrderItemID,
			ReceivedQuantity:    receivedQty,
			RejectedQuantity:    rejectedQty,
			RejectionReason:     itemBody.RejectionReason,
		})
	}

	return &service.CreateGoodsReceiptInput{
		PurchaseOrderID: body.PurchaseOrderID,
		ReceivedBy:      body.ReceivedBy,
		ReceivedAt:      receivedAt,
		Notes:           body.Notes,
		Items:           items,
	}, nil
}

func ToGoodsReceiptResponse(
	grn *domain.GoodsReceipt,
) dto.GoodsReceiptResponse {
	return dto.GoodsReceiptResponse{
		ID:              grn.ID,
		GRNNumber:       grn.GRNNumber,
		PurchaseOrderID: grn.PurchaseOrderID,
		VendorID:        grn.VendorID,
		ReceivedBy:      grn.ReceivedBy,
		ReceivedAt:      grn.ReceivedAt,
		Status:          grn.Status,
		Notes:           grn.Notes,
		CreatedAt:       grn.CreatedAt,
		UpdatedAt:       grn.UpdatedAt,
	}
}

func ToGoodsReceiptItemResponse(
	item *domain.GoodsReceiptItem,
) dto.GoodsReceiptItemResponse {
	return dto.GoodsReceiptItemResponse{
		ID:                  item.ID,
		GoodsReceiptID:      item.GoodsReceiptID,
		PurchaseOrderItemID: item.PurchaseOrderItemID,
		ItemID:              item.ItemID,
		ReceivedQuantity:    item.ReceivedQuantity.String(),
		RejectedQuantity:    item.RejectedQuantity.String(),
		RejectionReason:     item.RejectionReason,
		CreatedAt:           item.CreatedAt,
	}
}

func ToCreateGoodsReceiptResponse(
	grn *domain.GoodsReceipt,
	items []*domain.GoodsReceiptItem,
) dto.CreateGoodsReceiptResponse {
	itemResponses := make([]dto.GoodsReceiptItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, ToGoodsReceiptItemResponse(item))
	}

	return dto.CreateGoodsReceiptResponse{
		GoodsReceipt: ToGoodsReceiptResponse(grn),
		Items:        itemResponses,
	}
}

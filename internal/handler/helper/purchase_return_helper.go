package helper

import (
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/shopspring/decimal"
)

func ParseCreatePurchaseReturn(
	body *dto.CreatePurchaseReturnBody,
) (*service.CreatePurchaseReturnInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	var returnDate *time.Time
	if body.ReturnDate != nil && *body.ReturnDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *body.ReturnDate)
		if err != nil {
			parsedDate, err = time.Parse(time.RFC3339, *body.ReturnDate)
			if err != nil {
				return nil, fmt.Errorf("invalid return_date (expected YYYY-MM-DD): %w", err)
			}
		}
		returnDate = &parsedDate
	}

	items := make([]service.CreatePurchaseReturnItemInput, 0, len(body.Items))
	for i, itemBody := range body.Items {
		quantity, err := decimal.NewFromString(itemBody.ReturnQuantity)
		if err != nil {
			return nil, fmt.Errorf("invalid return_quantity for item %d: %w", i+1, err)
		}

		items = append(items, service.CreatePurchaseReturnItemInput{
			PurchaseOrderItemID: itemBody.PurchaseOrderItemID,
			ReturnQuantity:      quantity,
			Reason:              itemBody.Reason,
		})
	}

	return &service.CreatePurchaseReturnInput{
		PurchaseOrderID: body.PurchaseOrderID,
		GoodsReceiptID:  body.GoodsReceiptID,
		ReturnDate:      returnDate,
		Reason:          body.Reason,
		Carrier:         body.Carrier,
		TrackingNumber:  body.TrackingNumber,
		CreatedBy:       body.CreatedBy,
		Notes:           body.Notes,
		Items:           items,
	}, nil
}

func MapPurchaseReturnToResponse(
	ret *domain.PurchaseReturn,
) dto.PurchaseReturnResponse {
	return dto.PurchaseReturnResponse{
		ApprovedBy: ret.ApprovedBy, ApprovedAt: ret.ApprovedAt,
		ID:                ret.ID,
		ReturnNumber:      ret.ReturnNumber,
		PurchaseOrderID:   ret.PurchaseOrderID,
		GoodsReceiptID:    ret.GoodsReceiptID,
		VendorID:          ret.VendorID,
		Status:            ret.Status,
		ReturnDate:        ret.ReturnDate,
		TotalReturnAmount: ret.TotalReturnAmount.StringFixed(2),
		Reason:            ret.Reason,
		Carrier:           ret.Carrier,
		TrackingNumber:    ret.TrackingNumber,
		CreatedBy:         ret.CreatedBy,
		Notes:             ret.Notes,
		CreatedAt:         ret.CreatedAt,
		UpdatedAt:         ret.UpdatedAt,
	}
}

func MapPurchaseReturnItemToResponse(
	item *domain.PurchaseReturnItem,
) dto.PurchaseReturnItemResponse {
	return dto.PurchaseReturnItemResponse{
		ID:                  item.ID,
		PurchaseReturnID:    item.PurchaseReturnID,
		PurchaseOrderItemID: item.PurchaseOrderItemID,
		ItemID:              item.ItemID,
		Description:         item.Description,
		ReturnQuantity:      item.ReturnQuantity.String(),
		UnitPrice:           item.UnitPrice.StringFixed(2),
		TotalPrice:          item.TotalPrice.StringFixed(2),
		Reason:              item.Reason,
		CreatedAt:           item.CreatedAt,
	}
}

func MapPurchaseReturnDetailToResponse(
	detail *service.PurchaseReturnDetail,
) dto.PurchaseReturnDetailResponse {
	itemsResp := make([]dto.PurchaseReturnItemResponse, 0, len(detail.Items))
	for _, item := range detail.Items {
		itemsResp = append(itemsResp, MapPurchaseReturnItemToResponse(item))
	}

	return dto.PurchaseReturnDetailResponse{
		PurchaseReturn: MapPurchaseReturnToResponse(detail.Return),
		Items:          itemsResp,
	}
}

func MapPurchaseReturnListToResponse(
	returns []*domain.PurchaseReturn,
	total int64,
	limit int,
	offset int,
) dto.ListPurchaseReturnsResponse {
	respList := make([]dto.PurchaseReturnResponse, 0, len(returns))
	for _, ret := range returns {
		respList = append(respList, MapPurchaseReturnToResponse(ret))
	}

	return dto.ListPurchaseReturnsResponse{
		PurchaseReturns: respList,
		Total:           total,
		Limit:           limit,
		Offset:          offset,
	}
}

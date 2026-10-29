package helper

import (
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/shopspring/decimal"
)

func ParseCreateVendorBill(
	body *dto.CreateVendorBillBody,
) (*service.CreateVendorBillInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	billDate, err := time.Parse("2006-01-02", body.BillDate)
	if err != nil {
		billDate, err = time.Parse(time.RFC3339, body.BillDate)
		if err != nil {
			return nil, fmt.Errorf("invalid bill_date (expected YYYY-MM-DD): %w", err)
		}
	}

	var dueDate *time.Time
	if body.DueDate != nil && *body.DueDate != "" {
		parsedDue, err := time.Parse("2006-01-02", *body.DueDate)
		if err != nil {
			parsedDue, err = time.Parse(time.RFC3339, *body.DueDate)
			if err != nil {
				return nil, fmt.Errorf("invalid due_date (expected YYYY-MM-DD): %w", err)
			}
		}
		dueDate = &parsedDue
	}

	currency := body.Currency
	if currency == "" {
		currency = "USD"
	}

	items := make([]*domain.VendorBillItem, 0, len(body.Items))
	for i, itemBody := range body.Items {
		quantity, err := decimal.NewFromString(itemBody.Quantity)
		if err != nil {
			return nil, fmt.Errorf("invalid quantity for item %d: %w", i+1, err)
		}

		unitPrice, err := decimal.NewFromString(itemBody.UnitPrice)
		if err != nil {
			return nil, fmt.Errorf("invalid unit_price for item %d: %w", i+1, err)
		}

		taxRate := decimal.Zero
		if itemBody.TaxRate != nil && *itemBody.TaxRate != "" {
			parsedTax, err := decimal.NewFromString(*itemBody.TaxRate)
			if err != nil {
				return nil, fmt.Errorf("invalid tax_rate for item %d: %w", i+1, err)
			}
			taxRate = parsedTax
		}

		items = append(items, &domain.VendorBillItem{
			PurchaseOrderItemID: itemBody.PurchaseOrderItemID,
			Description:         itemBody.Description,
			Quantity:            quantity,
			UnitPrice:           unitPrice,
			TaxRate:             taxRate,
		})
	}

	return &service.CreateVendorBillInput{
		BillNumber:      body.BillNumber,
		PurchaseOrderID: body.PurchaseOrderID,
		BillDate:        billDate,
		DueDate:         dueDate,
		Currency:        currency,
		Notes:           body.Notes,
		Items:           items,
	}, nil
}

func ToVendorBillResponse(
	bill *domain.VendorBill,
) dto.VendorBillResponse {
	return dto.VendorBillResponse{
		ID:                 bill.ID,
		InternalBillNumber: bill.InternalBillNumber,
		BillNumber:         bill.BillNumber,
		PurchaseOrderID:    bill.PurchaseOrderID,
		VendorID:           bill.VendorID,
		BillDate:           bill.BillDate,
		DueDate:            bill.DueDate,
		Currency:           bill.Currency,
		Subtotal:           bill.Subtotal.String(),
		TaxAmount:          bill.TaxAmount.String(),
		TotalAmount:        bill.TotalAmount.String(),
		MatchStatus:        bill.MatchStatus,
		Status:             bill.Status,
		DiscrepancyReason:  bill.DiscrepancyReason,
		Notes:              bill.Notes,
		CreatedAt:          bill.CreatedAt,
		UpdatedAt:          bill.UpdatedAt,
	}
}

func ToVendorBillItemResponse(
	item *domain.VendorBillItem,
) dto.VendorBillItemResponse {
	return dto.VendorBillItemResponse{
		ID:                  item.ID,
		VendorBillID:        item.VendorBillID,
		PurchaseOrderItemID: item.PurchaseOrderItemID,
		ItemID:              item.ItemID,
		Description:         item.Description,
		Quantity:            item.Quantity.String(),
		UnitPrice:           item.UnitPrice.String(),
		TaxRate:             item.TaxRate.String(),
		TotalPrice:          item.TotalPrice.String(),
		MatchedQuantity:     item.MatchedQuantity.String(),
		MatchStatus:         item.MatchStatus,
		CreatedAt:           item.CreatedAt,
	}
}

func ToCreateVendorBillResponse(
	bill *domain.VendorBill,
	items []*domain.VendorBillItem,
) dto.CreateVendorBillResponse {
	itemResponses := make([]dto.VendorBillItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, ToVendorBillItemResponse(item))
	}

	return dto.CreateVendorBillResponse{
		VendorBill: ToVendorBillResponse(bill),
		Items:      itemResponses,
	}
}

func ToMatchBillResponse(
	result *service.MatchResult,
) dto.MatchBillResponse {
	itemResponses := make([]dto.VendorBillItemResponse, 0, len(result.Items))
	for _, item := range result.Items {
		itemResponses = append(itemResponses, ToVendorBillItemResponse(item))
	}

	return dto.MatchBillResponse{
		VendorBill:    ToVendorBillResponse(result.Bill),
		Items:         itemResponses,
		IsMatched:     result.IsMatched,
		Discrepancies: result.Discrepancies,
	}
}

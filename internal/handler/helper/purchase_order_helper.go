package helper

import (
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/shopspring/decimal"
)

func ParseCreatePurchaseOrder(
	body *dto.CreatePurchaseOrderBody,
) (*service.CreatePurchaseOrderInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	var expectedDeliveryDate *time.Time
	if body.ExpectedDeliveryDate != nil && *body.ExpectedDeliveryDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *body.ExpectedDeliveryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expected_delivery_date: %w", err)
		}
		expectedDeliveryDate = &parsedDate
	}

	currency := body.Currency
	if currency == "" {
		currency = "USD"
	}

	order := &domain.PurchaseOrder{
		PurchaseRequestID:    body.PurchaseRequestID,
		VendorID:             body.VendorID,
		ExpectedDeliveryDate: expectedDeliveryDate,
		Currency:             currency,
		PaymentTerms:         body.PaymentTerms,
		ShippingAddress:      body.ShippingAddress,
		Notes:                body.Notes,
	}

	items := make([]*domain.PurchaseOrderItem, 0, len(body.Items))
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

		items = append(items, &domain.PurchaseOrderItem{
			ItemID:      itemBody.ItemID,
			Description: itemBody.Description,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
			TaxRate:     taxRate,
		})
	}

	return &service.CreatePurchaseOrderInput{
		Order: order,
		Items: items,
	}, nil
}

func ParseConvertPRToPO(
	body *dto.ConvertPRToPOBody,
	prID string,
) (*service.ConvertPRToPOInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	var expectedDeliveryDate *time.Time
	if body.ExpectedDeliveryDate != nil && *body.ExpectedDeliveryDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *body.ExpectedDeliveryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expected_delivery_date: %w", err)
		}
		expectedDeliveryDate = &parsedDate
	}

	currency := body.Currency
	if currency == "" {
		currency = "USD"
	}

	return &service.ConvertPRToPOInput{
		PurchaseRequestID:    prID,
		VendorID:             body.VendorID,
		ExpectedDeliveryDate: expectedDeliveryDate,
		Currency:             currency,
		PaymentTerms:         body.PaymentTerms,
		ShippingAddress:      body.ShippingAddress,
		Notes:                body.Notes,
	}, nil
}

func ParseUpdatePurchaseOrder(
	body *dto.UpdatePurchaseOrderBody,
) (*domain.PurchaseOrder, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	var expectedDeliveryDate *time.Time
	if body.ExpectedDeliveryDate != nil && *body.ExpectedDeliveryDate != "" {
		parsedDate, err := time.Parse("2006-01-02", *body.ExpectedDeliveryDate)
		if err != nil {
			return nil, fmt.Errorf("invalid expected_delivery_date: %w", err)
		}
		expectedDeliveryDate = &parsedDate
	}

	return &domain.PurchaseOrder{
		ExpectedDeliveryDate: expectedDeliveryDate,
		Currency:             body.Currency,
		PaymentTerms:         body.PaymentTerms,
		ShippingAddress:      body.ShippingAddress,
		Notes:                body.Notes,
	}, nil
}

func ParseCreatePurchaseOrderItem(
	body *dto.CreatePurchaseOrderItemBody,
) (*domain.PurchaseOrderItem, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	quantity, err := decimal.NewFromString(body.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	unitPrice, err := decimal.NewFromString(body.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid unit_price: %w", err)
	}

	taxRate := decimal.Zero
	if body.TaxRate != nil && *body.TaxRate != "" {
		parsedTax, err := decimal.NewFromString(*body.TaxRate)
		if err != nil {
			return nil, fmt.Errorf("invalid tax_rate: %w", err)
		}
		taxRate = parsedTax
	}

	return &domain.PurchaseOrderItem{
		ItemID:      body.ItemID,
		Description: body.Description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TaxRate:     taxRate,
	}, nil
}

func ParseUpdatePurchaseOrderItem(
	body *dto.UpdatePurchaseOrderItemBody,
) (*domain.PurchaseOrderItem, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	quantity, err := decimal.NewFromString(body.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	unitPrice, err := decimal.NewFromString(body.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid unit_price: %w", err)
	}

	taxRate := decimal.Zero
	if body.TaxRate != nil && *body.TaxRate != "" {
		parsedTax, err := decimal.NewFromString(*body.TaxRate)
		if err != nil {
			return nil, fmt.Errorf("invalid tax_rate: %w", err)
		}
		taxRate = parsedTax
	}

	return &domain.PurchaseOrderItem{
		ItemID:      body.ItemID,
		Description: body.Description,
		Quantity:    quantity,
		UnitPrice:   unitPrice,
		TaxRate:     taxRate,
	}, nil
}

func ToPurchaseOrderResponse(
	order *domain.PurchaseOrder,
) dto.PurchaseOrderResponse {
	return dto.PurchaseOrderResponse{
		ApprovedBy: order.ApprovedBy, ApprovedAt: order.ApprovedAt, CreatedBy: order.CreatedBy,
		ID:                   order.ID,
		PONumber:             order.PONumber,
		PurchaseRequestID:    order.PurchaseRequestID,
		VendorID:             order.VendorID,
		Status:               order.Status,
		OrderDate:            order.OrderDate,
		ExpectedDeliveryDate: order.ExpectedDeliveryDate,
		Currency:             order.Currency,
		Subtotal:             order.Subtotal.String(),
		TaxAmount:            order.TaxAmount.String(),
		TotalAmount:          order.TotalAmount.String(),
		PaymentTerms:         order.PaymentTerms,
		ShippingAddress:      order.ShippingAddress,
		Notes:                order.Notes,
		CreatedAt:            order.CreatedAt,
		UpdatedAt:            order.UpdatedAt,
	}
}

func ToPurchaseOrderItemResponse(
	item *domain.PurchaseOrderItem,
) dto.PurchaseOrderItemResponse {
	return dto.PurchaseOrderItemResponse{
		ID:               item.ID,
		PurchaseOrderID:  item.PurchaseOrderID,
		ItemID:           item.ItemID,
		Description:      item.Description,
		Quantity:         item.Quantity.String(),
		UnitPrice:        item.UnitPrice.String(),
		TaxRate:          item.TaxRate.String(),
		TotalPrice:       item.TotalPrice.String(),
		ReceivedQuantity: item.ReceivedQuantity.String(),
		BilledQuantity:   item.BilledQuantity.String(),
	}
}

func ToCreatePurchaseOrderResponse(
	order *domain.PurchaseOrder,
	items []*domain.PurchaseOrderItem,
) dto.CreatePurchaseOrderResponse {
	itemResponses := make([]dto.PurchaseOrderItemResponse, 0, len(items))
	for _, item := range items {
		itemResponses = append(itemResponses, ToPurchaseOrderItemResponse(item))
	}

	return dto.CreatePurchaseOrderResponse{
		PurchaseOrder: ToPurchaseOrderResponse(order),
		Items:         itemResponses,
	}
}

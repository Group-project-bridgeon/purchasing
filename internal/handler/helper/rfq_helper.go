package helper

import (
	"fmt"
	"strings"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/shopspring/decimal"
)

func ParseCreateRFQ(
	body *dto.CreateRFQBody,
) (*service.CreateRFQInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	deadline, err := time.Parse(time.RFC3339, body.SubmissionDeadline)
	if err != nil {
		deadline, err = time.Parse("2006-01-02", body.SubmissionDeadline)
		if err != nil {
			return nil, fmt.Errorf("invalid submission_deadline (expected RFC3339 or YYYY-MM-DD): %w", err)
		}
	}

	items := make([]service.CreateRFQItemInput, 0, len(body.Items))
	for i, item := range body.Items {
		qty, err := decimal.NewFromString(item.TargetQuantity)
		if err != nil {
			return nil, fmt.Errorf("invalid target_quantity for item %d: %w", i+1, err)
		}

		var deliveryDate *time.Time
		if item.TargetDeliveryDate != nil && *item.TargetDeliveryDate != "" {
			d, err := time.Parse("2006-01-02", *item.TargetDeliveryDate)
			if err != nil {
				d, err = time.Parse(time.RFC3339, *item.TargetDeliveryDate)
				if err != nil {
					return nil, fmt.Errorf("invalid target_delivery_date for item %d: %w", i+1, err)
				}
			}
			deliveryDate = &d
		}

		items = append(items, service.CreateRFQItemInput{
			ItemID:             item.ItemID,
			Description:        item.Description,
			TargetQuantity:     qty,
			TargetDeliveryDate: deliveryDate,
		})
	}

	return &service.CreateRFQInput{
		Title:              body.Title,
		PurchaseRequestID:  body.PurchaseRequestID,
		SubmissionDeadline: deadline,
		CreatedBy:          body.CreatedBy,
		Notes:              body.Notes,
		Items:              items,
	}, nil
}

func ParseCreateRFQFromPR(
	body *dto.CreateRFQFromPRBody,
) (*service.CreateRFQFromPRInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	deadline, err := time.Parse(time.RFC3339, body.SubmissionDeadline)
	if err != nil {
		deadline, err = time.Parse("2006-01-02", body.SubmissionDeadline)
		if err != nil {
			return nil, fmt.Errorf("invalid submission_deadline (expected RFC3339 or YYYY-MM-DD): %w", err)
		}
	}

	return &service.CreateRFQFromPRInput{
		Title:              body.Title,
		SubmissionDeadline: deadline,
		CreatedBy:          body.CreatedBy,
		Notes:              body.Notes,
	}, nil
}

func ParseSubmitVendorQuotation(
	body *dto.SubmitVendorQuotationBody,
	rfqID string,
) (*service.SubmitVendorQuotationInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	validUntil, err := time.Parse("2006-01-02", body.ValidUntil)
	if err != nil {
		validUntil, err = time.Parse(time.RFC3339, body.ValidUntil)
		if err != nil {
			return nil, fmt.Errorf("invalid valid_until date (expected YYYY-MM-DD): %w", err)
		}
	}

	var quoteDate *time.Time
	if body.QuotationDate != nil && *body.QuotationDate != "" {
		qd, err := time.Parse("2006-01-02", *body.QuotationDate)
		if err != nil {
			qd, err = time.Parse(time.RFC3339, *body.QuotationDate)
			if err != nil {
				return nil, fmt.Errorf("invalid quotation_date (expected YYYY-MM-DD): %w", err)
			}
		}
		quoteDate = &qd
	}

	leadTime := 0
	if body.LeadTimeDays != nil {
		leadTime = *body.LeadTimeDays
	}

	currency := "USD"
	if body.Currency != nil && strings.TrimSpace(*body.Currency) != "" {
		currency = strings.TrimSpace(*body.Currency)
	}

	items := make([]service.SubmitVendorQuotationItemInput, 0, len(body.Items))
	for i, item := range body.Items {
		qty, err := decimal.NewFromString(item.OfferedQuantity)
		if err != nil {
			return nil, fmt.Errorf("invalid offered_quantity for item %d: %w", i+1, err)
		}

		unitPrice, err := decimal.NewFromString(item.UnitPrice)
		if err != nil {
			return nil, fmt.Errorf("invalid unit_price for item %d: %w", i+1, err)
		}

		taxRate := decimal.Zero
		if item.TaxRate != nil && strings.TrimSpace(*item.TaxRate) != "" {
			tr, err := decimal.NewFromString(*item.TaxRate)
			if err != nil {
				return nil, fmt.Errorf("invalid tax_rate for item %d: %w", i+1, err)
			}
			taxRate = tr
		}

		items = append(items, service.SubmitVendorQuotationItemInput{
			RFQItemID:       item.RFQItemID,
			ItemID:          item.ItemID,
			Description:     item.Description,
			OfferedQuantity: qty,
			UnitPrice:       unitPrice,
			TaxRate:         taxRate,
			LeadTimeDays:    item.LeadTimeDays,
		})
	}

	return &service.SubmitVendorQuotationInput{
		RFQID:          rfqID,
		VendorID:       body.VendorID,
		QuoteReference: body.QuoteReference,
		QuotationDate:  quoteDate,
		ValidUntil:     validUntil,
		Currency:       currency,
		PaymentTerms:   body.PaymentTerms,
		LeadTimeDays:   leadTime,
		Notes:          body.Notes,
		Items:          items,
	}, nil
}

func MapRFQToResponse(rfq *domain.RFQ) dto.RFQResponse {
	return dto.RFQResponse{
		ID:                 rfq.ID,
		RFQNumber:          rfq.RFQNumber,
		PurchaseRequestID:  rfq.PurchaseRequestID,
		Title:              rfq.Title,
		Status:             rfq.Status,
		SubmissionDeadline: rfq.SubmissionDeadline,
		CreatedBy:          rfq.CreatedBy,
		Notes:              rfq.Notes,
		CreatedAt:          rfq.CreatedAt,
		UpdatedAt:          rfq.UpdatedAt,
	}
}

func MapRFQItemToResponse(item *domain.RFQItem) dto.RFQItemResponse {
	return dto.RFQItemResponse{
		ID:                 item.ID,
		RFQID:              item.RFQID,
		ItemID:             item.ItemID,
		Description:        item.Description,
		TargetQuantity:     item.TargetQuantity.String(),
		TargetDeliveryDate: item.TargetDeliveryDate,
		CreatedAt:          item.CreatedAt,
	}
}

func MapRFQDetailToResponse(detail *service.RFQDetail) dto.RFQDetailResponse {
	items := make([]dto.RFQItemResponse, 0, len(detail.Items))
	for _, item := range detail.Items {
		items = append(items, MapRFQItemToResponse(item))
	}
	return dto.RFQDetailResponse{
		RFQ:   MapRFQToResponse(detail.RFQ),
		Items: items,
	}
}

func MapRFQListToResponse(rfqs []*domain.RFQ, total int64, limit int, offset int) dto.ListRFQsResponse {
	list := make([]dto.RFQResponse, 0, len(rfqs))
	for _, rfq := range rfqs {
		list = append(list, MapRFQToResponse(rfq))
	}
	return dto.ListRFQsResponse{
		RFQs:   list,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}
}

func MapVendorQuotationToResponse(q *domain.VendorQuotation) dto.VendorQuotationResponse {
	return dto.VendorQuotationResponse{
		ID:             q.ID,
		RFQID:          q.RFQID,
		VendorID:       q.VendorID,
		QuoteReference: q.QuoteReference,
		QuotationDate:  q.QuotationDate,
		ValidUntil:     q.ValidUntil,
		Currency:       q.Currency,
		Subtotal:       q.Subtotal.StringFixed(2),
		TaxAmount:      q.TaxAmount.StringFixed(2),
		TotalAmount:    q.TotalAmount.StringFixed(2),
		PaymentTerms:   q.PaymentTerms,
		LeadTimeDays:   q.LeadTimeDays,
		Status:         q.Status,
		Notes:          q.Notes,
		CreatedAt:      q.CreatedAt,
		UpdatedAt:      q.UpdatedAt,
	}
}

func MapVendorQuotationItemToResponse(item *domain.VendorQuotationItem) dto.VendorQuotationItemResponse {
	return dto.VendorQuotationItemResponse{
		ID:                item.ID,
		VendorQuotationID: item.VendorQuotationID,
		RFQItemID:         item.RFQItemID,
		ItemID:            item.ItemID,
		Description:       item.Description,
		OfferedQuantity:   item.OfferedQuantity.String(),
		UnitPrice:         item.UnitPrice.StringFixed(2),
		TaxRate:           item.TaxRate.StringFixed(2),
		TotalPrice:        item.TotalPrice.StringFixed(2),
		LeadTimeDays:      item.LeadTimeDays,
		CreatedAt:         item.CreatedAt,
	}
}

func MapVendorQuotationDetailToResponse(detail *service.VendorQuotationDetail) dto.VendorQuotationDetailResponse {
	items := make([]dto.VendorQuotationItemResponse, 0, len(detail.Items))
	for _, item := range detail.Items {
		items = append(items, MapVendorQuotationItemToResponse(item))
	}
	return dto.VendorQuotationDetailResponse{
		Quotation: MapVendorQuotationToResponse(detail.Quotation),
		Items:     items,
	}
}

func MapComparisonMatrixToResponse(matrix *service.QuotationComparisonMatrix) dto.QuotationComparisonMatrixResponse {
	quotes := make([]dto.VendorQuotationResponse, 0, len(matrix.Quotations))
	for _, q := range matrix.Quotations {
		quotes = append(quotes, MapVendorQuotationToResponse(q))
	}
	return dto.QuotationComparisonMatrixResponse{
		RFQID:                      matrix.RFQID,
		LowestBidQuotationID:       matrix.LowestBidQuotationID,
		FastestLeadTimeQuotationID: matrix.FastestLeadTimeQuotationID,
		Quotations:                 quotes,
	}
}

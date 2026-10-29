package dto

import (
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type CreateRFQBody struct {
	Title              string              `json:"title"`
	PurchaseRequestID  *string             `json:"purchase_request_id,omitempty"`
	SubmissionDeadline string              `json:"submission_deadline"`
	CreatedBy          string              `json:"created_by"`
	Notes              *string             `json:"notes,omitempty"`
	Items              []CreateRFQItemBody `json:"items"`
}

type CreateRFQItemBody struct {
	ItemID             string  `json:"item_id"`
	Description        string  `json:"description"`
	TargetQuantity     string  `json:"target_quantity"`
	TargetDeliveryDate *string `json:"target_delivery_date,omitempty"`
}

type CreateRFQFromPRBody struct {
	Title              string  `json:"title"`
	SubmissionDeadline string  `json:"submission_deadline"`
	CreatedBy          string  `json:"created_by"`
	Notes              *string `json:"notes,omitempty"`
}

type UpdateRFQBody struct {
	Title              *string `json:"title,omitempty"`
	SubmissionDeadline *string `json:"submission_deadline,omitempty"`
	Notes              *string `json:"notes,omitempty"`
}

type RFQResponse struct {
	ID                 string           `json:"id"`
	RFQNumber          string           `json:"rfq_number"`
	PurchaseRequestID  *string          `json:"purchase_request_id,omitempty"`
	Title              string           `json:"title"`
	Status             domain.RFQStatus `json:"status"`
	SubmissionDeadline time.Time        `json:"submission_deadline"`
	CreatedBy          string           `json:"created_by"`
	Notes              *string          `json:"notes,omitempty"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
}

type RFQItemResponse struct {
	ID                 string     `json:"id"`
	RFQID              string     `json:"rfq_id"`
	ItemID             string     `json:"item_id"`
	Description        string     `json:"description"`
	TargetQuantity     string     `json:"target_quantity"`
	TargetDeliveryDate *time.Time `json:"target_delivery_date,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
}

type RFQDetailResponse struct {
	RFQ   RFQResponse       `json:"rfq"`
	Items []RFQItemResponse `json:"items"`
}

type ListRFQsResponse struct {
	RFQs   []RFQResponse `json:"rfqs"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

type SubmitVendorQuotationBody struct {
	VendorID       string                          `json:"vendor_id"`
	QuoteReference string                          `json:"quote_reference"`
	QuotationDate  *string                         `json:"quotation_date,omitempty"`
	ValidUntil     string                          `json:"valid_until"`
	Currency       *string                         `json:"currency,omitempty"`
	PaymentTerms   *string                         `json:"payment_terms,omitempty"`
	LeadTimeDays   *int                            `json:"lead_time_days,omitempty"`
	Notes          *string                         `json:"notes,omitempty"`
	Items          []SubmitVendorQuotationItemBody `json:"items"`
}

type SubmitVendorQuotationItemBody struct {
	RFQItemID       string  `json:"rfq_item_id"`
	ItemID          string  `json:"item_id"`
	Description     string  `json:"description"`
	OfferedQuantity string  `json:"offered_quantity"`
	UnitPrice       string  `json:"unit_price"`
	TaxRate         *string `json:"tax_rate,omitempty"`
	LeadTimeDays    *int    `json:"lead_time_days,omitempty"`
}

type VendorQuotationResponse struct {
	ID             string                 `json:"id"`
	RFQID          string                 `json:"rfq_id"`
	VendorID       string                 `json:"vendor_id"`
	QuoteReference string                 `json:"quote_reference"`
	QuotationDate  time.Time              `json:"quotation_date"`
	ValidUntil     time.Time              `json:"valid_until"`
	Currency       string                 `json:"currency"`
	Subtotal       string                 `json:"subtotal"`
	TaxAmount      string                 `json:"tax_amount"`
	TotalAmount    string                 `json:"total_amount"`
	PaymentTerms   *string                `json:"payment_terms,omitempty"`
	LeadTimeDays   int                    `json:"lead_time_days"`
	Status         domain.QuotationStatus `json:"status"`
	Notes          *string                `json:"notes,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

type VendorQuotationItemResponse struct {
	ID                string    `json:"id"`
	VendorQuotationID string    `json:"vendor_quotation_id"`
	RFQItemID         string    `json:"rfq_item_id"`
	ItemID            string    `json:"item_id"`
	Description       string    `json:"description"`
	OfferedQuantity   string    `json:"offered_quantity"`
	UnitPrice         string    `json:"unit_price"`
	TaxRate           string    `json:"tax_rate"`
	TotalPrice        string    `json:"total_price"`
	LeadTimeDays      *int      `json:"lead_time_days,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

type VendorQuotationDetailResponse struct {
	Quotation VendorQuotationResponse       `json:"quotation"`
	Items     []VendorQuotationItemResponse `json:"items"`
}

type QuotationComparisonMatrixResponse struct {
	RFQID                      string                    `json:"rfq_id"`
	LowestBidQuotationID       *string                   `json:"lowest_bid_quotation_id,omitempty"`
	FastestLeadTimeQuotationID *string                   `json:"fastest_lead_time_quotation_id,omitempty"`
	Quotations                 []VendorQuotationResponse `json:"quotations"`
}

type AwardQuotationResponse struct {
	AwardedQuotationID string `json:"awarded_quotation_id"`
	PurchaseOrderID    string `json:"purchase_order_id"`
	PONumber           string `json:"po_number"`
	Status             string `json:"status"`
}

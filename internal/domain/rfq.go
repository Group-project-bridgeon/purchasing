package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type RFQStatus string

const (
	RFQStatusDraft     RFQStatus = "draft"
	RFQStatusPublished RFQStatus = "published"
	RFQStatusClosed    RFQStatus = "closed"
	RFQStatusCancelled RFQStatus = "cancelled"
)

type RFQ struct {
	ID                 string
	RFQNumber          string
	PurchaseRequestID  *string
	Title              string
	Status             RFQStatus
	SubmissionDeadline time.Time
	CreatedBy          string
	Notes              *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type RFQItem struct {
	ID                 string
	RFQID              string
	ItemID             string
	Description        string
	TargetQuantity     decimal.Decimal
	TargetDeliveryDate *time.Time
	CreatedAt          time.Time
}

type QuotationStatus string

const (
	QuotationStatusDraft     QuotationStatus = "draft"
	QuotationStatusSubmitted QuotationStatus = "submitted"
	QuotationStatusAwarded   QuotationStatus = "awarded"
	QuotationStatusRejected  QuotationStatus = "rejected"
)

type VendorQuotation struct {
	ID             string
	RFQID          string
	VendorID       string
	QuoteReference string
	QuotationDate  time.Time
	ValidUntil     time.Time
	Currency       string
	Subtotal       decimal.Decimal
	TaxAmount      decimal.Decimal
	TotalAmount    decimal.Decimal
	PaymentTerms   *string
	LeadTimeDays   int
	Status         QuotationStatus
	Notes          *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type VendorQuotationItem struct {
	ID                string
	VendorQuotationID string
	RFQItemID         string
	ItemID            string
	Description       string
	OfferedQuantity   decimal.Decimal
	UnitPrice         decimal.Decimal
	TaxRate           decimal.Decimal
	TotalPrice        decimal.Decimal
	LeadTimeDays      *int
	CreatedAt         time.Time
}

package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type PurchaseRequestStatus string

const (
	PurchaseRequestStatusDraft     PurchaseRequestStatus = "draft"
	PurchaseRequestStatusSubmitted PurchaseRequestStatus = "submitted"
	PurchaseRequestStatusApproved  PurchaseRequestStatus = "approved"
	PurchaseRequestStatusRejected  PurchaseRequestStatus = "rejected"
)

type PurchaseRequest struct {
	ID            string
	RequestNumber string

	RequestedBy  string
	DepartmentID *string

	Status PurchaseRequestStatus

	RequestedAt time.Time
	RequiredBy  *time.Time

	Purpose string
	Notes   *string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type PurchaseRequestItem struct {
	ID                string
	PurchaseRequestID string
	ItemID            string

	Description string

	Quantity           decimal.Decimal
	EstimatedUnitPrice *decimal.Decimal

	CreatedAt time.Time
}

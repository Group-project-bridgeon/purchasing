package dto

import (
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type CreatePurchaseRequestBody struct {
	RequestedBy  string                          `json:"requested_by"`
	DepartmentID *string                         `json:"department_id"`
	RequiredBy   *string                         `json:"required_by"`
	Purpose      string                          `json:"purpose"`
	Notes        *string                         `json:"notes"`
	Items        []CreatePurchaseRequestItemBody `json:"items"`
}

type CreatePurchaseRequestItemBody struct {
	ItemID             string  `json:"item_id"`
	Description        string  `json:"description"`
	Quantity           string  `json:"quantity"`
	EstimatedUnitPrice *string `json:"estimated_unit_price"`
}

type PurchaseRequestResponse struct {
	ID            string                       `json:"id"`
	RequestNumber string                       `json:"request_number"`
	RequestedBy   string                       `json:"requested_by"`
	DepartmentID  *string                      `json:"department_id,omitempty"`
	Status        domain.PurchaseRequestStatus `json:"status"`
	RequestedAt   time.Time                    `json:"requested_at"`
	RequiredBy    *time.Time                   `json:"required_by,omitempty"`
	Purpose       string                       `json:"purpose"`
	Notes         *string                      `json:"notes,omitempty"`
	CreatedAt     time.Time                    `json:"created_at"`
	UpdatedAt     time.Time                    `json:"updated_at"`
}

type PurchaseRequestItemResponse struct {
	ID                 string  `json:"id"`
	PurchaseRequestID  string  `json:"purchase_request_id"`
	ItemID             string  `json:"item_id"`
	Description        string  `json:"description"`
	Quantity           string  `json:"quantity"`
	EstimatedUnitPrice *string `json:"estimated_unit_price,omitempty"`
}

type CreatePurchaseRequestResponse struct {
	PurchaseRequest PurchaseRequestResponse       `json:"purchase_request"`
	Items           []PurchaseRequestItemResponse `json:"items"`
}

type ListPurchaseRequestsResponse struct {
	PurchaseRequests []PurchaseRequestResponse `json:"purchase_requests"`
	Limit            int                       `json:"limit"`
	Offset           int                       `json:"offset"`
}

type UpdatePurchaseRequestBody struct {
	RequiredBy *string `json:"required_by"`
	Purpose    string  `json:"purpose"`
	Notes      *string `json:"notes"`
}

type UpdatePurchaseRequestItemBody struct {
	ItemID             string  `json:"item_id"`
	Description        string  `json:"description"`
	Quantity           string  `json:"quantity"`
	EstimatedUnitPrice *string `json:"estimated_unit_price"`
}

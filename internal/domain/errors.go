package domain

import "errors"

var (
	ErrPurchaseRequestNotFound      = errors.New("purchase request not found")
	ErrPurchaseRequestNotEditable   = errors.New("purchase request cannot be updated in its current status")
	ErrPurchaseRequestCannotSubmit  = errors.New("purchase request cannot be submitted in its current status")
	ErrPurchaseRequestCannotApprove = errors.New("purchase request cannot be approved in its current status")
	ErrPurchaseRequestCannotReject  = errors.New("purchase request cannot be rejected in its current status")
)

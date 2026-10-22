package domain

import "errors"

var (
	ErrPurchaseRequestNotFound      = errors.New("purchase request not found")
	ErrPurchaseRequestNotEditable   = errors.New("purchase request cannot be updated in its current status")
	ErrPurchaseRequestCannotSubmit  = errors.New("purchase request cannot be submitted in its current status")
	ErrPurchaseRequestCannotApprove = errors.New("purchase request cannot be approved in its current status")
	ErrPurchaseRequestCannotReject  = errors.New("purchase request cannot be rejected in its current status")

	ErrPurchaseOrderNotFound      = errors.New("purchase order not found")
	ErrPurchaseOrderNotEditable   = errors.New("purchase order cannot be updated in its current status")
	ErrPurchaseOrderCannotApprove = errors.New("purchase order cannot be approved in its current status")
	ErrPurchaseOrderCannotIssue   = errors.New("purchase order cannot be issued in its current status")
	ErrPurchaseOrderCannotCancel  = errors.New("purchase order cannot be cancelled in its current status")
	ErrPurchaseOrderCannotClose   = errors.New("purchase order cannot be closed in its current status")

	ErrGoodsReceiptNotFound          = errors.New("goods receipt not found")
	ErrInvalidPOStatusForReceiving   = errors.New("purchase order is not in a valid status for receiving")
	ErrExceedsOrderedQuantity        = errors.New("received quantity exceeds remaining ordered quantity")
	ErrGoodsReceiptCannotCancel      = errors.New("goods receipt cannot be cancelled in its current status")

	ErrVendorBillNotFound       = errors.New("vendor bill not found")
	ErrVendorBillNotEditable    = errors.New("vendor bill cannot be updated in its current status")
	ErrVendorBillCannotApprove  = errors.New("vendor bill must be matched before it can be approved")
	ErrVendorBillCannotCancel   = errors.New("vendor bill cannot be cancelled in its current status")
	ErrDuplicateVendorInvoice   = errors.New("an invoice with this bill number already exists for this vendor")
	ErrInvalidPOStatusForBilling = errors.New("purchase order is not in a valid status for billing")
)

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

const purchaseOrderDocumentType = "purchase_order"

type CreatePurchaseOrderInput struct {
	Order *domain.PurchaseOrder
	Items []*domain.PurchaseOrderItem
}

type ConvertPRToPOInput struct {
	PurchaseRequestID    string
	VendorID             string
	ExpectedDeliveryDate *time.Time
	Currency             string
	PaymentTerms         *string
	ShippingAddress      *string
	Notes                *string
}

type PurchaseOrderService interface {
	CreatePurchaseOrder(
		ctx context.Context,
		input *CreatePurchaseOrderInput,
	) (*domain.PurchaseOrder, error)

	CreatePurchaseOrderFromPR(
		ctx context.Context,
		input *ConvertPRToPOInput,
	) (*domain.PurchaseOrder, error)

	GetPurchaseOrderByID(
		ctx context.Context,
		id string,
	) (*domain.PurchaseOrder, error)

	GetPurchaseOrderByNumber(
		ctx context.Context,
		poNumber string,
	) (*domain.PurchaseOrder, error)

	ListPurchaseOrders(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.PurchaseOrder, error)

	ListPurchaseOrdersByStatus(
		ctx context.Context,
		status domain.PurchaseOrderStatus,
		limit int,
		offset int,
	) ([]*domain.PurchaseOrder, error)

	ListPurchaseOrdersByVendor(
		ctx context.Context,
		vendorID string,
		limit int,
		offset int,
	) ([]*domain.PurchaseOrder, error)

	UpdatePurchaseOrder(
		ctx context.Context,
		order *domain.PurchaseOrder,
	) (*domain.PurchaseOrder, error)

	ApprovePurchaseOrder(
		ctx context.Context,
		id string,
	) (*domain.PurchaseOrder, error)

	IssuePurchaseOrder(
		ctx context.Context,
		id string,
	) (*domain.PurchaseOrder, error)

	CancelPurchaseOrder(
		ctx context.Context,
		id string,
	) (*domain.PurchaseOrder, error)

	GetPurchaseOrderItems(
		ctx context.Context,
		purchaseOrderID string,
	) ([]*domain.PurchaseOrderItem, error)

	AddPurchaseOrderItem(
		ctx context.Context,
		purchaseOrderID string,
		item *domain.PurchaseOrderItem,
	) (*domain.PurchaseOrderItem, error)

	UpdatePurchaseOrderItem(
		ctx context.Context,
		item *domain.PurchaseOrderItem,
	) (*domain.PurchaseOrderItem, error)

	DeletePurchaseOrderItem(
		ctx context.Context,
		id string,
	) error
}

type purchaseOrderService struct {
	repository     repository.PurchaseOrderRepository
	itemRepository repository.PurchaseOrderItemRepository
	transaction    repository.PurchaseOrderTransactionManager
}

func NewPurchaseOrderService(
	repository repository.PurchaseOrderRepository,
	itemRepository repository.PurchaseOrderItemRepository,
	transaction repository.PurchaseOrderTransactionManager,
) PurchaseOrderService {
	return &purchaseOrderService{
		repository:     repository,
		itemRepository: itemRepository,
		transaction:    transaction,
	}
}

func calculateOrderTotals(order *domain.PurchaseOrder, items []*domain.PurchaseOrderItem) {
	subtotal := decimal.Zero
	totalTax := decimal.Zero
	hundred := decimal.NewFromInt(100)

	for _, item := range items {
		lineSubtotal := item.Quantity.Mul(item.UnitPrice).Round(2)
		lineTax := decimal.Zero
		if !item.TaxRate.IsZero() {
			lineTax = lineSubtotal.Mul(item.TaxRate).Div(hundred).Round(2)
		}
		item.TotalPrice = lineSubtotal.Add(lineTax)

		subtotal = subtotal.Add(lineSubtotal)
		totalTax = totalTax.Add(lineTax)
	}

	order.Subtotal = subtotal
	order.TaxAmount = totalTax
	order.TotalAmount = subtotal.Add(totalTax)
}

func (s *purchaseOrderService) CreatePurchaseOrder(
	ctx context.Context,
	input *CreatePurchaseOrderInput,
) (*domain.PurchaseOrder, error) {
	if input == nil {
		return nil, fmt.Errorf("create purchase order input cannot be nil")
	}

	if input.Order == nil {
		return nil, fmt.Errorf("purchase order cannot be nil")
	}

	if input.Order.VendorID == "" {
		return nil, fmt.Errorf("vendor id cannot be empty")
	}

	if len(input.Items) == 0 {
		return nil, fmt.Errorf("purchase order must contain at least one item")
	}

	for i, item := range input.Items {
		if item == nil {
			return nil, fmt.Errorf("purchase order item %d cannot be nil", i+1)
		}
		if item.Quantity.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("purchase order item %d quantity must be greater than zero", i+1)
		}
		if item.UnitPrice.IsNegative() {
			return nil, fmt.Errorf("purchase order item %d unit price cannot be negative", i+1)
		}
		if item.TaxRate.IsNegative() {
			return nil, fmt.Errorf("purchase order item %d tax rate cannot be negative", i+1)
		}
	}

	calculateOrderTotals(input.Order, input.Items)

	var createdOrder *domain.PurchaseOrder

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			poRepo repository.PurchaseOrderRepository,
			itemRepo repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
			prRepo repository.PurchaseRequestRepository,
			prItemRepo repository.PurchaseRequestItemRepository,
		) error {
			year := time.Now().Year()

			lastNumber, err := docNumRepo.NextDocumentNumber(
				ctx,
				purchaseOrderDocumentType,
				year,
			)
			if err != nil {
				return fmt.Errorf("generate purchase order number: %w", err)
			}

			orderCopy := *input.Order
			orderCopy.PONumber = fmt.Sprintf("PO-%d-%06d", year, lastNumber)
			orderCopy.Status = domain.PurchaseOrderStatusDraft

			if orderCopy.Currency == "" {
				orderCopy.Currency = "USD"
			}

			savedOrder, err := poRepo.CreatePurchaseOrder(ctx, &orderCopy)
			if err != nil {
				return fmt.Errorf("create purchase order: %w", err)
			}

			createdOrder = savedOrder

			for i, item := range input.Items {
				itemCopy := *item
				itemCopy.PurchaseOrderID = createdOrder.ID

				if _, err := itemRepo.CreatePurchaseOrderItem(ctx, &itemCopy); err != nil {
					return fmt.Errorf("create purchase order item %d: %w", i+1, err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}

func (s *purchaseOrderService) CreatePurchaseOrderFromPR(
	ctx context.Context,
	input *ConvertPRToPOInput,
) (*domain.PurchaseOrder, error) {
	if input == nil {
		return nil, fmt.Errorf("convert pr input cannot be nil")
	}

	if input.PurchaseRequestID == "" {
		return nil, fmt.Errorf("purchase request id cannot be empty")
	}

	if input.VendorID == "" {
		return nil, fmt.Errorf("vendor id cannot be empty")
	}

	var createdOrder *domain.PurchaseOrder

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			poRepo repository.PurchaseOrderRepository,
			itemRepo repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
			prRepo repository.PurchaseRequestRepository,
			prItemRepo repository.PurchaseRequestItemRepository,
		) error {
			pr, err := prRepo.GetPurchaseRequestByID(ctx, input.PurchaseRequestID)
			if err != nil {
				return fmt.Errorf("get purchase request: %w", err)
			}

			if pr.Status != domain.PurchaseRequestStatusApproved {
				return fmt.Errorf("cannot create purchase order from purchase request in %q status (must be approved)", pr.Status)
			}

			prItems, err := prItemRepo.GetPurchaseRequestItems(ctx, pr.ID)
			if err != nil {
				return fmt.Errorf("get purchase request items: %w", err)
			}

			if len(prItems) == 0 {
				return fmt.Errorf("purchase request contains no items to convert")
			}

			poItems := make([]*domain.PurchaseOrderItem, 0, len(prItems))
			for _, prItem := range prItems {
				unitPrice := decimal.Zero
				if prItem.EstimatedUnitPrice != nil {
					unitPrice = *prItem.EstimatedUnitPrice
				}

				poItems = append(poItems, &domain.PurchaseOrderItem{
					ItemID:      prItem.ItemID,
					Description: prItem.Description,
					Quantity:    prItem.Quantity,
					UnitPrice:   unitPrice,
					TaxRate:     decimal.Zero,
				})
			}

			po := &domain.PurchaseOrder{
				PurchaseRequestID:    &pr.ID,
				VendorID:             input.VendorID,
				ExpectedDeliveryDate: input.ExpectedDeliveryDate,
				Currency:             input.Currency,
				PaymentTerms:         input.PaymentTerms,
				ShippingAddress:      input.ShippingAddress,
				Notes:                input.Notes,
			}
			if po.Currency == "" {
				po.Currency = "USD"
			}

			calculateOrderTotals(po, poItems)

			year := time.Now().Year()
			lastNumber, err := docNumRepo.NextDocumentNumber(
				ctx,
				purchaseOrderDocumentType,
				year,
			)
			if err != nil {
				return fmt.Errorf("generate purchase order number: %w", err)
			}

			po.PONumber = fmt.Sprintf("PO-%d-%06d", year, lastNumber)
			po.Status = domain.PurchaseOrderStatusDraft

			savedOrder, err := poRepo.CreatePurchaseOrder(ctx, po)
			if err != nil {
				return fmt.Errorf("create purchase order: %w", err)
			}

			createdOrder = savedOrder

			for i, item := range poItems {
				item.PurchaseOrderID = createdOrder.ID
				if _, err := itemRepo.CreatePurchaseOrderItem(ctx, item); err != nil {
					return fmt.Errorf("create purchase order item %d: %w", i+1, err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}

func (s *purchaseOrderService) GetPurchaseOrderByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrder, error) {
	return s.repository.GetPurchaseOrderByID(ctx, id)
}

func (s *purchaseOrderService) GetPurchaseOrderByNumber(
	ctx context.Context,
	poNumber string,
) (*domain.PurchaseOrder, error) {
	return s.repository.GetPurchaseOrderByNumber(ctx, poNumber)
}

func (s *purchaseOrderService) ListPurchaseOrders(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	return s.repository.ListPurchaseOrders(ctx, limit, offset)
}

func (s *purchaseOrderService) ListPurchaseOrdersByStatus(
	ctx context.Context,
	status domain.PurchaseOrderStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	return s.repository.ListPurchaseOrdersByStatus(ctx, status, limit, offset)
}

func (s *purchaseOrderService) ListPurchaseOrdersByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	return s.repository.ListPurchaseOrdersByVendor(ctx, vendorID, limit, offset)
}

func (s *purchaseOrderService) UpdatePurchaseOrder(
	ctx context.Context,
	order *domain.PurchaseOrder,
) (*domain.PurchaseOrder, error) {
	if order == nil {
		return nil, fmt.Errorf("purchase order cannot be nil")
	}

	current, err := s.repository.GetPurchaseOrderByID(ctx, order.ID)
	if err != nil {
		return nil, fmt.Errorf("get purchase order: %w", err)
	}

	if current.Status != domain.PurchaseOrderStatusDraft {
		return nil, domain.ErrPurchaseOrderNotEditable
	}

	return s.repository.UpdatePurchaseOrder(ctx, order)
}

func (s *purchaseOrderService) ApprovePurchaseOrder(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrder, error) {
	current, err := s.repository.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get purchase order: %w", err)
	}

	if current.Status != domain.PurchaseOrderStatusDraft &&
		current.Status != domain.PurchaseOrderStatusPendingApproval {
		return nil, domain.ErrPurchaseOrderCannotApprove
	}

	return s.repository.UpdatePurchaseOrderStatus(
		ctx,
		id,
		domain.PurchaseOrderStatusApproved,
	)
}

func (s *purchaseOrderService) IssuePurchaseOrder(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrder, error) {
	current, err := s.repository.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get purchase order: %w", err)
	}

	if current.Status != domain.PurchaseOrderStatusApproved {
		return nil, domain.ErrPurchaseOrderCannotIssue
	}

	return s.repository.UpdatePurchaseOrderStatus(
		ctx,
		id,
		domain.PurchaseOrderStatusIssued,
	)
}

func (s *purchaseOrderService) CancelPurchaseOrder(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrder, error) {
	current, err := s.repository.GetPurchaseOrderByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get purchase order: %w", err)
	}

	switch current.Status {
	case domain.PurchaseOrderStatusPartiallyReceived,
		domain.PurchaseOrderStatusReceived,
		domain.PurchaseOrderStatusBilled,
		domain.PurchaseOrderStatusClosed:
		return nil, domain.ErrPurchaseOrderCannotCancel
	}

	return s.repository.UpdatePurchaseOrderStatus(
		ctx,
		id,
		domain.PurchaseOrderStatusCancelled,
	)
}

func (s *purchaseOrderService) GetPurchaseOrderItems(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.PurchaseOrderItem, error) {
	return s.itemRepository.GetPurchaseOrderItems(ctx, purchaseOrderID)
}

func (s *purchaseOrderService) AddPurchaseOrderItem(
	ctx context.Context,
	purchaseOrderID string,
	item *domain.PurchaseOrderItem,
) (*domain.PurchaseOrderItem, error) {
	if purchaseOrderID == "" {
		return nil, fmt.Errorf("purchase order id cannot be empty")
	}

	if item == nil {
		return nil, fmt.Errorf("purchase order item cannot be nil")
	}

	if item.Quantity.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	if item.UnitPrice.IsNegative() {
		return nil, fmt.Errorf("unit price cannot be negative")
	}

	if item.TaxRate.IsNegative() {
		return nil, fmt.Errorf("tax rate cannot be negative")
	}

	current, err := s.repository.GetPurchaseOrderByID(ctx, purchaseOrderID)
	if err != nil {
		return nil, fmt.Errorf("get purchase order: %w", err)
	}

	if current.Status != domain.PurchaseOrderStatusDraft {
		return nil, domain.ErrPurchaseOrderNotEditable
	}

	lineSubtotal := item.Quantity.Mul(item.UnitPrice).Round(2)
	lineTax := decimal.Zero
	if !item.TaxRate.IsZero() {
		lineTax = lineSubtotal.Mul(item.TaxRate).Div(decimal.NewFromInt(100)).Round(2)
	}
	item.TotalPrice = lineSubtotal.Add(lineTax)
	item.PurchaseOrderID = purchaseOrderID

	created, err := s.itemRepository.CreatePurchaseOrderItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("create purchase order item: %w", err)
	}

	return created, nil
}

func (s *purchaseOrderService) UpdatePurchaseOrderItem(
	ctx context.Context,
	item *domain.PurchaseOrderItem,
) (*domain.PurchaseOrderItem, error) {
	if item == nil {
		return nil, fmt.Errorf("purchase order item cannot be nil")
	}

	if item.Quantity.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("quantity must be greater than zero")
	}

	if item.UnitPrice.IsNegative() {
		return nil, fmt.Errorf("unit price cannot be negative")
	}

	if item.TaxRate.IsNegative() {
		return nil, fmt.Errorf("tax rate cannot be negative")
	}

	lineSubtotal := item.Quantity.Mul(item.UnitPrice).Round(2)
	lineTax := decimal.Zero
	if !item.TaxRate.IsZero() {
		lineTax = lineSubtotal.Mul(item.TaxRate).Div(decimal.NewFromInt(100)).Round(2)
	}
	item.TotalPrice = lineSubtotal.Add(lineTax)

	updated, err := s.itemRepository.UpdatePurchaseOrderItem(ctx, item)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPurchaseOrderNotEditable
		}
		return nil, fmt.Errorf("update purchase order item: %w", err)
	}

	return updated, nil
}

func (s *purchaseOrderService) DeletePurchaseOrderItem(
	ctx context.Context,
	id string,
) error {
	if id == "" {
		return fmt.Errorf("purchase order item id cannot be empty")
	}

	if err := s.itemRepository.DeletePurchaseOrderItem(ctx, id); err != nil {
		return fmt.Errorf("delete purchase order item: %w", err)
	}

	return nil
}

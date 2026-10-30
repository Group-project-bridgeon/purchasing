package service

import (
	"context"
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

const goodsReceiptDocumentType = "goods_receipt"

type CreateGoodsReceiptInput struct {
	PurchaseOrderID string
	ReceivedBy      string
	ReceivedAt      *time.Time
	Notes           *string
	Items           []*domain.GoodsReceiptItem
}

type GoodsReceiptService interface {
	CreateGoodsReceipt(
		ctx context.Context,
		input *CreateGoodsReceiptInput,
	) (*domain.GoodsReceipt, error)

	GetGoodsReceiptByID(
		ctx context.Context,
		id string,
	) (*domain.GoodsReceipt, error)

	GetGoodsReceiptByNumber(
		ctx context.Context,
		grnNumber string,
	) (*domain.GoodsReceipt, error)

	ListGoodsReceipts(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.GoodsReceipt, error)

	ListGoodsReceiptsByPO(
		ctx context.Context,
		purchaseOrderID string,
	) ([]*domain.GoodsReceipt, error)

	ListGoodsReceiptsByVendor(
		ctx context.Context,
		vendorID string,
		limit int,
		offset int,
	) ([]*domain.GoodsReceipt, error)

	GetGoodsReceiptItems(
		ctx context.Context,
		goodsReceiptID string,
	) ([]*domain.GoodsReceiptItem, error)

	CancelGoodsReceipt(
		ctx context.Context,
		id string,
	) (*domain.GoodsReceipt, error)
}

type goodsReceiptService struct {
	repository     repository.GoodsReceiptRepository
	itemRepository repository.GoodsReceiptItemRepository
	transaction    repository.GoodsReceiptTransactionManager
}

func NewGoodsReceiptService(
	repository repository.GoodsReceiptRepository,
	itemRepository repository.GoodsReceiptItemRepository,
	transaction repository.GoodsReceiptTransactionManager,
) GoodsReceiptService {
	return &goodsReceiptService{
		repository:     repository,
		itemRepository: itemRepository,
		transaction:    transaction,
	}
}

func (s *goodsReceiptService) CreateGoodsReceipt(
	ctx context.Context,
	input *CreateGoodsReceiptInput,
) (*domain.GoodsReceipt, error) {
	if input == nil {
		return nil, fmt.Errorf("create goods receipt input cannot be nil")
	}

	if input.PurchaseOrderID == "" {
		return nil, fmt.Errorf("purchase order id cannot be empty")
	}

	if input.ReceivedBy == "" {
		return nil, fmt.Errorf("received by cannot be empty")
	}

	if len(input.Items) == 0 {
		return nil, fmt.Errorf("goods receipt must contain at least one item")
	}

	for i, item := range input.Items {
		if item == nil {
			return nil, fmt.Errorf("goods receipt item %d cannot be nil", i+1)
		}
		if item.PurchaseOrderItemID == "" {
			return nil, fmt.Errorf("goods receipt item %d purchase order item id cannot be empty", i+1)
		}
		if item.ReceivedQuantity.IsNegative() {
			return nil, fmt.Errorf("goods receipt item %d received quantity cannot be negative", i+1)
		}
		if item.RejectedQuantity.IsNegative() {
			return nil, fmt.Errorf("goods receipt item %d rejected quantity cannot be negative", i+1)
		}
		if item.ReceivedQuantity.Add(item.RejectedQuantity).LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("goods receipt item %d total quantity (received + rejected) must be greater than zero", i+1)
		}
	}

	var createdGRN *domain.GoodsReceipt

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			grnRepo repository.GoodsReceiptRepository,
			grnItemRepo repository.GoodsReceiptItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
		) error {
			po, err := poRepo.GetPurchaseOrderByID(ctx, input.PurchaseOrderID)
			if err != nil {
				return fmt.Errorf("get purchase order: %w", err)
			}

			if po.Status != domain.PurchaseOrderStatusIssued &&
				po.Status != domain.PurchaseOrderStatusPartiallyReceived {
				return fmt.Errorf("%w (current status: %q)", domain.ErrInvalidPOStatusForReceiving, po.Status)
			}

			poItems, err := poItemRepo.GetPurchaseOrderItems(ctx, po.ID)
			if err != nil {
				return fmt.Errorf("get purchase order items: %w", err)
			}

			poItemMap := make(map[string]*domain.PurchaseOrderItem, len(poItems))
			for _, pi := range poItems {
				poItemMap[pi.ID] = pi
			}

			for i, grnItem := range input.Items {
				poItem, ok := poItemMap[grnItem.PurchaseOrderItemID]
				if !ok {
					return fmt.Errorf("purchase order item %s does not belong to purchase order %s", grnItem.PurchaseOrderItemID, po.ID)
				}

				grnItem.ItemID = poItem.ItemID

				maxAllowed := poItem.Quantity.Sub(poItem.ReceivedQuantity)
				if grnItem.ReceivedQuantity.GreaterThan(maxAllowed) {
					return fmt.Errorf(
						"%w: item %q ordered %s, already received %s, attempted to receive %s",
						domain.ErrExceedsOrderedQuantity,
						poItem.Description,
						poItem.Quantity,
						poItem.ReceivedQuantity,
						grnItem.ReceivedQuantity,
					)
				}
				_ = i
			}

			year := time.Now().Year()
			lastNumber, err := docNumRepo.NextDocumentNumber(
				ctx,
				goodsReceiptDocumentType,
				year,
			)
			if err != nil {
				return fmt.Errorf("generate goods receipt number: %w", err)
			}

			receivedAt := time.Now()
			if input.ReceivedAt != nil {
				receivedAt = *input.ReceivedAt
			}

			grn := &domain.GoodsReceipt{
				GRNNumber:       fmt.Sprintf("GRN-%d-%06d", year, lastNumber),
				PurchaseOrderID: po.ID,
				VendorID:        po.VendorID,
				ReceivedBy:      input.ReceivedBy,
				ReceivedAt:      receivedAt,
				Status:          domain.GoodsReceiptStatusCompleted,
				Notes:           input.Notes,
			}

			savedGRN, err := grnRepo.CreateGoodsReceipt(ctx, grn)
			if err != nil {
				return fmt.Errorf("create goods receipt: %w", err)
			}

			createdGRN = savedGRN

			for _, grnItem := range input.Items {
				grnItem.GoodsReceiptID = createdGRN.ID

				if _, err := grnItemRepo.CreateGoodsReceiptItem(ctx, grnItem); err != nil {
					return fmt.Errorf("create goods receipt item: %w", err)
				}

				poItem := poItemMap[grnItem.PurchaseOrderItemID]
				newReceivedQty := poItem.ReceivedQuantity.Add(grnItem.ReceivedQuantity)

				if _, err := poItemRepo.UpdatePurchaseOrderItemFulfillment(
					ctx,
					poItem.ID,
					newReceivedQty,
					poItem.BilledQuantity,
				); err != nil {
					return fmt.Errorf("update purchase order item fulfillment: %w", err)
				}

				poItem.ReceivedQuantity = newReceivedQty
			}

			allFullyReceived := true
			hasAnyReceived := false

			for _, pi := range poItems {
				if pi.ReceivedQuantity.GreaterThan(decimal.Zero) {
					hasAnyReceived = true
				}
				if pi.ReceivedQuantity.LessThan(pi.Quantity) {
					allFullyReceived = false
				}
			}

			var newStatus domain.PurchaseOrderStatus
			if allFullyReceived {
				newStatus = domain.PurchaseOrderStatusReceived
			} else if hasAnyReceived {
				newStatus = domain.PurchaseOrderStatusPartiallyReceived
			} else {
				newStatus = domain.PurchaseOrderStatusIssued
			}

			if newStatus != po.Status {
				if _, err := poRepo.UpdatePurchaseOrderStatus(ctx, po.ID, newStatus); err != nil {
					return fmt.Errorf("update purchase order status: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return createdGRN, nil
}

func (s *goodsReceiptService) GetGoodsReceiptByID(
	ctx context.Context,
	id string,
) (*domain.GoodsReceipt, error) {
	return s.repository.GetGoodsReceiptByID(ctx, id)
}

func (s *goodsReceiptService) GetGoodsReceiptByNumber(
	ctx context.Context,
	grnNumber string,
) (*domain.GoodsReceipt, error) {
	return s.repository.GetGoodsReceiptByNumber(ctx, grnNumber)
}

func (s *goodsReceiptService) ListGoodsReceipts(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.GoodsReceipt, error) {
	return s.repository.ListGoodsReceipts(ctx, limit, offset)
}

func (s *goodsReceiptService) ListGoodsReceiptsByPO(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.GoodsReceipt, error) {
	return s.repository.ListGoodsReceiptsByPO(ctx, purchaseOrderID)
}

func (s *goodsReceiptService) ListGoodsReceiptsByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.GoodsReceipt, error) {
	return s.repository.ListGoodsReceiptsByVendor(ctx, vendorID, limit, offset)
}

func (s *goodsReceiptService) GetGoodsReceiptItems(
	ctx context.Context,
	goodsReceiptID string,
) ([]*domain.GoodsReceiptItem, error) {
	return s.itemRepository.GetGoodsReceiptItems(ctx, goodsReceiptID)
}

func (s *goodsReceiptService) CancelGoodsReceipt(
	ctx context.Context,
	id string,
) (*domain.GoodsReceipt, error) {
	var cancelledGRN *domain.GoodsReceipt

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			grnRepo repository.GoodsReceiptRepository,
			grnItemRepo repository.GoodsReceiptItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
		) error {
			grn, err := grnRepo.GetGoodsReceiptByID(ctx, id)
			if err != nil {
				return fmt.Errorf("get goods receipt: %w", err)
			}

			if grn.Status != domain.GoodsReceiptStatusCompleted {
				return domain.ErrGoodsReceiptCannotCancel
			}

			po, err := poRepo.GetPurchaseOrderByID(ctx, grn.PurchaseOrderID)
			if err != nil {
				return fmt.Errorf("get purchase order: %w", err)
			}

			if po.Status == domain.PurchaseOrderStatusBilled || po.Status == domain.PurchaseOrderStatusClosed {
				return fmt.Errorf("cannot cancel goods receipt for a purchase order in %q status", po.Status)
			}

			grnItems, err := grnItemRepo.GetGoodsReceiptItems(ctx, grn.ID)
			if err != nil {
				return fmt.Errorf("get goods receipt items: %w", err)
			}

			poItems, err := poItemRepo.GetPurchaseOrderItems(ctx, po.ID)
			if err != nil {
				return fmt.Errorf("get purchase order items: %w", err)
			}

			poItemMap := make(map[string]*domain.PurchaseOrderItem, len(poItems))
			for _, pi := range poItems {
				poItemMap[pi.ID] = pi
			}

			for _, gi := range grnItems {
				pi, ok := poItemMap[gi.PurchaseOrderItemID]
				if ok {
					newReceivedQty := pi.ReceivedQuantity.Sub(gi.ReceivedQuantity)
					if newReceivedQty.IsNegative() {
						newReceivedQty = decimal.Zero
					}

					if _, err := poItemRepo.UpdatePurchaseOrderItemFulfillment(
						ctx,
						pi.ID,
						newReceivedQty,
						pi.BilledQuantity,
					); err != nil {
						return fmt.Errorf("revert purchase order item fulfillment: %w", err)
					}

					pi.ReceivedQuantity = newReceivedQty
				}
			}

			updatedGRN, err := grnRepo.UpdateGoodsReceiptStatus(ctx, grn.ID, domain.GoodsReceiptStatusCancelled)
			if err != nil {
				return fmt.Errorf("cancel goods receipt: %w", err)
			}

			cancelledGRN = updatedGRN

			hasAnyReceived := false
			for _, pi := range poItems {
				if pi.ReceivedQuantity.GreaterThan(decimal.Zero) {
					hasAnyReceived = true
					break
				}
			}

			var newStatus domain.PurchaseOrderStatus
			if hasAnyReceived {
				newStatus = domain.PurchaseOrderStatusPartiallyReceived
			} else {
				newStatus = domain.PurchaseOrderStatusIssued
			}

			if newStatus != po.Status {
				if _, err := poRepo.UpdatePurchaseOrderStatus(ctx, po.ID, newStatus); err != nil {
					return fmt.Errorf("update purchase order status: %w", err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return cancelledGRN, nil
}

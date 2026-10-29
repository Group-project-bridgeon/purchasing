package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

const purchaseReturnDocumentType = "purchase_return"

type CreatePurchaseReturnItemInput struct {
	PurchaseOrderItemID string
	ReturnQuantity      decimal.Decimal
	Reason              *string
}

type CreatePurchaseReturnInput struct {
	PurchaseOrderID string
	GoodsReceiptID  *string
	ReturnDate      *time.Time
	Reason          string
	Carrier         *string
	TrackingNumber  *string
	CreatedBy       string
	Notes           *string
	Items           []CreatePurchaseReturnItemInput
}

type UpdatePurchaseReturnInput struct {
	Reason         *string
	Carrier        *string
	TrackingNumber *string
	Notes          *string
}

type PurchaseReturnDetail struct {
	Return *domain.PurchaseReturn
	Items  []*domain.PurchaseReturnItem
}

type PurchaseReturnService interface {
	CreatePurchaseReturn(
		ctx context.Context,
		input *CreatePurchaseReturnInput,
	) (*PurchaseReturnDetail, error)

	GetPurchaseReturnByID(
		ctx context.Context,
		id string,
	) (*PurchaseReturnDetail, error)

	GetPurchaseReturnByNumber(
		ctx context.Context,
		returnNumber string,
	) (*PurchaseReturnDetail, error)

	ListPurchaseReturns(
		ctx context.Context,
		poID *string,
		vendorID *string,
		status *domain.PurchaseReturnStatus,
		limit int,
		offset int,
	) ([]*domain.PurchaseReturn, int64, error)

	UpdatePurchaseReturn(
		ctx context.Context,
		id string,
		input *UpdatePurchaseReturnInput,
	) (*domain.PurchaseReturn, error)

	ApprovePurchaseReturn(
		ctx context.Context,
		id string,
	) (*domain.PurchaseReturn, error)

	ShipPurchaseReturn(
		ctx context.Context,
		id string,
		carrier string,
		trackingNumber string,
	) (*domain.PurchaseReturn, error)

	CompletePurchaseReturn(
		ctx context.Context,
		id string,
	) (*PurchaseReturnDetail, error)

	CancelPurchaseReturn(
		ctx context.Context,
		id string,
	) (*domain.PurchaseReturn, error)
}

type purchaseReturnService struct {
	returnRepo     repository.PurchaseReturnRepository
	returnItemRepo repository.PurchaseReturnItemRepository
	poRepo         repository.PurchaseOrderRepository
	poItemRepo     repository.PurchaseOrderItemRepository
	docNumRepo     repository.DocumentNumberRepository
	txManager      repository.PurchaseReturnTransactionManager
}

var _ PurchaseReturnService = (*purchaseReturnService)(nil)

func NewPurchaseReturnService(
	returnRepo repository.PurchaseReturnRepository,
	returnItemRepo repository.PurchaseReturnItemRepository,
	poRepo repository.PurchaseOrderRepository,
	poItemRepo repository.PurchaseOrderItemRepository,
	docNumRepo repository.DocumentNumberRepository,
	txManager repository.PurchaseReturnTransactionManager,
) PurchaseReturnService {
	return &purchaseReturnService{
		returnRepo:     returnRepo,
		returnItemRepo: returnItemRepo,
		poRepo:         poRepo,
		poItemRepo:     poItemRepo,
		docNumRepo:     docNumRepo,
		txManager:      txManager,
	}
}

func (s *purchaseReturnService) CreatePurchaseReturn(
	ctx context.Context,
	input *CreatePurchaseReturnInput,
) (*PurchaseReturnDetail, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.PurchaseOrderID) == "" {
		return nil, fmt.Errorf("purchase order id is required")
	}

	if strings.TrimSpace(input.CreatedBy) == "" {
		return nil, fmt.Errorf("created by user id is required")
	}

	if strings.TrimSpace(input.Reason) == "" {
		return nil, fmt.Errorf("return reason is required")
	}

	if len(input.Items) == 0 {
		return nil, fmt.Errorf("at least one return item is required")
	}

	po, err := s.poRepo.GetPurchaseOrderByID(ctx, input.PurchaseOrderID)
	if err != nil {
		return nil, fmt.Errorf("get purchase order: %w", err)
	}

	switch po.Status {
	case domain.PurchaseOrderStatusIssued,
		domain.PurchaseOrderStatusPartiallyReceived,
		domain.PurchaseOrderStatusReceived,
		domain.PurchaseOrderStatusBilled:
	default:
		return nil, fmt.Errorf("%w: current status %q", domain.ErrInvalidPOStatusForReturn, po.Status)
	}

	poItems, err := s.poItemRepo.GetPurchaseOrderItems(ctx, po.ID)
	if err != nil {
		return nil, fmt.Errorf("get purchase order items: %w", err)
	}

	poItemMap := make(map[string]*domain.PurchaseOrderItem, len(poItems))
	for _, item := range poItems {
		poItemMap[item.ID] = item
	}

	returnDate := time.Now()
	if input.ReturnDate != nil && !input.ReturnDate.IsZero() {
		returnDate = *input.ReturnDate
	}

	totalReturnAmount := decimal.Zero
	preparedItems := make([]*domain.PurchaseReturnItem, 0, len(input.Items))

	for i, itemInput := range input.Items {
		if strings.TrimSpace(itemInput.PurchaseOrderItemID) == "" {
			return nil, fmt.Errorf("items[%d]: purchase order item id is required", i)
		}

		if itemInput.ReturnQuantity.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("items[%d]: return quantity must be greater than zero", i)
		}

		poItem, exists := poItemMap[itemInput.PurchaseOrderItemID]
		if !exists {
			return nil, fmt.Errorf("items[%d]: purchase order item %q does not belong to purchase order %q", i, itemInput.PurchaseOrderItemID, po.ID)
		}

		alreadyReturned, err := s.returnItemRepo.GetTotalReturnedQuantityForPOItem(ctx, poItem.ID)
		if err != nil {
			return nil, fmt.Errorf("items[%d]: get existing returned quantity: %w", i, err)
		}

		availableToReturn := poItem.ReceivedQuantity.Sub(alreadyReturned)
		if itemInput.ReturnQuantity.GreaterThan(availableToReturn) {
			return nil, fmt.Errorf(
				"items[%d]: %w: requested %s, but available returnable is %s (received: %s, previously returned: %s)",
				i,
				domain.ErrExceedsReturnableQuantity,
				itemInput.ReturnQuantity.String(),
				availableToReturn.String(),
				poItem.ReceivedQuantity.String(),
				alreadyReturned.String(),
			)
		}

		lineTotal := itemInput.ReturnQuantity.Mul(poItem.UnitPrice)
		totalReturnAmount = totalReturnAmount.Add(lineTotal)

		preparedItems = append(preparedItems, &domain.PurchaseReturnItem{
			PurchaseOrderItemID: poItem.ID,
			ItemID:              poItem.ItemID,
			Description:         poItem.Description,
			ReturnQuantity:      itemInput.ReturnQuantity,
			UnitPrice:           poItem.UnitPrice,
			TotalPrice:          lineTotal,
			Reason:              itemInput.Reason,
		})
	}

	var createdReturn *domain.PurchaseReturn
	var createdItems []*domain.PurchaseReturnItem

	err = s.txManager.WithinTransaction(
		ctx,
		func(
			retRepo repository.PurchaseReturnRepository,
			retItemRepo repository.PurchaseReturnItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
		) error {
			year := returnDate.Year()
			nextSeq, err := docNumRepo.NextDocumentNumber(ctx, purchaseReturnDocumentType, year)
			if err != nil {
				return fmt.Errorf("generate return number: %w", err)
			}

			retNumber := fmt.Sprintf("RET-%d-%06d", year, nextSeq)

			ret := &domain.PurchaseReturn{
				ReturnNumber:      retNumber,
				PurchaseOrderID:   po.ID,
				GoodsReceiptID:    input.GoodsReceiptID,
				VendorID:          po.VendorID,
				Status:            domain.PurchaseReturnStatusDraft,
				ReturnDate:        returnDate,
				TotalReturnAmount: totalReturnAmount,
				Reason:            input.Reason,
				Carrier:           input.Carrier,
				TrackingNumber:    input.TrackingNumber,
				CreatedBy:         input.CreatedBy,
				Notes:             input.Notes,
			}

			savedReturn, err := retRepo.CreatePurchaseReturn(ctx, ret)
			if err != nil {
				return fmt.Errorf("create purchase return: %w", err)
			}

			savedItems := make([]*domain.PurchaseReturnItem, 0, len(preparedItems))
			for _, item := range preparedItems {
				item.PurchaseReturnID = savedReturn.ID
				savedItem, err := retItemRepo.CreatePurchaseReturnItem(ctx, item)
				if err != nil {
					return fmt.Errorf("create purchase return item: %w", err)
				}
				savedItems = append(savedItems, savedItem)
			}

			createdReturn = savedReturn
			createdItems = savedItems
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &PurchaseReturnDetail{
		Return: createdReturn,
		Items:  createdItems,
	}, nil
}

func (s *purchaseReturnService) GetPurchaseReturnByID(
	ctx context.Context,
	id string,
) (*PurchaseReturnDetail, error) {
	ret, err := s.returnRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.returnItemRepo.GetPurchaseReturnItems(ctx, ret.ID)
	if err != nil {
		return nil, fmt.Errorf("get return items: %w", err)
	}

	return &PurchaseReturnDetail{
		Return: ret,
		Items:  items,
	}, nil
}

func (s *purchaseReturnService) GetPurchaseReturnByNumber(
	ctx context.Context,
	returnNumber string,
) (*PurchaseReturnDetail, error) {
	ret, err := s.returnRepo.GetPurchaseReturnByNumber(ctx, returnNumber)
	if err != nil {
		return nil, err
	}

	items, err := s.returnItemRepo.GetPurchaseReturnItems(ctx, ret.ID)
	if err != nil {
		return nil, fmt.Errorf("get return items: %w", err)
	}

	return &PurchaseReturnDetail{
		Return: ret,
		Items:  items,
	}, nil
}

func (s *purchaseReturnService) ListPurchaseReturns(
	ctx context.Context,
	poID *string,
	vendorID *string,
	status *domain.PurchaseReturnStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseReturn, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	returns, err := s.returnRepo.ListPurchaseReturns(ctx, poID, vendorID, status, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list purchase returns: %w", err)
	}

	total, err := s.returnRepo.CountPurchaseReturns(ctx, poID, vendorID, status)
	if err != nil {
		return nil, 0, fmt.Errorf("count purchase returns: %w", err)
	}

	return returns, total, nil
}

func (s *purchaseReturnService) UpdatePurchaseReturn(
	ctx context.Context,
	id string,
	input *UpdatePurchaseReturnInput,
) (*domain.PurchaseReturn, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	ret, err := s.returnRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ret.Status != domain.PurchaseReturnStatusDraft {
		return nil, domain.ErrPurchaseReturnNotEditable
	}

	if input.Reason != nil && strings.TrimSpace(*input.Reason) != "" {
		ret.Reason = strings.TrimSpace(*input.Reason)
	}

	if input.Carrier != nil {
		ret.Carrier = input.Carrier
	}

	if input.TrackingNumber != nil {
		ret.TrackingNumber = input.TrackingNumber
	}

	if input.Notes != nil {
		ret.Notes = input.Notes
	}

	return s.returnRepo.UpdatePurchaseReturn(ctx, ret)
}

func (s *purchaseReturnService) ApprovePurchaseReturn(
	ctx context.Context,
	id string,
) (*domain.PurchaseReturn, error) {
	ret, err := s.returnRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ret.Status != domain.PurchaseReturnStatusDraft {
		return nil, domain.ErrPurchaseReturnCannotApprove
	}

	return s.returnRepo.UpdatePurchaseReturnStatus(ctx, id, domain.PurchaseReturnStatusApproved)
}

func (s *purchaseReturnService) ShipPurchaseReturn(
	ctx context.Context,
	id string,
	carrier string,
	trackingNumber string,
) (*domain.PurchaseReturn, error) {
	carrier = strings.TrimSpace(carrier)
	trackingNumber = strings.TrimSpace(trackingNumber)

	if carrier == "" {
		return nil, fmt.Errorf("carrier is required for shipping")
	}

	if trackingNumber == "" {
		return nil, fmt.Errorf("tracking number is required for shipping")
	}

	ret, err := s.returnRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ret.Status != domain.PurchaseReturnStatusApproved {
		return nil, domain.ErrPurchaseReturnCannotShip
	}

	return s.returnRepo.ShipPurchaseReturn(ctx, id, carrier, trackingNumber)
}

func (s *purchaseReturnService) CompletePurchaseReturn(
	ctx context.Context,
	id string,
) (*PurchaseReturnDetail, error) {
	ret, err := s.returnRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if ret.Status != domain.PurchaseReturnStatusShipped && ret.Status != domain.PurchaseReturnStatusApproved {
		return nil, domain.ErrPurchaseReturnCannotComplete
	}

	returnItems, err := s.returnItemRepo.GetPurchaseReturnItems(ctx, ret.ID)
	if err != nil {
		return nil, fmt.Errorf("get return items: %w", err)
	}

	var completedReturn *domain.PurchaseReturn

	err = s.txManager.WithinTransaction(
		ctx,
		func(
			retRepo repository.PurchaseReturnRepository,
			retItemRepo repository.PurchaseReturnItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
		) error {
			po, err := poRepo.GetPurchaseOrderByID(ctx, ret.PurchaseOrderID)
			if err != nil {
				return fmt.Errorf("get purchase order: %w", err)
			}

			poItems, err := poItemRepo.GetPurchaseOrderItems(ctx, po.ID)
			if err != nil {
				return fmt.Errorf("get po items: %w", err)
			}

			poItemMap := make(map[string]*domain.PurchaseOrderItem, len(poItems))
			for _, item := range poItems {
				poItemMap[item.ID] = item
			}

			for _, retItem := range returnItems {
				poItem, ok := poItemMap[retItem.PurchaseOrderItemID]
				if !ok {
					continue
				}

				newReceivedQty := poItem.ReceivedQuantity.Sub(retItem.ReturnQuantity)
				if newReceivedQty.LessThan(decimal.Zero) {
					newReceivedQty = decimal.Zero
				}

				if _, err := poItemRepo.UpdatePurchaseOrderItemFulfillment(
					ctx,
					poItem.ID,
					newReceivedQty,
					poItem.BilledQuantity,
				); err != nil {
					return fmt.Errorf("update po item quantities: %w", err)
				}
				poItem.ReceivedQuantity = newReceivedQty
			}

			totalOrdered := decimal.Zero
			totalReceived := decimal.Zero
			for _, item := range poItems {
				totalOrdered = totalOrdered.Add(item.Quantity)
				totalReceived = totalReceived.Add(item.ReceivedQuantity)
			}

			var newPOStatus domain.PurchaseOrderStatus
			if totalReceived.GreaterThanOrEqual(totalOrdered) {
				newPOStatus = domain.PurchaseOrderStatusReceived
			} else if totalReceived.GreaterThan(decimal.Zero) {
				newPOStatus = domain.PurchaseOrderStatusPartiallyReceived
			} else {
				newPOStatus = domain.PurchaseOrderStatusIssued
			}

			if po.Status != newPOStatus && po.Status != domain.PurchaseOrderStatusBilled {
				if _, err := poRepo.UpdatePurchaseOrderStatus(ctx, po.ID, newPOStatus); err != nil {
					return fmt.Errorf("update po status: %w", err)
				}
			}

			updated, err := retRepo.UpdatePurchaseReturnStatus(ctx, ret.ID, domain.PurchaseReturnStatusCompleted)
			if err != nil {
				return fmt.Errorf("update return status: %w", err)
			}

			completedReturn = updated
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &PurchaseReturnDetail{
		Return: completedReturn,
		Items:  returnItems,
	}, nil
}

func (s *purchaseReturnService) CancelPurchaseReturn(
	ctx context.Context,
	id string,
) (*domain.PurchaseReturn, error) {
	ret, err := s.returnRepo.GetPurchaseReturnByID(ctx, id)
	if err != nil {
		return nil, err
	}

	switch ret.Status {
	case domain.PurchaseReturnStatusDraft, domain.PurchaseReturnStatusApproved:
		return s.returnRepo.UpdatePurchaseReturnStatus(ctx, id, domain.PurchaseReturnStatusCancelled)
	default:
		return nil, domain.ErrPurchaseReturnCannotCancel
	}
}

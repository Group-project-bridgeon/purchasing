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

const vendorBillDocumentType = "vendor_bill"

type CreateVendorBillInput struct {
	BillNumber      string
	PurchaseOrderID string
	BillDate        time.Time
	DueDate         *time.Time
	Currency        string
	Notes           *string
	Items           []*domain.VendorBillItem
}

type MatchResult struct {
	Bill          *domain.VendorBill
	Items         []*domain.VendorBillItem
	IsMatched     bool
	Discrepancies []string
}

type VendorBillService interface {
	CreateVendorBill(
		ctx context.Context,
		input *CreateVendorBillInput,
	) (*domain.VendorBill, error)

	GetVendorBillByID(
		ctx context.Context,
		id string,
	) (*domain.VendorBill, error)

	GetVendorBillByInternalNumber(
		ctx context.Context,
		internalBillNumber string,
	) (*domain.VendorBill, error)

	ListVendorBills(
		ctx context.Context,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByPO(
		ctx context.Context,
		purchaseOrderID string,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByVendor(
		ctx context.Context,
		vendorID string,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByStatus(
		ctx context.Context,
		status domain.BillStatus,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	ListVendorBillsByMatchStatus(
		ctx context.Context,
		matchStatus domain.BillMatchStatus,
		limit int,
		offset int,
	) ([]*domain.VendorBill, error)

	GetVendorBillItems(
		ctx context.Context,
		vendorBillID string,
	) ([]*domain.VendorBillItem, error)

	MatchVendorBill(
		ctx context.Context,
		id string,
	) (*MatchResult, error)

	ApproveVendorBill(
		ctx context.Context,
		id string,
	) (*domain.VendorBill, error)

	CancelVendorBill(
		ctx context.Context,
		id string,
	) (*domain.VendorBill, error)
}

type vendorBillService struct {
	repository     repository.VendorBillRepository
	itemRepository repository.VendorBillItemRepository
	transaction    repository.VendorBillTransactionManager
}

func NewVendorBillService(
	repository repository.VendorBillRepository,
	itemRepository repository.VendorBillItemRepository,
	transaction repository.VendorBillTransactionManager,
) VendorBillService {
	return &vendorBillService{
		repository:     repository,
		itemRepository: itemRepository,
		transaction:    transaction,
	}
}

func calculateBillTotals(bill *domain.VendorBill, items []*domain.VendorBillItem) {
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

	bill.Subtotal = subtotal
	bill.TaxAmount = totalTax
	bill.TotalAmount = subtotal.Add(totalTax)
}

func (s *vendorBillService) CreateVendorBill(
	ctx context.Context,
	input *CreateVendorBillInput,
) (*domain.VendorBill, error) {
	if input == nil {
		return nil, fmt.Errorf("create vendor bill input cannot be nil")
	}

	if strings.TrimSpace(input.BillNumber) == "" {
		return nil, fmt.Errorf("bill number cannot be empty")
	}

	if input.PurchaseOrderID == "" {
		return nil, fmt.Errorf("purchase order id cannot be empty")
	}

	if len(input.Items) == 0 {
		return nil, fmt.Errorf("vendor bill must contain at least one item")
	}

	for i, item := range input.Items {
		if item == nil {
			return nil, fmt.Errorf("vendor bill item %d cannot be nil", i+1)
		}
		if item.PurchaseOrderItemID == "" {
			return nil, fmt.Errorf("vendor bill item %d purchase order item id cannot be empty", i+1)
		}
		if item.Quantity.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("vendor bill item %d quantity must be greater than zero", i+1)
		}
		if item.UnitPrice.IsNegative() {
			return nil, fmt.Errorf("vendor bill item %d unit price cannot be negative", i+1)
		}
		if item.TaxRate.IsNegative() {
			return nil, fmt.Errorf("vendor bill item %d tax rate cannot be negative", i+1)
		}
	}

	bill := &domain.VendorBill{
		BillNumber:      input.BillNumber,
		PurchaseOrderID: input.PurchaseOrderID,
		BillDate:        input.BillDate,
		DueDate:         input.DueDate,
		Currency:        input.Currency,
		Notes:           input.Notes,
		Status:          domain.BillStatusDraft,
		MatchStatus:     domain.BillMatchStatusUnmatched,
	}
	if bill.Currency == "" {
		bill.Currency = "USD"
	}

	calculateBillTotals(bill, input.Items)

	var createdBill *domain.VendorBill

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			billRepo repository.VendorBillRepository,
			billItemRepo repository.VendorBillItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			grnItemRepo repository.GoodsReceiptItemRepository,
			docNumRepo repository.DocumentNumberRepository,
		) error {
			po, err := poRepo.GetPurchaseOrderByID(ctx, bill.PurchaseOrderID)
			if err != nil {
				return fmt.Errorf("get purchase order: %w", err)
			}

			switch po.Status {
			case domain.PurchaseOrderStatusIssued,
				domain.PurchaseOrderStatusPartiallyReceived,
				domain.PurchaseOrderStatusReceived:
			default:
				return fmt.Errorf("%w: current status %q", domain.ErrInvalidPOStatusForBilling, po.Status)
			}

			bill.VendorID = po.VendorID

			existing, err := billRepo.GetVendorBillByVendorInvoice(ctx, po.VendorID, bill.BillNumber)
			if err != nil {
				return fmt.Errorf("check existing invoice: %w", err)
			}
			if existing != nil {
				return domain.ErrDuplicateVendorInvoice
			}

			year := time.Now().Year()
			lastNumber, err := docNumRepo.NextDocumentNumber(
				ctx,
				vendorBillDocumentType,
				year,
			)
			if err != nil {
				return fmt.Errorf("generate vendor bill number: %w", err)
			}

			bill.InternalBillNumber = fmt.Sprintf("BILL-%d-%06d", year, lastNumber)

			savedBill, err := billRepo.CreateVendorBill(ctx, bill)
			if err != nil {
				return fmt.Errorf("create vendor bill: %w", err)
			}

			createdBill = savedBill

			poItems, err := poItemRepo.GetPurchaseOrderItems(ctx, po.ID)
			if err != nil {
				return fmt.Errorf("get purchase order items: %w", err)
			}

			poItemMap := make(map[string]*domain.PurchaseOrderItem, len(poItems))
			for _, pi := range poItems {
				poItemMap[pi.ID] = pi
			}

			for i, item := range input.Items {
				poItem, ok := poItemMap[item.PurchaseOrderItemID]
				if !ok {
					return fmt.Errorf("item %d purchase order item %s does not belong to purchase order %s", i+1, item.PurchaseOrderItemID, po.ID)
				}

				item.VendorBillID = createdBill.ID
				item.ItemID = poItem.ItemID
				item.MatchStatus = domain.BillItemMatchStatusUnmatched
				item.MatchedQuantity = decimal.Zero

				if item.Description == "" {
					item.Description = poItem.Description
				}

				if _, err := billItemRepo.CreateVendorBillItem(ctx, item); err != nil {
					return fmt.Errorf("create vendor bill item %d: %w", i+1, err)
				}
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	// Run initial 3-way match
	_, _ = s.MatchVendorBill(ctx, createdBill.ID)

	return s.repository.GetVendorBillByID(ctx, createdBill.ID)
}

func (s *vendorBillService) MatchVendorBill(
	ctx context.Context,
	id string,
) (*MatchResult, error) {
	var result MatchResult

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			billRepo repository.VendorBillRepository,
			billItemRepo repository.VendorBillItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			grnItemRepo repository.GoodsReceiptItemRepository,
			docNumRepo repository.DocumentNumberRepository,
		) error {
			bill, err := billRepo.GetVendorBillByID(ctx, id)
			if err != nil {
				return fmt.Errorf("get vendor bill: %w", err)
			}

			if bill.Status == domain.BillStatusCancelled {
				return fmt.Errorf("cannot match a cancelled vendor bill")
			}

			po, err := poRepo.GetPurchaseOrderByID(ctx, bill.PurchaseOrderID)
			if err != nil {
				return fmt.Errorf("get purchase order: %w", err)
			}

			billItems, err := billItemRepo.GetVendorBillItems(ctx, bill.ID)
			if err != nil {
				return fmt.Errorf("get vendor bill items: %w", err)
			}

			poItems, err := poItemRepo.GetPurchaseOrderItems(ctx, po.ID)
			if err != nil {
				return fmt.Errorf("get purchase order items: %w", err)
			}

			poItemMap := make(map[string]*domain.PurchaseOrderItem, len(poItems))
			for _, pi := range poItems {
				poItemMap[pi.ID] = pi
			}

			allMatched := true
			var discrepancies []string

			for _, bi := range billItems {
				poItem, ok := poItemMap[bi.PurchaseOrderItemID]
				if !ok {
					allMatched = false
					discrepancy := fmt.Sprintf("PO Item %s not found in PO %s", bi.PurchaseOrderItemID, po.ID)
					discrepancies = append(discrepancies, discrepancy)
					_, _ = billItemRepo.UpdateVendorBillItemMatch(ctx, bi.ID, decimal.Zero, domain.BillItemMatchStatusUnmatched)
					continue
				}

				// Price check: billed unit price <= agreed PO unit price
				if bi.UnitPrice.GreaterThan(poItem.UnitPrice) {
					allMatched = false
					discrepancy := fmt.Sprintf("Price variance on %q: billed at %s %s, exceeding agreed PO price of %s %s",
						bi.Description, bi.UnitPrice, bill.Currency, poItem.UnitPrice, po.Currency)
					discrepancies = append(discrepancies, discrepancy)
					_, _ = billItemRepo.UpdateVendorBillItemMatch(ctx, bi.ID, decimal.Zero, domain.BillItemMatchStatusPriceVariance)
					continue
				}

				// Quantity check: billed quantity <= physical received quantity from GRNs
				// Available for this bill = received_qty - (billed_qty - currently_matched_qty)
				availableReceived := poItem.ReceivedQuantity.Sub(poItem.BilledQuantity.Sub(bi.MatchedQuantity))
				if bi.Quantity.GreaterThan(availableReceived) {
					allMatched = false
					discrepancy := fmt.Sprintf("Quantity variance on %q: billed for %s units, but only %s units received (ordered: %s)",
						bi.Description, bi.Quantity, poItem.ReceivedQuantity, poItem.Quantity)
					discrepancies = append(discrepancies, discrepancy)
					_, _ = billItemRepo.UpdateVendorBillItemMatch(ctx, bi.ID, decimal.Zero, domain.BillItemMatchStatusQuantityVariance)
					continue
				}

				// Item passes 3-Way Match
				_, err = billItemRepo.UpdateVendorBillItemMatch(
					ctx,
					bi.ID,
					bi.Quantity,
					domain.BillItemMatchStatusMatched,
				)
				if err != nil {
					return fmt.Errorf("update bill item match: %w", err)
				}
				bi.MatchedQuantity = bi.Quantity
				bi.MatchStatus = domain.BillItemMatchStatusMatched
			}

			var matchStatus domain.BillMatchStatus
			var discrepancyReason *string

			if allMatched {
				matchStatus = domain.BillMatchStatusMatched
				discrepancyReason = nil

				// Update billed_quantity on PO items
				for _, bi := range billItems {
					poItem := poItemMap[bi.PurchaseOrderItemID]
					newBilledQty := poItem.BilledQuantity.Add(bi.Quantity)

					if _, err := poItemRepo.UpdatePurchaseOrderItemFulfillment(
						ctx,
						poItem.ID,
						poItem.ReceivedQuantity,
						newBilledQty,
					); err != nil {
						return fmt.Errorf("update po item billed quantity: %w", err)
					}
					poItem.BilledQuantity = newBilledQty
				}

				// Check if PO is now fully billed and fully received
				allFullyBilled := true
				for _, pi := range poItems {
					if pi.BilledQuantity.LessThan(pi.Quantity) {
						allFullyBilled = false
						break
					}
				}

				if allFullyBilled && po.Status == domain.PurchaseOrderStatusReceived {
					if _, err := poRepo.UpdatePurchaseOrderStatus(ctx, po.ID, domain.PurchaseOrderStatusBilled); err != nil {
						return fmt.Errorf("update po status to billed: %w", err)
					}
				}
			} else {
				matchStatus = domain.BillMatchStatusDiscrepant
				reason := strings.Join(discrepancies, "; ")
				discrepancyReason = &reason
			}

			updatedBill, err := billRepo.UpdateVendorBillMatchStatus(ctx, bill.ID, matchStatus, discrepancyReason)
			if err != nil {
				return fmt.Errorf("update vendor bill match status: %w", err)
			}

			result = MatchResult{
				Bill:          updatedBill,
				Items:         billItems,
				IsMatched:     allMatched,
				Discrepancies: discrepancies,
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *vendorBillService) ApproveVendorBill(
	ctx context.Context,
	id string,
) (*domain.VendorBill, error) {
	bill, err := s.repository.GetVendorBillByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get vendor bill: %w", err)
	}

	if bill.MatchStatus != domain.BillMatchStatusMatched {
		return nil, domain.ErrVendorBillCannotApprove
	}

	return s.repository.UpdateVendorBillStatus(ctx, id, domain.BillStatusApproved)
}

func (s *vendorBillService) CancelVendorBill(
	ctx context.Context,
	id string,
) (*domain.VendorBill, error) {
	var cancelledBill *domain.VendorBill

	err := s.transaction.WithinTransaction(
		ctx,
		func(
			billRepo repository.VendorBillRepository,
			billItemRepo repository.VendorBillItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			grnItemRepo repository.GoodsReceiptItemRepository,
			docNumRepo repository.DocumentNumberRepository,
		) error {
			bill, err := billRepo.GetVendorBillByID(ctx, id)
			if err != nil {
				return fmt.Errorf("get vendor bill: %w", err)
			}

			if bill.Status == domain.BillStatusPaid {
				return fmt.Errorf("cannot cancel a paid vendor bill")
			}

			billItems, err := billItemRepo.GetVendorBillItems(ctx, bill.ID)
			if err != nil {
				return fmt.Errorf("get vendor bill items: %w", err)
			}

			po, err := poRepo.GetPurchaseOrderByID(ctx, bill.PurchaseOrderID)
			if err != nil {
				return fmt.Errorf("get purchase order: %w", err)
			}

			poItems, err := poItemRepo.GetPurchaseOrderItems(ctx, po.ID)
			if err != nil {
				return fmt.Errorf("get purchase order items: %w", err)
			}

			poItemMap := make(map[string]*domain.PurchaseOrderItem, len(poItems))
			for _, pi := range poItems {
				poItemMap[pi.ID] = pi
			}

			// If bill was matched, reverse billed_quantity on PO items
			if bill.MatchStatus == domain.BillMatchStatusMatched {
				for _, bi := range billItems {
					poItem, ok := poItemMap[bi.PurchaseOrderItemID]
					if ok {
						newBilledQty := poItem.BilledQuantity.Sub(bi.MatchedQuantity)
						if newBilledQty.IsNegative() {
							newBilledQty = decimal.Zero
						}

						if _, err := poItemRepo.UpdatePurchaseOrderItemFulfillment(
							ctx,
							poItem.ID,
							poItem.ReceivedQuantity,
							newBilledQty,
						); err != nil {
							return fmt.Errorf("revert billed quantity on po item: %w", err)
						}
						poItem.BilledQuantity = newBilledQty
					}
				}

				// If PO was marked billed, revert back to received
				if po.Status == domain.PurchaseOrderStatusBilled {
					if _, err := poRepo.UpdatePurchaseOrderStatus(ctx, po.ID, domain.PurchaseOrderStatusReceived); err != nil {
						return fmt.Errorf("revert purchase order status: %w", err)
					}
				}
			}

			updated, err := billRepo.UpdateVendorBillStatus(ctx, bill.ID, domain.BillStatusCancelled)
			if err != nil {
				return fmt.Errorf("cancel vendor bill: %w", err)
			}

			cancelledBill = updated
			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return cancelledBill, nil
}

func (s *vendorBillService) GetVendorBillByID(
	ctx context.Context,
	id string,
) (*domain.VendorBill, error) {
	return s.repository.GetVendorBillByID(ctx, id)
}

func (s *vendorBillService) GetVendorBillByInternalNumber(
	ctx context.Context,
	internalBillNumber string,
) (*domain.VendorBill, error) {
	return s.repository.GetVendorBillByInternalNumber(ctx, internalBillNumber)
}

func (s *vendorBillService) ListVendorBills(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return s.repository.ListVendorBills(ctx, limit, offset)
}

func (s *vendorBillService) ListVendorBillsByPO(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.VendorBill, error) {
	return s.repository.ListVendorBillsByPO(ctx, purchaseOrderID)
}

func (s *vendorBillService) ListVendorBillsByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return s.repository.ListVendorBillsByVendor(ctx, vendorID, limit, offset)
}

func (s *vendorBillService) ListVendorBillsByStatus(
	ctx context.Context,
	status domain.BillStatus,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return s.repository.ListVendorBillsByStatus(ctx, status, limit, offset)
}

func (s *vendorBillService) ListVendorBillsByMatchStatus(
	ctx context.Context,
	matchStatus domain.BillMatchStatus,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return s.repository.ListVendorBillsByMatchStatus(ctx, matchStatus, limit, offset)
}

func (s *vendorBillService) GetVendorBillItems(
	ctx context.Context,
	vendorBillID string,
) ([]*domain.VendorBillItem, error) {
	return s.itemRepository.GetVendorBillItems(ctx, vendorBillID)
}

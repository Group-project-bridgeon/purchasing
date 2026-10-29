package service

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

type mockVendorBillRepository struct {
	createdBill *domain.VendorBill
	getBill     *domain.VendorBill
	bills       []*domain.VendorBill
	existing    *domain.VendorBill
	createErr   error
	getErr      error
}

func (m *mockVendorBillRepository) CreateVendorBill(
	ctx context.Context,
	bill *domain.VendorBill,
) (*domain.VendorBill, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdBill = bill
	bill.ID = "bill-123"
	return bill, nil
}

func (m *mockVendorBillRepository) GetVendorBillByID(
	ctx context.Context,
	id string,
) (*domain.VendorBill, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getBill, nil
}

func (m *mockVendorBillRepository) GetVendorBillByInternalNumber(
	ctx context.Context,
	internalBillNumber string,
) (*domain.VendorBill, error) {
	return m.getBill, nil
}

func (m *mockVendorBillRepository) GetVendorBillByVendorInvoice(
	ctx context.Context,
	vendorID string,
	billNumber string,
) (*domain.VendorBill, error) {
	return m.existing, nil
}

func (m *mockVendorBillRepository) ListVendorBills(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return m.bills, nil
}

func (m *mockVendorBillRepository) ListVendorBillsByPO(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.VendorBill, error) {
	return m.bills, nil
}

func (m *mockVendorBillRepository) ListVendorBillsByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return m.bills, nil
}

func (m *mockVendorBillRepository) ListVendorBillsByStatus(
	ctx context.Context,
	status domain.BillStatus,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return m.bills, nil
}

func (m *mockVendorBillRepository) ListVendorBillsByMatchStatus(
	ctx context.Context,
	matchStatus domain.BillMatchStatus,
	limit int,
	offset int,
) ([]*domain.VendorBill, error) {
	return m.bills, nil
}

func (m *mockVendorBillRepository) UpdateVendorBill(
	ctx context.Context,
	bill *domain.VendorBill,
) (*domain.VendorBill, error) {
	return bill, nil
}

func (m *mockVendorBillRepository) UpdateVendorBillMatchStatus(
	ctx context.Context,
	id string,
	matchStatus domain.BillMatchStatus,
	discrepancyReason *string,
) (*domain.VendorBill, error) {
	if m.getBill != nil {
		m.getBill.MatchStatus = matchStatus
		m.getBill.DiscrepancyReason = discrepancyReason
		return m.getBill, nil
	}
	return &domain.VendorBill{ID: id, MatchStatus: matchStatus, DiscrepancyReason: discrepancyReason}, nil
}

func (m *mockVendorBillRepository) UpdateVendorBillStatus(
	ctx context.Context,
	id string,
	status domain.BillStatus,
) (*domain.VendorBill, error) {
	if m.getBill != nil {
		m.getBill.Status = status
		return m.getBill, nil
	}
	return &domain.VendorBill{ID: id, Status: status}, nil
}

type mockVendorBillItemRepository struct {
	createdItems []*domain.VendorBillItem
	items        []*domain.VendorBillItem
	createErr    error
}

func (m *mockVendorBillItemRepository) CreateVendorBillItem(
	ctx context.Context,
	item *domain.VendorBillItem,
) (*domain.VendorBillItem, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdItems = append(m.createdItems, item)
	item.ID = fmt.Sprintf("vbi-%d", len(m.createdItems))
	return item, nil
}

func (m *mockVendorBillItemRepository) GetVendorBillItems(
	ctx context.Context,
	vendorBillID string,
) ([]*domain.VendorBillItem, error) {
	return m.items, nil
}

func (m *mockVendorBillItemRepository) GetVendorBillItemByID(
	ctx context.Context,
	id string,
) (*domain.VendorBillItem, error) {
	return nil, nil
}

func (m *mockVendorBillItemRepository) UpdateVendorBillItemMatch(
	ctx context.Context,
	id string,
	matchedQuantity decimal.Decimal,
	matchStatus domain.BillItemMatchStatus,
) (*domain.VendorBillItem, error) {
	for _, it := range m.items {
		if it.ID == id {
			it.MatchedQuantity = matchedQuantity
			it.MatchStatus = matchStatus
			return it, nil
		}
	}
	return &domain.VendorBillItem{ID: id, MatchedQuantity: matchedQuantity, MatchStatus: matchStatus}, nil
}

type mockBillTransactionManager struct {
	billRepo    repository.VendorBillRepository
	itemRepo    repository.VendorBillItemRepository
	poRepo      repository.PurchaseOrderRepository
	poItemRepo  repository.PurchaseOrderItemRepository
	grnItemRepo repository.GoodsReceiptItemRepository
	docNumRepo  repository.DocumentNumberRepository
}

func (m *mockBillTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		repository.VendorBillRepository,
		repository.VendorBillItemRepository,
		repository.PurchaseOrderRepository,
		repository.PurchaseOrderItemRepository,
		repository.GoodsReceiptItemRepository,
		repository.DocumentNumberRepository,
	) error,
) error {
	return fn(m.billRepo, m.itemRepo, m.poRepo, m.poItemRepo, m.grnItemRepo, m.docNumRepo)
}

func TestCreateVendorBill_Validation(t *testing.T) {
	service := &vendorBillService{}

	t.Run("nil input", func(t *testing.T) {
		_, err := service.CreateVendorBill(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "input cannot be nil") {
			t.Fatalf("expected nil input error, got %v", err)
		}
	})

	t.Run("empty bill number", func(t *testing.T) {
		_, err := service.CreateVendorBill(context.Background(), &CreateVendorBillInput{
			BillNumber: "",
		})
		if err == nil || !strings.Contains(err.Error(), "bill number cannot be empty") {
			t.Fatalf("expected bill number error, got %v", err)
		}
	})

	t.Run("empty purchase order id", func(t *testing.T) {
		_, err := service.CreateVendorBill(context.Background(), &CreateVendorBillInput{
			BillNumber:      "INV-001",
			PurchaseOrderID: "",
		})
		if err == nil || !strings.Contains(err.Error(), "purchase order id cannot be empty") {
			t.Fatalf("expected purchase order id error, got %v", err)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := service.CreateVendorBill(context.Background(), &CreateVendorBillInput{
			BillNumber:      "INV-001",
			PurchaseOrderID: "po-1",
			Items:           []*domain.VendorBillItem{},
		})
		if err == nil || !strings.Contains(err.Error(), "at least one item") {
			t.Fatalf("expected items error, got %v", err)
		}
	})
}

func TestCreateVendorBill_DuplicateInvoice(t *testing.T) {
	billRepo := &mockVendorBillRepository{
		existing: &domain.VendorBill{ID: "bill-existing"},
	}
	itemRepo := &mockVendorBillItemRepository{}
	poRepo := &mockPurchaseOrderRepository{
		getOrder: &domain.PurchaseOrder{
			ID:       "po-1",
			VendorID: "vendor-1",
			Status:   domain.PurchaseOrderStatusIssued,
		},
	}
	poItemRepo := &mockPurchaseOrderItemRepository{}
	grnItemRepo := &mockGoodsReceiptItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}

	tx := &mockBillTransactionManager{
		billRepo:    billRepo,
		itemRepo:    itemRepo,
		poRepo:      poRepo,
		poItemRepo:  poItemRepo,
		grnItemRepo: grnItemRepo,
		docNumRepo:  docNumRepo,
	}

	service := NewVendorBillService(billRepo, itemRepo, tx)

	input := &CreateVendorBillInput{
		BillNumber:      "INV-DUP-100",
		PurchaseOrderID: "po-1",
		Items: []*domain.VendorBillItem{
			{
				PurchaseOrderItemID: "poi-1",
				Quantity:            decimal.NewFromInt(5),
				UnitPrice:           decimal.NewFromInt(10),
			},
		},
	}

	_, err := service.CreateVendorBill(context.Background(), input)
	if err != domain.ErrDuplicateVendorInvoice {
		t.Fatalf("expected ErrDuplicateVendorInvoice, got %v", err)
	}
}

func TestMatchVendorBill_PriceVariance(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:       "po-1",
		VendorID: "vendor-1",
		Currency: "USD",
		Status:   domain.PurchaseOrderStatusReceived,
	}
	poItem := &domain.PurchaseOrderItem{
		ID:               "poi-1",
		PurchaseOrderID:  "po-1",
		Description:      "Heavy Duty Motor",
		Quantity:         decimal.NewFromInt(5),
		UnitPrice:        decimal.NewFromInt(100), // Agreed price: $100
		ReceivedQuantity: decimal.NewFromInt(5),
		BilledQuantity:   decimal.Zero,
	}

	bill := &domain.VendorBill{
		ID:              "bill-1",
		PurchaseOrderID: "po-1",
		VendorID:        "vendor-1",
		Currency:        "USD",
		Status:          domain.BillStatusDraft,
		MatchStatus:     domain.BillMatchStatusUnmatched,
	}
	billItem := &domain.VendorBillItem{
		ID:                  "bi-1",
		VendorBillID:        "bill-1",
		PurchaseOrderItemID: "poi-1",
		Description:         "Heavy Duty Motor",
		Quantity:            decimal.NewFromInt(5),
		UnitPrice:           decimal.NewFromInt(120), // Seller billed $120! Price variance!
	}

	billRepo := &mockVendorBillRepository{getBill: bill}
	itemRepo := &mockVendorBillItemRepository{items: []*domain.VendorBillItem{billItem}}
	poRepo := &mockPurchaseOrderRepository{getOrder: po}
	poItemRepo := &mockPurchaseOrderItemRepository{items: []*domain.PurchaseOrderItem{poItem}}
	grnItemRepo := &mockGoodsReceiptItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}

	tx := &mockBillTransactionManager{
		billRepo:    billRepo,
		itemRepo:    itemRepo,
		poRepo:      poRepo,
		poItemRepo:  poItemRepo,
		grnItemRepo: grnItemRepo,
		docNumRepo:  docNumRepo,
	}

	service := NewVendorBillService(billRepo, itemRepo, tx)

	matchResult, err := service.MatchVendorBill(context.Background(), "bill-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matchResult.IsMatched {
		t.Fatal("expected match to fail due to price variance")
	}
	if bill.MatchStatus != domain.BillMatchStatusDiscrepant {
		t.Fatalf("expected discrepant status, got %s", bill.MatchStatus)
	}
	if billItem.MatchStatus != domain.BillItemMatchStatusPriceVariance {
		t.Fatalf("expected item match status price_variance, got %s", billItem.MatchStatus)
	}
	if len(matchResult.Discrepancies) == 0 {
		t.Fatal("expected discrepancy details to be present")
	}
}

func TestMatchVendorBill_QuantityVariance(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:       "po-1",
		VendorID: "vendor-1",
		Currency: "USD",
		Status:   domain.PurchaseOrderStatusPartiallyReceived,
	}
	poItem := &domain.PurchaseOrderItem{
		ID:               "poi-1",
		PurchaseOrderID:  "po-1",
		Description:      "Heavy Duty Motor",
		Quantity:         decimal.NewFromInt(10),
		UnitPrice:        decimal.NewFromInt(100),
		ReceivedQuantity: decimal.NewFromInt(3), // Warehouse only received 3!
		BilledQuantity:   decimal.Zero,
	}

	bill := &domain.VendorBill{
		ID:              "bill-1",
		PurchaseOrderID: "po-1",
		VendorID:        "vendor-1",
		Currency:        "USD",
		Status:          domain.BillStatusDraft,
		MatchStatus:     domain.BillMatchStatusUnmatched,
	}
	billItem := &domain.VendorBillItem{
		ID:                  "bi-1",
		VendorBillID:        "bill-1",
		PurchaseOrderItemID: "poi-1",
		Description:         "Heavy Duty Motor",
		Quantity:            decimal.NewFromInt(5), // Seller invoices for 5! (5 > 3)
		UnitPrice:           decimal.NewFromInt(100),
	}

	billRepo := &mockVendorBillRepository{getBill: bill}
	itemRepo := &mockVendorBillItemRepository{items: []*domain.VendorBillItem{billItem}}
	poRepo := &mockPurchaseOrderRepository{getOrder: po}
	poItemRepo := &mockPurchaseOrderItemRepository{items: []*domain.PurchaseOrderItem{poItem}}
	grnItemRepo := &mockGoodsReceiptItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}

	tx := &mockBillTransactionManager{
		billRepo:    billRepo,
		itemRepo:    itemRepo,
		poRepo:      poRepo,
		poItemRepo:  poItemRepo,
		grnItemRepo: grnItemRepo,
		docNumRepo:  docNumRepo,
	}

	service := NewVendorBillService(billRepo, itemRepo, tx)

	matchResult, err := service.MatchVendorBill(context.Background(), "bill-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matchResult.IsMatched {
		t.Fatal("expected match to fail due to quantity variance")
	}
	if bill.MatchStatus != domain.BillMatchStatusDiscrepant {
		t.Fatalf("expected discrepant status, got %s", bill.MatchStatus)
	}
	if billItem.MatchStatus != domain.BillItemMatchStatusQuantityVariance {
		t.Fatalf("expected item match status quantity_variance, got %s", billItem.MatchStatus)
	}
}

func TestMatchVendorBill_SuccessAndPOBilledTransition(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:       "po-1",
		VendorID: "vendor-1",
		Currency: "USD",
		Status:   domain.PurchaseOrderStatusReceived,
	}
	poItem := &domain.PurchaseOrderItem{
		ID:               "poi-1",
		PurchaseOrderID:  "po-1",
		Description:      "Copper Wiring",
		Quantity:         decimal.NewFromInt(10),
		UnitPrice:        decimal.NewFromInt(25),
		ReceivedQuantity: decimal.NewFromInt(10),
		BilledQuantity:   decimal.Zero,
	}

	bill := &domain.VendorBill{
		ID:              "bill-1",
		PurchaseOrderID: "po-1",
		VendorID:        "vendor-1",
		Currency:        "USD",
		Status:          domain.BillStatusDraft,
		MatchStatus:     domain.BillMatchStatusUnmatched,
	}
	billItem := &domain.VendorBillItem{
		ID:                  "bi-1",
		VendorBillID:        "bill-1",
		PurchaseOrderItemID: "poi-1",
		Description:         "Copper Wiring",
		Quantity:            decimal.NewFromInt(10),
		UnitPrice:           decimal.NewFromInt(25),
	}

	billRepo := &mockVendorBillRepository{getBill: bill}
	itemRepo := &mockVendorBillItemRepository{items: []*domain.VendorBillItem{billItem}}
	poRepo := &mockPurchaseOrderRepository{getOrder: po}
	poItemRepo := &mockPurchaseOrderItemRepository{items: []*domain.PurchaseOrderItem{poItem}}
	grnItemRepo := &mockGoodsReceiptItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}

	tx := &mockBillTransactionManager{
		billRepo:    billRepo,
		itemRepo:    itemRepo,
		poRepo:      poRepo,
		poItemRepo:  poItemRepo,
		grnItemRepo: grnItemRepo,
		docNumRepo:  docNumRepo,
	}

	service := NewVendorBillService(billRepo, itemRepo, tx)

	matchResult, err := service.MatchVendorBill(context.Background(), "bill-1")
	if err != nil {
		t.Fatalf("unexpected error on match: %v", err)
	}

	if !matchResult.IsMatched {
		t.Fatalf("expected match to pass, got discrepancies: %v", matchResult.Discrepancies)
	}
	if bill.MatchStatus != domain.BillMatchStatusMatched {
		t.Fatalf("expected matched status, got %s", bill.MatchStatus)
	}
	if !poItem.BilledQuantity.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected po item billed quantity 10, got %s", poItem.BilledQuantity)
	}
	// PO should have transitioned to billed!
	if po.Status != domain.PurchaseOrderStatusBilled {
		t.Fatalf("expected PO status billed, got %s", po.Status)
	}

	// Approve bill
	approved, err := service.ApproveVendorBill(context.Background(), "bill-1")
	if err != nil {
		t.Fatalf("unexpected error approving bill: %v", err)
	}
	if approved.Status != domain.BillStatusApproved {
		t.Fatalf("expected approved status, got %s", approved.Status)
	}
}

func TestCancelVendorBill_Rollback(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:       "po-1",
		VendorID: "vendor-1",
		Currency: "USD",
		Status:   domain.PurchaseOrderStatusBilled,
	}
	poItem := &domain.PurchaseOrderItem{
		ID:               "poi-1",
		PurchaseOrderID:  "po-1",
		Quantity:         decimal.NewFromInt(10),
		ReceivedQuantity: decimal.NewFromInt(10),
		BilledQuantity:   decimal.NewFromInt(10),
	}

	bill := &domain.VendorBill{
		ID:              "bill-1",
		PurchaseOrderID: "po-1",
		VendorID:        "vendor-1",
		Status:          domain.BillStatusApproved,
		MatchStatus:     domain.BillMatchStatusMatched,
	}
	billItem := &domain.VendorBillItem{
		ID:                  "bi-1",
		VendorBillID:        "bill-1",
		PurchaseOrderItemID: "poi-1",
		Quantity:            decimal.NewFromInt(10),
		MatchedQuantity:     decimal.NewFromInt(10),
	}

	billRepo := &mockVendorBillRepository{getBill: bill}
	itemRepo := &mockVendorBillItemRepository{items: []*domain.VendorBillItem{billItem}}
	poRepo := &mockPurchaseOrderRepository{getOrder: po}
	poItemRepo := &mockPurchaseOrderItemRepository{items: []*domain.PurchaseOrderItem{poItem}}
	grnItemRepo := &mockGoodsReceiptItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}

	tx := &mockBillTransactionManager{
		billRepo:    billRepo,
		itemRepo:    itemRepo,
		poRepo:      poRepo,
		poItemRepo:  poItemRepo,
		grnItemRepo: grnItemRepo,
		docNumRepo:  docNumRepo,
	}

	service := NewVendorBillService(billRepo, itemRepo, tx)

	cancelled, err := service.CancelVendorBill(context.Background(), "bill-1")
	if err != nil {
		t.Fatalf("unexpected error cancelling bill: %v", err)
	}

	if cancelled.Status != domain.BillStatusCancelled {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}
	// Billed quantity should be rolled back to 0
	if !poItem.BilledQuantity.Equal(decimal.Zero) {
		t.Fatalf("expected po item billed quantity 0, got %s", poItem.BilledQuantity)
	}
	// PO status should revert from billed back to received
	if po.Status != domain.PurchaseOrderStatusReceived {
		t.Fatalf("expected PO status received, got %s", po.Status)
	}
}

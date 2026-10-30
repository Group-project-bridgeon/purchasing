package service

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

type mockGoodsReceiptRepository struct {
	createdGRN *domain.GoodsReceipt
	getGRN     *domain.GoodsReceipt
	receipts   []*domain.GoodsReceipt
	createErr  error
	getErr     error
}

func (m *mockGoodsReceiptRepository) CreateGoodsReceipt(
	ctx context.Context,
	receipt *domain.GoodsReceipt,
) (*domain.GoodsReceipt, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdGRN = receipt
	receipt.ID = "grn-123"
	return receipt, nil
}

func (m *mockGoodsReceiptRepository) GetGoodsReceiptByID(
	ctx context.Context,
	id string,
) (*domain.GoodsReceipt, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getGRN, nil
}

func (m *mockGoodsReceiptRepository) GetGoodsReceiptByNumber(
	ctx context.Context,
	grnNumber string,
) (*domain.GoodsReceipt, error) {
	return m.getGRN, nil
}

func (m *mockGoodsReceiptRepository) ListGoodsReceipts(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.GoodsReceipt, error) {
	return m.receipts, nil
}

func (m *mockGoodsReceiptRepository) ListGoodsReceiptsByPO(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.GoodsReceipt, error) {
	return m.receipts, nil
}

func (m *mockGoodsReceiptRepository) ListGoodsReceiptsByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.GoodsReceipt, error) {
	return m.receipts, nil
}

func (m *mockGoodsReceiptRepository) UpdateGoodsReceiptStatus(
	ctx context.Context,
	id string,
	status domain.GoodsReceiptStatus,
) (*domain.GoodsReceipt, error) {
	if m.getGRN != nil {
		m.getGRN.Status = status
		return m.getGRN, nil
	}
	return &domain.GoodsReceipt{ID: id, Status: status}, nil
}

type mockGoodsReceiptItemRepository struct {
	createdItems []*domain.GoodsReceiptItem
	items        []*domain.GoodsReceiptItem
	createErr    error
}

func (m *mockGoodsReceiptItemRepository) CreateGoodsReceiptItem(
	ctx context.Context,
	item *domain.GoodsReceiptItem,
) (*domain.GoodsReceiptItem, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdItems = append(m.createdItems, item)
	item.ID = fmt.Sprintf("gri-%d", len(m.createdItems))
	return item, nil
}

func (m *mockGoodsReceiptItemRepository) GetGoodsReceiptItems(
	ctx context.Context,
	goodsReceiptID string,
) ([]*domain.GoodsReceiptItem, error) {
	return m.items, nil
}

func (m *mockGoodsReceiptItemRepository) GetGoodsReceiptItemsByPOItemID(
	ctx context.Context,
	purchaseOrderItemID string,
) ([]*domain.GoodsReceiptItem, error) {
	return m.items, nil
}

type mockGRNTransactionManager struct {
	grnRepo    repository.GoodsReceiptRepository
	itemRepo   repository.GoodsReceiptItemRepository
	poRepo     repository.PurchaseOrderRepository
	poItemRepo repository.PurchaseOrderItemRepository
	docNumRepo repository.DocumentNumberRepository
}

func (m *mockGRNTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		repository.GoodsReceiptRepository,
		repository.GoodsReceiptItemRepository,
		repository.PurchaseOrderRepository,
		repository.PurchaseOrderItemRepository,
		repository.DocumentNumberRepository,
	) error,
) error {
	return fn(m.grnRepo, m.itemRepo, m.poRepo, m.poItemRepo, m.docNumRepo)
}

func TestCreateGoodsReceipt_Validation(t *testing.T) {
	service := &goodsReceiptService{}

	t.Run("nil input", func(t *testing.T) {
		_, err := service.CreateGoodsReceipt(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "input cannot be nil") {
			t.Fatalf("expected nil input error, got %v", err)
		}
	})

	t.Run("empty purchase order id", func(t *testing.T) {
		_, err := service.CreateGoodsReceipt(context.Background(), &CreateGoodsReceiptInput{
			PurchaseOrderID: "",
		})
		if err == nil || !strings.Contains(err.Error(), "purchase order id cannot be empty") {
			t.Fatalf("expected purchase order id error, got %v", err)
		}
	})

	t.Run("empty received by", func(t *testing.T) {
		_, err := service.CreateGoodsReceipt(context.Background(), &CreateGoodsReceiptInput{
			PurchaseOrderID: "po-1",
			ReceivedBy:      "",
		})
		if err == nil || !strings.Contains(err.Error(), "received by cannot be empty") {
			t.Fatalf("expected received by error, got %v", err)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := service.CreateGoodsReceipt(context.Background(), &CreateGoodsReceiptInput{
			PurchaseOrderID: "po-1",
			ReceivedBy:      "user-1",
			Items:           []*domain.GoodsReceiptItem{},
		})
		if err == nil || !strings.Contains(err.Error(), "at least one item") {
			t.Fatalf("expected items error, got %v", err)
		}
	})
}

func TestCreateGoodsReceipt_POStatusGuard(t *testing.T) {
	grnRepo := &mockGoodsReceiptRepository{}
	itemRepo := &mockGoodsReceiptItemRepository{}
	poRepo := &mockPurchaseOrderRepository{
		getOrder: &domain.PurchaseOrder{
			ID:     "po-draft",
			Status: domain.PurchaseOrderStatusDraft, // Not issued!
		},
	}
	poItemRepo := &mockPurchaseOrderItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}
	tx := &mockGRNTransactionManager{
		grnRepo:    grnRepo,
		itemRepo:   itemRepo,
		poRepo:     poRepo,
		poItemRepo: poItemRepo,
		docNumRepo: docNumRepo,
	}

	service := NewGoodsReceiptService(grnRepo, itemRepo, tx)

	input := &CreateGoodsReceiptInput{
		PurchaseOrderID: "po-draft",
		ReceivedBy:      "user-1",
		Items: []*domain.GoodsReceiptItem{
			{
				PurchaseOrderItemID: "poi-1",
				ReceivedQuantity:    decimal.NewFromInt(5),
			},
		},
	}

	_, err := service.CreateGoodsReceipt(context.Background(), input)
	if err == nil || !strings.Contains(err.Error(), "not in a valid status for receiving") {
		t.Fatalf("expected invalid po status error, got %v", err)
	}
}

func TestCreateGoodsReceipt_OverReceiptGuard(t *testing.T) {
	grnRepo := &mockGoodsReceiptRepository{}
	itemRepo := &mockGoodsReceiptItemRepository{}
	poRepo := &mockPurchaseOrderRepository{
		getOrder: &domain.PurchaseOrder{
			ID:       "po-issued",
			VendorID: "vendor-1",
			Status:   domain.PurchaseOrderStatusIssued,
		},
	}
	poItemRepo := &mockPurchaseOrderItemRepository{
		items: []*domain.PurchaseOrderItem{
			{
				ID:               "poi-1",
				PurchaseOrderID:  "po-issued",
				Quantity:         decimal.NewFromInt(10), // Ordered 10
				ReceivedQuantity: decimal.NewFromInt(8),  // Already received 8
			},
		},
	}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}
	tx := &mockGRNTransactionManager{
		grnRepo:    grnRepo,
		itemRepo:   itemRepo,
		poRepo:     poRepo,
		poItemRepo: poItemRepo,
		docNumRepo: docNumRepo,
	}

	service := NewGoodsReceiptService(grnRepo, itemRepo, tx)

	// Trying to receive 5 more (8 + 5 = 13 > 10)
	input := &CreateGoodsReceiptInput{
		PurchaseOrderID: "po-issued",
		ReceivedBy:      "user-1",
		Items: []*domain.GoodsReceiptItem{
			{
				PurchaseOrderItemID: "poi-1",
				ReceivedQuantity:    decimal.NewFromInt(5),
			},
		},
	}

	_, err := service.CreateGoodsReceipt(context.Background(), input)
	if err == nil || !strings.Contains(err.Error(), "exceeds remaining ordered quantity") {
		t.Fatalf("expected over-receipt error, got %v", err)
	}
}

func TestCreateGoodsReceipt_PartialAndFullReceipt(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:       "po-100",
		VendorID: "vendor-abc",
		Status:   domain.PurchaseOrderStatusIssued,
	}
	poItem := &domain.PurchaseOrderItem{
		ID:               "poi-100",
		PurchaseOrderID:  "po-100",
		ItemID:           "item-steel",
		Quantity:         decimal.NewFromInt(10),
		ReceivedQuantity: decimal.Zero,
	}

	grnRepo := &mockGoodsReceiptRepository{}
	itemRepo := &mockGoodsReceiptItemRepository{}
	poRepo := &mockPurchaseOrderRepository{getOrder: po}
	poItemRepo := &mockPurchaseOrderItemRepository{items: []*domain.PurchaseOrderItem{poItem}}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 5}
	tx := &mockGRNTransactionManager{
		grnRepo:    grnRepo,
		itemRepo:   itemRepo,
		poRepo:     poRepo,
		poItemRepo: poItemRepo,
		docNumRepo: docNumRepo,
	}

	service := NewGoodsReceiptService(grnRepo, itemRepo, tx)

	// Step 1: Partial delivery of 4 units
	input1 := &CreateGoodsReceiptInput{
		PurchaseOrderID: "po-100",
		ReceivedBy:      "user-warehouse",
		Items: []*domain.GoodsReceiptItem{
			{
				PurchaseOrderItemID: "poi-100",
				ReceivedQuantity:    decimal.NewFromInt(4),
				RejectedQuantity:    decimal.NewFromInt(1), // 1 rejected
			},
		},
	}

	grn1, err := service.CreateGoodsReceipt(context.Background(), input1)
	if err != nil {
		t.Fatalf("unexpected error on partial receipt: %v", err)
	}

	expectedGRNNumber := fmt.Sprintf("GRN-%d-%06d", time.Now().Year(), 5)
	if grn1.GRNNumber != expectedGRNNumber {
		t.Fatalf("expected grn number %s, got %s", expectedGRNNumber, grn1.GRNNumber)
	}
	if po.Status != domain.PurchaseOrderStatusPartiallyReceived {
		t.Fatalf("expected PO status partially_received, got %s", po.Status)
	}
	if !poItem.ReceivedQuantity.Equal(decimal.NewFromInt(4)) {
		t.Fatalf("expected PO item received quantity 4, got %s", poItem.ReceivedQuantity)
	}

	// Step 2: Final delivery of remaining 6 units
	docNumRepo.nextNumber = 6
	input2 := &CreateGoodsReceiptInput{
		PurchaseOrderID: "po-100",
		ReceivedBy:      "user-warehouse",
		Items: []*domain.GoodsReceiptItem{
			{
				PurchaseOrderItemID: "poi-100",
				ReceivedQuantity:    decimal.NewFromInt(6),
			},
		},
	}

	_, err = service.CreateGoodsReceipt(context.Background(), input2)
	if err != nil {
		t.Fatalf("unexpected error on final receipt: %v", err)
	}

	if po.Status != domain.PurchaseOrderStatusReceived {
		t.Fatalf("expected PO status received, got %s", po.Status)
	}
	if !poItem.ReceivedQuantity.Equal(decimal.NewFromInt(10)) {
		t.Fatalf("expected PO item received quantity 10, got %s", poItem.ReceivedQuantity)
	}
}

func TestCancelGoodsReceipt_Success(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:       "po-1",
		VendorID: "vendor-1",
		Status:   domain.PurchaseOrderStatusReceived,
	}
	poItem := &domain.PurchaseOrderItem{
		ID:               "poi-1",
		PurchaseOrderID:  "po-1",
		Quantity:         decimal.NewFromInt(10),
		ReceivedQuantity: decimal.NewFromInt(10),
	}
	grn := &domain.GoodsReceipt{
		ID:              "grn-1",
		PurchaseOrderID: "po-1",
		Status:          domain.GoodsReceiptStatusCompleted,
	}
	grnItem := &domain.GoodsReceiptItem{
		ID:                  "gri-1",
		GoodsReceiptID:      "grn-1",
		PurchaseOrderItemID: "poi-1",
		ReceivedQuantity:    decimal.NewFromInt(10),
	}

	grnRepo := &mockGoodsReceiptRepository{getGRN: grn}
	itemRepo := &mockGoodsReceiptItemRepository{items: []*domain.GoodsReceiptItem{grnItem}}
	poRepo := &mockPurchaseOrderRepository{getOrder: po}
	poItemRepo := &mockPurchaseOrderItemRepository{items: []*domain.PurchaseOrderItem{poItem}}
	tx := &mockGRNTransactionManager{
		grnRepo:    grnRepo,
		itemRepo:   itemRepo,
		poRepo:     poRepo,
		poItemRepo: poItemRepo,
	}

	service := NewGoodsReceiptService(grnRepo, itemRepo, tx)

	cancelled, err := service.CancelGoodsReceipt(context.Background(), "grn-1")
	if err != nil {
		t.Fatalf("unexpected error cancelling grn: %v", err)
	}

	if cancelled.Status != domain.GoodsReceiptStatusCancelled {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}
	// Received qty should revert from 10 to 0
	if !poItem.ReceivedQuantity.Equal(decimal.Zero) {
		t.Fatalf("expected PO item received qty 0, got %s", poItem.ReceivedQuantity)
	}
	// PO status should revert to issued
	if po.Status != domain.PurchaseOrderStatusIssued {
		t.Fatalf("expected PO status issued, got %s", po.Status)
	}
}

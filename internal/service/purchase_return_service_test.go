package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

type mockPurchaseReturnRepository struct {
	createdReturn *domain.PurchaseReturn
	getReturn     *domain.PurchaseReturn
	returns       []*domain.PurchaseReturn
	count         int64
	createErr     error
	getErr        error
}

func (m *mockPurchaseReturnRepository) CreatePurchaseReturn(
	ctx context.Context,
	ret *domain.PurchaseReturn,
) (*domain.PurchaseReturn, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	ret.ID = "ret-123"
	m.createdReturn = ret
	return ret, nil
}

func (m *mockPurchaseReturnRepository) GetPurchaseReturnByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseReturn, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.getReturn != nil {
		return m.getReturn, nil
	}
	return nil, domain.ErrPurchaseReturnNotFound
}

func (m *mockPurchaseReturnRepository) GetPurchaseReturnByNumber(
	ctx context.Context,
	returnNumber string,
) (*domain.PurchaseReturn, error) {
	return m.getReturn, nil
}

func (m *mockPurchaseReturnRepository) ListPurchaseReturns(
	ctx context.Context,
	poID *string,
	vendorID *string,
	status *domain.PurchaseReturnStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseReturn, error) {
	return m.returns, nil
}

func (m *mockPurchaseReturnRepository) CountPurchaseReturns(
	ctx context.Context,
	poID *string,
	vendorID *string,
	status *domain.PurchaseReturnStatus,
) (int64, error) {
	return m.count, nil
}

func (m *mockPurchaseReturnRepository) UpdatePurchaseReturn(
	ctx context.Context,
	ret *domain.PurchaseReturn,
) (*domain.PurchaseReturn, error) {
	m.getReturn = ret
	return ret, nil
}

func (m *mockPurchaseReturnRepository) UpdatePurchaseReturnStatus(
	ctx context.Context,
	id string,
	status domain.PurchaseReturnStatus,
) (*domain.PurchaseReturn, error) {
	if m.getReturn != nil {
		m.getReturn.Status = status
		return m.getReturn, nil
	}
	return &domain.PurchaseReturn{ID: id, Status: status}, nil
}

func (m *mockPurchaseReturnRepository) ShipPurchaseReturn(
	ctx context.Context,
	id string,
	carrier string,
	trackingNumber string,
) (*domain.PurchaseReturn, error) {
	if m.getReturn != nil {
		m.getReturn.Status = domain.PurchaseReturnStatusShipped
		m.getReturn.Carrier = &carrier
		m.getReturn.TrackingNumber = &trackingNumber
		return m.getReturn, nil
	}
	return &domain.PurchaseReturn{
		ID:             id,
		Status:         domain.PurchaseReturnStatusShipped,
		Carrier:        &carrier,
		TrackingNumber: &trackingNumber,
	}, nil
}

type mockPurchaseReturnItemRepository struct {
	createdItems   []*domain.PurchaseReturnItem
	items          []*domain.PurchaseReturnItem
	totalReturned  decimal.Decimal
	totalReturnErr error
}

func (m *mockPurchaseReturnItemRepository) CreatePurchaseReturnItem(
	ctx context.Context,
	item *domain.PurchaseReturnItem,
) (*domain.PurchaseReturnItem, error) {
	item.ID = "item-ret-1"
	m.createdItems = append(m.createdItems, item)
	return item, nil
}

func (m *mockPurchaseReturnItemRepository) GetPurchaseReturnItems(
	ctx context.Context,
	purchaseReturnID string,
) ([]*domain.PurchaseReturnItem, error) {
	return m.items, nil
}

func (m *mockPurchaseReturnItemRepository) GetPurchaseReturnItemByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseReturnItem, error) {
	return nil, nil
}

func (m *mockPurchaseReturnItemRepository) GetTotalReturnedQuantityForPOItem(
	ctx context.Context,
	purchaseOrderItemID string,
) (decimal.Decimal, error) {
	if m.totalReturnErr != nil {
		return decimal.Zero, m.totalReturnErr
	}
	return m.totalReturned, nil
}

type mockReturnTransactionManager struct {
	returnRepo     repository.PurchaseReturnRepository
	returnItemRepo repository.PurchaseReturnItemRepository
	poRepo         repository.PurchaseOrderRepository
	poItemRepo     repository.PurchaseOrderItemRepository
	docNumRepo     repository.DocumentNumberRepository
}

func (m *mockReturnTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		repository.PurchaseReturnRepository,
		repository.PurchaseReturnItemRepository,
		repository.PurchaseOrderRepository,
		repository.PurchaseOrderItemRepository,
		repository.DocumentNumberRepository,
	) error,
) error {
	return fn(m.returnRepo, m.returnItemRepo, m.poRepo, m.poItemRepo, m.docNumRepo)
}

func TestCreatePurchaseReturn_Validation(t *testing.T) {
	service := &purchaseReturnService{}

	t.Run("nil input", func(t *testing.T) {
		_, err := service.CreatePurchaseReturn(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "input cannot be nil") {
			t.Fatalf("expected nil input error, got %v", err)
		}
	})

	t.Run("empty purchase order id", func(t *testing.T) {
		_, err := service.CreatePurchaseReturn(context.Background(), &CreatePurchaseReturnInput{
			PurchaseOrderID: "",
		})
		if err == nil || !strings.Contains(err.Error(), "purchase order id is required") {
			t.Fatalf("expected purchase order id error, got %v", err)
		}
	})

	t.Run("empty created by", func(t *testing.T) {
		_, err := service.CreatePurchaseReturn(context.Background(), &CreatePurchaseReturnInput{
			PurchaseOrderID: "po-1",
			CreatedBy:       "",
		})
		if err == nil || !strings.Contains(err.Error(), "created by user id is required") {
			t.Fatalf("expected created by error, got %v", err)
		}
	})

	t.Run("empty reason", func(t *testing.T) {
		_, err := service.CreatePurchaseReturn(context.Background(), &CreatePurchaseReturnInput{
			PurchaseOrderID: "po-1",
			CreatedBy:       "user-1",
			Reason:          "",
		})
		if err == nil || !strings.Contains(err.Error(), "return reason is required") {
			t.Fatalf("expected reason error, got %v", err)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := service.CreatePurchaseReturn(context.Background(), &CreatePurchaseReturnInput{
			PurchaseOrderID: "po-1",
			CreatedBy:       "user-1",
			Reason:          "Damaged goods",
			Items:           []CreatePurchaseReturnItemInput{},
		})
		if err == nil || !strings.Contains(err.Error(), "at least one return item is required") {
			t.Fatalf("expected at least one item error, got %v", err)
		}
	})
}

func TestCreatePurchaseReturn_InvalidPOStatus(t *testing.T) {
	poRepo := &mockPurchaseOrderRepository{
		getOrder: &domain.PurchaseOrder{
			ID:       "po-1",
			Status:   domain.PurchaseOrderStatusDraft,
			VendorID: "ven-1",
		},
	}
	service := &purchaseReturnService{
		poRepo: poRepo,
	}

	_, err := service.CreatePurchaseReturn(context.Background(), &CreatePurchaseReturnInput{
		PurchaseOrderID: "po-1",
		CreatedBy:       "user-1",
		Reason:          "Damaged goods",
		Items: []CreatePurchaseReturnItemInput{
			{
				PurchaseOrderItemID: "poi-1",
				ReturnQuantity:      decimal.NewFromInt(5),
			},
		},
	})

	if err == nil || !strings.Contains(err.Error(), domain.ErrInvalidPOStatusForReturn.Error()) {
		t.Fatalf("expected ErrInvalidPOStatusForReturn, got %v", err)
	}
}

func TestCreatePurchaseReturn_ExceedsReturnableQuantity(t *testing.T) {
	poRepo := &mockPurchaseOrderRepository{
		getOrder: &domain.PurchaseOrder{
			ID:       "po-1",
			Status:   domain.PurchaseOrderStatusReceived,
			VendorID: "ven-1",
		},
	}
	poItemRepo := &mockPurchaseOrderItemRepository{
		items: []*domain.PurchaseOrderItem{
			{
				ID:               "poi-1",
				PurchaseOrderID:  "po-1",
				ItemID:           "item-1",
				Description:      "Steel bolts",
				Quantity:         decimal.NewFromInt(100),
				UnitPrice:        decimal.NewFromFloat(2.50),
				ReceivedQuantity: decimal.NewFromInt(10), // Only 10 received
			},
		},
	}
	retItemRepo := &mockPurchaseReturnItemRepository{
		totalReturned: decimal.NewFromInt(4), // 4 already returned, so only 6 available
	}

	service := &purchaseReturnService{
		poRepo:         poRepo,
		poItemRepo:     poItemRepo,
		returnItemRepo: retItemRepo,
	}

	// Attempting to return 8 when only 6 are available
	_, err := service.CreatePurchaseReturn(context.Background(), &CreatePurchaseReturnInput{
		PurchaseOrderID: "po-1",
		CreatedBy:       "user-1",
		Reason:          "Defective batch",
		Items: []CreatePurchaseReturnItemInput{
			{
				PurchaseOrderItemID: "poi-1",
				ReturnQuantity:      decimal.NewFromInt(8),
			},
		},
	})

	if err == nil || !strings.Contains(err.Error(), domain.ErrExceedsReturnableQuantity.Error()) {
		t.Fatalf("expected ErrExceedsReturnableQuantity, got %v", err)
	}
}

func TestCreatePurchaseReturn_Success(t *testing.T) {
	poRepo := &mockPurchaseOrderRepository{
		getOrder: &domain.PurchaseOrder{
			ID:       "po-1",
			Status:   domain.PurchaseOrderStatusPartiallyReceived,
			VendorID: "ven-1",
		},
	}
	poItemRepo := &mockPurchaseOrderItemRepository{
		items: []*domain.PurchaseOrderItem{
			{
				ID:               "poi-1",
				PurchaseOrderID:  "po-1",
				ItemID:           "item-1",
				Description:      "Widget A",
				Quantity:         decimal.NewFromInt(50),
				UnitPrice:        decimal.NewFromFloat(10.00),
				ReceivedQuantity: decimal.NewFromInt(20),
			},
		},
	}
	retRepo := &mockPurchaseReturnRepository{}
	retItemRepo := &mockPurchaseReturnItemRepository{
		totalReturned: decimal.Zero,
	}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 42}
	txManager := &mockReturnTransactionManager{
		returnRepo:     retRepo,
		returnItemRepo: retItemRepo,
		poRepo:         poRepo,
		poItemRepo:     poItemRepo,
		docNumRepo:     docNumRepo,
	}

	service := NewPurchaseReturnService(
		retRepo,
		retItemRepo,
		poRepo,
		poItemRepo,
		docNumRepo,
		txManager,
	)

	returnDate := time.Date(2026, 10, 22, 0, 0, 0, 0, time.UTC)
	detail, err := service.CreatePurchaseReturn(context.Background(), &CreatePurchaseReturnInput{
		PurchaseOrderID: "po-1",
		CreatedBy:       "user-1",
		Reason:          "Cracked casing during transit",
		ReturnDate:      &returnDate,
		Items: []CreatePurchaseReturnItemInput{
			{
				PurchaseOrderItemID: "poi-1",
				ReturnQuantity:      decimal.NewFromInt(5),
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error creating return: %v", err)
	}

	if detail.Return.ReturnNumber != "RET-2026-000042" {
		t.Errorf("expected return number RET-2026-000042, got %s", detail.Return.ReturnNumber)
	}
	if detail.Return.Status != domain.PurchaseReturnStatusDraft {
		t.Errorf("expected status draft, got %s", detail.Return.Status)
	}
	// Total return amount: 5 qty * $10.00 = $50.00
	expectedTotal := decimal.NewFromFloat(50.00)
	if !detail.Return.TotalReturnAmount.Equal(expectedTotal) {
		t.Errorf("expected total return amount %s, got %s", expectedTotal, detail.Return.TotalReturnAmount)
	}

	if len(detail.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(detail.Items))
	}
	if detail.Items[0].Description != "Widget A" {
		t.Errorf("expected description 'Widget A', got %s", detail.Items[0].Description)
	}
}

func TestPurchaseReturn_LifecycleTransitions(t *testing.T) {
	retRepo := &mockPurchaseReturnRepository{
		getReturn: &domain.PurchaseReturn{
			ID:     "ret-1",
			Status: domain.PurchaseReturnStatusDraft,
		},
	}
	service := &purchaseReturnService{
		returnRepo: retRepo,
	}

	// 1. Approve from draft -> success
	approved, err := service.ApprovePurchaseReturn(context.Background(), "ret-1")
	if err != nil {
		t.Fatalf("unexpected error approving return: %v", err)
	}
	if approved.Status != domain.PurchaseReturnStatusApproved {
		t.Errorf("expected approved status, got %s", approved.Status)
	}

	// 2. Cannot approve if not draft
	_, err = service.ApprovePurchaseReturn(context.Background(), "ret-1")
	if err == nil || !strings.Contains(err.Error(), domain.ErrPurchaseReturnCannotApprove.Error()) {
		t.Fatalf("expected ErrPurchaseReturnCannotApprove, got %v", err)
	}

	// 3. Ship from approved -> success
	shipped, err := service.ShipPurchaseReturn(context.Background(), "ret-1", "FedEx", "TRACK123456")
	if err != nil {
		t.Fatalf("unexpected error shipping return: %v", err)
	}
	if shipped.Status != domain.PurchaseReturnStatusShipped {
		t.Errorf("expected shipped status, got %s", shipped.Status)
	}

	// 4. Cannot cancel once shipped
	_, err = service.CancelPurchaseReturn(context.Background(), "ret-1")
	if err == nil || !strings.Contains(err.Error(), domain.ErrPurchaseReturnCannotCancel.Error()) {
		t.Fatalf("expected ErrPurchaseReturnCannotCancel, got %v", err)
	}

	// 5. Cancel from draft -> success
	retRepo.getReturn.Status = domain.PurchaseReturnStatusDraft
	cancelled, err := service.CancelPurchaseReturn(context.Background(), "ret-1")
	if err != nil {
		t.Fatalf("unexpected error cancelling return: %v", err)
	}
	if cancelled.Status != domain.PurchaseReturnStatusCancelled {
		t.Errorf("expected cancelled status, got %s", cancelled.Status)
	}
}

func TestCompletePurchaseReturn_Success(t *testing.T) {
	po := &domain.PurchaseOrder{
		ID:       "po-1",
		Status:   domain.PurchaseOrderStatusReceived, // currently fully received
		VendorID: "ven-1",
	}
	poItem := &domain.PurchaseOrderItem{
		ID:               "poi-1",
		PurchaseOrderID:  "po-1",
		Quantity:         decimal.NewFromInt(100),
		ReceivedQuantity: decimal.NewFromInt(100), // 100 received
	}

	ret := &domain.PurchaseReturn{
		ID:              "ret-1",
		PurchaseOrderID: "po-1",
		Status:          domain.PurchaseReturnStatusShipped,
	}

	retItem := &domain.PurchaseReturnItem{
		ID:                  "pri-1",
		PurchaseReturnID:    "ret-1",
		PurchaseOrderItemID: "poi-1",
		ReturnQuantity:      decimal.NewFromInt(20), // returning 20
	}

	poRepo := &mockPurchaseOrderRepository{getOrder: po}
	poItemRepo := &mockPurchaseOrderItemRepository{items: []*domain.PurchaseOrderItem{poItem}}
	retRepo := &mockPurchaseReturnRepository{getReturn: ret}
	retItemRepo := &mockPurchaseReturnItemRepository{items: []*domain.PurchaseReturnItem{retItem}}
	txManager := &mockReturnTransactionManager{
		returnRepo:     retRepo,
		returnItemRepo: retItemRepo,
		poRepo:         poRepo,
		poItemRepo:     poItemRepo,
	}

	service := NewPurchaseReturnService(
		retRepo,
		retItemRepo,
		poRepo,
		poItemRepo,
		nil,
		txManager,
	)

	detail, err := service.CompletePurchaseReturn(context.Background(), "ret-1")
	if err != nil {
		t.Fatalf("unexpected error completing return: %v", err)
	}

	if detail.Return.Status != domain.PurchaseReturnStatusCompleted {
		t.Errorf("expected return status completed, got %s", detail.Return.Status)
	}

	// PO item received quantity must be decremented: 100 - 20 = 80
	if !poItem.ReceivedQuantity.Equal(decimal.NewFromInt(80)) {
		t.Errorf("expected po item received quantity 80, got %s", poItem.ReceivedQuantity)
	}

	// PO status must have transitioned from received to partially_received!
	if po.Status != domain.PurchaseOrderStatusPartiallyReceived {
		t.Errorf("expected PO status partially_received, got %s", po.Status)
	}
}

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

type mockPurchaseOrderRepository struct {
	createdOrder *domain.PurchaseOrder
	getOrder     *domain.PurchaseOrder
	updatedOrder *domain.PurchaseOrder
	updatedStatus domain.PurchaseOrderStatus
	getErr       error
	createErr    error
	updateErr    error
}

func (m *mockPurchaseOrderRepository) CreatePurchaseOrder(
	ctx context.Context,
	order *domain.PurchaseOrder,
) (*domain.PurchaseOrder, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdOrder = order
	order.ID = "po-123"
	return order, nil
}

func (m *mockPurchaseOrderRepository) GetPurchaseOrderByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrder, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getOrder, nil
}

func (m *mockPurchaseOrderRepository) GetPurchaseOrderByNumber(
	ctx context.Context,
	poNumber string,
) (*domain.PurchaseOrder, error) {
	return m.getOrder, nil
}

func (m *mockPurchaseOrderRepository) ListPurchaseOrders(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	return []*domain.PurchaseOrder{m.getOrder}, nil
}

func (m *mockPurchaseOrderRepository) ListPurchaseOrdersByStatus(
	ctx context.Context,
	status domain.PurchaseOrderStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	return []*domain.PurchaseOrder{m.getOrder}, nil
}

func (m *mockPurchaseOrderRepository) ListPurchaseOrdersByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	return []*domain.PurchaseOrder{m.getOrder}, nil
}

func (m *mockPurchaseOrderRepository) UpdatePurchaseOrder(
	ctx context.Context,
	order *domain.PurchaseOrder,
) (*domain.PurchaseOrder, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.updatedOrder = order
	return order, nil
}

func (m *mockPurchaseOrderRepository) UpdatePurchaseOrderStatus(
	ctx context.Context,
	id string,
	status domain.PurchaseOrderStatus,
) (*domain.PurchaseOrder, error) {
	if m.updateErr != nil {
		return nil, m.updateErr
	}
	m.updatedStatus = status
	if m.getOrder != nil {
		m.getOrder.Status = status
		return m.getOrder, nil
	}
	return &domain.PurchaseOrder{ID: id, Status: status}, nil
}

type mockPurchaseOrderItemRepository struct {
	createdItems []*domain.PurchaseOrderItem
	items        []*domain.PurchaseOrderItem
	createErr    error
}

func (m *mockPurchaseOrderItemRepository) CreatePurchaseOrderItem(
	ctx context.Context,
	item *domain.PurchaseOrderItem,
) (*domain.PurchaseOrderItem, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	m.createdItems = append(m.createdItems, item)
	item.ID = fmt.Sprintf("poi-%d", len(m.createdItems))
	return item, nil
}

func (m *mockPurchaseOrderItemRepository) GetPurchaseOrderItems(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.PurchaseOrderItem, error) {
	return m.items, nil
}

func (m *mockPurchaseOrderItemRepository) GetPurchaseOrderItemByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrderItem, error) {
	return nil, nil
}

func (m *mockPurchaseOrderItemRepository) UpdatePurchaseOrderItem(
	ctx context.Context,
	item *domain.PurchaseOrderItem,
) (*domain.PurchaseOrderItem, error) {
	return item, nil
}

func (m *mockPurchaseOrderItemRepository) DeletePurchaseOrderItem(
	ctx context.Context,
	id string,
) error {
	return nil
}

func (m *mockPurchaseOrderItemRepository) UpdatePurchaseOrderItemFulfillment(
	ctx context.Context,
	id string,
	receivedQuantity decimal.Decimal,
	billedQuantity decimal.Decimal,
) (*domain.PurchaseOrderItem, error) {
	return nil, nil
}

type mockPOTansactionManager struct {
	poRepo     repository.PurchaseOrderRepository
	itemRepo   repository.PurchaseOrderItemRepository
	docNumRepo repository.DocumentNumberRepository
	prRepo     repository.PurchaseRequestRepository
	prItemRepo repository.PurchaseRequestItemRepository
}

func (m *mockPOTansactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		repository.PurchaseOrderRepository,
		repository.PurchaseOrderItemRepository,
		repository.DocumentNumberRepository,
		repository.PurchaseRequestRepository,
		repository.PurchaseRequestItemRepository,
	) error,
) error {
	return fn(m.poRepo, m.itemRepo, m.docNumRepo, m.prRepo, m.prItemRepo)
}

func TestCreatePurchaseOrder_Validation(t *testing.T) {
	service := &purchaseOrderService{}

	t.Run("nil input", func(t *testing.T) {
		_, err := service.CreatePurchaseOrder(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "input cannot be nil") {
			t.Fatalf("expected nil input error, got %v", err)
		}
	})

	t.Run("nil order", func(t *testing.T) {
		_, err := service.CreatePurchaseOrder(context.Background(), &CreatePurchaseOrderInput{})
		if err == nil || !strings.Contains(err.Error(), "purchase order cannot be nil") {
			t.Fatalf("expected nil order error, got %v", err)
		}
	})

	t.Run("empty vendor id", func(t *testing.T) {
		_, err := service.CreatePurchaseOrder(context.Background(), &CreatePurchaseOrderInput{
			Order: &domain.PurchaseOrder{VendorID: ""},
		})
		if err == nil || !strings.Contains(err.Error(), "vendor id cannot be empty") {
			t.Fatalf("expected vendor id error, got %v", err)
		}
	})

	t.Run("no items", func(t *testing.T) {
		_, err := service.CreatePurchaseOrder(context.Background(), &CreatePurchaseOrderInput{
			Order: &domain.PurchaseOrder{VendorID: "vendor-1"},
		})
		if err == nil || !strings.Contains(err.Error(), "at least one item") {
			t.Fatalf("expected at least one item error, got %v", err)
		}
	})

	t.Run("invalid item quantity", func(t *testing.T) {
		_, err := service.CreatePurchaseOrder(context.Background(), &CreatePurchaseOrderInput{
			Order: &domain.PurchaseOrder{VendorID: "vendor-1"},
			Items: []*domain.PurchaseOrderItem{
				{Quantity: decimal.Zero},
			},
		})
		if err == nil || !strings.Contains(err.Error(), "greater than zero") {
			t.Fatalf("expected quantity error, got %v", err)
		}
	})

	t.Run("negative unit price", func(t *testing.T) {
		_, err := service.CreatePurchaseOrder(context.Background(), &CreatePurchaseOrderInput{
			Order: &domain.PurchaseOrder{VendorID: "vendor-1"},
			Items: []*domain.PurchaseOrderItem{
				{
					Quantity:  decimal.NewFromInt(1),
					UnitPrice: decimal.NewFromInt(-5),
				},
			},
		})
		if err == nil || !strings.Contains(err.Error(), "unit price cannot be negative") {
			t.Fatalf("expected unit price error, got %v", err)
		}
	})
}

func TestCreatePurchaseOrder_Success(t *testing.T) {
	poRepo := &mockPurchaseOrderRepository{}
	itemRepo := &mockPurchaseOrderItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 42}
	tx := &mockPOTansactionManager{
		poRepo:     poRepo,
		itemRepo:   itemRepo,
		docNumRepo: docNumRepo,
	}

	service := NewPurchaseOrderService(poRepo, itemRepo, tx)

	qty := decimal.NewFromInt(10)
	price := decimal.NewFromInt(20)
	tax := decimal.NewFromInt(10) // 10% tax

	input := &CreatePurchaseOrderInput{
		Order: &domain.PurchaseOrder{
			VendorID: "vendor-xyz",
		},
		Items: []*domain.PurchaseOrderItem{
			{
				ItemID:      "item-1",
				Description: "Widget A",
				Quantity:    qty,
				UnitPrice:   price,
				TaxRate:     tax,
			},
		},
	}

	created, err := service.CreatePurchaseOrder(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedPONumber := fmt.Sprintf("PO-%d-%06d", time.Now().Year(), 42)
	if created.PONumber != expectedPONumber {
		t.Fatalf("expected PO number %s, got %s", expectedPONumber, created.PONumber)
	}

	// 10 * 20 = 200 subtotal, 10% tax = 20, total = 220
	if !created.Subtotal.Equal(decimal.NewFromInt(200)) {
		t.Fatalf("expected subtotal 200, got %s", created.Subtotal)
	}
	if !created.TaxAmount.Equal(decimal.NewFromInt(20)) {
		t.Fatalf("expected tax amount 20, got %s", created.TaxAmount)
	}
	if !created.TotalAmount.Equal(decimal.NewFromInt(220)) {
		t.Fatalf("expected total amount 220, got %s", created.TotalAmount)
	}
	if created.Status != domain.PurchaseOrderStatusDraft {
		t.Fatalf("expected status draft, got %s", created.Status)
	}
	if len(itemRepo.createdItems) != 1 {
		t.Fatalf("expected 1 created item, got %d", len(itemRepo.createdItems))
	}
}

func TestCreatePurchaseOrderFromPR_Success(t *testing.T) {
	poRepo := &mockPurchaseOrderRepository{}
	itemRepo := &mockPurchaseOrderItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}
	unitPrice := decimal.NewFromInt(50)
	prRepo := &mockPurchaseRequestRepository{
		getRequest: &domain.PurchaseRequest{
			ID:     "pr-approved-123",
			Status: domain.PurchaseRequestStatusApproved,
		},
	}
	prItemRepo := &mockPurchaseRequestItemRepository{
		items: []*domain.PurchaseRequestItem{
			{
				ItemID:             "item-abc",
				Description:        "Steel Bar",
				Quantity:           decimal.NewFromInt(5),
				EstimatedUnitPrice: &unitPrice,
			},
		},
	}

	tx := &mockPOTansactionManager{
		poRepo:     poRepo,
		itemRepo:   itemRepo,
		docNumRepo: docNumRepo,
		prRepo:     prRepo,
		prItemRepo: prItemRepo,
	}

	service := NewPurchaseOrderService(poRepo, itemRepo, tx)

	input := &ConvertPRToPOInput{
		PurchaseRequestID: "pr-approved-123",
		VendorID:          "vendor-999",
	}

	order, err := service.CreatePurchaseOrderFromPR(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if order.PurchaseRequestID == nil || *order.PurchaseRequestID != "pr-approved-123" {
		t.Fatalf("expected PR ID pr-approved-123, got %v", order.PurchaseRequestID)
	}
	if order.VendorID != "vendor-999" {
		t.Fatalf("expected vendor id vendor-999, got %s", order.VendorID)
	}
	// 5 * 50 = 250
	if !order.Subtotal.Equal(decimal.NewFromInt(250)) {
		t.Fatalf("expected subtotal 250, got %s", order.Subtotal)
	}
}

func TestCreatePurchaseOrderFromPR_NotApproved(t *testing.T) {
	poRepo := &mockPurchaseOrderRepository{}
	itemRepo := &mockPurchaseOrderItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 1}
	prRepo := &mockPurchaseRequestRepository{
		getRequest: &domain.PurchaseRequest{
			ID:     "pr-draft",
			Status: domain.PurchaseRequestStatusDraft,
		},
	}
	tx := &mockPOTansactionManager{
		poRepo:     poRepo,
		itemRepo:   itemRepo,
		docNumRepo: docNumRepo,
		prRepo:     prRepo,
	}

	service := NewPurchaseOrderService(poRepo, itemRepo, tx)

	input := &ConvertPRToPOInput{
		PurchaseRequestID: "pr-draft",
		VendorID:          "vendor-1",
	}

	_, err := service.CreatePurchaseOrderFromPR(context.Background(), input)
	if err == nil || !strings.Contains(err.Error(), "must be approved") {
		t.Fatalf("expected must be approved error, got %v", err)
	}
}

func TestPurchaseOrder_StatusTransitions(t *testing.T) {
	poRepo := &mockPurchaseOrderRepository{
		getOrder: &domain.PurchaseOrder{
			ID:     "po-1",
			Status: domain.PurchaseOrderStatusDraft,
		},
	}
	service := NewPurchaseOrderService(poRepo, nil, nil)

	// Approve draft -> approved
	approved, err := service.ApprovePurchaseOrder(context.Background(), "po-1")
	if err != nil {
		t.Fatalf("failed to approve: %v", err)
	}
	if approved.Status != domain.PurchaseOrderStatusApproved {
		t.Fatalf("expected approved status, got %s", approved.Status)
	}

	// Issue approved -> issued
	issued, err := service.IssuePurchaseOrder(context.Background(), "po-1")
	if err != nil {
		t.Fatalf("failed to issue: %v", err)
	}
	if issued.Status != domain.PurchaseOrderStatusIssued {
		t.Fatalf("expected issued status, got %s", issued.Status)
	}

	// Cancel issued -> cancelled
	cancelled, err := service.CancelPurchaseOrder(context.Background(), "po-1")
	if err != nil {
		t.Fatalf("failed to cancel: %v", err)
	}
	if cancelled.Status != domain.PurchaseOrderStatusCancelled {
		t.Fatalf("expected cancelled status, got %s", cancelled.Status)
	}

	// Cannot cancel if received
	poRepo.getOrder.Status = domain.PurchaseOrderStatusReceived
	_, err = service.CancelPurchaseOrder(context.Background(), "po-1")
	if err != domain.ErrPurchaseOrderCannotCancel {
		t.Fatalf("expected ErrPurchaseOrderCannotCancel, got %v", err)
	}
}

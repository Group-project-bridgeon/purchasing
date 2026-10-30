package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
	"github.com/shopspring/decimal"
)

func TestCreatePurchaseRequest_NilInput(t *testing.T) {
	service := &purchaseRequestService{}

	_, err := service.CreatePurchaseRequest(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "input cannot be nil") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePurchaseRequest_NilRequest(t *testing.T) {
	service := &purchaseRequestService{}

	input := &CreatePurchaseRequestInput{}

	_, err := service.CreatePurchaseRequest(context.Background(), input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "purchase request cannot be nil") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePurchaseRequest_NoItems(t *testing.T) {
	service := &purchaseRequestService{}

	input := &CreatePurchaseRequestInput{
		Request: &domain.PurchaseRequest{},
	}

	_, err := service.CreatePurchaseRequest(context.Background(), input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "at least one item") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePurchaseRequest_NilItem(t *testing.T) {
	service := &purchaseRequestService{}

	input := &CreatePurchaseRequestInput{
		Request: &domain.PurchaseRequest{},
		Items: []*domain.PurchaseRequestItem{
			nil,
		},
	}

	_, err := service.CreatePurchaseRequest(context.Background(), input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "item 1 cannot be nil") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePurchaseRequest_InvalidQuantity(t *testing.T) {
	service := &purchaseRequestService{}

	input := &CreatePurchaseRequestInput{
		Request: &domain.PurchaseRequest{},
		Items: []*domain.PurchaseRequestItem{
			{
				Quantity: decimal.Zero,
			},
		},
	}

	_, err := service.CreatePurchaseRequest(context.Background(), input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "positive quantity") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePurchaseRequest_NegativeQuantity(t *testing.T) {
	service := &purchaseRequestService{}

	input := &CreatePurchaseRequestInput{
		Request: &domain.PurchaseRequest{},
		Items: []*domain.PurchaseRequestItem{
			{
				Quantity: decimal.NewFromInt(-1),
			},
		},
	}

	_, err := service.CreatePurchaseRequest(context.Background(), input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "positive quantity") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCreatePurchaseRequest_NegativeEstimatedUnitPrice(t *testing.T) {
	service := &purchaseRequestService{}

	negativePrice := decimal.NewFromInt(-100)

	input := &CreatePurchaseRequestInput{
		Request: &domain.PurchaseRequest{},
		Items: []*domain.PurchaseRequestItem{
			{
				Quantity:           decimal.NewFromInt(1),
				EstimatedUnitPrice: &negativePrice,
			},
		},
	}

	_, err := service.CreatePurchaseRequest(context.Background(), input)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "negative estimated unit price") {
		t.Fatalf("unexpected error: %v", err)
	}
}

type mockTransactionManager struct {
	requestRepository        repository.PurchaseRequestRepository
	itemRepository           repository.PurchaseRequestItemRepository
	documentNumberRepository repository.DocumentNumberRepository
	called                   bool
	err                      error
}

func (m *mockTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		repository.PurchaseRequestRepository,
		repository.PurchaseRequestItemRepository,
		repository.DocumentNumberRepository,
	) error,
) error {
	m.called = true

	if m.err != nil {
		return m.err
	}

	return fn(
		m.requestRepository,
		m.itemRepository,
		m.documentNumberRepository,
	)
}

type mockDocumentNumberRepository struct {
	nextNumber int64
	err        error
}

func (m *mockDocumentNumberRepository) NextDocumentNumber(
	ctx context.Context,
	documentType string,
	year int,
) (int64, error) {
	if m.err != nil {
		return 0, m.err
	}

	return m.nextNumber, nil
}

type mockPurchaseRequestRepository struct {
	createdRequest *domain.PurchaseRequest
	createErr      error
	createCalled   bool
	getRequest     *domain.PurchaseRequest
	getErr         error
}

func (m *mockPurchaseRequestRepository) CreatePurchaseRequest(
	ctx context.Context,
	request *domain.PurchaseRequest,
) (*domain.PurchaseRequest, error) {
	m.createCalled = true

	if m.createErr != nil {
		return nil, m.createErr
	}

	return request, nil
}

func (m *mockPurchaseRequestRepository) GetPurchaseRequestByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getRequest, nil
}

func (m *mockPurchaseRequestRepository) GetPurchaseRequestByNumber(
	ctx context.Context,
	requestNumber string,
) (*domain.PurchaseRequest, error) {
	return nil, nil
}

func (m *mockPurchaseRequestRepository) ListPurchaseRequests(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.PurchaseRequest, error) {
	return nil, nil
}

func (m *mockPurchaseRequestRepository) ListPurchaseRequestsByStatus(
	ctx context.Context,
	status domain.PurchaseRequestStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseRequest, error) {
	return nil, nil
}

func (m *mockPurchaseRequestRepository) UpdatePurchaseRequest(
	ctx context.Context,
	request *domain.PurchaseRequest,
) (*domain.PurchaseRequest, error) {
	return nil, nil
}

func (m *mockPurchaseRequestRepository) SubmitPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	return nil, nil
}

func (m *mockPurchaseRequestRepository) ApprovePurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	return nil, nil
}

func (m *mockPurchaseRequestRepository) RejectPurchaseRequest(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequest, error) {
	return nil, nil
}

type mockPurchaseRequestItemRepository struct {
	createdItems []*domain.PurchaseRequestItem
	items        []*domain.PurchaseRequestItem
	createErr    error
	getErr       error
}

func (m *mockPurchaseRequestItemRepository) CreatePurchaseRequestItem(
	ctx context.Context,
	item *domain.PurchaseRequestItem,
) (*domain.PurchaseRequestItem, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}

	m.createdItems = append(m.createdItems, item)

	return item, nil
}

func (m *mockPurchaseRequestItemRepository) GetPurchaseRequestItems(
	ctx context.Context,
	purchaseRequestID string,
) ([]*domain.PurchaseRequestItem, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.items, nil
}

func (m *mockPurchaseRequestItemRepository) GetPurchaseRequestItemByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequestItem, error) {
	return nil, nil
}

func (m *mockPurchaseRequestItemRepository) UpdatePurchaseRequestItem(
	ctx context.Context,
	item *domain.PurchaseRequestItem,
) (*domain.PurchaseRequestItem, error) {
	return nil, nil
}

func (m *mockPurchaseRequestItemRepository) DeletePurchaseRequestItem(
	ctx context.Context,
	id string,
) error {
	return nil
}

func TestCreatePurchaseRequest_Success(t *testing.T) {
	request := &domain.PurchaseRequest{
		ID:          "request-123",
		RequestedBy: "user-123",
		Purpose:     "Office equipment",
	}

	items := []*domain.PurchaseRequestItem{
		{
			ItemID:      "item-1",
			Description: "Laptop",
			Quantity:    decimal.NewFromInt(2),
		},
		{
			ItemID:      "item-2",
			Description: "Monitor",
			Quantity:    decimal.NewFromInt(2),
		},
	}

	requestRepository := &mockPurchaseRequestRepository{}

	itemRepository := &mockPurchaseRequestItemRepository{}

	documentNumberRepository := &mockDocumentNumberRepository{
		nextNumber: 1,
	}

	transaction := &mockTransactionManager{
		requestRepository:        requestRepository,
		itemRepository:           itemRepository,
		documentNumberRepository: documentNumberRepository,
	}

	service := NewPurchaseRequestService(
		requestRepository,
		itemRepository,
		transaction,
	)

	input := &CreatePurchaseRequestInput{
		Request: request,
		Items:   items,
	}

	result, err := service.CreatePurchaseRequest(
		context.Background(),
		input,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result == nil {
		t.Fatal("expected created request, got nil")
	}

	expectedRequestNumber := fmt.Sprintf(
		"PR-%d-%06d",
		time.Now().Year(),
		1,
	)

	if result.RequestNumber != expectedRequestNumber {
		t.Fatalf(
			"expected request number %q, got %q",
			expectedRequestNumber,
			result.RequestNumber,
		)
	}

	if result.Status != domain.PurchaseRequestStatusDraft {
		t.Fatalf(
			"expected status %q, got %q",
			domain.PurchaseRequestStatusDraft,
			result.Status,
		)
	}

	if !transaction.called {
		t.Fatal("expected transaction to be called")
	}

	if !requestRepository.createCalled {
		t.Fatal("expected request repository to be called")
	}

	if len(itemRepository.createdItems) != 2 {
		t.Fatalf(
			"expected 2 items to be created, got %d",
			len(itemRepository.createdItems),
		)
	}

	for _, item := range itemRepository.createdItems {
		if item.PurchaseRequestID != request.ID {
			t.Fatalf(
				"expected item request ID %q, got %q",
				request.ID,
				item.PurchaseRequestID,
			)
		}
	}
}

func TestCreatePurchaseRequest_RequestCreationFails(t *testing.T) {
	expectedErr := errors.New("database error")

	requestRepository := &mockPurchaseRequestRepository{
		createErr: expectedErr,
	}

	itemRepository := &mockPurchaseRequestItemRepository{}

	documentNumberRepository := &mockDocumentNumberRepository{
		nextNumber: 1,
	}

	transaction := &mockTransactionManager{
		requestRepository:        requestRepository,
		itemRepository:           itemRepository,
		documentNumberRepository: documentNumberRepository,
	}

	service := NewPurchaseRequestService(
		requestRepository,
		itemRepository,
		transaction,
	)

	input := &CreatePurchaseRequestInput{
		Request: &domain.PurchaseRequest{
			Purpose: "Office equipment",
		},
		Items: []*domain.PurchaseRequestItem{
			{
				ItemID:      "item-1",
				Description: "Laptop",
				Quantity:    decimal.NewFromInt(1),
			},
		},
	}

	result, err := service.CreatePurchaseRequest(
		context.Background(),
		input,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}

	if !strings.Contains(err.Error(), "create purchase request") {
		t.Fatalf("expected create purchase request error, got %v", err)
	}

	if len(itemRepository.createdItems) != 0 {
		t.Fatalf(
			"expected no items to be created, got %d",
			len(itemRepository.createdItems),
		)
	}
}

func TestCreatePurchaseRequest_ItemCreationFails(t *testing.T) {
	expectedErr := errors.New("item database error")

	request := &domain.PurchaseRequest{
		ID:          "request-123",
		RequestedBy: "user-123",
		Purpose:     "Office equipment",
	}

	requestRepository := &mockPurchaseRequestRepository{}

	itemRepository := &mockPurchaseRequestItemRepository{
		createErr: expectedErr,
	}

	documentNumberRepository := &mockDocumentNumberRepository{
		nextNumber: 1,
	}

	transaction := &mockTransactionManager{
		requestRepository:        requestRepository,
		itemRepository:           itemRepository,
		documentNumberRepository: documentNumberRepository,
	}

	service := NewPurchaseRequestService(
		requestRepository,
		itemRepository,
		transaction,
	)

	input := &CreatePurchaseRequestInput{
		Request: request,
		Items: []*domain.PurchaseRequestItem{
			{
				ItemID:      "item-1",
				Description: "Laptop",
				Quantity:    decimal.NewFromInt(1),
			},
		},
	}

	result, err := service.CreatePurchaseRequest(
		context.Background(),
		input,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %v", result)
	}

	if !strings.Contains(err.Error(), "create purchase request item 1") {
		t.Fatalf("expected item creation error, got %v", err)
	}
}

func TestAddPurchaseRequestItem_EmptyID(t *testing.T) {
	service := &purchaseRequestService{}

	_, err := service.AddPurchaseRequestItem(context.Background(), "", &domain.PurchaseRequestItem{})
	if err == nil || !strings.Contains(err.Error(), "cannot be empty") {
		t.Fatalf("expected empty id error, got %v", err)
	}
}

func TestAddPurchaseRequestItem_NilItem(t *testing.T) {
	service := &purchaseRequestService{}

	_, err := service.AddPurchaseRequestItem(context.Background(), "pr-1", nil)
	if err == nil || !strings.Contains(err.Error(), "cannot be nil") {
		t.Fatalf("expected nil item error, got %v", err)
	}
}

func TestAddPurchaseRequestItem_InvalidQuantity(t *testing.T) {
	service := &purchaseRequestService{}

	_, err := service.AddPurchaseRequestItem(context.Background(), "pr-1", &domain.PurchaseRequestItem{
		Quantity: decimal.Zero,
	})
	if err == nil || !strings.Contains(err.Error(), "greater than zero") {
		t.Fatalf("expected quantity error, got %v", err)
	}
}

func TestAddPurchaseRequestItem_NegativeEstimatedUnitPrice(t *testing.T) {
	service := &purchaseRequestService{}
	neg := decimal.NewFromInt(-10)

	_, err := service.AddPurchaseRequestItem(context.Background(), "pr-1", &domain.PurchaseRequestItem{
		Quantity:           decimal.NewFromInt(5),
		EstimatedUnitPrice: &neg,
	})
	if err == nil || !strings.Contains(err.Error(), "cannot be negative") {
		t.Fatalf("expected negative price error, got %v", err)
	}
}

func TestAddPurchaseRequestItem_NotDraft(t *testing.T) {
	requestRepo := &mockPurchaseRequestRepository{
		getRequest: &domain.PurchaseRequest{
			ID:     "pr-1",
			Status: domain.PurchaseRequestStatusSubmitted,
		},
	}
	itemRepo := &mockPurchaseRequestItemRepository{}

	service := NewPurchaseRequestService(requestRepo, itemRepo, nil)

	_, err := service.AddPurchaseRequestItem(context.Background(), "pr-1", &domain.PurchaseRequestItem{
		Quantity: decimal.NewFromInt(5),
	})
	if !errors.Is(err, domain.ErrPurchaseRequestNotEditable) {
		t.Fatal("expected typed draft-only lifecycle rejection")
	}
}

func TestAddPurchaseRequestItem_Success(t *testing.T) {
	requestRepo := &mockPurchaseRequestRepository{
		getRequest: &domain.PurchaseRequest{
			ID:     "pr-1",
			Status: domain.PurchaseRequestStatusDraft,
		},
	}
	itemRepo := &mockPurchaseRequestItemRepository{}

	service := NewPurchaseRequestService(requestRepo, itemRepo, nil)

	created, err := service.AddPurchaseRequestItem(context.Background(), "pr-1", &domain.PurchaseRequestItem{
		ItemID:      "item-123",
		Description: "Steel pipes",
		Quantity:    decimal.NewFromInt(10),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if created.PurchaseRequestID != "pr-1" {
		t.Fatalf("expected PR ID pr-1, got %s", created.PurchaseRequestID)
	}
	if len(itemRepo.createdItems) != 1 {
		t.Fatalf("expected 1 created item in repo, got %d", len(itemRepo.createdItems))
	}
}

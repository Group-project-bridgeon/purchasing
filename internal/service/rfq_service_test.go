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

type mockRFQRepository struct {
	createdRFQ *domain.RFQ
	getRFQ     *domain.RFQ
	rfqs       []*domain.RFQ
	count      int64
	createErr  error
	getErr     error
}

func (m *mockRFQRepository) CreateRFQ(
	ctx context.Context,
	rfq *domain.RFQ,
) (*domain.RFQ, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	rfq.ID = "rfq-123"
	m.createdRFQ = rfq
	return rfq, nil
}

func (m *mockRFQRepository) GetRFQByID(
	ctx context.Context,
	id string,
) (*domain.RFQ, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.getRFQ != nil {
		return m.getRFQ, nil
	}
	return nil, domain.ErrRFQNotFound
}

func (m *mockRFQRepository) GetRFQByNumber(
	ctx context.Context,
	rfqNumber string,
) (*domain.RFQ, error) {
	return m.getRFQ, nil
}

func (m *mockRFQRepository) ListRFQs(
	ctx context.Context,
	status *domain.RFQStatus,
	purchaseRequestID *string,
	limit int,
	offset int,
) ([]*domain.RFQ, error) {
	return m.rfqs, nil
}

func (m *mockRFQRepository) CountRFQs(
	ctx context.Context,
	status *domain.RFQStatus,
	purchaseRequestID *string,
) (int64, error) {
	return m.count, nil
}

func (m *mockRFQRepository) UpdateRFQ(
	ctx context.Context,
	rfq *domain.RFQ,
) (*domain.RFQ, error) {
	m.getRFQ = rfq
	return rfq, nil
}

func (m *mockRFQRepository) UpdateRFQStatus(
	ctx context.Context,
	id string,
	status domain.RFQStatus,
) (*domain.RFQ, error) {
	if m.getRFQ != nil {
		m.getRFQ.Status = status
		return m.getRFQ, nil
	}
	return &domain.RFQ{ID: id, Status: status}, nil
}

type mockRFQItemRepository struct {
	createdItems []*domain.RFQItem
	items        []*domain.RFQItem
}

func (m *mockRFQItemRepository) CreateRFQItem(
	ctx context.Context,
	item *domain.RFQItem,
) (*domain.RFQItem, error) {
	item.ID = "rfq-item-1"
	m.createdItems = append(m.createdItems, item)
	return item, nil
}

func (m *mockRFQItemRepository) GetRFQItems(
	ctx context.Context,
	rfqID string,
) ([]*domain.RFQItem, error) {
	return m.items, nil
}

func (m *mockRFQItemRepository) GetRFQItemByID(
	ctx context.Context,
	id string,
) (*domain.RFQItem, error) {
	return nil, nil
}

type mockVendorQuotationRepository struct {
	createdQuotation *domain.VendorQuotation
	getQuotation     *domain.VendorQuotation
	quotations       []*domain.VendorQuotation
	existingRef      *domain.VendorQuotation
	rejectedOther    bool
}

func (m *mockVendorQuotationRepository) CreateVendorQuotation(
	ctx context.Context,
	q *domain.VendorQuotation,
) (*domain.VendorQuotation, error) {
	q.ID = "quote-123"
	m.createdQuotation = q
	return q, nil
}

func (m *mockVendorQuotationRepository) GetVendorQuotationByID(
	ctx context.Context,
	id string,
) (*domain.VendorQuotation, error) {
	if m.getQuotation != nil {
		return m.getQuotation, nil
	}
	return nil, domain.ErrQuotationNotFound
}

func (m *mockVendorQuotationRepository) GetVendorQuotationByRef(
	ctx context.Context,
	rfqID string,
	vendorID string,
	ref string,
) (*domain.VendorQuotation, error) {
	return m.existingRef, nil
}

func (m *mockVendorQuotationRepository) ListVendorQuotationsByRFQ(
	ctx context.Context,
	rfqID string,
) ([]*domain.VendorQuotation, error) {
	return m.quotations, nil
}

func (m *mockVendorQuotationRepository) UpdateVendorQuotationStatus(
	ctx context.Context,
	id string,
	status domain.QuotationStatus,
) (*domain.VendorQuotation, error) {
	if m.getQuotation != nil {
		m.getQuotation.Status = status
		return m.getQuotation, nil
	}
	return &domain.VendorQuotation{ID: id, Status: status}, nil
}

func (m *mockVendorQuotationRepository) RejectOtherVendorQuotations(
	ctx context.Context,
	rfqID string,
	awardedQuotationID string,
) error {
	m.rejectedOther = true
	return nil
}

type mockVendorQuotationItemRepository struct {
	createdItems []*domain.VendorQuotationItem
	items        []*domain.VendorQuotationItem
}

func (m *mockVendorQuotationItemRepository) CreateVendorQuotationItem(
	ctx context.Context,
	item *domain.VendorQuotationItem,
) (*domain.VendorQuotationItem, error) {
	item.ID = "q-item-1"
	m.createdItems = append(m.createdItems, item)
	return item, nil
}

func (m *mockVendorQuotationItemRepository) GetVendorQuotationItems(
	ctx context.Context,
	vendorQuotationID string,
) ([]*domain.VendorQuotationItem, error) {
	return m.items, nil
}

type mockRFQTransactionManager struct {
	rfqRepo       repository.RFQRepository
	rfqItemRepo   repository.RFQItemRepository
	quotationRepo repository.VendorQuotationRepository
	quoteItemRepo repository.VendorQuotationItemRepository
	poRepo        repository.PurchaseOrderRepository
	poItemRepo    repository.PurchaseOrderItemRepository
	docNumRepo    repository.DocumentNumberRepository
	prRepo        repository.PurchaseRequestRepository
	prItemRepo    repository.PurchaseRequestItemRepository
}

func (m *mockRFQTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		repository.RFQRepository,
		repository.RFQItemRepository,
		repository.VendorQuotationRepository,
		repository.VendorQuotationItemRepository,
		repository.PurchaseOrderRepository,
		repository.PurchaseOrderItemRepository,
		repository.DocumentNumberRepository,
		repository.PurchaseRequestRepository,
		repository.PurchaseRequestItemRepository,
	) error,
) error {
	return fn(
		m.rfqRepo,
		m.rfqItemRepo,
		m.quotationRepo,
		m.quoteItemRepo,
		m.poRepo,
		m.poItemRepo,
		m.docNumRepo,
		m.prRepo,
		m.prItemRepo,
	)
}

func TestCreateRFQ_Validation(t *testing.T) {
	service := &rfqService{}

	t.Run("nil input", func(t *testing.T) {
		_, err := service.CreateRFQ(context.Background(), nil)
		if err == nil || !strings.Contains(err.Error(), "input cannot be nil") {
			t.Fatalf("expected nil input error, got %v", err)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		_, err := service.CreateRFQ(context.Background(), &CreateRFQInput{
			Title: "",
		})
		if err == nil || !strings.Contains(err.Error(), "rfq title is required") {
			t.Fatalf("expected title error, got %v", err)
		}
	})

	t.Run("empty created_by", func(t *testing.T) {
		_, err := service.CreateRFQ(context.Background(), &CreateRFQInput{
			Title:     "Laptops sourcing",
			CreatedBy: "",
		})
		if err == nil || !strings.Contains(err.Error(), "created by user id is required") {
			t.Fatalf("expected created by error, got %v", err)
		}
	})

	t.Run("past deadline", func(t *testing.T) {
		_, err := service.CreateRFQ(context.Background(), &CreateRFQInput{
			Title:              "Laptops sourcing",
			CreatedBy:          "user-1",
			SubmissionDeadline: time.Now().Add(-1 * time.Hour),
		})
		if err == nil || !strings.Contains(err.Error(), "submission deadline must be a future date/time") {
			t.Fatalf("expected future deadline error, got %v", err)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		_, err := service.CreateRFQ(context.Background(), &CreateRFQInput{
			Title:              "Laptops sourcing",
			CreatedBy:          "user-1",
			SubmissionDeadline: time.Now().Add(48 * time.Hour),
			Items:              []CreateRFQItemInput{},
		})
		if err == nil || !strings.Contains(err.Error(), "at least one rfq line item is required") {
			t.Fatalf("expected items error, got %v", err)
		}
	})
}

func TestCreateRFQ_Success(t *testing.T) {
	rfqRepo := &mockRFQRepository{}
	rfqItemRepo := &mockRFQItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 15}
	txManager := &mockRFQTransactionManager{
		rfqRepo:     rfqRepo,
		rfqItemRepo: rfqItemRepo,
		docNumRepo:  docNumRepo,
	}

	service := NewRFQService(
		rfqRepo,
		rfqItemRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		docNumRepo,
		txManager,
	)

	deadline := time.Now().Add(72 * time.Hour)
	detail, err := service.CreateRFQ(context.Background(), &CreateRFQInput{
		Title:              "Office Furniture Q4",
		CreatedBy:          "buyer-1",
		SubmissionDeadline: deadline,
		Items: []CreateRFQItemInput{
			{
				ItemID:         "item-chair",
				Description:    "Ergonomic Mesh Chair",
				TargetQuantity: decimal.NewFromInt(25),
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error creating RFQ: %v", err)
	}

	if detail.RFQ.RFQNumber != "RFQ-2026-000015" {
		t.Errorf("expected RFQ number RFQ-2026-000015, got %s", detail.RFQ.RFQNumber)
	}
	if detail.RFQ.Status != domain.RFQStatusDraft {
		t.Errorf("expected status draft, got %s", detail.RFQ.Status)
	}
	if len(detail.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(detail.Items))
	}
	if detail.Items[0].Description != "Ergonomic Mesh Chair" {
		t.Errorf("expected description 'Ergonomic Mesh Chair', got %s", detail.Items[0].Description)
	}
}

func TestCreateRFQFromPR_Success(t *testing.T) {
	prDesc := "Dell XPS 15"
	pr := &domain.PurchaseRequest{
		ID:            "pr-1",
		RequestNumber: "PR-2026-000005",
		Status:        domain.PurchaseRequestStatusApproved,
	}
	prItems := []*domain.PurchaseRequestItem{
		{
			ID:          "pri-1",
			ItemID:      "item-dell",
			Description: prDesc,
			Quantity:    decimal.NewFromInt(10),
		},
	}

	prRepo := &mockPurchaseRequestRepository{getRequest: pr}
	prItemRepo := &mockPurchaseRequestItemRepository{items: prItems}
	rfqRepo := &mockRFQRepository{}
	rfqItemRepo := &mockRFQItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 20}
	txManager := &mockRFQTransactionManager{
		rfqRepo:     rfqRepo,
		rfqItemRepo: rfqItemRepo,
		docNumRepo:  docNumRepo,
	}

	service := NewRFQService(
		rfqRepo,
		rfqItemRepo,
		nil,
		nil,
		nil,
		nil,
		prRepo,
		prItemRepo,
		docNumRepo,
		txManager,
	)

	detail, err := service.CreateRFQFromPR(context.Background(), "pr-1", &CreateRFQFromPRInput{
		SubmissionDeadline: time.Now().Add(48 * time.Hour),
		CreatedBy:          "buyer-1",
	})

	if err != nil {
		t.Fatalf("unexpected error creating RFQ from PR: %v", err)
	}

	if *detail.RFQ.PurchaseRequestID != "pr-1" {
		t.Errorf("expected purchase request id pr-1, got %v", detail.RFQ.PurchaseRequestID)
	}
	if len(detail.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(detail.Items))
	}
	if detail.Items[0].Description != "Dell XPS 15" {
		t.Errorf("expected item description 'Dell XPS 15', got %s", detail.Items[0].Description)
	}
}

func TestRFQ_PublishAndClose(t *testing.T) {
	rfq := &domain.RFQ{
		ID:                 "rfq-1",
		Status:             domain.RFQStatusDraft,
		SubmissionDeadline: time.Now().Add(24 * time.Hour),
	}
	rfqRepo := &mockRFQRepository{getRFQ: rfq}
	service := &rfqService{rfqRepo: rfqRepo}

	// 1. Publish from draft -> success
	published, err := service.PublishRFQ(context.Background(), "rfq-1")
	if err != nil {
		t.Fatalf("unexpected error publishing RFQ: %v", err)
	}
	if published.Status != domain.RFQStatusPublished {
		t.Errorf("expected published status, got %s", published.Status)
	}

	// 2. Close from published -> success
	closed, err := service.CloseRFQ(context.Background(), "rfq-1")
	if err != nil {
		t.Fatalf("unexpected error closing RFQ: %v", err)
	}
	if closed.Status != domain.RFQStatusClosed {
		t.Errorf("expected closed status, got %s", closed.Status)
	}
}

func TestSubmitVendorQuotation_Success(t *testing.T) {
	rfq := &domain.RFQ{
		ID:                 "rfq-1",
		Status:             domain.RFQStatusPublished,
		SubmissionDeadline: time.Now().Add(48 * time.Hour),
	}
	rfqItems := []*domain.RFQItem{
		{
			ID:             "rfq-item-1",
			RFQID:          "rfq-1",
			ItemID:         "item-1",
			TargetQuantity: decimal.NewFromInt(10),
		},
	}

	rfqRepo := &mockRFQRepository{getRFQ: rfq}
	rfqItemRepo := &mockRFQItemRepository{items: rfqItems}
	quoteRepo := &mockVendorQuotationRepository{}
	quoteItemRepo := &mockVendorQuotationItemRepository{}
	txManager := &mockRFQTransactionManager{
		quotationRepo: quoteRepo,
		quoteItemRepo: quoteItemRepo,
	}

	service := NewRFQService(
		rfqRepo,
		rfqItemRepo,
		quoteRepo,
		quoteItemRepo,
		nil,
		nil,
		nil,
		nil,
		nil,
		txManager,
	)

	validUntil := time.Now().Add(30 * 24 * time.Hour)
	leadTime := 7
	taxRate := decimal.NewFromFloat(5.0)

	detail, err := service.SubmitVendorQuotation(context.Background(), &SubmitVendorQuotationInput{
		RFQID:          "rfq-1",
		VendorID:       "vendor-acme",
		QuoteReference: "Q-ACME-2026-001",
		ValidUntil:     validUntil,
		LeadTimeDays:   leadTime,
		Items: []SubmitVendorQuotationItemInput{
			{
				RFQItemID:       "rfq-item-1",
				ItemID:          "item-1",
				Description:     "High quality parts",
				OfferedQuantity: decimal.NewFromInt(10),
				UnitPrice:       decimal.NewFromFloat(100.00),
				TaxRate:         taxRate,
			},
		},
	})

	if err != nil {
		t.Fatalf("unexpected error submitting quotation: %v", err)
	}

	// Subtotal: 10 * 100 = 1000.00
	// Tax: 1000 * 5% = 50.00
	// Total: 1050.00
	expectedSubtotal := decimal.NewFromFloat(1000.00)
	expectedTotal := decimal.NewFromFloat(1050.00)

	if !detail.Quotation.Subtotal.Equal(expectedSubtotal) {
		t.Errorf("expected subtotal %s, got %s", expectedSubtotal, detail.Quotation.Subtotal)
	}
	if !detail.Quotation.TotalAmount.Equal(expectedTotal) {
		t.Errorf("expected total amount %s, got %s", expectedTotal, detail.Quotation.TotalAmount)
	}
	if detail.Quotation.Status != domain.QuotationStatusSubmitted {
		t.Errorf("expected status submitted, got %s", detail.Quotation.Status)
	}
}

func TestCompareRFQQuotations(t *testing.T) {
	quotations := []*domain.VendorQuotation{
		{
			ID:           "q-1",
			VendorID:     "v-1",
			TotalAmount:  decimal.NewFromFloat(5000.00),
			LeadTimeDays: 14,
		},
		{
			ID:           "q-2",
			VendorID:     "v-2",
			TotalAmount:  decimal.NewFromFloat(4200.00), // Lowest price!
			LeadTimeDays: 20,
		},
		{
			ID:           "q-3",
			VendorID:     "v-3",
			TotalAmount:  decimal.NewFromFloat(4800.00),
			LeadTimeDays: 5, // Fastest lead time!
		},
	}

	quoteRepo := &mockVendorQuotationRepository{quotations: quotations}
	service := &rfqService{quotationRepo: quoteRepo}

	matrix, err := service.CompareRFQQuotations(context.Background(), "rfq-1")
	if err != nil {
		t.Fatalf("unexpected error comparing quotations: %v", err)
	}

	if matrix.LowestBidQuotationID == nil || *matrix.LowestBidQuotationID != "q-2" {
		t.Errorf("expected lowest bid q-2, got %v", matrix.LowestBidQuotationID)
	}

	if matrix.FastestLeadTimeQuotationID == nil || *matrix.FastestLeadTimeQuotationID != "q-3" {
		t.Errorf("expected fastest lead time q-3, got %v", matrix.FastestLeadTimeQuotationID)
	}
}

func TestAwardVendorQuotation_Success(t *testing.T) {
	rfq := &domain.RFQ{
		ID:        "rfq-1",
		RFQNumber: "RFQ-2026-000001",
		Status:    domain.RFQStatusPublished,
	}

	quote := &domain.VendorQuotation{
		ID:             "q-winner",
		RFQID:          "rfq-1",
		VendorID:       "vendor-best",
		QuoteReference: "QUOTE-BEST-99",
		Currency:       "USD",
		Subtotal:       decimal.NewFromFloat(1000.00),
		TaxAmount:      decimal.NewFromFloat(50.00),
		TotalAmount:    decimal.NewFromFloat(1050.00),
		LeadTimeDays:   7,
		Status:         domain.QuotationStatusSubmitted,
	}

	quoteItems := []*domain.VendorQuotationItem{
		{
			ID:                "qi-1",
			VendorQuotationID: "q-winner",
			ItemID:            "item-1",
			Description:       "Super Bolt",
			OfferedQuantity:   decimal.NewFromInt(100),
			UnitPrice:         decimal.NewFromFloat(10.00),
			TaxRate:           decimal.NewFromFloat(5.00),
			TotalPrice:        decimal.NewFromFloat(1050.00),
		},
	}

	rfqRepo := &mockRFQRepository{getRFQ: rfq}
	quoteRepo := &mockVendorQuotationRepository{getQuotation: quote}
	quoteItemRepo := &mockVendorQuotationItemRepository{items: quoteItems}
	poRepo := &mockPurchaseOrderRepository{}
	poItemRepo := &mockPurchaseOrderItemRepository{}
	docNumRepo := &mockDocumentNumberRepository{nextNumber: 99}

	txManager := &mockRFQTransactionManager{
		rfqRepo:       rfqRepo,
		quotationRepo: quoteRepo,
		poRepo:        poRepo,
		poItemRepo:    poItemRepo,
		docNumRepo:    docNumRepo,
	}

	service := NewRFQService(
		rfqRepo,
		nil,
		quoteRepo,
		quoteItemRepo,
		poRepo,
		poItemRepo,
		nil,
		nil,
		docNumRepo,
		txManager,
	)

	result, err := service.AwardVendorQuotation(context.Background(), "rfq-1", "q-winner")
	if err != nil {
		t.Fatalf("unexpected error awarding quotation: %v", err)
	}

	if result.AwardedQuotationID != "q-winner" {
		t.Errorf("expected awarded quotation q-winner, got %s", result.AwardedQuotationID)
	}
	if result.PONumber != "PO-2026-000099" {
		t.Errorf("expected PO number PO-2026-000099, got %s", result.PONumber)
	}
	if quote.Status != domain.QuotationStatusAwarded {
		t.Errorf("expected quotation status awarded, got %s", quote.Status)
	}
	if !quoteRepo.rejectedOther {
		t.Errorf("expected competing quotations to be rejected")
	}
	if rfq.Status != domain.RFQStatusClosed {
		t.Errorf("expected rfq status closed, got %s", rfq.Status)
	}
}

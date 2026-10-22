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

const (
	rfqDocumentType         = "rfq"
	rfqPurchaseOrderDocType = "purchase_order"
)

type CreateRFQItemInput struct {
	ItemID             string
	Description        string
	TargetQuantity     decimal.Decimal
	TargetDeliveryDate *time.Time
}

type CreateRFQInput struct {
	Title              string
	PurchaseRequestID  *string
	SubmissionDeadline time.Time
	CreatedBy          string
	Notes              *string
	Items              []CreateRFQItemInput
}

type CreateRFQFromPRInput struct {
	Title              string
	SubmissionDeadline time.Time
	CreatedBy          string
	Notes              *string
}

type UpdateRFQInput struct {
	Title              *string
	SubmissionDeadline *time.Time
	Notes              *string
}

type RFQDetail struct {
	RFQ   *domain.RFQ
	Items []*domain.RFQItem
}

type SubmitVendorQuotationItemInput struct {
	RFQItemID       string
	ItemID          string
	Description     string
	OfferedQuantity decimal.Decimal
	UnitPrice       decimal.Decimal
	TaxRate         decimal.Decimal
	LeadTimeDays    *int
}

type SubmitVendorQuotationInput struct {
	RFQID          string
	VendorID       string
	QuoteReference string
	QuotationDate  *time.Time
	ValidUntil     time.Time
	Currency       string
	PaymentTerms   *string
	LeadTimeDays   int
	Notes          *string
	Items          []SubmitVendorQuotationItemInput
}

type VendorQuotationDetail struct {
	Quotation *domain.VendorQuotation
	Items     []*domain.VendorQuotationItem
}

type QuotationComparisonMatrix struct {
	RFQID                      string
	LowestBidQuotationID       *string
	FastestLeadTimeQuotationID *string
	Quotations                 []*domain.VendorQuotation
}

type AwardQuotationResult struct {
	AwardedQuotationID string
	PurchaseOrderID    string
	PONumber           string
	Status             string
}

type RFQService interface {
	CreateRFQ(
		ctx context.Context,
		input *CreateRFQInput,
	) (*RFQDetail, error)

	CreateRFQFromPR(
		ctx context.Context,
		purchaseRequestID string,
		input *CreateRFQFromPRInput,
	) (*RFQDetail, error)

	GetRFQByID(
		ctx context.Context,
		id string,
	) (*RFQDetail, error)

	GetRFQByNumber(
		ctx context.Context,
		rfqNumber string,
	) (*RFQDetail, error)

	ListRFQs(
		ctx context.Context,
		status *domain.RFQStatus,
		purchaseRequestID *string,
		limit int,
		offset int,
	) ([]*domain.RFQ, int64, error)

	UpdateRFQ(
		ctx context.Context,
		id string,
		input *UpdateRFQInput,
	) (*domain.RFQ, error)

	PublishRFQ(
		ctx context.Context,
		id string,
	) (*domain.RFQ, error)

	CloseRFQ(
		ctx context.Context,
		id string,
	) (*domain.RFQ, error)

	CancelRFQ(
		ctx context.Context,
		id string,
	) (*domain.RFQ, error)

	SubmitVendorQuotation(
		ctx context.Context,
		input *SubmitVendorQuotationInput,
	) (*VendorQuotationDetail, error)

	GetVendorQuotationByID(
		ctx context.Context,
		id string,
	) (*VendorQuotationDetail, error)

	ListVendorQuotationsByRFQ(
		ctx context.Context,
		rfqID string,
	) ([]*domain.VendorQuotation, error)

	CompareRFQQuotations(
		ctx context.Context,
		rfqID string,
	) (*QuotationComparisonMatrix, error)

	AwardVendorQuotation(
		ctx context.Context,
		rfqID string,
		quotationID string,
	) (*AwardQuotationResult, error)
}

type rfqService struct {
	rfqRepo       repository.RFQRepository
	rfqItemRepo   repository.RFQItemRepository
	quotationRepo repository.VendorQuotationRepository
	quoteItemRepo repository.VendorQuotationItemRepository
	poRepo        repository.PurchaseOrderRepository
	poItemRepo    repository.PurchaseOrderItemRepository
	prRepo        repository.PurchaseRequestRepository
	prItemRepo    repository.PurchaseRequestItemRepository
	docNumRepo    repository.DocumentNumberRepository
	txManager     repository.RFQTransactionManager
}

var _ RFQService = (*rfqService)(nil)

func NewRFQService(
	rfqRepo repository.RFQRepository,
	rfqItemRepo repository.RFQItemRepository,
	quotationRepo repository.VendorQuotationRepository,
	quoteItemRepo repository.VendorQuotationItemRepository,
	poRepo repository.PurchaseOrderRepository,
	poItemRepo repository.PurchaseOrderItemRepository,
	prRepo repository.PurchaseRequestRepository,
	prItemRepo repository.PurchaseRequestItemRepository,
	docNumRepo repository.DocumentNumberRepository,
	txManager repository.RFQTransactionManager,
) RFQService {
	return &rfqService{
		rfqRepo:       rfqRepo,
		rfqItemRepo:   rfqItemRepo,
		quotationRepo: quotationRepo,
		quoteItemRepo: quoteItemRepo,
		poRepo:        poRepo,
		poItemRepo:    poItemRepo,
		prRepo:        prRepo,
		prItemRepo:    prItemRepo,
		docNumRepo:    docNumRepo,
		txManager:     txManager,
	}
}

func (s *rfqService) CreateRFQ(
	ctx context.Context,
	input *CreateRFQInput,
) (*RFQDetail, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.Title) == "" {
		return nil, fmt.Errorf("rfq title is required")
	}

	if strings.TrimSpace(input.CreatedBy) == "" {
		return nil, fmt.Errorf("created by user id is required")
	}

	if input.SubmissionDeadline.IsZero() || input.SubmissionDeadline.Before(time.Now()) {
		return nil, fmt.Errorf("submission deadline must be a future date/time")
	}

	if len(input.Items) == 0 {
		return nil, fmt.Errorf("at least one rfq line item is required")
	}

	for i, item := range input.Items {
		if strings.TrimSpace(item.ItemID) == "" {
			return nil, fmt.Errorf("items[%d]: item id is required", i)
		}
		if strings.TrimSpace(item.Description) == "" {
			return nil, fmt.Errorf("items[%d]: description is required", i)
		}
		if item.TargetQuantity.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("items[%d]: target quantity must be greater than zero", i)
		}
	}

	var createdRFQ *domain.RFQ
	var createdItems []*domain.RFQItem

	err := s.txManager.WithinTransaction(
		ctx,
		func(
			rfqRepo repository.RFQRepository,
			rfqItemRepo repository.RFQItemRepository,
			_ repository.VendorQuotationRepository,
			_ repository.VendorQuotationItemRepository,
			_ repository.PurchaseOrderRepository,
			_ repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
			_ repository.PurchaseRequestRepository,
			_ repository.PurchaseRequestItemRepository,
		) error {
			year := time.Now().Year()
			nextSeq, err := docNumRepo.NextDocumentNumber(ctx, rfqDocumentType, year)
			if err != nil {
				return fmt.Errorf("generate rfq number: %w", err)
			}

			rfqNumber := fmt.Sprintf("RFQ-%d-%06d", year, nextSeq)

			rfq := &domain.RFQ{
				RFQNumber:          rfqNumber,
				PurchaseRequestID:  input.PurchaseRequestID,
				Title:              strings.TrimSpace(input.Title),
				Status:             domain.RFQStatusDraft,
				SubmissionDeadline: input.SubmissionDeadline,
				CreatedBy:          input.CreatedBy,
				Notes:              input.Notes,
			}

			savedRFQ, err := rfqRepo.CreateRFQ(ctx, rfq)
			if err != nil {
				return fmt.Errorf("create rfq: %w", err)
			}

			items := make([]*domain.RFQItem, 0, len(input.Items))
			for _, itemInput := range input.Items {
				item := &domain.RFQItem{
					RFQID:              savedRFQ.ID,
					ItemID:             itemInput.ItemID,
					Description:        strings.TrimSpace(itemInput.Description),
					TargetQuantity:     itemInput.TargetQuantity,
					TargetDeliveryDate: itemInput.TargetDeliveryDate,
				}

				savedItem, err := rfqItemRepo.CreateRFQItem(ctx, item)
				if err != nil {
					return fmt.Errorf("create rfq item: %w", err)
				}
				items = append(items, savedItem)
			}

			createdRFQ = savedRFQ
			createdItems = items
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &RFQDetail{
		RFQ:   createdRFQ,
		Items: createdItems,
	}, nil
}

func (s *rfqService) CreateRFQFromPR(
	ctx context.Context,
	purchaseRequestID string,
	input *CreateRFQFromPRInput,
) (*RFQDetail, error) {
	if strings.TrimSpace(purchaseRequestID) == "" {
		return nil, fmt.Errorf("purchase request id is required")
	}

	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	pr, err := s.prRepo.GetPurchaseRequestByID(ctx, purchaseRequestID)
	if err != nil {
		return nil, fmt.Errorf("get purchase request: %w", err)
	}

	if pr.Status != domain.PurchaseRequestStatusApproved {
		return nil, fmt.Errorf("purchase request must be approved to create an RFQ (current status %q)", pr.Status)
	}

	prItems, err := s.prItemRepo.GetPurchaseRequestItems(ctx, pr.ID)
	if err != nil {
		return nil, fmt.Errorf("get purchase request items: %w", err)
	}

	if len(prItems) == 0 {
		return nil, fmt.Errorf("purchase request has no line items")
	}

	rfqItems := make([]CreateRFQItemInput, 0, len(prItems))
	for _, item := range prItems {
		desc := item.Description
		if strings.TrimSpace(desc) == "" {
			desc = "Item " + item.ItemID
		}

		rfqItems = append(rfqItems, CreateRFQItemInput{
			ItemID:         item.ItemID,
			Description:    desc,
			TargetQuantity: item.Quantity,
		})
	}

	title := input.Title
	if strings.TrimSpace(title) == "" {
		title = fmt.Sprintf("Sourcing for PR %s", pr.RequestNumber)
	}

	return s.CreateRFQ(ctx, &CreateRFQInput{
		Title:              title,
		PurchaseRequestID:  &pr.ID,
		SubmissionDeadline: input.SubmissionDeadline,
		CreatedBy:          input.CreatedBy,
		Notes:              input.Notes,
		Items:              rfqItems,
	})
}

func (s *rfqService) GetRFQByID(
	ctx context.Context,
	id string,
) (*RFQDetail, error) {
	rfq, err := s.rfqRepo.GetRFQByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.rfqItemRepo.GetRFQItems(ctx, rfq.ID)
	if err != nil {
		return nil, fmt.Errorf("get rfq items: %w", err)
	}

	return &RFQDetail{
		RFQ:   rfq,
		Items: items,
	}, nil
}

func (s *rfqService) GetRFQByNumber(
	ctx context.Context,
	rfqNumber string,
) (*RFQDetail, error) {
	rfq, err := s.rfqRepo.GetRFQByNumber(ctx, rfqNumber)
	if err != nil {
		return nil, err
	}

	items, err := s.rfqItemRepo.GetRFQItems(ctx, rfq.ID)
	if err != nil {
		return nil, fmt.Errorf("get rfq items: %w", err)
	}

	return &RFQDetail{
		RFQ:   rfq,
		Items: items,
	}, nil
}

func (s *rfqService) ListRFQs(
	ctx context.Context,
	status *domain.RFQStatus,
	purchaseRequestID *string,
	limit int,
	offset int,
) ([]*domain.RFQ, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rfqs, err := s.rfqRepo.ListRFQs(ctx, status, purchaseRequestID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list rfqs: %w", err)
	}

	total, err := s.rfqRepo.CountRFQs(ctx, status, purchaseRequestID)
	if err != nil {
		return nil, 0, fmt.Errorf("count rfqs: %w", err)
	}

	return rfqs, total, nil
}

func (s *rfqService) UpdateRFQ(
	ctx context.Context,
	id string,
	input *UpdateRFQInput,
) (*domain.RFQ, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	rfq, err := s.rfqRepo.GetRFQByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rfq.Status != domain.RFQStatusDraft {
		return nil, domain.ErrRFQNotEditable
	}

	if input.Title != nil && strings.TrimSpace(*input.Title) != "" {
		rfq.Title = strings.TrimSpace(*input.Title)
	}

	if input.SubmissionDeadline != nil {
		if input.SubmissionDeadline.Before(time.Now()) {
			return nil, fmt.Errorf("submission deadline must be a future date/time")
		}
		rfq.SubmissionDeadline = *input.SubmissionDeadline
	}

	if input.Notes != nil {
		rfq.Notes = input.Notes
	}

	return s.rfqRepo.UpdateRFQ(ctx, rfq)
}

func (s *rfqService) PublishRFQ(
	ctx context.Context,
	id string,
) (*domain.RFQ, error) {
	rfq, err := s.rfqRepo.GetRFQByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rfq.Status != domain.RFQStatusDraft {
		return nil, domain.ErrRFQCannotPublish
	}

	if time.Now().After(rfq.SubmissionDeadline) {
		return nil, domain.ErrRFQDeadlinePassed
	}

	return s.rfqRepo.UpdateRFQStatus(ctx, id, domain.RFQStatusPublished)
}

func (s *rfqService) CloseRFQ(
	ctx context.Context,
	id string,
) (*domain.RFQ, error) {
	rfq, err := s.rfqRepo.GetRFQByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if rfq.Status != domain.RFQStatusPublished && rfq.Status != domain.RFQStatusDraft {
		return nil, domain.ErrRFQCannotClose
	}

	return s.rfqRepo.UpdateRFQStatus(ctx, id, domain.RFQStatusClosed)
}

func (s *rfqService) CancelRFQ(
	ctx context.Context,
	id string,
) (*domain.RFQ, error) {
	rfq, err := s.rfqRepo.GetRFQByID(ctx, id)
	if err != nil {
		return nil, err
	}

	switch rfq.Status {
	case domain.RFQStatusDraft, domain.RFQStatusPublished:
		return s.rfqRepo.UpdateRFQStatus(ctx, id, domain.RFQStatusCancelled)
	default:
		return nil, domain.ErrRFQCannotCancel
	}
}

func (s *rfqService) SubmitVendorQuotation(
	ctx context.Context,
	input *SubmitVendorQuotationInput,
) (*VendorQuotationDetail, error) {
	if input == nil {
		return nil, fmt.Errorf("input cannot be nil")
	}

	if strings.TrimSpace(input.RFQID) == "" {
		return nil, fmt.Errorf("rfq id is required")
	}

	if strings.TrimSpace(input.VendorID) == "" {
		return nil, fmt.Errorf("vendor id is required")
	}

	if strings.TrimSpace(input.QuoteReference) == "" {
		return nil, fmt.Errorf("quote reference is required")
	}

	if input.ValidUntil.IsZero() || input.ValidUntil.Before(time.Now()) {
		return nil, fmt.Errorf("quote valid_until date must be a future date")
	}

	if len(input.Items) == 0 {
		return nil, fmt.Errorf("at least one quotation item is required")
	}

	rfq, err := s.rfqRepo.GetRFQByID(ctx, input.RFQID)
	if err != nil {
		return nil, fmt.Errorf("get rfq: %w", err)
	}

	if rfq.Status != domain.RFQStatusPublished {
		return nil, fmt.Errorf("%w: current status %q", domain.ErrRFQNotAcceptingQuotes, rfq.Status)
	}

	if time.Now().After(rfq.SubmissionDeadline) {
		return nil, domain.ErrRFQDeadlinePassed
	}

	existing, err := s.quotationRepo.GetVendorQuotationByRef(ctx, rfq.ID, input.VendorID, input.QuoteReference)
	if err != nil {
		return nil, fmt.Errorf("check existing quotation: %w", err)
	}
	if existing != nil {
		return nil, domain.ErrDuplicateVendorQuotation
	}

	rfqItems, err := s.rfqItemRepo.GetRFQItems(ctx, rfq.ID)
	if err != nil {
		return nil, fmt.Errorf("get rfq items: %w", err)
	}

	rfqItemMap := make(map[string]*domain.RFQItem, len(rfqItems))
	for _, item := range rfqItems {
		rfqItemMap[item.ID] = item
	}

	subtotal := decimal.Zero
	taxTotal := decimal.Zero
	preparedItems := make([]*domain.VendorQuotationItem, 0, len(input.Items))

	for i, itemInput := range input.Items {
		if strings.TrimSpace(itemInput.RFQItemID) == "" {
			return nil, fmt.Errorf("items[%d]: rfq item id is required", i)
		}

		if _, exists := rfqItemMap[itemInput.RFQItemID]; !exists {
			return nil, fmt.Errorf("items[%d]: rfq item %q does not belong to rfq %q", i, itemInput.RFQItemID, rfq.ID)
		}

		if itemInput.OfferedQuantity.LessThanOrEqual(decimal.Zero) {
			return nil, fmt.Errorf("items[%d]: offered quantity must be greater than zero", i)
		}

		if itemInput.UnitPrice.LessThan(decimal.Zero) {
			return nil, fmt.Errorf("items[%d]: unit price cannot be negative", i)
		}

		lineSubtotal := itemInput.OfferedQuantity.Mul(itemInput.UnitPrice)
		taxRate := itemInput.TaxRate
		if taxRate.LessThan(decimal.Zero) {
			taxRate = decimal.Zero
		}

		lineTax := lineSubtotal.Mul(taxRate).Div(decimal.NewFromInt(100))
		lineTotal := lineSubtotal.Add(lineTax)

		subtotal = subtotal.Add(lineSubtotal)
		taxTotal = taxTotal.Add(lineTax)

		preparedItems = append(preparedItems, &domain.VendorQuotationItem{
			RFQItemID:       itemInput.RFQItemID,
			ItemID:          itemInput.ItemID,
			Description:     itemInput.Description,
			OfferedQuantity: itemInput.OfferedQuantity,
			UnitPrice:       itemInput.UnitPrice,
			TaxRate:         taxRate,
			TotalPrice:      lineTotal,
			LeadTimeDays:    itemInput.LeadTimeDays,
		})
	}

	totalAmount := subtotal.Add(taxTotal)

	quoteDate := time.Now()
	if input.QuotationDate != nil && !input.QuotationDate.IsZero() {
		quoteDate = *input.QuotationDate
	}

	currency := input.Currency
	if strings.TrimSpace(currency) == "" {
		currency = "USD"
	}

	var createdQuotation *domain.VendorQuotation
	var createdItems []*domain.VendorQuotationItem

	err = s.txManager.WithinTransaction(
		ctx,
		func(
			_ repository.RFQRepository,
			_ repository.RFQItemRepository,
			quotationRepo repository.VendorQuotationRepository,
			quoteItemRepo repository.VendorQuotationItemRepository,
			_ repository.PurchaseOrderRepository,
			_ repository.PurchaseOrderItemRepository,
			_ repository.DocumentNumberRepository,
			_ repository.PurchaseRequestRepository,
			_ repository.PurchaseRequestItemRepository,
		) error {
			q := &domain.VendorQuotation{
				RFQID:          rfq.ID,
				VendorID:       input.VendorID,
				QuoteReference: strings.TrimSpace(input.QuoteReference),
				QuotationDate:  quoteDate,
				ValidUntil:     input.ValidUntil,
				Currency:       currency,
				Subtotal:       subtotal,
				TaxAmount:      taxTotal,
				TotalAmount:    totalAmount,
				PaymentTerms:   input.PaymentTerms,
				LeadTimeDays:   input.LeadTimeDays,
				Status:         domain.QuotationStatusSubmitted,
				Notes:          input.Notes,
			}

			savedQuotation, err := quotationRepo.CreateVendorQuotation(ctx, q)
			if err != nil {
				return fmt.Errorf("create vendor quotation: %w", err)
			}

			savedItems := make([]*domain.VendorQuotationItem, 0, len(preparedItems))
			for _, item := range preparedItems {
				item.VendorQuotationID = savedQuotation.ID
				savedItem, err := quoteItemRepo.CreateVendorQuotationItem(ctx, item)
				if err != nil {
					return fmt.Errorf("create quotation item: %w", err)
				}
				savedItems = append(savedItems, savedItem)
			}

			createdQuotation = savedQuotation
			createdItems = savedItems
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &VendorQuotationDetail{
		Quotation: createdQuotation,
		Items:     createdItems,
	}, nil
}

func (s *rfqService) GetVendorQuotationByID(
	ctx context.Context,
	id string,
) (*VendorQuotationDetail, error) {
	q, err := s.quotationRepo.GetVendorQuotationByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.quoteItemRepo.GetVendorQuotationItems(ctx, q.ID)
	if err != nil {
		return nil, fmt.Errorf("get quotation items: %w", err)
	}

	return &VendorQuotationDetail{
		Quotation: q,
		Items:     items,
	}, nil
}

func (s *rfqService) ListVendorQuotationsByRFQ(
	ctx context.Context,
	rfqID string,
) ([]*domain.VendorQuotation, error) {
	return s.quotationRepo.ListVendorQuotationsByRFQ(ctx, rfqID)
}

func (s *rfqService) CompareRFQQuotations(
	ctx context.Context,
	rfqID string,
) (*QuotationComparisonMatrix, error) {
	quotations, err := s.quotationRepo.ListVendorQuotationsByRFQ(ctx, rfqID)
	if err != nil {
		return nil, fmt.Errorf("list quotations: %w", err)
	}

	matrix := &QuotationComparisonMatrix{
		RFQID:      rfqID,
		Quotations: quotations,
	}

	if len(quotations) == 0 {
		return matrix, nil
	}

	var lowestBidID string
	var lowestAmount decimal.Decimal

	var fastestLeadID string
	var minLeadTime int = -1

	for i, q := range quotations {
		if i == 0 || q.TotalAmount.LessThan(lowestAmount) {
			lowestAmount = q.TotalAmount
			lowestBidID = q.ID
		}

		if minLeadTime == -1 || q.LeadTimeDays < minLeadTime {
			minLeadTime = q.LeadTimeDays
			fastestLeadID = q.ID
		}
	}

	if lowestBidID != "" {
		matrix.LowestBidQuotationID = &lowestBidID
	}
	if fastestLeadID != "" {
		matrix.FastestLeadTimeQuotationID = &fastestLeadID
	}

	return matrix, nil
}

func (s *rfqService) AwardVendorQuotation(
	ctx context.Context,
	rfqID string,
	quotationID string,
) (*AwardQuotationResult, error) {
	rfq, err := s.rfqRepo.GetRFQByID(ctx, rfqID)
	if err != nil {
		return nil, fmt.Errorf("get rfq: %w", err)
	}

	if rfq.Status == domain.RFQStatusCancelled {
		return nil, fmt.Errorf("cannot award a cancelled RFQ")
	}

	quoteDetail, err := s.GetVendorQuotationByID(ctx, quotationID)
	if err != nil {
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if quoteDetail.Quotation.RFQID != rfq.ID {
		return nil, fmt.Errorf("quotation %q does not belong to rfq %q", quotationID, rfqID)
	}

	if quoteDetail.Quotation.Status != domain.QuotationStatusSubmitted {
		return nil, fmt.Errorf("%w: quotation status is %q", domain.ErrQuotationCannotAward, quoteDetail.Quotation.Status)
	}

	var createdPO *domain.PurchaseOrder

	err = s.txManager.WithinTransaction(
		ctx,
		func(
			rfqRepo repository.RFQRepository,
			_ repository.RFQItemRepository,
			quotationRepo repository.VendorQuotationRepository,
			_ repository.VendorQuotationItemRepository,
			poRepo repository.PurchaseOrderRepository,
			poItemRepo repository.PurchaseOrderItemRepository,
			docNumRepo repository.DocumentNumberRepository,
			_ repository.PurchaseRequestRepository,
			_ repository.PurchaseRequestItemRepository,
		) error {
			if _, err := quotationRepo.UpdateVendorQuotationStatus(ctx, quotationID, domain.QuotationStatusAwarded); err != nil {
				return fmt.Errorf("update quotation status to awarded: %w", err)
			}

			if err := quotationRepo.RejectOtherVendorQuotations(ctx, rfq.ID, quotationID); err != nil {
				return fmt.Errorf("reject other quotations: %w", err)
			}

			if _, err := rfqRepo.UpdateRFQStatus(ctx, rfq.ID, domain.RFQStatusClosed); err != nil {
				return fmt.Errorf("close rfq: %w", err)
			}

			year := time.Now().Year()
			nextSeq, err := docNumRepo.NextDocumentNumber(ctx, rfqPurchaseOrderDocType, year)
			if err != nil {
				return fmt.Errorf("generate po number: %w", err)
			}

			poNumber := fmt.Sprintf("PO-%d-%06d", year, nextSeq)

			poNotes := fmt.Sprintf("Awarded from RFQ %s (Quote ref: %s)", rfq.RFQNumber, quoteDetail.Quotation.QuoteReference)
			if quoteDetail.Quotation.Notes != nil {
				poNotes = fmt.Sprintf("%s. %s", poNotes, *quoteDetail.Quotation.Notes)
			}

			deliveryDate := time.Now().AddDate(0, 0, quoteDetail.Quotation.LeadTimeDays)

			po := &domain.PurchaseOrder{
				PONumber:             poNumber,
				PurchaseRequestID:    rfq.PurchaseRequestID,
				VendorID:             quoteDetail.Quotation.VendorID,
				Status:               domain.PurchaseOrderStatusDraft,
				OrderDate:            time.Now(),
				ExpectedDeliveryDate: &deliveryDate,
				Currency:             quoteDetail.Quotation.Currency,
				Subtotal:             quoteDetail.Quotation.Subtotal,
				TaxAmount:            quoteDetail.Quotation.TaxAmount,
				TotalAmount:          quoteDetail.Quotation.TotalAmount,
				PaymentTerms:         quoteDetail.Quotation.PaymentTerms,
				Notes:                &poNotes,
			}

			savedPO, err := poRepo.CreatePurchaseOrder(ctx, po)
			if err != nil {
				return fmt.Errorf("create purchase order: %w", err)
			}

			for _, qItem := range quoteDetail.Items {
				poItem := &domain.PurchaseOrderItem{
					PurchaseOrderID:  savedPO.ID,
					ItemID:           qItem.ItemID,
					Description:      qItem.Description,
					Quantity:         qItem.OfferedQuantity,
					UnitPrice:        qItem.UnitPrice,
					TaxRate:          qItem.TaxRate,
					TotalPrice:       qItem.TotalPrice,
					ReceivedQuantity: decimal.Zero,
					BilledQuantity:   decimal.Zero,
				}

				if _, err := poItemRepo.CreatePurchaseOrderItem(ctx, poItem); err != nil {
					return fmt.Errorf("create po item: %w", err)
				}
			}

			createdPO = savedPO
			return nil
		},
	)
	if err != nil {
		return nil, err
	}

	return &AwardQuotationResult{
		AwardedQuotationID: quotationID,
		PurchaseOrderID:    createdPO.ID,
		PONumber:           createdPO.PONumber,
		Status:             string(createdPO.Status),
	}, nil
}

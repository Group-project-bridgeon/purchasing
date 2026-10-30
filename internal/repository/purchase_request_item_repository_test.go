package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

func TestPostgresPurchaseRequestItemRepository_CreatePurchaseRequestItem(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-ITEM-TEST-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Item repository test",
	}

	createdRequest, err := requestRepo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	price := decimal.NewFromFloat(1250.50)

	item := &domain.PurchaseRequestItem{
		PurchaseRequestID:  createdRequest.ID,
		ItemID:             "33333333-3333-3333-3333-333333333333",
		Description:        "Test laptop",
		Quantity:           decimal.NewFromFloat(2.500),
		EstimatedUnitPrice: &price,
	}

	result, err := itemRepo.CreatePurchaseRequestItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create purchase request item: %v", err)
	}

	if result == nil {
		t.Fatal("expected item, got nil")
	}

	if result.ID == "" {
		t.Fatal("expected generated item ID, got empty ID")
	}

	if result.PurchaseRequestID != createdRequest.ID {
		t.Fatalf(
			"expected request ID %q, got %q",
			createdRequest.ID,
			result.PurchaseRequestID,
		)
	}

	if result.ItemID != item.ItemID {
		t.Fatalf(
			"expected item ID %q, got %q",
			item.ItemID,
			result.ItemID,
		)
	}

	if !result.Quantity.Equal(item.Quantity) {
		t.Fatalf(
			"expected quantity %s, got %s",
			item.Quantity,
			result.Quantity,
		)
	}

	if result.EstimatedUnitPrice == nil {
		t.Fatal("expected estimated unit price, got nil")
	}

	if !result.EstimatedUnitPrice.Equal(*item.EstimatedUnitPrice) {
		t.Fatalf(
			"expected price %s, got %s",
			*item.EstimatedUnitPrice,
			*result.EstimatedUnitPrice,
		)
	}
}

func TestPostgresPurchaseRequestItemRepository_GetPurchaseRequestItemByID(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-ITEM-TEST-002",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Get item by ID test",
	}

	createdRequest, err := requestRepo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	item := &domain.PurchaseRequestItem{
		PurchaseRequestID: createdRequest.ID,
		ItemID:            "33333333-3333-3333-3333-333333333333",
		Description:       "Test monitor",
		Quantity:          decimal.NewFromInt(2),
	}

	createdItem, err := itemRepo.CreatePurchaseRequestItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	result, err := itemRepo.GetPurchaseRequestItemByID(
		ctx,
		createdItem.ID,
	)
	if err != nil {
		t.Fatalf("failed to get item: %v", err)
	}

	if result == nil {
		t.Fatal("expected item, got nil")
	}

	if result.ID != createdItem.ID {
		t.Fatalf(
			"expected ID %q, got %q",
			createdItem.ID,
			result.ID,
		)
	}

	if result.Description != createdItem.Description {
		t.Fatalf(
			"expected description %q, got %q",
			createdItem.Description,
			result.Description,
		)
	}
}

func TestPostgresPurchaseRequestItemRepository_GetPurchaseRequestItems(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-ITEM-TEST-003",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "List items test",
	}

	createdRequest, err := requestRepo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	for i := 1; i <= 3; i++ {
		item := &domain.PurchaseRequestItem{
			PurchaseRequestID: createdRequest.ID,
			ItemID:            "33333333-3333-3333-3333-333333333333",
			Description:       "Test item",
			Quantity:          decimal.NewFromInt(int64(i)),
		}

		if _, err := itemRepo.CreatePurchaseRequestItem(ctx, item); err != nil {
			t.Fatalf(
				"failed to create item %d: %v",
				i,
				err,
			)
		}
	}

	result, err := itemRepo.GetPurchaseRequestItems(
		ctx,
		createdRequest.ID,
	)
	if err != nil {
		t.Fatalf("failed to get request items: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf(
			"expected 3 items, got %d",
			len(result),
		)
	}

	for i, item := range result {
		if item.PurchaseRequestID != createdRequest.ID {
			t.Fatalf(
				"item %d has wrong request ID: %q",
				i+1,
				item.PurchaseRequestID,
			)
		}
	}
}

func TestPostgresPurchaseRequestItemRepository_UpdatePurchaseRequestItem(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-ITEM-UPDATE-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Update item test",
	}

	createdRequest, err := requestRepo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	item := &domain.PurchaseRequestItem{
		PurchaseRequestID: createdRequest.ID,
		ItemID:            "33333333-3333-3333-3333-333333333333",
		Description:       "Original item",
		Quantity:          decimal.NewFromInt(1),
	}

	createdItem, err := itemRepo.CreatePurchaseRequestItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	createdItem.Description = "Updated item"
	createdItem.Quantity = decimal.NewFromInt(5)

	result, err := itemRepo.UpdatePurchaseRequestItem(ctx, createdItem)
	if err != nil {
		t.Fatalf("failed to update item: %v", err)
	}

	if result.Description != "Updated item" {
		t.Fatalf(
			"expected description %q, got %q",
			"Updated item",
			result.Description,
		)
	}

	if !result.Quantity.Equal(decimal.NewFromInt(5)) {
		t.Fatalf(
			"expected quantity 5, got %s",
			result.Quantity,
		)
	}
}

func TestPostgresPurchaseRequestItemRepository_DeletePurchaseRequestItem(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-ITEM-DELETE-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Delete item test",
	}

	createdRequest, err := requestRepo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	item := &domain.PurchaseRequestItem{
		PurchaseRequestID: createdRequest.ID,
		ItemID:            "33333333-3333-3333-3333-333333333333",
		Description:       "Item to delete",
		Quantity:          decimal.NewFromInt(1),
	}

	createdItem, err := itemRepo.CreatePurchaseRequestItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if err := itemRepo.DeletePurchaseRequestItem(ctx, createdItem.ID); err != nil {
		t.Fatalf("failed to delete item: %v", err)
	}

	_, err = itemRepo.GetPurchaseRequestItemByID(ctx, createdItem.ID)
	if err == nil {
		t.Fatal("expected item to be deleted")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf(
			"expected pgx.ErrNoRows, got %v",
			err,
		)
	}
}

func TestPostgresPurchaseRequestItemRepository_CannotUpdateSubmittedRequestItem(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-ITEM-LOCK-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Submitted item test",
	}

	createdRequest, err := requestRepo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	item := &domain.PurchaseRequestItem{
		PurchaseRequestID: createdRequest.ID,
		ItemID:            "33333333-3333-3333-3333-333333333333",
		Description:       "Locked item",
		Quantity:          decimal.NewFromInt(1),
	}

	createdItem, err := itemRepo.CreatePurchaseRequestItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if _, err := requestRepo.SubmitPurchaseRequest(ctx, createdRequest.ID); err != nil {
		t.Fatalf("failed to submit request: %v", err)
	}

	createdItem.Description = "Should not update"

	_, err = itemRepo.UpdatePurchaseRequestItem(ctx, createdItem)
	if err == nil {
		t.Fatal("expected error when updating item of submitted request")
	}

	if !errors.Is(err, domain.ErrPurchaseRequestNotEditable) {
		t.Fatal("expected explicit non-draft lifecycle rejection")
	}
}

func TestPostgresPurchaseRequestItemRepository_CannotDeleteSubmittedRequestItem(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-ITEM-LOCK-002",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Submitted delete test",
	}

	createdRequest, err := requestRepo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	item := &domain.PurchaseRequestItem{
		PurchaseRequestID: createdRequest.ID,
		ItemID:            "33333333-3333-3333-3333-333333333333",
		Description:       "Locked delete item",
		Quantity:          decimal.NewFromInt(1),
	}

	createdItem, err := itemRepo.CreatePurchaseRequestItem(ctx, item)
	if err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	if _, err := requestRepo.SubmitPurchaseRequest(ctx, createdRequest.ID); err != nil {
		t.Fatalf("failed to submit request: %v", err)
	}

	err = itemRepo.DeletePurchaseRequestItem(ctx, createdItem.ID)
	if !errors.Is(err, domain.ErrPurchaseRequestNotEditable) {
		t.Fatal("expected explicit non-draft delete rejection")
	}

	// A rejected delete must preserve its item.
	result, err := itemRepo.GetPurchaseRequestItemByID(
		ctx,
		createdItem.ID,
	)
	if err != nil {
		t.Fatalf("expected item to still exist: %v", err)
	}

	if result.ID != createdItem.ID {
		t.Fatalf(
			"expected item ID %q, got %q",
			createdItem.ID,
			result.ID,
		)
	}
}

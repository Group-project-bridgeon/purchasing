package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
)

func TestPostgresTransactionManager_Commit(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	transactionManager := NewPostgresTransactionManager(db)

	ctx := context.Background()

	requestID := ""

	err := transactionManager.WithinTransaction(
		ctx,
		func(
			requestRepository PurchaseRequestRepository,
			itemRepository PurchaseRequestItemRepository,
			_ DocumentNumberRepository,
		) error {
			request := &domain.PurchaseRequest{
				RequestNumber: "PR-TX-COMMIT-001",
				RequestedBy:   "11111111-1111-1111-1111-111111111111",
				Purpose:       "Transaction commit test",
			}

			createdRequest, err := requestRepository.CreatePurchaseRequest(
				ctx,
				request,
			)
			if err != nil {
				return err
			}

			requestID = createdRequest.ID

			item := &domain.PurchaseRequestItem{
				PurchaseRequestID: createdRequest.ID,
				ItemID:            "33333333-3333-3333-3333-333333333333",
				Description:       "Transaction test item",
				Quantity:          decimal.NewFromInt(2),
			}

			_, err = itemRepository.CreatePurchaseRequestItem(
				ctx,
				item,
			)

			return err
		},
	)

	if err != nil {
		t.Fatalf("expected transaction to commit, got error: %v", err)
	}

	requestRepo := NewPostgresPurchaseRequestRepository(db)
	itemRepo := NewPostgresPurchaseRequestItemRepository(db)

	request, err := requestRepo.GetPurchaseRequestByID(
		ctx,
		requestID,
	)
	if err != nil {
		t.Fatalf("expected committed request to exist: %v", err)
	}

	if request.RequestNumber != "PR-TX-COMMIT-001" {
		t.Fatalf(
			"unexpected request number %q",
			request.RequestNumber,
		)
	}

	items, err := itemRepo.GetPurchaseRequestItems(
		ctx,
		requestID,
	)
	if err != nil {
		t.Fatalf("failed to get committed items: %v", err)
	}

	if len(items) != 1 {
		t.Fatalf(
			"expected 1 committed item, got %d",
			len(items),
		)
	}
}

func TestPostgresTransactionManager_Rollback(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	transactionManager := NewPostgresTransactionManager(db)

	ctx := context.Background()

	const requestNumber = "PR-TX-ROLLBACK-001"

	err := transactionManager.WithinTransaction(
		ctx,
		func(
			requestRepository PurchaseRequestRepository,
			itemRepository PurchaseRequestItemRepository,
			_ DocumentNumberRepository,
		) error {
			request := &domain.PurchaseRequest{
				RequestNumber: requestNumber,
				RequestedBy:   "11111111-1111-1111-1111-111111111111",
				Purpose:       "Transaction rollback test",
			}

			createdRequest, err := requestRepository.CreatePurchaseRequest(
				ctx,
				request,
			)
			if err != nil {
				return err
			}

			item := &domain.PurchaseRequestItem{
				PurchaseRequestID: createdRequest.ID,
				ItemID:            "33333333-3333-3333-3333-333333333333",
				Description:       "This should be rolled back",
				Quantity:          decimal.NewFromInt(1),
			}

			if _, err := itemRepository.CreatePurchaseRequestItem(
				ctx,
				item,
			); err != nil {
				return err
			}

			return pgx.ErrTxClosed
		},
	)

	if err == nil {
		t.Fatal("expected transaction to return an error")
	}

	requestRepo := NewPostgresPurchaseRequestRepository(db)

	_, err = requestRepo.GetPurchaseRequestByNumber(
		ctx,
		requestNumber,
	)

	if err == nil {
		t.Fatal("expected rolled-back request to not exist")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf(
			"expected pgx.ErrNoRows after rollback, got %v",
			err,
		)
	}
}
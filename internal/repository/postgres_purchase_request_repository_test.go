package repository

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testDatabaseURL = "postgres://postgres:postgres@localhost:5433/purchasing?sslmode=disable"

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	db, err := pgxpool.New(ctx, testDatabaseURL)
	if err != nil {
		t.Fatalf("failed to create database pool: %v", err)
	}

	if err := db.Ping(ctx); err != nil {
		db.Close()
		t.Fatalf("failed to connect to database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func cleanupPurchaseRequests(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"DELETE FROM purchase_requests",
	)
	if err != nil {
		t.Fatalf("failed to clean purchase requests: %v", err)
	}
}

func TestPostgresPurchaseRequestRepository_CreatePurchaseRequest(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)

	ctx := context.Background()

	requestedBy := "11111111-1111-1111-1111-111111111111"
	departmentID := "22222222-2222-2222-2222-222222222222"

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-000001",
		RequestedBy:   requestedBy,
		DepartmentID:  &departmentID,
		Status:        domain.PurchaseRequestStatusDraft,
		RequiredBy: func() *time.Time {
			value := time.Now().AddDate(0, 0, 7)
			return &value
		}(),
		Purpose: "Repository integration test",
	}

	result, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	if result == nil {
		t.Fatal("expected purchase request, got nil")
	}

	if result.ID == "" {
		t.Fatal("expected generated ID, got empty ID")
	}

	if result.RequestNumber != request.RequestNumber {
		t.Fatalf(
			"expected request number %q, got %q",
			request.RequestNumber,
			result.RequestNumber,
		)
	}

	if result.RequestedBy != requestedBy {
		t.Fatalf(
			"expected requested by %q, got %q",
			requestedBy,
			result.RequestedBy,
		)
	}

	if result.Status != domain.PurchaseRequestStatusDraft {
		t.Fatalf(
			"expected status %q, got %q",
			domain.PurchaseRequestStatusDraft,
			result.Status,
		)
	}

	if result.Purpose != request.Purpose {
		t.Fatalf(
			"expected purpose %q, got %q",
			request.Purpose,
			result.Purpose,
		)
	}
}

func TestPostgresPurchaseRequestRepository_GetPurchaseRequestByID(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)

	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-000002",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Get by ID integration test",
	}

	created, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	result, err := repo.GetPurchaseRequestByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get purchase request: %v", err)
	}

	if result == nil {
		t.Fatal("expected purchase request, got nil")
	}

	if result.ID != created.ID {
		t.Fatalf(
			"expected ID %q, got %q",
			created.ID,
			result.ID,
		)
	}

	if result.RequestNumber != created.RequestNumber {
		t.Fatalf(
			"expected request number %q, got %q",
			created.RequestNumber,
			result.RequestNumber,
		)
	}
}

func TestPostgresPurchaseRequestRepository_GetPurchaseRequestByNumber(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-000003",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Get by number integration test",
	}

	created, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	result, err := repo.GetPurchaseRequestByNumber(
		ctx,
		created.RequestNumber,
	)
	if err != nil {
		t.Fatalf("failed to get purchase request by number: %v", err)
	}

	if result == nil {
		t.Fatal("expected purchase request, got nil")
	}

	if result.ID != created.ID {
		t.Fatalf(
			"expected ID %q, got %q",
			created.ID,
			result.ID,
		)
	}
}

func TestPostgresPurchaseRequestRepository_ListPurchaseRequests(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		request := &domain.PurchaseRequest{
			RequestNumber: fmt.Sprintf("PR-TEST-LIST-%03d", i),
			RequestedBy:   "11111111-1111-1111-1111-111111111111",
			Purpose:       "List integration test",
		}

		if _, err := repo.CreatePurchaseRequest(ctx, request); err != nil {
			t.Fatalf(
				"failed to create test request %d: %v",
				i,
				err,
			)
		}
	}

	result, err := repo.ListPurchaseRequests(ctx, 10, 0)
	if err != nil {
		t.Fatalf("failed to list purchase requests: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf(
			"expected 3 purchase requests, got %d",
			len(result),
		)
	}
}

func TestPostgresPurchaseRequestRepository_ListPurchaseRequestsByStatus(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		request := &domain.PurchaseRequest{
			RequestNumber: fmt.Sprintf("PR-TEST-STATUS-%03d", i),
			RequestedBy:   "11111111-1111-1111-1111-111111111111",
			Purpose:       "Status list integration test",
		}

		created, err := repo.CreatePurchaseRequest(ctx, request)
		if err != nil {
			t.Fatalf(
				"failed to create test request %d: %v",
				i,
				err,
			)
		}

		if i == 1 {
			if _, err := repo.SubmitPurchaseRequest(ctx, created.ID); err != nil {
				t.Fatalf("failed to submit request: %v", err)
			}
		}
	}

	result, err := repo.ListPurchaseRequestsByStatus(
		ctx,
		domain.PurchaseRequestStatusDraft,
		10,
		0,
	)
	if err != nil {
		t.Fatalf(
			"failed to list purchase requests by status: %v",
			err,
		)
	}

	if len(result) != 2 {
		t.Fatalf(
			"expected 2 draft requests, got %d",
			len(result),
		)
	}

	for _, request := range result {
		if request.Status != domain.PurchaseRequestStatusDraft {
			t.Fatalf(
				"expected draft status, got %q",
				request.Status,
			)
		}
	}
}

func TestPostgresPurchaseRequestRepository_UpdatePurchaseRequest(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-UPDATE-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Original purpose",
	}

	created, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	updatedPurpose := "Updated purpose"
	updatedNotes := "Updated notes"

	created.Purpose = updatedPurpose
	created.Notes = &updatedNotes

	result, err := repo.UpdatePurchaseRequest(ctx, created)
	if err != nil {
		t.Fatalf("failed to update purchase request: %v", err)
	}

	if result.Purpose != updatedPurpose {
		t.Fatalf(
			"expected purpose %q, got %q",
			updatedPurpose,
			result.Purpose,
		)
	}

	if result.Notes == nil {
		t.Fatal("expected notes, got nil")
	}

	if *result.Notes != updatedNotes {
		t.Fatalf(
			"expected notes %q, got %q",
			updatedNotes,
			*result.Notes,
		)
	}

	if result.Status != domain.PurchaseRequestStatusDraft {
		t.Fatalf(
			"expected status draft, got %q",
			result.Status,
		)
	}
}

func TestPostgresPurchaseRequestRepository_SubmitPurchaseRequest(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-SUBMIT-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Submit integration test",
	}

	created, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	result, err := repo.SubmitPurchaseRequest(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to submit purchase request: %v", err)
	}

	if result.Status != domain.PurchaseRequestStatusSubmitted {
		t.Fatalf(
			"expected status submitted, got %q",
			result.Status,
		)
	}
}

func TestPostgresPurchaseRequestRepository_ApprovePurchaseRequest(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-APPROVE-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Approve integration test",
	}

	created, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	if _, err := repo.SubmitPurchaseRequest(ctx, created.ID); err != nil {
		t.Fatalf("failed to submit purchase request: %v", err)
	}

	result, err := repo.ApprovePurchaseRequest(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to approve purchase request: %v", err)
	}

	if result.Status != domain.PurchaseRequestStatusApproved {
		t.Fatalf(
			"expected status approved, got %q",
			result.Status,
		)
	}
}

func TestPostgresPurchaseRequestRepository_RejectPurchaseRequest(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-REJECT-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Reject integration test",
	}

	created, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	if _, err := repo.SubmitPurchaseRequest(ctx, created.ID); err != nil {
		t.Fatalf("failed to submit purchase request: %v", err)
	}

	result, err := repo.RejectPurchaseRequest(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to reject purchase request: %v", err)
	}

	if result.Status != domain.PurchaseRequestStatusRejected {
		t.Fatalf(
			"expected status rejected, got %q",
			result.Status,
		)
	}
}

func TestPostgresPurchaseRequestRepository_CannotApproveDraftRequest(t *testing.T) {
	db := setupTestDB(t)
	cleanupPurchaseRequests(t, db)

	repo := NewPostgresPurchaseRequestRepository(db)
	ctx := context.Background()

	request := &domain.PurchaseRequest{
		RequestNumber: "PR-TEST-INVALID-001",
		RequestedBy:   "11111111-1111-1111-1111-111111111111",
		Purpose:       "Invalid transition test",
	}

	created, err := repo.CreatePurchaseRequest(ctx, request)
	if err != nil {
		t.Fatalf("failed to create purchase request: %v", err)
	}

	_, err = repo.ApprovePurchaseRequest(ctx, created.ID)
	if err == nil {
		t.Fatal("expected error when approving draft request")
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		t.Fatalf(
			"expected pgx.ErrNoRows, got %v",
			err,
		)
	}
}

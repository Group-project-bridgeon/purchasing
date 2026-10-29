package repository

import (
	"context"
	"testing"
)

func TestPostgresDocumentNumberRepository_NextDocumentNumber(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()

	// Clean the counter table before the test.
	_, err := db.Exec(
		ctx,
		"DELETE FROM document_number_counters",
	)
	if err != nil {
		t.Fatalf("failed to clean document number counters: %v", err)
	}

	repository := NewPostgresDocumentNumberRepository(db)

	// First purchase request number for 2026.
	number, err := repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate first document number: %v", err)
	}

	if number != 1 {
		t.Fatalf(
			"expected first number to be 1, got %d",
			number,
		)
	}

	// Second purchase request number for 2026.
	number, err = repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate second document number: %v", err)
	}

	if number != 2 {
		t.Fatalf(
			"expected second number to be 2, got %d",
			number,
		)
	}

	// Third purchase request number for 2026.
	number, err = repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate third document number: %v", err)
	}

	if number != 3 {
		t.Fatalf(
			"expected third number to be 3, got %d",
			number,
		)
	}
}

func TestPostgresDocumentNumberRepository_YearlyReset(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"DELETE FROM document_number_counters",
	)
	if err != nil {
		t.Fatalf("failed to clean document number counters: %v", err)
	}

	repository := NewPostgresDocumentNumberRepository(db)

	// Generate numbers for 2026.
	number, err := repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate 2026 number: %v", err)
	}

	if number != 1 {
		t.Fatalf(
			"expected 2026 first number to be 1, got %d",
			number,
		)
	}

	number, err = repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate second 2026 number: %v", err)
	}

	if number != 2 {
		t.Fatalf(
			"expected 2026 second number to be 2, got %d",
			number,
		)
	}

	// New year must start from 1.
	number, err = repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2027,
	)
	if err != nil {
		t.Fatalf("failed to generate 2027 number: %v", err)
	}

	if number != 1 {
		t.Fatalf(
			"expected 2027 first number to be 1, got %d",
			number,
		)
	}
}

func TestPostgresDocumentNumberRepository_DocumentTypeIsolation(t *testing.T) {
	db := setupTestDB(t)

	ctx := context.Background()

	_, err := db.Exec(
		ctx,
		"DELETE FROM document_number_counters",
	)
	if err != nil {
		t.Fatalf("failed to clean document number counters: %v", err)
	}

	repository := NewPostgresDocumentNumberRepository(db)

	// First purchase request for 2026.
	number, err := repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate purchase request number: %v", err)
	}

	if number != 1 {
		t.Fatalf(
			"expected purchase request number to be 1, got %d",
			number,
		)
	}

	// First purchase order for the same year
	// should have its own counter.
	number, err = repository.NextDocumentNumber(
		ctx,
		"purchase_order",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate purchase order number: %v", err)
	}

	if number != 1 {
		t.Fatalf(
			"expected purchase order number to be 1, got %d",
			number,
		)
	}

	// Second purchase request should still be 2.
	number, err = repository.NextDocumentNumber(
		ctx,
		"purchase_request",
		2026,
	)
	if err != nil {
		t.Fatalf("failed to generate second purchase request number: %v", err)
	}

	if number != 2 {
		t.Fatalf(
			"expected second purchase request number to be 2, got %d",
			number,
		)
	}
}

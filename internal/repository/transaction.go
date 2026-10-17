package repository

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionManager interface {
	WithinTransaction(
		ctx context.Context,
		fn func(
			PurchaseRequestRepository,
			PurchaseRequestItemRepository,
			DocumentNumberRepository,
		) error,
	) error
}

type postgresTransactionManager struct {
	db *pgxpool.Pool
}

func NewPostgresTransactionManager(
	db *pgxpool.Pool,
) TransactionManager {
	return &postgresTransactionManager{
		db: db,
	}
}

func (m *postgresTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		PurchaseRequestRepository,
		PurchaseRequestItemRepository,
		DocumentNumberRepository,
	) error,
) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	requestRepository := NewPostgresPurchaseRequestRepositoryWithTx(tx)
	itemRepository := NewPostgresPurchaseRequestItemRepositoryWithTx(tx)
	documentNumberRepository := NewPostgresDocumentNumberRepositoryWithTx(tx)

	if err := fn(
		requestRepository,
		itemRepository,
		documentNumberRepository,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
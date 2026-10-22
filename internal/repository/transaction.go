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

type PurchaseOrderTransactionManager interface {
	WithinTransaction(
		ctx context.Context,
		fn func(
			PurchaseOrderRepository,
			PurchaseOrderItemRepository,
			DocumentNumberRepository,
			PurchaseRequestRepository,
			PurchaseRequestItemRepository,
		) error,
	) error
}

type postgresPurchaseOrderTransactionManager struct {
	db *pgxpool.Pool
}

func NewPostgresPurchaseOrderTransactionManager(
	db *pgxpool.Pool,
) PurchaseOrderTransactionManager {
	return &postgresPurchaseOrderTransactionManager{
		db: db,
	}
}

func (m *postgresPurchaseOrderTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		PurchaseOrderRepository,
		PurchaseOrderItemRepository,
		DocumentNumberRepository,
		PurchaseRequestRepository,
		PurchaseRequestItemRepository,
	) error,
) error {
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	poRepository := NewPostgresPurchaseOrderRepositoryWithTx(tx)
	itemRepository := NewPostgresPurchaseOrderItemRepositoryWithTx(tx)
	documentNumberRepository := NewPostgresDocumentNumberRepositoryWithTx(tx)
	prRepository := NewPostgresPurchaseRequestRepositoryWithTx(tx)
	prItemRepository := NewPostgresPurchaseRequestItemRepositoryWithTx(tx)

	if err := fn(
		poRepository,
		itemRepository,
		documentNumberRepository,
		prRepository,
		prItemRepository,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

type GoodsReceiptTransactionManager interface {
	WithinTransaction(
		ctx context.Context,
		fn func(
			GoodsReceiptRepository,
			GoodsReceiptItemRepository,
			PurchaseOrderRepository,
			PurchaseOrderItemRepository,
			DocumentNumberRepository,
		) error,
	) error
}

type postgresGoodsReceiptTransactionManager struct {
	db *pgxpool.Pool
}

func NewPostgresGoodsReceiptTransactionManager(
	db *pgxpool.Pool,
) GoodsReceiptTransactionManager {
	return &postgresGoodsReceiptTransactionManager{
		db: db,
	}
}

func (m *postgresGoodsReceiptTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(
		GoodsReceiptRepository,
		GoodsReceiptItemRepository,
		PurchaseOrderRepository,
		PurchaseOrderItemRepository,
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

	grnRepository := NewPostgresGoodsReceiptRepositoryWithTx(tx)
	grnItemRepository := NewPostgresGoodsReceiptItemRepositoryWithTx(tx)
	poRepository := NewPostgresPurchaseOrderRepositoryWithTx(tx)
	poItemRepository := NewPostgresPurchaseOrderItemRepositoryWithTx(tx)
	documentNumberRepository := NewPostgresDocumentNumberRepositoryWithTx(tx)

	if err := fn(
		grnRepository,
		grnItemRepository,
		poRepository,
		poItemRepository,
		documentNumberRepository,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

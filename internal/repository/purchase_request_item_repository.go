package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type PurchaseRequestItemRepository interface {
	CreatePurchaseRequestItem(
		ctx context.Context,
		item *domain.PurchaseRequestItem,
	) (*domain.PurchaseRequestItem, error)

	GetPurchaseRequestItems(
		ctx context.Context,
		purchaseRequestID string,
	) ([]*domain.PurchaseRequestItem, error)

	GetPurchaseRequestItemByID(
		ctx context.Context,
		id string,
	) (*domain.PurchaseRequestItem, error)

	UpdatePurchaseRequestItem(
		ctx context.Context,
		item *domain.PurchaseRequestItem,
	) (*domain.PurchaseRequestItem, error)

	DeletePurchaseRequestItem(
		ctx context.Context,
		id string,
	) error
}

type postgresPurchaseRequestItemRepository struct {
	db      *pgxpool.Pool
	queries *generated.Queries
}

var _ PurchaseRequestItemRepository = (*postgresPurchaseRequestItemRepository)(nil)

func NewPostgresPurchaseRequestItemRepository(
	db *pgxpool.Pool,
) PurchaseRequestItemRepository {
	return &postgresPurchaseRequestItemRepository{
		db:      db,
		queries: generated.New(db),
	}
}

func NewPostgresPurchaseRequestItemRepositoryWithTx(
	tx pgx.Tx,
) PurchaseRequestItemRepository {
	return &postgresPurchaseRequestItemRepository{
		queries: generated.New(tx),
	}
}

func (r *postgresPurchaseRequestItemRepository) CreatePurchaseRequestItem(
	ctx context.Context,
	item *domain.PurchaseRequestItem,
) (*domain.PurchaseRequestItem, error) {
	purchaseRequestID := pgtype.UUID{}

	if err := purchaseRequestID.Scan(item.PurchaseRequestID); err != nil {
		return nil, fmt.Errorf("invalid purchase request id: %w", err)
	}

	itemID := pgtype.UUID{}

	if err := itemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	quantity, err := decimalToNumeric(item.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	var estimatedUnitPrice pgtype.Numeric

	if item.EstimatedUnitPrice != nil {
		estimatedUnitPrice, err = decimalToNumeric(*item.EstimatedUnitPrice)
		if err != nil {
			return nil, fmt.Errorf("invalid estimated unit price: %w", err)
		}
	}

	created, err := r.queries.CreatePurchaseRequestItem(
		ctx,
		generated.CreatePurchaseRequestItemParams{
			PurchaseRequestID:  purchaseRequestID,
			ItemID:             itemID,
			Description:        item.Description,
			Quantity:           quantity,
			EstimatedUnitPrice: estimatedUnitPrice,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create purchase request item: %w", err)
	}

	result, err := mapPurchaseRequestItemToDomain(created)
	if err != nil {
		return nil, fmt.Errorf("map purchase request item: %w", err)
	}

	return result, nil
}

func decimalToNumeric(value decimal.Decimal) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric

	if err := numeric.Scan(value.String()); err != nil {
		return pgtype.Numeric{}, err
	}

	return numeric, nil
}

func mapPurchaseRequestItemToDomain(
	item generated.PurchaseRequestItem,
) (*domain.PurchaseRequestItem, error) {
	quantity, err := numericToDecimal(item.Quantity)
	if err != nil {
		return nil, fmt.Errorf("convert quantity: %w", err)
	}

	result := &domain.PurchaseRequestItem{
		ID:                item.ID.String(),
		PurchaseRequestID: item.PurchaseRequestID.String(),
		ItemID:            item.ItemID.String(),
		Description:       item.Description,
		Quantity:          quantity,
		CreatedAt:         item.CreatedAt.Time,
	}

	if item.EstimatedUnitPrice.Valid {
		price, err := numericToDecimal(item.EstimatedUnitPrice)
		if err != nil {
			return nil, fmt.Errorf("convert estimated unit price: %w", err)
		}

		result.EstimatedUnitPrice = &price
	}

	return result, nil
}

func numericToDecimal(value pgtype.Numeric) (decimal.Decimal, error) {
	if !value.Valid {
		return decimal.Zero, fmt.Errorf("numeric value is NULL")
	}

	if value.NaN {
		return decimal.Zero, fmt.Errorf("numeric value is NaN")
	}

	if value.Int == nil {
		return decimal.Zero, fmt.Errorf("numeric value has no integer representation")
	}

	return decimal.NewFromBigInt(
		value.Int,
		value.Exp,
	), nil
}

func (r *postgresPurchaseRequestItemRepository) DeletePurchaseRequestItem(
	ctx context.Context,
	id string,
) error {
	itemID := pgtype.UUID{}

	if err := itemID.Scan(id); err != nil {
		return fmt.Errorf("invalid purchase request item id: %w", err)
	}

	err := r.queries.DeletePurchaseRequestItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("delete purchase request item: %w", err)
	}

	return nil
}

func (r *postgresPurchaseRequestItemRepository) GetPurchaseRequestItemByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseRequestItem, error) {
	itemID := pgtype.UUID{}

	if err := itemID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase request item id: %w", err)
	}

	item, err := r.queries.GetPurchaseRequestItemByID(ctx, itemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase request item not found: %w", err)
		}

		return nil, fmt.Errorf("get purchase request item by id: %w", err)
	}

	result, err := mapPurchaseRequestItemToDomain(item)
	if err != nil {
		return nil, fmt.Errorf("map purchase request item: %w", err)
	}

	return result, nil
}

func (r *postgresPurchaseRequestItemRepository) GetPurchaseRequestItems(
	ctx context.Context,
	purchaseRequestID string,
) ([]*domain.PurchaseRequestItem, error) {
	requestID := pgtype.UUID{}

	if err := requestID.Scan(purchaseRequestID); err != nil {
		return nil, fmt.Errorf("invalid purchase request id: %w", err)
	}

	items, err := r.queries.GetPurchaseRequestItems(ctx, requestID)
	if err != nil {
		return nil, fmt.Errorf("get purchase request items: %w", err)
	}

	result := make([]*domain.PurchaseRequestItem, 0, len(items))

	for _, item := range items {
		mapped, err := mapPurchaseRequestItemToDomain(item)
		if err != nil {
			return nil, fmt.Errorf("map purchase request item: %w", err)
		}

		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresPurchaseRequestItemRepository) UpdatePurchaseRequestItem(
	ctx context.Context,
	item *domain.PurchaseRequestItem,
) (*domain.PurchaseRequestItem, error) {
	itemID := pgtype.UUID{}

	if err := itemID.Scan(item.ID); err != nil {
		return nil, fmt.Errorf("invalid purchase request item id: %w", err)
	}

	productItemID := pgtype.UUID{}

	if err := productItemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	quantity, err := decimalToNumeric(item.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	var estimatedUnitPrice pgtype.Numeric

	if item.EstimatedUnitPrice != nil {
		estimatedUnitPrice, err = decimalToNumeric(*item.EstimatedUnitPrice)
		if err != nil {
			return nil, fmt.Errorf("invalid estimated unit price: %w", err)
		}
	}

	updated, err := r.queries.UpdatePurchaseRequestItem(
		ctx,
		generated.UpdatePurchaseRequestItemParams{
			ID:                 itemID,
			ItemID:             productItemID,
			Description:        item.Description,
			Quantity:           quantity,
			EstimatedUnitPrice: estimatedUnitPrice,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf(
				"purchase request item not found or not editable: %w",
				err,
			)
		}

		return nil, fmt.Errorf("update purchase request item: %w", err)
	}

	result, err := mapPurchaseRequestItemToDomain(updated)
	if err != nil {
		return nil, fmt.Errorf("map purchase request item: %w", err)
	}

	return result, nil
}

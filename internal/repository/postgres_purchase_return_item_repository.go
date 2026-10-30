package repository

import (
	"context"
	"fmt"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
)

type postgresPurchaseReturnItemRepository struct {
	queries *generated.Queries
}

var _ PurchaseReturnItemRepository = (*postgresPurchaseReturnItemRepository)(nil)

func NewPostgresPurchaseReturnItemRepository(
	db *pgxpool.Pool,
) PurchaseReturnItemRepository {
	return &postgresPurchaseReturnItemRepository{
		queries: generated.New(db),
	}
}

func NewPostgresPurchaseReturnItemRepositoryWithTx(
	tx pgx.Tx,
) PurchaseReturnItemRepository {
	return &postgresPurchaseReturnItemRepository{
		queries: generated.New(tx),
	}
}

func mapPurchaseReturnItemToDomain(
	item generated.PurchaseReturnItem,
) (*domain.PurchaseReturnItem, error) {
	qty, err := numericToDecimal(item.ReturnQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert return quantity: %w", err)
	}

	unitPrice, err := numericToDecimal(item.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("convert unit price: %w", err)
	}

	totalPrice, err := numericToDecimal(item.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("convert total price: %w", err)
	}

	result := &domain.PurchaseReturnItem{
		ID:                  item.ID.String(),
		PurchaseReturnID:    item.PurchaseReturnID.String(),
		PurchaseOrderItemID: item.PurchaseOrderItemID.String(),
		ItemID:              item.ItemID.String(),
		Description:         item.Description,
		ReturnQuantity:      qty,
		UnitPrice:           unitPrice,
		TotalPrice:          totalPrice,
		CreatedAt:           item.CreatedAt.Time,
	}

	if item.Reason.Valid {
		r := item.Reason.String
		result.Reason = &r
	}

	return result, nil
}

func (r *postgresPurchaseReturnItemRepository) CreatePurchaseReturnItem(
	ctx context.Context,
	item *domain.PurchaseReturnItem,
) (*domain.PurchaseReturnItem, error) {
	returnID := pgtype.UUID{}
	if err := returnID.Scan(item.PurchaseReturnID); err != nil {
		return nil, fmt.Errorf("invalid purchase return id: %w", err)
	}

	poItemID := pgtype.UUID{}
	if err := poItemID.Scan(item.PurchaseOrderItemID); err != nil {
		return nil, fmt.Errorf("invalid purchase order item id: %w", err)
	}

	itemID := pgtype.UUID{}
	if err := itemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	qtyNumeric := pgtype.Numeric{}
	if err := qtyNumeric.Scan(item.ReturnQuantity.String()); err != nil {
		return nil, fmt.Errorf("invalid return quantity: %w", err)
	}

	unitPriceNumeric := pgtype.Numeric{}
	if err := unitPriceNumeric.Scan(item.UnitPrice.String()); err != nil {
		return nil, fmt.Errorf("invalid unit price: %w", err)
	}

	totalPriceNumeric := pgtype.Numeric{}
	if err := totalPriceNumeric.Scan(item.TotalPrice.String()); err != nil {
		return nil, fmt.Errorf("invalid total price: %w", err)
	}

	var reason pgtype.Text
	if item.Reason != nil {
		reason = pgtype.Text{String: *item.Reason, Valid: true}
	}

	created, err := r.queries.CreatePurchaseReturnItem(
		ctx,
		generated.CreatePurchaseReturnItemParams{
			PurchaseReturnID:    returnID,
			PurchaseOrderItemID: poItemID,
			ItemID:              itemID,
			Description:         item.Description,
			ReturnQuantity:      qtyNumeric,
			UnitPrice:           unitPriceNumeric,
			TotalPrice:          totalPriceNumeric,
			Reason:              reason,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create purchase return item: %w", err)
	}

	return mapPurchaseReturnItemToDomain(created)
}

func (r *postgresPurchaseReturnItemRepository) GetPurchaseReturnItems(
	ctx context.Context,
	purchaseReturnID string,
) ([]*domain.PurchaseReturnItem, error) {
	returnID := pgtype.UUID{}
	if err := returnID.Scan(purchaseReturnID); err != nil {
		return nil, fmt.Errorf("invalid purchase return id: %w", err)
	}

	items, err := r.queries.GetPurchaseReturnItems(ctx, returnID)
	if err != nil {
		return nil, fmt.Errorf("get purchase return items: %w", err)
	}

	result := make([]*domain.PurchaseReturnItem, 0, len(items))
	for _, item := range items {
		mapped, err := mapPurchaseReturnItemToDomain(item)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresPurchaseReturnItemRepository) GetPurchaseReturnItemByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseReturnItem, error) {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase return item id: %w", err)
	}

	item, err := r.queries.GetPurchaseReturnItemByID(ctx, itemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase return item not found: %w", err)
		}
		return nil, fmt.Errorf("get purchase return item by id: %w", err)
	}

	return mapPurchaseReturnItemToDomain(item)
}

func (r *postgresPurchaseReturnItemRepository) GetTotalReturnedQuantityForPOItem(
	ctx context.Context,
	purchaseOrderItemID string,
) (decimal.Decimal, error) {
	poItemID := pgtype.UUID{}
	if err := poItemID.Scan(purchaseOrderItemID); err != nil {
		return decimal.Zero, fmt.Errorf("invalid purchase order item id: %w", err)
	}

	totalNumeric, err := r.queries.GetTotalReturnedQuantityForPOItem(ctx, poItemID)
	if err != nil {
		return decimal.Zero, fmt.Errorf("get total returned quantity for po item: %w", err)
	}

	return numericToDecimal(totalNumeric)
}

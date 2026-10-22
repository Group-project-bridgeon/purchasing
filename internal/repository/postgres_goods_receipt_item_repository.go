package repository

import (
	"context"
	"fmt"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresGoodsReceiptItemRepository struct {
	queries *generated.Queries
}

var _ GoodsReceiptItemRepository = (*postgresGoodsReceiptItemRepository)(nil)

func NewPostgresGoodsReceiptItemRepository(
	db *pgxpool.Pool,
) GoodsReceiptItemRepository {
	return &postgresGoodsReceiptItemRepository{
		queries: generated.New(db),
	}
}

func NewPostgresGoodsReceiptItemRepositoryWithTx(
	tx pgx.Tx,
) GoodsReceiptItemRepository {
	return &postgresGoodsReceiptItemRepository{
		queries: generated.New(tx),
	}
}

func mapGoodsReceiptItemToDomain(
	item generated.GoodsReceiptItem,
) (*domain.GoodsReceiptItem, error) {
	receivedQty, err := numericToDecimal(item.ReceivedQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert received quantity: %w", err)
	}

	rejectedQty, err := numericToDecimal(item.RejectedQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert rejected quantity: %w", err)
	}

	result := &domain.GoodsReceiptItem{
		ID:                  item.ID.String(),
		GoodsReceiptID:      item.GoodsReceiptID.String(),
		PurchaseOrderItemID: item.PurchaseOrderItemID.String(),
		ItemID:              item.ItemID.String(),
		ReceivedQuantity:    receivedQty,
		RejectedQuantity:    rejectedQty,
		CreatedAt:           item.CreatedAt.Time,
	}

	if item.RejectionReason.Valid {
		r := item.RejectionReason.String
		result.RejectionReason = &r
	}

	return result, nil
}

func (r *postgresGoodsReceiptItemRepository) CreateGoodsReceiptItem(
	ctx context.Context,
	item *domain.GoodsReceiptItem,
) (*domain.GoodsReceiptItem, error) {
	grnID := pgtype.UUID{}
	if err := grnID.Scan(item.GoodsReceiptID); err != nil {
		return nil, fmt.Errorf("invalid goods receipt id: %w", err)
	}

	poItemID := pgtype.UUID{}
	if err := poItemID.Scan(item.PurchaseOrderItemID); err != nil {
		return nil, fmt.Errorf("invalid purchase order item id: %w", err)
	}

	itemID := pgtype.UUID{}
	if err := itemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	receivedQty, err := decimalToNumeric(item.ReceivedQuantity)
	if err != nil {
		return nil, fmt.Errorf("invalid received quantity: %w", err)
	}

	rejectedQty, err := decimalToNumeric(item.RejectedQuantity)
	if err != nil {
		return nil, fmt.Errorf("invalid rejected quantity: %w", err)
	}

	var reason pgtype.Text
	if item.RejectionReason != nil {
		reason = pgtype.Text{String: *item.RejectionReason, Valid: true}
	}

	created, err := r.queries.CreateGoodsReceiptItem(
		ctx,
		generated.CreateGoodsReceiptItemParams{
			GoodsReceiptID:      grnID,
			PurchaseOrderItemID: poItemID,
			ItemID:              itemID,
			ReceivedQuantity:    receivedQty,
			RejectedQuantity:    rejectedQty,
			RejectionReason:     reason,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create goods receipt item: %w", err)
	}

	return mapGoodsReceiptItemToDomain(created)
}

func (r *postgresGoodsReceiptItemRepository) GetGoodsReceiptItems(
	ctx context.Context,
	goodsReceiptID string,
) ([]*domain.GoodsReceiptItem, error) {
	grnID := pgtype.UUID{}
	if err := grnID.Scan(goodsReceiptID); err != nil {
		return nil, fmt.Errorf("invalid goods receipt id: %w", err)
	}

	items, err := r.queries.GetGoodsReceiptItems(ctx, grnID)
	if err != nil {
		return nil, fmt.Errorf("get goods receipt items: %w", err)
	}

	result := make([]*domain.GoodsReceiptItem, 0, len(items))
	for _, it := range items {
		mapped, err := mapGoodsReceiptItemToDomain(it)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresGoodsReceiptItemRepository) GetGoodsReceiptItemsByPOItemID(
	ctx context.Context,
	purchaseOrderItemID string,
) ([]*domain.GoodsReceiptItem, error) {
	poItemID := pgtype.UUID{}
	if err := poItemID.Scan(purchaseOrderItemID); err != nil {
		return nil, fmt.Errorf("invalid purchase order item id: %w", err)
	}

	items, err := r.queries.GetGoodsReceiptItemsByPOItemID(ctx, poItemID)
	if err != nil {
		return nil, fmt.Errorf("get goods receipt items by po item id: %w", err)
	}

	result := make([]*domain.GoodsReceiptItem, 0, len(items))
	for _, it := range items {
		mapped, err := mapGoodsReceiptItemToDomain(it)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

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

type postgresPurchaseOrderItemRepository struct {
	queries *generated.Queries
}

var _ PurchaseOrderItemRepository = (*postgresPurchaseOrderItemRepository)(nil)

func NewPostgresPurchaseOrderItemRepository(
	db *pgxpool.Pool,
) PurchaseOrderItemRepository {
	return &postgresPurchaseOrderItemRepository{
		queries: generated.New(db),
	}
}

func NewPostgresPurchaseOrderItemRepositoryWithTx(
	tx pgx.Tx,
) PurchaseOrderItemRepository {
	return &postgresPurchaseOrderItemRepository{
		queries: generated.New(tx),
	}
}

func mapPurchaseOrderItemToDomain(
	item generated.PurchaseOrderItem,
) (*domain.PurchaseOrderItem, error) {
	quantity, err := numericToDecimal(item.Quantity)
	if err != nil {
		return nil, fmt.Errorf("convert quantity: %w", err)
	}

	unitPrice, err := numericToDecimal(item.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("convert unit price: %w", err)
	}

	taxRate, err := numericToDecimal(item.TaxRate)
	if err != nil {
		return nil, fmt.Errorf("convert tax rate: %w", err)
	}

	totalPrice, err := numericToDecimal(item.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("convert total price: %w", err)
	}

	receivedQuantity, err := numericToDecimal(item.ReceivedQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert received quantity: %w", err)
	}

	billedQuantity, err := numericToDecimal(item.BilledQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert billed quantity: %w", err)
	}

	return &domain.PurchaseOrderItem{
		ID:               item.ID.String(),
		PurchaseOrderID:  item.PurchaseOrderID.String(),
		ItemID:           item.ItemID.String(),
		Description:      item.Description,
		Quantity:         quantity,
		UnitPrice:        unitPrice,
		TaxRate:          taxRate,
		TotalPrice:       totalPrice,
		ReceivedQuantity: receivedQuantity,
		BilledQuantity:   billedQuantity,
		CreatedAt:        item.CreatedAt.Time,
	}, nil
}

func (r *postgresPurchaseOrderItemRepository) CreatePurchaseOrderItem(
	ctx context.Context,
	item *domain.PurchaseOrderItem,
) (*domain.PurchaseOrderItem, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(item.PurchaseOrderID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	itemID := pgtype.UUID{}
	if err := itemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	quantity, err := decimalToNumeric(item.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	unitPrice, err := decimalToNumeric(item.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid unit price: %w", err)
	}

	taxRate, err := decimalToNumeric(item.TaxRate)
	if err != nil {
		return nil, fmt.Errorf("invalid tax rate: %w", err)
	}

	totalPrice, err := decimalToNumeric(item.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid total price: %w", err)
	}

	created, err := r.queries.CreatePurchaseOrderItem(
		ctx,
		generated.CreatePurchaseOrderItemParams{
			PurchaseOrderID: poID,
			ItemID:          itemID,
			Description:     item.Description,
			Quantity:        quantity,
			UnitPrice:       unitPrice,
			TaxRate:         taxRate,
			TotalPrice:      totalPrice,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create purchase order item: %w", err)
	}

	return mapPurchaseOrderItemToDomain(created)
}

func (r *postgresPurchaseOrderItemRepository) GetPurchaseOrderItems(
	ctx context.Context,
	purchaseOrderID string,
) ([]*domain.PurchaseOrderItem, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(purchaseOrderID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	items, err := r.queries.GetPurchaseOrderItems(ctx, poID)
	if err != nil {
		return nil, fmt.Errorf("get purchase order items: %w", err)
	}

	result := make([]*domain.PurchaseOrderItem, 0, len(items))
	for _, item := range items {
		mapped, err := mapPurchaseOrderItemToDomain(item)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresPurchaseOrderItemRepository) GetPurchaseOrderItemByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrderItem, error) {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase order item id: %w", err)
	}

	item, err := r.queries.GetPurchaseOrderItemByID(ctx, itemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order item not found: %w", err)
		}
		return nil, fmt.Errorf("get purchase order item by id: %w", err)
	}

	return mapPurchaseOrderItemToDomain(item)
}

func (r *postgresPurchaseOrderItemRepository) UpdatePurchaseOrderItem(
	ctx context.Context,
	item *domain.PurchaseOrderItem,
) (*domain.PurchaseOrderItem, error) {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(item.ID); err != nil {
		return nil, fmt.Errorf("invalid purchase order item id: %w", err)
	}

	productItemID := pgtype.UUID{}
	if err := productItemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	quantity, err := decimalToNumeric(item.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	unitPrice, err := decimalToNumeric(item.UnitPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid unit price: %w", err)
	}

	taxRate, err := decimalToNumeric(item.TaxRate)
	if err != nil {
		return nil, fmt.Errorf("invalid tax rate: %w", err)
	}

	totalPrice, err := decimalToNumeric(item.TotalPrice)
	if err != nil {
		return nil, fmt.Errorf("invalid total price: %w", err)
	}

	updated, err := r.queries.UpdatePurchaseOrderItem(
		ctx,
		generated.UpdatePurchaseOrderItemParams{
			ID:          itemID,
			ItemID:      productItemID,
			Description: item.Description,
			Quantity:    quantity,
			UnitPrice:   unitPrice,
			TaxRate:     taxRate,
			TotalPrice:  totalPrice,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order item not found or not editable: %w", err)
		}
		return nil, fmt.Errorf("update purchase order item: %w", err)
	}

	return mapPurchaseOrderItemToDomain(updated)
}

func (r *postgresPurchaseOrderItemRepository) DeletePurchaseOrderItem(
	ctx context.Context,
	id string,
) error {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(id); err != nil {
		return fmt.Errorf("invalid purchase order item id: %w", err)
	}

	rows, err := r.queries.DeletePurchaseOrderItem(ctx, itemID)
	if err != nil {
		return fmt.Errorf("delete purchase order item: %w", err)
	}

	if rows == 0 {
		return domain.ErrPurchaseOrderNotEditable
	}
	return nil
}

func (r *postgresPurchaseOrderItemRepository) UpdatePurchaseOrderItemFulfillment(
	ctx context.Context,
	id string,
	receivedQuantity decimal.Decimal,
	billedQuantity decimal.Decimal,
) (*domain.PurchaseOrderItem, error) {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase order item id: %w", err)
	}

	receivedNum, err := decimalToNumeric(receivedQuantity)
	if err != nil {
		return nil, fmt.Errorf("invalid received quantity: %w", err)
	}

	billedNum, err := decimalToNumeric(billedQuantity)
	if err != nil {
		return nil, fmt.Errorf("invalid billed quantity: %w", err)
	}

	updated, err := r.queries.UpdatePurchaseOrderItemFulfillment(
		ctx,
		generated.UpdatePurchaseOrderItemFulfillmentParams{
			ID:               itemID,
			ReceivedQuantity: receivedNum,
			BilledQuantity:   billedNum,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("update purchase order item fulfillment: %w", err)
	}

	return mapPurchaseOrderItemToDomain(updated)
}

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

type postgresVendorBillItemRepository struct {
	queries *generated.Queries
}

var _ VendorBillItemRepository = (*postgresVendorBillItemRepository)(nil)

func NewPostgresVendorBillItemRepository(
	db *pgxpool.Pool,
) VendorBillItemRepository {
	return &postgresVendorBillItemRepository{
		queries: generated.New(db),
	}
}

func NewPostgresVendorBillItemRepositoryWithTx(
	tx pgx.Tx,
) VendorBillItemRepository {
	return &postgresVendorBillItemRepository{
		queries: generated.New(tx),
	}
}

func mapVendorBillItemToDomain(
	item generated.VendorBillItem,
) (*domain.VendorBillItem, error) {
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

	matchedQuantity, err := numericToDecimal(item.MatchedQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert matched quantity: %w", err)
	}

	return &domain.VendorBillItem{
		ID:                  item.ID.String(),
		VendorBillID:        item.VendorBillID.String(),
		PurchaseOrderItemID: item.PurchaseOrderItemID.String(),
		ItemID:              item.ItemID.String(),
		Description:         item.Description,
		Quantity:            quantity,
		UnitPrice:           unitPrice,
		TaxRate:             taxRate,
		TotalPrice:          totalPrice,
		MatchedQuantity:     matchedQuantity,
		MatchStatus:         domain.BillItemMatchStatus(item.MatchStatus),
		CreatedAt:           item.CreatedAt.Time,
	}, nil
}

func (r *postgresVendorBillItemRepository) CreateVendorBillItem(
	ctx context.Context,
	item *domain.VendorBillItem,
) (*domain.VendorBillItem, error) {
	billID := pgtype.UUID{}
	if err := billID.Scan(item.VendorBillID); err != nil {
		return nil, fmt.Errorf("invalid vendor bill id: %w", err)
	}

	poItemID := pgtype.UUID{}
	if err := poItemID.Scan(item.PurchaseOrderItemID); err != nil {
		return nil, fmt.Errorf("invalid purchase order item id: %w", err)
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

	matchedQuantity, err := decimalToNumeric(item.MatchedQuantity)
	if err != nil {
		return nil, fmt.Errorf("invalid matched quantity: %w", err)
	}

	matchStatus := string(item.MatchStatus)
	if matchStatus == "" {
		matchStatus = string(domain.BillItemMatchStatusUnmatched)
	}

	created, err := r.queries.CreateVendorBillItem(
		ctx,
		generated.CreateVendorBillItemParams{
			VendorBillID:        billID,
			PurchaseOrderItemID: poItemID,
			ItemID:              itemID,
			Description:         item.Description,
			Quantity:            quantity,
			UnitPrice:           unitPrice,
			TaxRate:             taxRate,
			TotalPrice:          totalPrice,
			MatchedQuantity:     matchedQuantity,
			MatchStatus:         matchStatus,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create vendor bill item: %w", err)
	}

	return mapVendorBillItemToDomain(created)
}

func (r *postgresVendorBillItemRepository) GetVendorBillItems(
	ctx context.Context,
	vendorBillID string,
) ([]*domain.VendorBillItem, error) {
	billID := pgtype.UUID{}
	if err := billID.Scan(vendorBillID); err != nil {
		return nil, fmt.Errorf("invalid vendor bill id: %w", err)
	}

	items, err := r.queries.GetVendorBillItems(ctx, billID)
	if err != nil {
		return nil, fmt.Errorf("get vendor bill items: %w", err)
	}

	result := make([]*domain.VendorBillItem, 0, len(items))
	for _, it := range items {
		mapped, err := mapVendorBillItemToDomain(it)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresVendorBillItemRepository) GetVendorBillItemByID(
	ctx context.Context,
	id string,
) (*domain.VendorBillItem, error) {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid vendor bill item id: %w", err)
	}

	item, err := r.queries.GetVendorBillItemByID(ctx, itemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor bill item not found: %w", err)
		}
		return nil, fmt.Errorf("get vendor bill item by id: %w", err)
	}

	return mapVendorBillItemToDomain(item)
}

func (r *postgresVendorBillItemRepository) UpdateVendorBillItemMatch(
	ctx context.Context,
	id string,
	matchedQuantity decimal.Decimal,
	matchStatus domain.BillItemMatchStatus,
) (*domain.VendorBillItem, error) {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid vendor bill item id: %w", err)
	}

	matchedNum, err := decimalToNumeric(matchedQuantity)
	if err != nil {
		return nil, fmt.Errorf("invalid matched quantity: %w", err)
	}

	updated, err := r.queries.UpdateVendorBillItemMatch(
		ctx,
		generated.UpdateVendorBillItemMatchParams{
			ID:              itemID,
			MatchedQuantity: matchedNum,
			MatchStatus:     string(matchStatus),
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("vendor bill item not found: %w", err)
		}
		return nil, fmt.Errorf("update vendor bill item match: %w", err)
	}

	return mapVendorBillItemToDomain(updated)
}

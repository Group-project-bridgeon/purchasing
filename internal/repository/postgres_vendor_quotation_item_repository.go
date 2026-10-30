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

type postgresVendorQuotationItemRepository struct {
	queries *generated.Queries
}

var _ VendorQuotationItemRepository = (*postgresVendorQuotationItemRepository)(nil)

func NewPostgresVendorQuotationItemRepository(
	db *pgxpool.Pool,
) VendorQuotationItemRepository {
	return &postgresVendorQuotationItemRepository{
		queries: generated.New(db),
	}
}

func NewPostgresVendorQuotationItemRepositoryWithTx(
	tx pgx.Tx,
) VendorQuotationItemRepository {
	return &postgresVendorQuotationItemRepository{
		queries: generated.New(tx),
	}
}

func mapVendorQuotationItemToDomain(
	item generated.VendorQuotationItem,
) (*domain.VendorQuotationItem, error) {
	qty, err := numericToDecimal(item.OfferedQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert offered quantity: %w", err)
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

	result := &domain.VendorQuotationItem{
		ID:                item.ID.String(),
		VendorQuotationID: item.VendorQuotationID.String(),
		RFQItemID:         item.RfqItemID.String(),
		ItemID:            item.ItemID.String(),
		Description:       item.Description,
		OfferedQuantity:   qty,
		UnitPrice:         unitPrice,
		TaxRate:           taxRate,
		TotalPrice:        totalPrice,
		CreatedAt:         item.CreatedAt.Time,
	}

	if item.LeadTimeDays.Valid {
		days := int(item.LeadTimeDays.Int32)
		result.LeadTimeDays = &days
	}

	return result, nil
}

func (r *postgresVendorQuotationItemRepository) CreateVendorQuotationItem(
	ctx context.Context,
	item *domain.VendorQuotationItem,
) (*domain.VendorQuotationItem, error) {
	quoteID := pgtype.UUID{}
	if err := quoteID.Scan(item.VendorQuotationID); err != nil {
		return nil, fmt.Errorf("invalid vendor quotation id: %w", err)
	}

	rfqItemID := pgtype.UUID{}
	if err := rfqItemID.Scan(item.RFQItemID); err != nil {
		return nil, fmt.Errorf("invalid rfq item id: %w", err)
	}

	itemID := pgtype.UUID{}
	if err := itemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	qtyNumeric := pgtype.Numeric{}
	if err := qtyNumeric.Scan(item.OfferedQuantity.String()); err != nil {
		return nil, fmt.Errorf("invalid offered quantity: %w", err)
	}

	unitPriceNumeric := pgtype.Numeric{}
	if err := unitPriceNumeric.Scan(item.UnitPrice.String()); err != nil {
		return nil, fmt.Errorf("invalid unit price: %w", err)
	}

	taxRateNumeric := pgtype.Numeric{}
	if err := taxRateNumeric.Scan(item.TaxRate.String()); err != nil {
		return nil, fmt.Errorf("invalid tax rate: %w", err)
	}

	totalPriceNumeric := pgtype.Numeric{}
	if err := totalPriceNumeric.Scan(item.TotalPrice.String()); err != nil {
		return nil, fmt.Errorf("invalid total price: %w", err)
	}

	var leadTime pgtype.Int4
	if item.LeadTimeDays != nil {
		leadTime = pgtype.Int4{Int32: int32(*item.LeadTimeDays), Valid: true}
	}

	created, err := r.queries.CreateVendorQuotationItem(
		ctx,
		generated.CreateVendorQuotationItemParams{
			VendorQuotationID: quoteID,
			RfqItemID:         rfqItemID,
			ItemID:            itemID,
			Description:       item.Description,
			OfferedQuantity:   qtyNumeric,
			UnitPrice:         unitPriceNumeric,
			TaxRate:           taxRateNumeric,
			TotalPrice:        totalPriceNumeric,
			LeadTimeDays:      leadTime,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create vendor quotation item: %w", err)
	}

	return mapVendorQuotationItemToDomain(created)
}

func (r *postgresVendorQuotationItemRepository) GetVendorQuotationItems(
	ctx context.Context,
	vendorQuotationID string,
) ([]*domain.VendorQuotationItem, error) {
	quoteID := pgtype.UUID{}
	if err := quoteID.Scan(vendorQuotationID); err != nil {
		return nil, fmt.Errorf("invalid vendor quotation id: %w", err)
	}

	rows, err := r.queries.GetVendorQuotationItems(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("get vendor quotation items: %w", err)
	}

	result := make([]*domain.VendorQuotationItem, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapVendorQuotationItemToDomain(row)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

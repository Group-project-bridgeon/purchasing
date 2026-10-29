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

type postgresRFQItemRepository struct {
	queries *generated.Queries
}

var _ RFQItemRepository = (*postgresRFQItemRepository)(nil)

func NewPostgresRFQItemRepository(
	db *pgxpool.Pool,
) RFQItemRepository {
	return &postgresRFQItemRepository{
		queries: generated.New(db),
	}
}

func NewPostgresRFQItemRepositoryWithTx(
	tx pgx.Tx,
) RFQItemRepository {
	return &postgresRFQItemRepository{
		queries: generated.New(tx),
	}
}

func mapRFQItemToDomain(
	item generated.RfqItem,
) (*domain.RFQItem, error) {
	qty, err := numericToDecimal(item.TargetQuantity)
	if err != nil {
		return nil, fmt.Errorf("convert target quantity: %w", err)
	}

	result := &domain.RFQItem{
		ID:             item.ID.String(),
		RFQID:          item.RfqID.String(),
		ItemID:         item.ItemID.String(),
		Description:    item.Description,
		TargetQuantity: qty,
		CreatedAt:      item.CreatedAt.Time,
	}

	if item.TargetDeliveryDate.Valid {
		date := item.TargetDeliveryDate.Time
		result.TargetDeliveryDate = &date
	}

	return result, nil
}

func (r *postgresRFQItemRepository) CreateRFQItem(
	ctx context.Context,
	item *domain.RFQItem,
) (*domain.RFQItem, error) {
	rfqID := pgtype.UUID{}
	if err := rfqID.Scan(item.RFQID); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	itemID := pgtype.UUID{}
	if err := itemID.Scan(item.ItemID); err != nil {
		return nil, fmt.Errorf("invalid item id: %w", err)
	}

	qtyNumeric := pgtype.Numeric{}
	if err := qtyNumeric.Scan(item.TargetQuantity.String()); err != nil {
		return nil, fmt.Errorf("invalid target quantity: %w", err)
	}

	var deliveryDate pgtype.Date
	if item.TargetDeliveryDate != nil {
		deliveryDate = pgtype.Date{Time: *item.TargetDeliveryDate, Valid: true}
	}

	created, err := r.queries.CreateRFQItem(
		ctx,
		generated.CreateRFQItemParams{
			RfqID:              rfqID,
			ItemID:             itemID,
			Description:        item.Description,
			TargetQuantity:     qtyNumeric,
			TargetDeliveryDate: deliveryDate,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create rfq item: %w", err)
	}

	return mapRFQItemToDomain(created)
}

func (r *postgresRFQItemRepository) GetRFQItems(
	ctx context.Context,
	rfqID string,
) ([]*domain.RFQItem, error) {
	id := pgtype.UUID{}
	if err := id.Scan(rfqID); err != nil {
		return nil, fmt.Errorf("invalid rfq id: %w", err)
	}

	rows, err := r.queries.GetRFQItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get rfq items: %w", err)
	}

	result := make([]*domain.RFQItem, 0, len(rows))
	for _, row := range rows {
		mapped, err := mapRFQItemToDomain(row)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresRFQItemRepository) GetRFQItemByID(
	ctx context.Context,
	id string,
) (*domain.RFQItem, error) {
	itemID := pgtype.UUID{}
	if err := itemID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid rfq item id: %w", err)
	}

	row, err := r.queries.GetRFQItemByID(ctx, itemID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("rfq item not found: %w", err)
		}
		return nil, fmt.Errorf("get rfq item by id: %w", err)
	}

	return mapRFQItemToDomain(row)
}

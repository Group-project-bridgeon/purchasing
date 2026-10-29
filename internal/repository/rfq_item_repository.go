package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type RFQItemRepository interface {
	CreateRFQItem(
		ctx context.Context,
		item *domain.RFQItem,
	) (*domain.RFQItem, error)

	GetRFQItems(
		ctx context.Context,
		rfqID string,
	) ([]*domain.RFQItem, error)

	GetRFQItemByID(
		ctx context.Context,
		id string,
	) (*domain.RFQItem, error)
}

package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type RFQRepository interface {
	CreateRFQ(
		ctx context.Context,
		rfq *domain.RFQ,
	) (*domain.RFQ, error)

	GetRFQByID(
		ctx context.Context,
		id string,
	) (*domain.RFQ, error)

	GetRFQByNumber(
		ctx context.Context,
		rfqNumber string,
	) (*domain.RFQ, error)

	ListRFQs(
		ctx context.Context,
		status *domain.RFQStatus,
		purchaseRequestID *string,
		limit int,
		offset int,
	) ([]*domain.RFQ, error)

	CountRFQs(
		ctx context.Context,
		status *domain.RFQStatus,
		purchaseRequestID *string,
	) (int64, error)

	UpdateRFQ(
		ctx context.Context,
		rfq *domain.RFQ,
	) (*domain.RFQ, error)

	UpdateRFQStatus(
		ctx context.Context,
		id string,
		status domain.RFQStatus,
	) (*domain.RFQ, error)
}

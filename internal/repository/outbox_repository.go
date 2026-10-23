package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type OutboxRepository interface {
	CreateEvent(ctx context.Context, event *domain.OutboxEvent) (*domain.OutboxEvent, error)
	GetPendingEvents(ctx context.Context, limit int32) ([]*domain.OutboxEvent, error)
	MarkPublished(ctx context.Context, id string) error
	MarkFailed(ctx context.Context, id string, lastError string) error
	ListEvents(ctx context.Context, filter domain.OutboxFilter) ([]*domain.OutboxEvent, error)
	GetEventByID(ctx context.Context, id string) (*domain.OutboxEvent, error)
	RetryEvent(ctx context.Context, id string) error
}

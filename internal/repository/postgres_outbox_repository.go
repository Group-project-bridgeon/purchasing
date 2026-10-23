package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresOutboxRepository struct {
	queries *generated.Queries
}

var _ OutboxRepository = (*postgresOutboxRepository)(nil)

func NewPostgresOutboxRepository(db *pgxpool.Pool) OutboxRepository {
	return &postgresOutboxRepository{
		queries: generated.New(db),
	}
}

func NewPostgresOutboxRepositoryWithTx(tx pgx.Tx) OutboxRepository {
	return &postgresOutboxRepository{
		queries: generated.New(tx),
	}
}

func mapOutboxEventToDomain(e generated.OutboxEvent) *domain.OutboxEvent {
	var lastError *string
	if e.LastError.Valid {
		lastError = &e.LastError.String
	}

	var publishedAt *time.Time
	if e.PublishedAt.Valid {
		publishedAt = &e.PublishedAt.Time
	}

	return &domain.OutboxEvent{
		ID:            e.ID.String(),
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		EventType:     e.EventType,
		Payload:       json.RawMessage(e.Payload),
		Status:        domain.OutboxStatus(e.Status),
		RetryCount:    e.RetryCount,
		LastError:     lastError,
		CreatedAt:     e.CreatedAt.Time,
		PublishedAt:   publishedAt,
	}
}

func (r *postgresOutboxRepository) CreateEvent(
	ctx context.Context,
	event *domain.OutboxEvent,
) (*domain.OutboxEvent, error) {
	created, err := r.queries.CreateOutboxEvent(ctx, generated.CreateOutboxEventParams{
		AggregateType: event.AggregateType,
		AggregateID:   event.AggregateID,
		EventType:     event.EventType,
		Payload:       []byte(event.Payload),
	})
	if err != nil {
		return nil, fmt.Errorf("create outbox event: %w", err)
	}

	return mapOutboxEventToDomain(created), nil
}

func (r *postgresOutboxRepository) GetPendingEvents(
	ctx context.Context,
	limit int32,
) ([]*domain.OutboxEvent, error) {
	events, err := r.queries.GetPendingOutboxEvents(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending outbox events: %w", err)
	}

	result := make([]*domain.OutboxEvent, len(events))
	for i, e := range events {
		result[i] = mapOutboxEventToDomain(e)
	}
	return result, nil
}

func (r *postgresOutboxRepository) MarkPublished(
	ctx context.Context,
	id string,
) error {
	var uuidVal pgtype.UUID
	if err := uuidVal.Scan(id); err != nil {
		return fmt.Errorf("invalid event id %q: %w", id, err)
	}

	if err := r.queries.MarkOutboxEventPublished(ctx, uuidVal); err != nil {
		return fmt.Errorf("mark outbox event published: %w", err)
	}
	return nil
}

func (r *postgresOutboxRepository) MarkFailed(
	ctx context.Context,
	id string,
	lastError string,
) error {
	var uuidVal pgtype.UUID
	if err := uuidVal.Scan(id); err != nil {
		return fmt.Errorf("invalid event id %q: %w", id, err)
	}

	arg := generated.MarkOutboxEventFailedParams{
		ID:        uuidVal,
		LastError: pgtype.Text{String: lastError, Valid: lastError != ""},
	}

	if err := r.queries.MarkOutboxEventFailed(ctx, arg); err != nil {
		return fmt.Errorf("mark outbox event failed: %w", err)
	}
	return nil
}

func (r *postgresOutboxRepository) ListEvents(
	ctx context.Context,
	filter domain.OutboxFilter,
) ([]*domain.OutboxEvent, error) {
	arg := generated.ListOutboxEventsParams{
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}

	if filter.Status != nil && *filter.Status != "" {
		arg.Status = pgtype.Text{String: *filter.Status, Valid: true}
	}
	if filter.AggregateType != nil && *filter.AggregateType != "" {
		arg.AggregateType = pgtype.Text{String: *filter.AggregateType, Valid: true}
	}

	events, err := r.queries.ListOutboxEvents(ctx, arg)
	if err != nil {
		return nil, fmt.Errorf("list outbox events: %w", err)
	}

	result := make([]*domain.OutboxEvent, len(events))
	for i, e := range events {
		result[i] = mapOutboxEventToDomain(e)
	}
	return result, nil
}

func (r *postgresOutboxRepository) GetEventByID(
	ctx context.Context,
	id string,
) (*domain.OutboxEvent, error) {
	var uuidVal pgtype.UUID
	if err := uuidVal.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid event id %q: %w", id, err)
	}

	event, err := r.queries.GetOutboxEventByID(ctx, uuidVal)
	if err != nil {
		return nil, fmt.Errorf("get outbox event by id: %w", err)
	}

	return mapOutboxEventToDomain(event), nil
}

func (r *postgresOutboxRepository) RetryEvent(
	ctx context.Context,
	id string,
) error {
	var uuidVal pgtype.UUID
	if err := uuidVal.Scan(id); err != nil {
		return fmt.Errorf("invalid event id %q: %w", id, err)
	}

	if err := r.queries.RetryOutboxEvent(ctx, uuidVal); err != nil {
		return fmt.Errorf("retry outbox event: %w", err)
	}
	return nil
}

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/metrics"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
)

// EventDispatcher defines the contract for sending published outbox events to downstream brokers or webhooks.
type EventDispatcher interface {
	Dispatch(ctx context.Context, event *domain.OutboxEvent) error
}

// LogEventDispatcher is a built-in dispatcher that logs outgoing domain events.
type LogEventDispatcher struct{}

func NewLogEventDispatcher() *LogEventDispatcher {
	return &LogEventDispatcher{}
}

func (d *LogEventDispatcher) Dispatch(ctx context.Context, event *domain.OutboxEvent) error {
	slog.InfoContext(
		ctx,
		"outbox_event_dispatched",
		slog.String("event_type", event.EventType),
		slog.String("aggregate_type", event.AggregateType),
		slog.String("aggregate_id", event.AggregateID),
		slog.String("event_id", event.ID),
	)
	return nil
}

// OutboxService manages outbox event querying, manual retries, and background publishing.
type OutboxService interface {
	RecordEvent(ctx context.Context, aggregateType, aggregateID, eventType string, payload interface{}) (*domain.OutboxEvent, error)
	ListEvents(ctx context.Context, filter domain.OutboxFilter) ([]*domain.OutboxEvent, error)
	GetEventByID(ctx context.Context, id string) (*domain.OutboxEvent, error)
	RetryEvent(ctx context.Context, id string) error
	ProcessBatch(ctx context.Context, batchSize int32) (int, error)
	StartBackgroundPublisher(ctx context.Context, pollInterval time.Duration)
	Stop()
}

type outboxService struct {
	repo       repository.OutboxRepository
	dispatcher EventDispatcher
	stopChan   chan struct{}
	wg         sync.WaitGroup
}

func NewOutboxService(
	repo repository.OutboxRepository,
	dispatcher EventDispatcher,
) OutboxService {
	if dispatcher == nil {
		dispatcher = NewLogEventDispatcher()
	}

	return &outboxService{
		repo:       repo,
		dispatcher: dispatcher,
		stopChan:   make(chan struct{}),
	}
}

func (s *outboxService) RecordEvent(
	ctx context.Context,
	aggregateType string,
	aggregateID string,
	eventType string,
	payload interface{},
) (*domain.OutboxEvent, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal outbox payload: %w", err)
	}

	event := &domain.OutboxEvent{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       json.RawMessage(payloadBytes),
		Status:        domain.OutboxStatusPending,
	}

	return s.repo.CreateEvent(ctx, event)
}

func (s *outboxService) ListEvents(
	ctx context.Context,
	filter domain.OutboxFilter,
) ([]*domain.OutboxEvent, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 20
	}
	return s.repo.ListEvents(ctx, filter)
}

func (s *outboxService) GetEventByID(
	ctx context.Context,
	id string,
) (*domain.OutboxEvent, error) {
	if id == "" {
		return nil, fmt.Errorf("event id is required")
	}
	return s.repo.GetEventByID(ctx, id)
}

func (s *outboxService) RetryEvent(
	ctx context.Context,
	id string,
) error {
	if id == "" {
		return fmt.Errorf("event id is required")
	}
	return s.repo.RetryEvent(ctx, id)
}

func (s *outboxService) ProcessBatch(
	ctx context.Context,
	batchSize int32,
) (int, error) {
	if batchSize <= 0 {
		batchSize = 50
	}

	start := time.Now()
	defer func() {
		metrics.RecordOutboxBatchDuration(time.Since(start))
	}()

	events, err := s.repo.GetPendingEvents(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("fetch pending events: %w", err)
	}

	dispatchedCount := 0
	for _, event := range events {
		dispatchErr := s.dispatcher.Dispatch(ctx, event)
		if dispatchErr != nil {
			metrics.RecordOutboxEvent("failed")
			slog.ErrorContext(ctx, "outbox_dispatch_failed",
				slog.String("event_id", event.ID),
				slog.String("event_type", event.EventType),
				slog.Any("error", dispatchErr),
			)
			if markErr := s.repo.MarkFailed(ctx, event.ID, dispatchErr.Error()); markErr != nil {
				slog.ErrorContext(ctx, "outbox_mark_failed_error",
					slog.String("event_id", event.ID),
					slog.Any("error", markErr),
				)
			}
		} else {
			if markErr := s.repo.MarkPublished(ctx, event.ID); markErr != nil {
				metrics.RecordOutboxEvent("failed")
				slog.ErrorContext(ctx, "outbox_mark_published_error",
					slog.String("event_id", event.ID),
					slog.Any("error", markErr),
				)
			} else {
				metrics.RecordOutboxEvent("dispatched")
				dispatchedCount++
			}
		}
	}

	return dispatchedCount, nil
}

func (s *outboxService) StartBackgroundPublisher(
	ctx context.Context,
	pollInterval time.Duration,
) {
	if pollInterval <= 0 {
		pollInterval = 3 * time.Second
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(pollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-s.stopChan:
				slog.InfoContext(ctx, "outbox_worker_stopping")
				return
			case <-ctx.Done():
				slog.InfoContext(ctx, "outbox_worker_context_cancelled")
				return
			case <-ticker.C:
				processCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				count, err := s.ProcessBatch(processCtx, 50)
				cancel()
				if err != nil {
					slog.ErrorContext(ctx, "outbox_batch_processing_error", slog.Any("error", err))
				} else if count > 0 {
					slog.InfoContext(ctx, "outbox_batch_dispatched", slog.Int("count", count))
				}
			}
		}
	}()
	slog.InfoContext(ctx, "outbox_worker_started", slog.Duration("interval", pollInterval))
}

func (s *outboxService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	slog.Info("outbox_worker_stopped_cleanly")
}

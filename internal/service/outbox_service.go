package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/logger"
	"github.com/Group-project-bridgeon/purchasing/internal/metrics"
	"github.com/Group-project-bridgeon/purchasing/internal/repository"
)

// EventDispatcher defines the contract for sending published outbox events to downstream brokers or webhooks.
type EventDispatcher interface {
	Dispatch(ctx context.Context, event *domain.OutboxEvent) error
}

var ErrDispatcherNotConfigured = errors.New("production outbox dispatcher is not configured")

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
	repo         repository.OutboxRepository
	dispatcher   EventDispatcher
	stopChan     chan struct{}
	stopOnce     sync.Once
	startOnce    sync.Once
	workerCancel context.CancelFunc
	workerMu     sync.Mutex
	wg           sync.WaitGroup
}

func NewOutboxService(
	repo repository.OutboxRepository,
	dispatcher EventDispatcher,
) OutboxService {

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
	if s.dispatcher == nil {
		return 0, ErrDispatcherNotConfigured
	}
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
				slog.String("component", "outbox"), slog.String("operation", "dispatch"), slog.String("request_id", logger.GetRequestID(ctx)),
				slog.String("event_id", event.ID),
				slog.String("reason", "delivery_failed"),
			)
			if markErr := s.repo.MarkFailed(ctx, event.ID, "delivery_failed"); markErr != nil {
				logger.Failure(ctx, "outbox", "mark_failed", markErr)
			}
		} else {
			if markErr := s.repo.MarkPublished(ctx, event.ID); markErr != nil {
				metrics.RecordOutboxEvent("failed")
				logger.Failure(ctx, "outbox", "mark_published", markErr)
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
	if s.dispatcher == nil {
		return
	}
	s.startOnce.Do(func() {
		workerCtx, stop := context.WithCancel(ctx)
		s.workerMu.Lock()
		s.workerCancel = stop
		s.workerMu.Unlock()
		ctx = workerCtx
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
					slog.InfoContext(ctx, "outbox_worker_stopping", "component", "outbox", "operation", "worker_stop", "reason", "stopping", "request_id", logger.GetRequestID(ctx))
					return
				case <-ctx.Done():
					slog.InfoContext(ctx, "outbox_worker_context_cancelled", "component", "outbox", "operation", "worker_stop", "reason", "context_canceled", "request_id", logger.GetRequestID(ctx))
					return
				case <-ticker.C:
					processCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
					count, err := s.ProcessBatch(processCtx, 50)
					cancel()
					if err != nil {
						logger.Failure(ctx, "outbox", "process_batch", err)
					} else if count > 0 {
						slog.InfoContext(ctx, "outbox_batch_dispatched", "component", "outbox", "operation", "process_batch", "reason", "completed", "request_id", logger.GetRequestID(ctx), "count", count)
					}
				}
			}
		}()
		slog.InfoContext(ctx, "outbox_worker_started", "component", "outbox", "operation", "worker_start", "reason", "configured_dispatcher", "request_id", logger.GetRequestID(ctx), "interval", pollInterval)
	})
}

func (s *outboxService) Stop() {
	s.stopOnce.Do(func() {
		close(s.stopChan)
		s.workerMu.Lock()
		if s.workerCancel != nil {
			s.workerCancel()
		}
		s.workerMu.Unlock()
	})
	s.wg.Wait()
	slog.Info("outbox_worker_stopped_cleanly", "component", "outbox", "operation", "worker_stop", "reason", "completed", "request_id", "")
}

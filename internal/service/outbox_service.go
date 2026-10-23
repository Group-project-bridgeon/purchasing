package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
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
	log.Printf(
		"[OUTBOX EVENT DISPATCHED] type=%s aggregate=%s:%s event_id=%s payload=%s",
		event.EventType,
		event.AggregateType,
		event.AggregateID,
		event.ID,
		string(event.Payload),
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

	events, err := s.repo.GetPendingEvents(ctx, batchSize)
	if err != nil {
		return 0, fmt.Errorf("fetch pending events: %w", err)
	}

	dispatchedCount := 0
	for _, event := range events {
		dispatchErr := s.dispatcher.Dispatch(ctx, event)
		if dispatchErr != nil {
			log.Printf("[OUTBOX ERROR] dispatch failed for event %s: %v", event.ID, dispatchErr)
			if markErr := s.repo.MarkFailed(ctx, event.ID, dispatchErr.Error()); markErr != nil {
				log.Printf("[OUTBOX ERROR] failed to mark event %s as failed: %v", event.ID, markErr)
			}
		} else {
			if markErr := s.repo.MarkPublished(ctx, event.ID); markErr != nil {
				log.Printf("[OUTBOX ERROR] failed to mark event %s as published: %v", event.ID, markErr)
			} else {
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
				log.Println("[OUTBOX] background publisher worker stopping...")
				return
			case <-ctx.Done():
				log.Println("[OUTBOX] context cancelled, background publisher worker stopping...")
				return
			case <-ticker.C:
				processCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				count, err := s.ProcessBatch(processCtx, 50)
				cancel()
				if err != nil {
					log.Printf("[OUTBOX] batch processing error: %v", err)
				} else if count > 0 {
					log.Printf("[OUTBOX] successfully dispatched %d pending events", count)
				}
			}
		}
	}()
	log.Printf("[OUTBOX] background publisher worker started (interval: %v)", pollInterval)
}

func (s *outboxService) Stop() {
	close(s.stopChan)
	s.wg.Wait()
	log.Println("[OUTBOX] background publisher worker shut down cleanly")
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type mockOutboxRepository struct {
	counter      int
	events       map[string]*domain.OutboxEvent
	publishedIDs []string
	failedIDs    map[string]string
}

func newMockOutboxRepository() *mockOutboxRepository {
	return &mockOutboxRepository{
		events:    make(map[string]*domain.OutboxEvent),
		failedIDs: make(map[string]string),
	}
}

func (m *mockOutboxRepository) CreateEvent(ctx context.Context, event *domain.OutboxEvent) (*domain.OutboxEvent, error) {
	m.counter++
	if event.ID == "" {
		event.ID = fmt.Sprintf("evt-mock-uuid-%d", m.counter)
	}
	event.CreatedAt = time.Now()
	m.events[event.ID] = event
	return event, nil
}

func (m *mockOutboxRepository) GetPendingEvents(ctx context.Context, limit int32) ([]*domain.OutboxEvent, error) {
	var list []*domain.OutboxEvent
	for _, e := range m.events {
		if e.Status == domain.OutboxStatusPending {
			list = append(list, e)
			if int32(len(list)) >= limit {
				break
			}
		}
	}
	return list, nil
}

func (m *mockOutboxRepository) MarkPublished(ctx context.Context, id string) error {
	e, ok := m.events[id]
	if !ok {
		return errors.New("event not found")
	}
	now := time.Now()
	e.Status = domain.OutboxStatusPublished
	e.PublishedAt = &now
	m.publishedIDs = append(m.publishedIDs, id)
	return nil
}

func (m *mockOutboxRepository) MarkFailed(ctx context.Context, id string, lastError string) error {
	e, ok := m.events[id]
	if !ok {
		return errors.New("event not found")
	}
	e.RetryCount++
	e.LastError = &lastError
	if e.RetryCount >= 5 {
		e.Status = domain.OutboxStatusFailed
	}
	m.failedIDs[id] = lastError
	return nil
}

func (m *mockOutboxRepository) ListEvents(ctx context.Context, filter domain.OutboxFilter) ([]*domain.OutboxEvent, error) {
	var list []*domain.OutboxEvent
	for _, e := range m.events {
		if filter.Status != nil && string(e.Status) != *filter.Status {
			continue
		}
		if filter.AggregateType != nil && e.AggregateType != *filter.AggregateType {
			continue
		}
		list = append(list, e)
	}
	return list, nil
}

func (m *mockOutboxRepository) GetEventByID(ctx context.Context, id string) (*domain.OutboxEvent, error) {
	e, ok := m.events[id]
	if !ok {
		return nil, errors.New("event not found")
	}
	return e, nil
}

func (m *mockOutboxRepository) RetryEvent(ctx context.Context, id string) error {
	e, ok := m.events[id]
	if !ok {
		return errors.New("event not found")
	}
	e.Status = domain.OutboxStatusPending
	e.RetryCount = 0
	e.LastError = nil
	return nil
}

type mockDispatcher struct {
	shouldFail bool
}

func (d *mockDispatcher) Dispatch(ctx context.Context, event *domain.OutboxEvent) error {
	if d.shouldFail {
		return errors.New("simulated broker connection failure")
	}
	return nil
}

func TestOutboxService_RecordEvent(t *testing.T) {
	repo := newMockOutboxRepository()
	svc := NewOutboxService(repo, nil)

	payload := map[string]interface{}{
		"order_number": "PO-2026-0001",
		"total_amount": "5000.00",
	}

	event, err := svc.RecordEvent(
		context.Background(),
		"purchase_order",
		"po-12345",
		"purchase_order.issued",
		payload,
	)
	if err != nil {
		t.Fatalf("unexpected error recording outbox event: %v", err)
	}

	if event.AggregateType != "purchase_order" {
		t.Errorf("expected aggregate_type 'purchase_order', got %q", event.AggregateType)
	}
	if event.EventType != "purchase_order.issued" {
		t.Errorf("expected event_type 'purchase_order.issued', got %q", event.EventType)
	}
	if event.Status != domain.OutboxStatusPending {
		t.Errorf("expected status 'pending', got %q", event.Status)
	}

	var parsedPayload map[string]interface{}
	if err := json.Unmarshal(event.Payload, &parsedPayload); err != nil {
		t.Fatalf("invalid json payload: %v", err)
	}
	if parsedPayload["order_number"] != "PO-2026-0001" {
		t.Errorf("expected payload order_number PO-2026-0001, got %v", parsedPayload["order_number"])
	}
}

func TestOutboxService_ProcessBatch_Success(t *testing.T) {
	repo := newMockOutboxRepository()
	dispatcher := &mockDispatcher{shouldFail: false}
	svc := NewOutboxService(repo, dispatcher)

	// Seed 2 pending events
	_, _ = svc.RecordEvent(context.Background(), "purchase_order", "po-1", "purchase_order.issued", "test1")
	_, _ = svc.RecordEvent(context.Background(), "goods_receipt", "grn-1", "goods_receipt.received", "test2")

	count, err := svc.ProcessBatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected batch error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 dispatched events, got %d", count)
	}
	if len(repo.publishedIDs) != 2 {
		t.Errorf("expected 2 published IDs in repo, got %d", len(repo.publishedIDs))
	}
}

func TestOutboxService_ProcessBatch_Failure(t *testing.T) {
	repo := newMockOutboxRepository()
	dispatcher := &mockDispatcher{shouldFail: true}
	svc := NewOutboxService(repo, dispatcher)

	event, _ := svc.RecordEvent(context.Background(), "vendor_bill", "bill-1", "vendor_bill.approved", "payload")

	count, err := svc.ProcessBatch(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected batch error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 dispatched events on failure, got %d", count)
	}
	if repo.failedIDs[event.ID] != "delivery_failed" {
		t.Errorf("expected failed error recorded, got %q", repo.failedIDs[event.ID])
	}
}

func TestOutboxService_RetryEvent(t *testing.T) {
	repo := newMockOutboxRepository()
	svc := NewOutboxService(repo, nil)

	event, _ := svc.RecordEvent(context.Background(), "purchase_return", "ret-1", "purchase_return.completed", "payload")
	_ = repo.MarkFailed(context.Background(), event.ID, "some error")

	// Verify it has failed state
	stored, _ := svc.GetEventByID(context.Background(), event.ID)
	if stored.RetryCount != 1 {
		t.Errorf("expected retry_count 1, got %d", stored.RetryCount)
	}

	// Retry
	err := svc.RetryEvent(context.Background(), event.ID)
	if err != nil {
		t.Fatalf("failed to retry event: %v", err)
	}

	retried, _ := svc.GetEventByID(context.Background(), event.ID)
	if retried.Status != domain.OutboxStatusPending {
		t.Errorf("expected status reset to 'pending', got %q", retried.Status)
	}
	if retried.RetryCount != 0 {
		t.Errorf("expected retry_count reset to 0, got %d", retried.RetryCount)
	}
}

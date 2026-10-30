package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

func TestUnconfiguredOutboxCannotAcknowledge(t *testing.T) {
	repo := newMockOutboxRepository()
	svc := NewOutboxService(repo, nil)
	event, err := svc.RecordEvent(context.Background(), "order", "id", "created", map[string]string{"test": "synthetic"})
	if err != nil {
		t.Fatal(err)
	}
	svc.StartBackgroundPublisher(context.Background(), time.Millisecond)
	if count, err := svc.ProcessBatch(context.Background(), 10); count != 0 || !errors.Is(err, ErrDispatcherNotConfigured) {
		t.Fatal("unconfigured delivery acknowledged")
	}
	svc.Stop()
	svc.Stop()
	if event.Status != domain.OutboxStatusPending || len(repo.publishedIDs) != 0 || event.RetryCount != 0 || event.PublishedAt != nil {
		t.Fatal("disabled outbox changed pending facts")
	}
}

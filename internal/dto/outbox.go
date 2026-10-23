package dto

import (
	"encoding/json"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
)

type OutboxEventResponse struct {
	ID            string              `json:"id"`
	AggregateType string              `json:"aggregate_type"`
	AggregateID   string              `json:"aggregate_id"`
	EventType     string              `json:"event_type"`
	Payload       json.RawMessage     `json:"payload"`
	Status        domain.OutboxStatus `json:"status"`
	RetryCount    int32               `json:"retry_count"`
	LastError     *string             `json:"last_error,omitempty"`
	CreatedAt     time.Time           `json:"created_at"`
	PublishedAt   *time.Time          `json:"published_at,omitempty"`
}

type ListOutboxEventsResponse struct {
	Events []OutboxEventResponse `json:"events"`
	Count  int                   `json:"count"`
}

func ToOutboxEventResponse(e *domain.OutboxEvent) OutboxEventResponse {
	return OutboxEventResponse{
		ID:            e.ID,
		AggregateType: e.AggregateType,
		AggregateID:   e.AggregateID,
		EventType:     e.EventType,
		Payload:       e.Payload,
		Status:        e.Status,
		RetryCount:    e.RetryCount,
		LastError:     e.LastError,
		CreatedAt:     e.CreatedAt,
		PublishedAt:   e.PublishedAt,
	}
}

func ToOutboxEventResponses(events []*domain.OutboxEvent) []OutboxEventResponse {
	result := make([]OutboxEventResponse, len(events))
	for i, e := range events {
		result[i] = ToOutboxEventResponse(e)
	}
	return result
}

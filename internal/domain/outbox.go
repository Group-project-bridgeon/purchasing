package domain

import (
	"encoding/json"
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "pending"
	OutboxStatusPublished OutboxStatus = "published"
	OutboxStatusFailed    OutboxStatus = "failed"
)

type OutboxEvent struct {
	ID            string          `json:"id"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	EventType     string          `json:"event_type"`
	Payload       json.RawMessage `json:"payload"`
	Status        OutboxStatus    `json:"status"`
	RetryCount    int32           `json:"retry_count"`
	LastError     *string         `json:"last_error,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
}

type OutboxFilter struct {
	Status        *string
	AggregateType *string
	Limit         int32
	Offset        int32
}

// Package actor carries the actual verified human across application/database layers.
package actor

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type identityKey struct{}

var ErrRequired = errors.New("verified human actor required")

func WithID(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, identityKey{}, id)
}
func ID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(identityKey{}).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, ErrRequired
	}
	return id, nil
}

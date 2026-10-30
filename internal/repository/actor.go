package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/actor"
	"github.com/jackc/pgx/v5/pgtype"
)

func requiredActor(ctx context.Context) (pgtype.UUID, error) {
	id, err := actor.ID(ctx)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

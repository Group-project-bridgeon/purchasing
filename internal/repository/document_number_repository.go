package repository

import (
	"context"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentNumberRepository interface {
	NextDocumentNumber(
		ctx context.Context,
		documentType string,
		year int,
	) (int64, error)
}

type postgresDocumentNumberRepository struct {
	queries *generated.Queries
}

var _ DocumentNumberRepository = (*postgresDocumentNumberRepository)(nil)

func NewPostgresDocumentNumberRepository(
	db *pgxpool.Pool,
) DocumentNumberRepository {
	return &postgresDocumentNumberRepository{
		queries: generated.New(db),
	}
}

func NewPostgresDocumentNumberRepositoryWithTx(
	tx pgx.Tx,
) DocumentNumberRepository {
	return &postgresDocumentNumberRepository{
		queries: generated.New(tx),
	}
}

func (r *postgresDocumentNumberRepository) NextDocumentNumber(
	ctx context.Context,
	documentType string,
	year int,
) (int64, error) {
	return r.queries.NextDocumentNumber(
		ctx,
		generated.NextDocumentNumberParams{
			DocumentType: documentType,
			Year:         int32(year),
		},
	)
}

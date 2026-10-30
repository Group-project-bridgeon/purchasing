package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

const RequiredSchemaVersion int64 = 9

var ErrSchemaNotReady = errors.New("required clean Purchasing schema is unavailable; apply reviewed service migrations")

type QueryRower interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

// CheckSchema is read-only. It explicitly rejects the legacy embedded-runner
// tracker; adoption requires an operator-reviewed backup/schema/history procedure.
func CheckSchema(ctx context.Context, db QueryRower) error {
	var shape int
	err := db.QueryRow(ctx, `SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='schema_migrations' AND ((column_name='version' AND data_type='bigint') OR (column_name='dirty' AND data_type='boolean'))`).Scan(&shape)
	if err != nil {
		return errors.Join(ErrSchemaNotReady, err)
	}
	if shape != 2 {
		return ErrSchemaNotReady
	}
	var version int64
	var dirty bool
	if err = db.QueryRow(ctx, `SELECT version,dirty FROM public.schema_migrations`).Scan(&version, &dirty); err != nil {
		return errors.Join(ErrSchemaNotReady, err)
	}
	if version != RequiredSchemaVersion || dirty {
		return ErrSchemaNotReady
	}
	var complete bool
	err = db.QueryRow(ctx, `SELECT bool_and(to_regclass('public.'||name) IS NOT NULL) FROM unnest($1::text[]) AS name`, []string{"purchase_requests", "purchase_request_items", "document_number_counters", "purchase_orders", "purchase_order_items", "goods_receipts", "goods_receipt_items", "vendor_bills", "vendor_bill_items", "purchase_returns", "purchase_return_items", "rfqs", "rfq_items", "vendor_quotations", "vendor_quotation_items", "outbox_events"}).Scan(&complete)
	if err != nil {
		return errors.Join(ErrSchemaNotReady, err)
	}
	if !complete {
		return ErrSchemaNotReady
	}
	return nil
}

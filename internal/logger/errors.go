package logger

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"syscall"

	"github.com/Group-project-bridgeon/purchasing/internal/actor"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

// StageError carries only caller-selected stable diagnostics. SafeHint must be
// constructed from known configuration validation text, never raw DB/key errors.
type StageError struct {
	Component, Operation, Reason, SafeHint string
	Err                                    error
}

func (e *StageError) Error() string { return e.Reason }
func (e *StageError) Unwrap() error { return e.Err }

// ErrorReason exposes stable categories only, never infrastructure error text.
func ErrorReason(err error) (string, string) {
	if errors.Is(err, syscall.EADDRINUSE) {
		return "address_in_use", ""
	}
	for _, stateErr := range []error{domain.ErrPurchaseRequestNotEditable, domain.ErrPurchaseOrderNotEditable, domain.ErrPurchaseReturnNotEditable, domain.ErrRFQNotEditable} {
		if errors.Is(err, stateErr) {
			return "document_state_conflict", ""
		}
	}
	if errors.Is(err, actor.ErrRequired) {
		return "forbidden", ""
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout", ""
	}
	if errors.Is(err, context.Canceled) {
		return "request_canceled", ""
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case "28P01", "28000":
			return "database_authentication_failed", pg.Code
		case "42501":
			return "forbidden", pg.Code
		case "23505":
			return "conflict", pg.Code
		case "23514", "23503", "40001", "40P01":
			return "document_state_conflict", pg.Code
		case "57014", "55P03":
			return "timeout", pg.Code
		default:
			return "database_operation_failed", pg.Code
		}
	}
	var network net.Error
	if errors.As(err, &network) {
		return "dependency_unavailable", ""
	}
	return "operation_failed", ""
}

func Failure(ctx context.Context, component, operation string, err error) {
	reason, state := ErrorReason(err)
	var stage *StageError
	if errors.As(err, &stage) {
		component = stage.Component
		operation = stage.Operation
		if reason == "operation_failed" || reason == "dependency_unavailable" {
			reason = stage.Reason
		}
	}
	attrs := []any{"component", component, "operation", operation, "reason", reason, "request_id", GetRequestID(ctx)}
	if stage != nil && stage.SafeHint != "" {
		attrs = append(attrs, "configuration_issue", stage.SafeHint)
	}
	if id, e := actor.ID(ctx); e == nil {
		attrs = append(attrs, "actor_user_id", id.String())
	}
	if state != "" {
		attrs = append(attrs, "sqlstate", state)
	}
	slog.ErrorContext(ctx, "operation_failed", attrs...)
}

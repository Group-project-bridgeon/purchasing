package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/actor"
	"github.com/Group-project-bridgeon/purchasing/internal/database"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func fixtureSQL(t *testing.T, db *pgxpool.Pool, query string, args ...any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.Exec(ctx, query, args...); err != nil {
		var pg *pgconn.PgError
		state := "none"
		if errors.As(err, &pg) {
			state = pg.Code
		}
		t.Fatalf("owned fixture SQL failed (SQLSTATE %s)", state)
	}
}
func requireSQLState(t *testing.T, err error, state string) {
	t.Helper()
	var pg *pgconn.PgError
	if !errors.As(err, &pg) || pg.Code != state {
		t.Fatalf("expected SQLSTATE %s from invariant rejection", state)
	}
}
func seedOrder(t *testing.T, db *pgxpool.Pool, creator uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	fixtureSQL(t, db, `INSERT INTO purchase_orders(id,po_number,vendor_id,status,currency,created_by) VALUES($1,$2,$3,'draft','USD',$4)`, id, "PO-"+id.String(), uuid.New(), creator)
	return id
}
func seedOrderItem(t *testing.T, db *pgxpool.Pool, parent uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	fixtureSQL(t, db, `INSERT INTO purchase_order_items(id,purchase_order_id,item_id,description,quantity,unit_price,tax_rate,total_price) VALUES($1,$2,$3,'fixture',1,1,0,1)`, id, parent, uuid.New())
	return id
}

func TestPrerequisiteApprovalsRequireDifferentPersistedCreator(t *testing.T) {
	db := setupTestDB(t)
	creator, approver := uuid.New(), uuid.New()
	base := context.Background()
	for _, table := range []string{"purchase_requests", "purchase_orders", "purchase_returns", "vendor_bills"} {
		t.Run(table, func(t *testing.T) {
			id := uuid.New()
			po := seedOrder(t, db, creator)
			switch table {
			case "purchase_requests":
				fixtureSQL(t, db, `INSERT INTO purchase_requests(id,request_number,requested_by,purpose) VALUES($1,$2,$3,'fixture')`, id, "PR-"+id.String(), creator)
				fixtureSQL(t, db, `UPDATE purchase_requests SET status='submitted' WHERE id=$1`, id)
			case "purchase_orders":
				id = po
			case "purchase_returns":
				fixtureSQL(t, db, `INSERT INTO purchase_returns(id,return_number,purchase_order_id,vendor_id,reason,created_by) VALUES($1,$2,$3,$4,'fixture',$5)`, id, "RT-"+id.String(), po, uuid.New(), creator)
			case "vendor_bills":
				fixtureSQL(t, db, `INSERT INTO vendor_bills(id,internal_bill_number,bill_number,purchase_order_id,vendor_id,bill_date,currency,match_status,created_by) VALUES($1,$2,$3,$4,$5,CURRENT_DATE,'USD','matched',$6)`, id, "VB-"+id.String(), id.String(), po, uuid.New(), creator)
			}
			approve := func(who uuid.UUID) (*string, *time.Time, error) {
				ctx := actor.WithID(base, who)
				switch table {
				case "purchase_requests":
					v, e := NewPostgresPurchaseRequestRepository(db).ApprovePurchaseRequest(ctx, id.String())
					if e != nil {
						return nil, nil, e
					}
					return v.ApprovedBy, v.ApprovedAt, nil
				case "purchase_orders":
					v, e := NewPostgresPurchaseOrderRepository(db).UpdatePurchaseOrderStatus(ctx, id.String(), domain.PurchaseOrderStatusApproved)
					if e != nil {
						return nil, nil, e
					}
					return v.ApprovedBy, v.ApprovedAt, nil
				case "purchase_returns":
					v, e := NewPostgresPurchaseReturnRepository(db).UpdatePurchaseReturnStatus(ctx, id.String(), domain.PurchaseReturnStatusApproved)
					if e != nil {
						return nil, nil, e
					}
					return v.ApprovedBy, v.ApprovedAt, nil
				default:
					v, e := NewPostgresVendorBillRepository(db).UpdateVendorBillStatus(ctx, id.String(), domain.BillStatusApproved)
					if e != nil {
						return nil, nil, e
					}
					return v.ApprovedBy, v.ApprovedAt, nil
				}
			}
			_, _, err := approve(creator)
			requireSQLState(t, err, "42501")
			by, at, err := approve(approver)
			if err != nil {
				t.Fatal("different authorized actor could not approve")
			}
			if by == nil || *by != approver.String() || at == nil {
				t.Fatal("approval actor/time not durable")
			}
			_, err = db.Exec(base, `UPDATE `+table+` SET approved_by=$2 WHERE id=$1`, id, creator)
			requireSQLState(t, err, "23514")
			field := "notes"
			if table == "purchase_requests" {
				field = "purpose"
			}
			_, err = db.Exec(base, `UPDATE `+table+` SET `+field+`='unauthorized edit' WHERE id=$1`, id)
			requireSQLState(t, err, "23514")
			_, err = db.Exec(base, `UPDATE `+table+` SET status='draft' WHERE id=$1`, id)
			requireSQLState(t, err, "23514")
		})
	}
}

func TestPrerequisiteItemsRejectEveryNonDraftMutation(t *testing.T) {
	db := setupTestDB(t)
	a, b := uuid.New(), uuid.New()
	ctx := context.Background()
	for _, kind := range []string{"request", "order"} {
		t.Run(kind, func(t *testing.T) {
			parent, item := uuid.New(), uuid.New()
			var table, parentColumn string
			if kind == "request" {
				table = "purchase_request_items"
				parentColumn = "purchase_request_id"
				fixtureSQL(t, db, `INSERT INTO purchase_requests(id,request_number,requested_by,purpose) VALUES($1,$2,$3,'fixture')`, parent, "PR-"+parent.String(), a)
				fixtureSQL(t, db, `INSERT INTO purchase_request_items(id,purchase_request_id,item_id,description,quantity) VALUES($1,$2,$3,'fixture',1)`, item, parent, uuid.New())
				fixtureSQL(t, db, `UPDATE purchase_requests SET status='submitted' WHERE id=$1`, parent)
				_, err := NewPostgresPurchaseRequestRepository(db).ApprovePurchaseRequest(actor.WithID(ctx, b), parent.String())
				if err != nil {
					t.Fatal("fixture request approval failed")
				}
			} else {
				table = "purchase_order_items"
				parentColumn = "purchase_order_id"
				parent = seedOrder(t, db, a)
				item = seedOrderItem(t, db, parent)
				fixtureSQL(t, db, `UPDATE purchase_orders SET status='approved',approved_by=$2 WHERE id=$1`, parent, b)
				fixtureSQL(t, db, `UPDATE purchase_orders SET status='issued' WHERE id=$1`, parent)
			}
			for _, query := range []string{`UPDATE ` + table + ` SET quantity=2 WHERE id=$1`, `DELETE FROM ` + table + ` WHERE id=$1`} {
				_, err := db.Exec(ctx, query, item)
				requireSQLState(t, err, "23514")
			}
			query := `INSERT INTO ` + table + `(id,` + parentColumn + `,item_id,description,quantity) VALUES($1,$2,$3,'blocked',1)`
			if kind == "order" {
				query = `INSERT INTO purchase_order_items(id,purchase_order_id,item_id,description,quantity,unit_price,total_price) VALUES($1,$2,$3,'blocked',1,1,1)`
			}
			_, err := db.Exec(ctx, query, uuid.New(), parent, uuid.New())
			requireSQLState(t, err, "23514")
			if kind == "order" {
				fixtureSQL(t, db, `UPDATE purchase_order_items SET received_quantity=1,billed_quantity=1 WHERE id=$1`, item)
			}
		})
	}
	// Header updates on the two other editable parent routes must also reject.
	po := seedOrder(t, db, a)
	rfq := uuid.New()
	fixtureSQL(t, db, `INSERT INTO rfqs(id,rfq_number,title,submission_deadline,created_by) VALUES($1,$2,'fixture',NOW()+INTERVAL '1 day',$3)`, rfq, "RFQ-"+rfq.String(), a)
	fixtureSQL(t, db, `UPDATE rfqs SET status='published' WHERE id=$1`, rfq)
	_, err := db.Exec(ctx, `UPDATE rfqs SET title='blocked' WHERE id=$1`, rfq)
	requireSQLState(t, err, "23514")
	if _, err := NewPostgresPurchaseOrderRepository(db).UpdatePurchaseOrderStatus(ctx, po.String(), domain.PurchaseOrderStatusApproved); err == nil {
		t.Fatal("missing trusted approval actor accepted")
	}
}

func TestPrerequisiteApprovalSerializesWithItemMutations(t *testing.T) {
	db := setupTestDB(t)
	for _, kind := range []string{"request", "order"} {
		t.Run(kind, func(t *testing.T) {
			for _, operation := range []string{"insert", "update", "delete", "header"} {
				t.Run(operation, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					a, b := uuid.New(), uuid.New()
					var parent, item uuid.UUID
					var table, parentColumn, header string
					if kind == "order" {
						parent = seedOrder(t, db, a)
						item = seedOrderItem(t, db, parent)
						table = "purchase_order_items"
						parentColumn = "purchase_order_id"
						header = "purchase_orders"
					} else {
						parent, item = uuid.New(), uuid.New()
						table = "purchase_request_items"
						parentColumn = "purchase_request_id"
						header = "purchase_requests"
						fixtureSQL(t, db, `INSERT INTO purchase_requests(id,request_number,requested_by,purpose) VALUES($1,$2,$3,'fixture')`, parent, "PR-"+parent.String(), a)
						fixtureSQL(t, db, `INSERT INTO purchase_request_items(id,purchase_request_id,item_id,description,quantity) VALUES($1,$2,$3,'fixture',1)`, item, parent, uuid.New())
						fixtureSQL(t, db, `UPDATE purchase_requests SET status='submitted' WHERE id=$1`, parent)
					}
					tx, err := db.Begin(ctx)
					if err != nil {
						t.Fatal("owned begin failed")
					}
					defer func() {
						cleanup, c := context.WithTimeout(context.Background(), 2*time.Second)
						defer c()
						_ = tx.Rollback(cleanup)
					}()
					if _, err = tx.Exec(ctx, `UPDATE `+header+` SET status='approved',approved_by=$2 WHERE id=$1`, parent, b); err != nil {
						t.Fatal("owned approval failed")
					}
					result := make(chan error, 1)
					started := make(chan uint32, 1)
					go func() {
						conn, err := db.Acquire(ctx)
						if err != nil {
							result <- err
							return
						}
						defer conn.Release()
						started <- conn.Conn().PgConn().PID()
						switch operation {
						case "insert":
							query := `INSERT INTO ` + table + `(` + parentColumn + `,item_id,description,quantity) VALUES($1,$2,'blocked',1)`
							if kind == "order" {
								query = `INSERT INTO purchase_order_items(purchase_order_id,item_id,description,quantity,unit_price,total_price) VALUES($1,$2,'blocked',1,1,1)`
							}
							_, err = conn.Exec(ctx, query, parent, uuid.New())
						case "update":
							_, err = conn.Exec(ctx, `UPDATE `+table+` SET quantity=2 WHERE id=$1`, item)
						case "delete":
							_, err = conn.Exec(ctx, `DELETE FROM `+table+` WHERE id=$1`, item)
						case "header":
							_, err = conn.Exec(ctx, `UPDATE `+header+` SET notes='blocked' WHERE id=$1`, parent)
						}
						result <- err
					}()
					var pid uint32
					select {
					case pid = <-started:
					case <-result:
						t.Fatal("worker failed before lock check")
					case <-ctx.Done():
						t.Fatal("worker startup timed out")
					}
					ticker := time.NewTicker(10 * time.Millisecond)
					defer ticker.Stop()
					waiting := false
					for !waiting {
						if err = db.QueryRow(ctx, `SELECT COALESCE(wait_event_type='Lock',false) FROM pg_stat_activity WHERE pid=$1`, pid).Scan(&waiting); err != nil {
							t.Fatal("owned lock observation failed")
						}
						if waiting {
							break
						}
						select {
						case <-result:
							t.Fatal("edit did not wait on the actual parent lock")
						case <-ctx.Done():
							t.Fatal("bounded lock observation timed out")
						case <-ticker.C:
						}
					}
					if err = tx.Commit(ctx); err != nil {
						t.Fatal("approval commit failed")
					}
					select {
					case err = <-result:
						requireSQLState(t, err, "23514")
					case <-ctx.Done():
						t.Fatal("bounded mutation did not finish")
					}
					var count int
					if err = db.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE `+parentColumn+`=$1 AND quantity=1`, parent).Scan(&count); err != nil || count != 1 {
						t.Fatal("approval race changed protected lines")
					}
				})
			}
		})
	}
}

func TestPrerequisiteSchemaAndUnconfiguredOutbox(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()
	if err := database.CheckSchema(ctx, db); err != nil {
		t.Fatal("required migrated schema unavailable")
	}
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = tx.Exec(ctx, `ALTER TABLE schema_migrations RENAME TO fixture_original_tracker;CREATE TABLE schema_migrations(version VARCHAR(100) PRIMARY KEY,applied_at TIMESTAMPTZ DEFAULT NOW())`); err != nil {
		t.Fatal("legacy tracker simulation failed")
	}
	if err = database.CheckSchema(ctx, tx); !errors.Is(err, database.ErrSchemaNotReady) {
		t.Fatal("embedded-runner tracker accepted")
	}
	if err = tx.Rollback(ctx); err != nil {
		t.Fatal("fixture rollback failed")
	}
	var legacy uuid.UUID
	if err = db.QueryRow(ctx, `SELECT id FROM purchase_orders WHERE po_number='LEGACY-UNKNOWN'`).Scan(&legacy); err != nil {
		t.Fatal("legacy migration preservation seed missing")
	}
	_, err = NewPostgresPurchaseOrderRepository(db).UpdatePurchaseOrderStatus(actor.WithID(ctx, uuid.New()), legacy.String(), domain.PurchaseOrderStatusApproved)
	requireSQLState(t, err, "23514")
}

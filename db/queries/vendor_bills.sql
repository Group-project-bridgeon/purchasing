-- name: CreateVendorBill :one
INSERT INTO vendor_bills (
    internal_bill_number,
    bill_number,
    purchase_order_id,
    vendor_id,
    bill_date,
    due_date,
    currency,
    subtotal,
    tax_amount,
    total_amount,
    match_status,
    status,
    discrepancy_reason,
    notes,
    created_by
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13,
    $14,
    sqlc.arg(created_by)
)
RETURNING *;


-- name: GetVendorBillByID :one
SELECT *
FROM vendor_bills
WHERE id = $1;


-- name: GetVendorBillByInternalNumber :one
SELECT *
FROM vendor_bills
WHERE internal_bill_number = $1;


-- name: GetVendorBillByVendorInvoice :one
SELECT *
FROM vendor_bills
WHERE vendor_id = $1
  AND bill_number = $2;


-- name: ListVendorBills :many
SELECT *
FROM vendor_bills
ORDER BY bill_date DESC
LIMIT $1
OFFSET $2;


-- name: ListVendorBillsByPO :many
SELECT *
FROM vendor_bills
WHERE purchase_order_id = $1
ORDER BY bill_date DESC;


-- name: ListVendorBillsByVendor :many
SELECT *
FROM vendor_bills
WHERE vendor_id = $1
ORDER BY bill_date DESC
LIMIT $2
OFFSET $3;


-- name: ListVendorBillsByStatus :many
SELECT *
FROM vendor_bills
WHERE status = $1
ORDER BY bill_date DESC
LIMIT $2
OFFSET $3;


-- name: ListVendorBillsByMatchStatus :many
SELECT *
FROM vendor_bills
WHERE match_status = $1
ORDER BY bill_date DESC
LIMIT $2
OFFSET $3;


-- name: UpdateVendorBill :one
UPDATE vendor_bills
SET
    bill_date = $2,
    due_date = $3,
    currency = $4,
    subtotal = $5,
    tax_amount = $6,
    total_amount = $7,
    notes = $8,
    updated_at = NOW()
WHERE id = $1
  AND status = 'draft'
RETURNING *;


-- name: UpdateVendorBillMatchStatus :one
UPDATE vendor_bills
SET
    match_status = $2,
    discrepancy_reason = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateVendorBillStatus :one
UPDATE vendor_bills
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: CreateVendorBillItem :one
INSERT INTO vendor_bill_items (
    vendor_bill_id,
    purchase_order_item_id,
    item_id,
    description,
    quantity,
    unit_price,
    tax_rate,
    total_price,
    matched_quantity,
    match_status
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10
)
RETURNING *;


-- name: GetVendorBillItems :many
SELECT *
FROM vendor_bill_items
WHERE vendor_bill_id = $1
ORDER BY created_at ASC;


-- name: GetVendorBillItemByID :one
SELECT *
FROM vendor_bill_items
WHERE id = $1;


-- name: UpdateVendorBillItemMatch :one
UPDATE vendor_bill_items
SET
    matched_quantity = $2,
    match_status = $3
WHERE id = $1
RETURNING *;

-- name: ApproveVendorBill :one
UPDATE vendor_bills SET status='approved', approved_by=sqlc.arg(approved_by), approved_at=NOW(), updated_at=NOW()
WHERE id=sqlc.arg(id) AND status='draft' AND match_status='matched' RETURNING *;

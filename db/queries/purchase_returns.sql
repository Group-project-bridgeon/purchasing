-- name: CreatePurchaseReturn :one
INSERT INTO purchase_returns (
    return_number,
    purchase_order_id,
    goods_receipt_id,
    vendor_id,
    status,
    return_date,
    total_return_amount,
    reason,
    carrier,
    tracking_number,
    created_by,
    notes
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    COALESCE($6::date, CURRENT_DATE),
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
)
RETURNING *;


-- name: GetPurchaseReturnByID :one
SELECT *
FROM purchase_returns
WHERE id = $1;


-- name: GetPurchaseReturnByNumber :one
SELECT *
FROM purchase_returns
WHERE return_number = $1;


-- name: ListPurchaseReturns :many
SELECT *
FROM purchase_returns
WHERE ($1::uuid IS NULL OR purchase_order_id = $1)
  AND ($2::uuid IS NULL OR vendor_id = $2)
  AND ($3::varchar IS NULL OR status = $3)
ORDER BY created_at DESC
LIMIT $4
OFFSET $5;


-- name: CountPurchaseReturns :one
SELECT count(*)
FROM purchase_returns
WHERE ($1::uuid IS NULL OR purchase_order_id = $1)
  AND ($2::uuid IS NULL OR vendor_id = $2)
  AND ($3::varchar IS NULL OR status = $3);


-- name: UpdatePurchaseReturn :one
UPDATE purchase_returns
SET
    reason = $2,
    carrier = $3,
    tracking_number = $4,
    notes = $5,
    updated_at = NOW()
WHERE id = $1 AND status = 'draft'
RETURNING *;


-- name: UpdatePurchaseReturnStatus :one
UPDATE purchase_returns
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: ShipPurchaseReturn :one
UPDATE purchase_returns
SET
    status = 'shipped',
    carrier = $2,
    tracking_number = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: CreatePurchaseReturnItem :one
INSERT INTO purchase_return_items (
    purchase_return_id,
    purchase_order_item_id,
    item_id,
    description,
    return_quantity,
    unit_price,
    total_price,
    reason
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8
)
RETURNING *;


-- name: GetPurchaseReturnItems :many
SELECT *
FROM purchase_return_items
WHERE purchase_return_id = $1
ORDER BY created_at ASC;


-- name: GetPurchaseReturnItemByID :one
SELECT *
FROM purchase_return_items
WHERE id = $1;


-- name: GetTotalReturnedQuantityForPOItem :one
SELECT COALESCE(SUM(pri.return_quantity), 0)::numeric(18, 3) AS total_returned
FROM purchase_return_items pri
JOIN purchase_returns pr ON pr.id = pri.purchase_return_id
WHERE pri.purchase_order_item_id = $1
  AND pr.status IN ('approved', 'shipped', 'completed');

-- name: ApprovePurchaseReturn :one
UPDATE purchase_returns SET status='approved', approved_by=sqlc.arg(approved_by), approved_at=NOW(), updated_at=NOW()
WHERE id=sqlc.arg(id) AND status='draft' RETURNING *;

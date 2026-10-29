-- name: CreateGoodsReceipt :one
INSERT INTO goods_receipts (
    grn_number,
    purchase_order_id,
    vendor_id,
    received_by,
    received_at,
    status,
    notes
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    COALESCE($5::timestamptz, NOW()),
    $6,
    $7
)
RETURNING *;


-- name: GetGoodsReceiptByID :one
SELECT *
FROM goods_receipts
WHERE id = $1;


-- name: GetGoodsReceiptByNumber :one
SELECT *
FROM goods_receipts
WHERE grn_number = $1;


-- name: ListGoodsReceipts :many
SELECT *
FROM goods_receipts
ORDER BY received_at DESC
LIMIT $1
OFFSET $2;


-- name: ListGoodsReceiptsByPO :many
SELECT *
FROM goods_receipts
WHERE purchase_order_id = $1
ORDER BY received_at DESC;


-- name: ListGoodsReceiptsByVendor :many
SELECT *
FROM goods_receipts
WHERE vendor_id = $1
ORDER BY received_at DESC
LIMIT $2
OFFSET $3;


-- name: UpdateGoodsReceiptStatus :one
UPDATE goods_receipts
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: CreateGoodsReceiptItem :one
INSERT INTO goods_receipt_items (
    goods_receipt_id,
    purchase_order_item_id,
    item_id,
    received_quantity,
    rejected_quantity,
    rejection_reason
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;


-- name: GetGoodsReceiptItems :many
SELECT *
FROM goods_receipt_items
WHERE goods_receipt_id = $1
ORDER BY created_at ASC;


-- name: GetGoodsReceiptItemsByPOItemID :many
SELECT *
FROM goods_receipt_items
WHERE purchase_order_item_id = $1
ORDER BY created_at ASC;

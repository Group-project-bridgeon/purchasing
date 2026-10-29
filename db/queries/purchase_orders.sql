-- name: CreatePurchaseOrder :one
INSERT INTO purchase_orders (
    po_number,
    purchase_request_id,
    vendor_id,
    status,
    order_date,
    expected_delivery_date,
    currency,
    subtotal,
    tax_amount,
    total_amount,
    payment_terms,
    shipping_address,
    notes
)
VALUES (
    $1,
    $2,
    $3,
    'draft',
    COALESCE($4::timestamptz, NOW()),
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12
)
RETURNING *;


-- name: GetPurchaseOrderByID :one
SELECT *
FROM purchase_orders
WHERE id = $1;


-- name: GetPurchaseOrderByNumber :one
SELECT *
FROM purchase_orders
WHERE po_number = $1;


-- name: ListPurchaseOrders :many
SELECT *
FROM purchase_orders
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;


-- name: ListPurchaseOrdersByStatus :many
SELECT *
FROM purchase_orders
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;


-- name: ListPurchaseOrdersByVendor :many
SELECT *
FROM purchase_orders
WHERE vendor_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;


-- name: UpdatePurchaseOrder :one
UPDATE purchase_orders
SET
    expected_delivery_date = $2,
    currency = $3,
    subtotal = $4,
    tax_amount = $5,
    total_amount = $6,
    payment_terms = $7,
    shipping_address = $8,
    notes = $9,
    updated_at = NOW()
WHERE id = $1
  AND status = 'draft'
RETURNING *;


-- name: UpdatePurchaseOrderStatus :one
UPDATE purchase_orders
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: CreatePurchaseOrderItem :one
INSERT INTO purchase_order_items (
    purchase_order_id,
    item_id,
    description,
    quantity,
    unit_price,
    tax_rate,
    total_price
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;


-- name: GetPurchaseOrderItems :many
SELECT *
FROM purchase_order_items
WHERE purchase_order_id = $1
ORDER BY created_at ASC;


-- name: GetPurchaseOrderItemByID :one
SELECT *
FROM purchase_order_items
WHERE id = $1;


-- name: UpdatePurchaseOrderItem :one
UPDATE purchase_order_items
SET
    item_id = $2,
    description = $3,
    quantity = $4,
    unit_price = $5,
    tax_rate = $6,
    total_price = $7
WHERE purchase_order_items.id = $1
  AND EXISTS (
      SELECT 1
      FROM purchase_orders po
      WHERE po.id = purchase_order_items.purchase_order_id
        AND po.status = 'draft'
  )
RETURNING *;


-- name: DeletePurchaseOrderItem :exec
DELETE FROM purchase_order_items
WHERE purchase_order_items.id = $1
  AND EXISTS (
      SELECT 1
      FROM purchase_orders po
      WHERE po.id = purchase_order_items.purchase_order_id
        AND po.status = 'draft'
  );


-- name: UpdatePurchaseOrderItemFulfillment :one
UPDATE purchase_order_items
SET
    received_quantity = $2,
    billed_quantity = $3
WHERE id = $1
RETURNING *;

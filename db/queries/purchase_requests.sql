
-- name: CreatePurchaseRequest :one
INSERT INTO purchase_requests (
    request_number,
    requested_by,
    department_id,
    status,
    requested_at,
    required_by,
    purpose,
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
    $7
)
RETURNING *;


-- name: GetPurchaseRequestByID :one
SELECT *
FROM purchase_requests
WHERE id = $1;


-- name: GetPurchaseRequestByNumber :one
SELECT *
FROM purchase_requests
WHERE request_number = $1;


-- name: ListPurchaseRequests :many
SELECT *
FROM purchase_requests
ORDER BY created_at DESC
LIMIT $1
OFFSET $2;


-- name: ListPurchaseRequestsByStatus :many
SELECT *
FROM purchase_requests
WHERE status = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;


-- name: UpdatePurchaseRequest :one
UPDATE purchase_requests
SET
    required_by = $2,
    purpose = $3,
    notes = $4,
    updated_at = NOW()
WHERE id = $1
  AND status = 'draft'
RETURNING *;


-- name: SubmitPurchaseRequest :one
UPDATE purchase_requests
SET
    status = 'submitted',
    requested_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND status = 'draft'
RETURNING *;


-- name: ApprovePurchaseRequest :one
UPDATE purchase_requests
SET
    status = 'approved',
 approved_by = sqlc.arg(approved_by),
 approved_at = NOW(),
    updated_at = NOW()
WHERE id = $1
  AND status = 'submitted'
RETURNING *;


-- name: RejectPurchaseRequest :one
UPDATE purchase_requests
SET
    status = 'rejected',
    updated_at = NOW()
WHERE id = $1
  AND status = 'submitted'
RETURNING *;


-- name: CreatePurchaseRequestItem :one
INSERT INTO purchase_request_items (
    purchase_request_id,
    item_id,
    description,
    quantity,
    estimated_unit_price
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;


-- name: GetPurchaseRequestItems :many
SELECT *
FROM purchase_request_items
WHERE purchase_request_id = $1
ORDER BY created_at ASC;


-- name: GetPurchaseRequestItemByID :one
SELECT *
FROM purchase_request_items
WHERE id = $1;


-- name: UpdatePurchaseRequestItem :one
UPDATE purchase_request_items
SET
    item_id = $2,
    description = $3,
    quantity = $4,
    estimated_unit_price = $5
WHERE purchase_request_items.id = $1
  AND EXISTS (
      SELECT 1
      FROM purchase_requests pr
      WHERE pr.id = purchase_request_items.purchase_request_id
        AND pr.status = 'draft'
  )
RETURNING *;

-- name: DeletePurchaseRequestItem :execrows
DELETE FROM purchase_request_items
WHERE purchase_request_items.id = $1
  AND EXISTS (
      SELECT 1
      FROM purchase_requests pr
      WHERE pr.id = purchase_request_items.purchase_request_id
        AND pr.status = 'draft'
  );


-- name: NextDocumentNumber :one
INSERT INTO document_number_counters (
    document_type,
    year,
    last_number
)
VALUES (
    $1,
    $2,
    1
)
ON CONFLICT (document_type, year)
DO UPDATE SET
    last_number = document_number_counters.last_number + 1
RETURNING last_number;


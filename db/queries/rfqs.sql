-- name: CreateRFQ :one
INSERT INTO rfqs (
    rfq_number,
    purchase_request_id,
    title,
    status,
    submission_deadline,
    created_by,
    notes
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


-- name: GetRFQByID :one
SELECT *
FROM rfqs
WHERE id = $1;


-- name: GetRFQByNumber :one
SELECT *
FROM rfqs
WHERE rfq_number = $1;


-- name: ListRFQs :many
SELECT *
FROM rfqs
WHERE ($1::varchar IS NULL OR status = $1)
  AND ($2::uuid IS NULL OR purchase_request_id = $2)
ORDER BY created_at DESC
LIMIT $3
OFFSET $4;


-- name: CountRFQs :one
SELECT count(*)
FROM rfqs
WHERE ($1::varchar IS NULL OR status = $1)
  AND ($2::uuid IS NULL OR purchase_request_id = $2);


-- name: UpdateRFQ :one
UPDATE rfqs
SET
    title = $2,
    submission_deadline = $3,
    notes = $4,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: UpdateRFQStatus :one
UPDATE rfqs
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: CreateRFQItem :one
INSERT INTO rfq_items (
    rfq_id,
    item_id,
    description,
    target_quantity,
    target_delivery_date
)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;


-- name: GetRFQItems :many
SELECT *
FROM rfq_items
WHERE rfq_id = $1
ORDER BY created_at ASC;


-- name: GetRFQItemByID :one
SELECT *
FROM rfq_items
WHERE id = $1;


-- name: CreateVendorQuotation :one
INSERT INTO vendor_quotations (
    rfq_id,
    vendor_id,
    quote_reference,
    quotation_date,
    valid_until,
    currency,
    subtotal,
    tax_amount,
    total_amount,
    payment_terms,
    lead_time_days,
    status,
    notes
)
VALUES (
    $1,
    $2,
    $3,
    COALESCE($4::date, CURRENT_DATE),
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11,
    $12,
    $13
)
RETURNING *;


-- name: GetVendorQuotationByID :one
SELECT *
FROM vendor_quotations
WHERE id = $1;


-- name: GetVendorQuotationByRef :one
SELECT *
FROM vendor_quotations
WHERE rfq_id = $1
  AND vendor_id = $2
  AND quote_reference = $3;


-- name: ListVendorQuotationsByRFQ :many
SELECT *
FROM vendor_quotations
WHERE rfq_id = $1
ORDER BY total_amount ASC, lead_time_days ASC;


-- name: UpdateVendorQuotationStatus :one
UPDATE vendor_quotations
SET
    status = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: RejectOtherVendorQuotations :exec
UPDATE vendor_quotations
SET
    status = 'rejected',
    updated_at = NOW()
WHERE rfq_id = $1
  AND id != $2
  AND status != 'rejected';


-- name: CreateVendorQuotationItem :one
INSERT INTO vendor_quotation_items (
    vendor_quotation_id,
    rfq_item_id,
    item_id,
    description,
    offered_quantity,
    unit_price,
    tax_rate,
    total_price,
    lead_time_days
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
    $9
)
RETURNING *;


-- name: GetVendorQuotationItems :many
SELECT *
FROM vendor_quotation_items
WHERE vendor_quotation_id = $1
ORDER BY created_at ASC;

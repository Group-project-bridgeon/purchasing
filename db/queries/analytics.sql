-- name: GetVendorDeliveryMetrics :one
SELECT
    COUNT(DISTINCT gr.id) AS total_receipts,
    COUNT(DISTINCT CASE WHEN po.expected_delivery_date IS NOT NULL AND gr.received_at::date <= po.expected_delivery_date THEN gr.id END) AS on_time_receipts,
    COUNT(DISTINCT poi.id) AS total_ordered_items,
    COUNT(DISTINCT CASE WHEN poi.received_quantity >= poi.quantity THEN poi.id END) AS fulfilled_items,
    COALESCE(SUM(gri.received_quantity), 0)::numeric(18, 3) AS total_received_quantity,
    COALESCE(SUM(gri.rejected_quantity), 0)::numeric(18, 3) AS total_rejected_quantity
FROM goods_receipts gr
JOIN purchase_orders po ON po.id = gr.purchase_order_id
LEFT JOIN goods_receipt_items gri ON gri.goods_receipt_id = gr.id
LEFT JOIN purchase_order_items poi ON poi.purchase_order_id = po.id
WHERE gr.vendor_id = $1
  AND gr.status != 'cancelled'
  AND ($2::timestamptz IS NULL OR gr.received_at >= $2)
  AND ($3::timestamptz IS NULL OR gr.received_at <= $3);


-- name: GetVendorReturnMetrics :one
SELECT
    COUNT(DISTINCT pr.id) AS total_returns,
    COALESCE(SUM(pri.return_quantity), 0)::numeric(18, 3) AS total_returned_quantity,
    COALESCE(SUM(pr.total_return_amount), 0)::numeric(18, 2) AS total_returned_amount
FROM purchase_returns pr
LEFT JOIN purchase_return_items pri ON pri.purchase_return_id = pr.id
WHERE pr.vendor_id = $1
  AND pr.status IN ('approved', 'shipped', 'completed')
  AND ($2::date IS NULL OR pr.return_date >= $2)
  AND ($3::date IS NULL OR pr.return_date <= $3);


-- name: GetVendorBillingMetrics :one
SELECT
    COUNT(vb.id) AS total_bills,
    COUNT(CASE WHEN vb.match_status = 'matched' THEN 1 END) AS matched_bills,
    COUNT(CASE WHEN vb.match_status = 'discrepant' THEN 1 END) AS discrepant_bills,
    COALESCE(SUM(vb.total_amount), 0)::numeric(18, 2) AS total_invoiced_amount
FROM vendor_bills vb
WHERE vb.vendor_id = $1
  AND vb.status != 'cancelled'
  AND ($2::date IS NULL OR vb.bill_date >= $2)
  AND ($3::date IS NULL OR vb.bill_date <= $3);


-- name: GetVendorOrderSummary :one
SELECT
    COUNT(po.id) AS total_orders,
    COALESCE(SUM(po.total_amount), 0)::numeric(18, 2) AS total_committed_amount
FROM purchase_orders po
WHERE po.vendor_id = $1
  AND po.status != 'cancelled'
  AND ($2::timestamptz IS NULL OR po.order_date >= $2)
  AND ($3::timestamptz IS NULL OR po.order_date <= $3);


-- name: GetDashboardSummary :one
SELECT
    COUNT(id) AS total_orders,
    COALESCE(SUM(total_amount), 0)::numeric(18, 2) AS total_committed_spend,
    COUNT(CASE WHEN status = 'draft' THEN 1 END) AS draft_orders,
    COUNT(CASE WHEN status = 'pending_approval' THEN 1 END) AS pending_approval_orders,
    COUNT(CASE WHEN status = 'approved' THEN 1 END) AS approved_orders,
    COUNT(CASE WHEN status = 'issued' THEN 1 END) AS issued_orders,
    COUNT(CASE WHEN status = 'partially_received' THEN 1 END) AS partially_received_orders,
    COUNT(CASE WHEN status = 'received' THEN 1 END) AS received_orders,
    COUNT(CASE WHEN status = 'billed' THEN 1 END) AS billed_orders,
    COUNT(CASE WHEN status = 'closed' THEN 1 END) AS closed_orders,
    COUNT(CASE WHEN status = 'cancelled' THEN 1 END) AS cancelled_orders
FROM purchase_orders
WHERE ($1::timestamptz IS NULL OR order_date >= $1)
  AND ($2::timestamptz IS NULL OR order_date <= $2);


-- name: GetDashboardLiabilities :one
SELECT
    COALESCE(SUM(GREATEST(0, poi.received_quantity - poi.billed_quantity) * poi.unit_price), 0)::numeric(18, 2) AS unbilled_receiving_liability,
    COALESCE(SUM(GREATEST(0, poi.quantity - poi.received_quantity) * poi.unit_price), 0)::numeric(18, 2) AS pending_delivery_exposure
FROM purchase_order_items poi
JOIN purchase_orders po ON po.id = poi.purchase_order_id
WHERE po.status IN ('issued', 'partially_received', 'received')
  AND ($1::timestamptz IS NULL OR po.order_date >= $1)
  AND ($2::timestamptz IS NULL OR po.order_date <= $2);


-- name: GetSpendByVendor :many
SELECT
    po.vendor_id,
    COUNT(DISTINCT po.id) AS total_orders,
    COALESCE(SUM(po.total_amount), 0)::numeric(18, 2) AS committed_spend
FROM purchase_orders po
WHERE po.status != 'cancelled'
  AND ($1::timestamptz IS NULL OR po.order_date >= $1)
  AND ($2::timestamptz IS NULL OR po.order_date <= $2)
GROUP BY po.vendor_id
ORDER BY committed_spend DESC
LIMIT $3;


-- name: ListActiveVendors :many
SELECT DISTINCT vendor_id
FROM purchase_orders
WHERE status != 'cancelled'
ORDER BY vendor_id;

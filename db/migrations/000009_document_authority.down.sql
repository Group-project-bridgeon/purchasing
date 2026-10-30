BEGIN;
DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM purchase_orders WHERE created_by IS NOT NULL OR approved_by IS NOT NULL)
 OR EXISTS(SELECT 1 FROM vendor_bills WHERE created_by IS NOT NULL OR approved_by IS NOT NULL)
 OR EXISTS(SELECT 1 FROM purchase_requests WHERE approved_by IS NOT NULL)
 OR EXISTS(SELECT 1 FROM purchase_returns WHERE approved_by IS NOT NULL) THEN
  RAISE EXCEPTION 'document authority history requires reviewed forward recovery';
 END IF;
END $$;
DROP TRIGGER purchasing_request_item_authority ON purchase_request_items;
DROP TRIGGER purchasing_order_item_authority ON purchase_order_items;
DROP FUNCTION purchasing_guard_item();
DROP TRIGGER purchasing_request_authority ON purchase_requests;
DROP TRIGGER purchasing_order_authority ON purchase_orders;
DROP TRIGGER purchasing_bill_authority ON vendor_bills;
DROP TRIGGER purchasing_return_authority ON purchase_returns;
DROP TRIGGER purchasing_rfq_authority ON rfqs;
DROP FUNCTION purchasing_guard_document();
ALTER TABLE purchase_requests DROP COLUMN approved_by,DROP COLUMN approved_at;
ALTER TABLE purchase_orders DROP COLUMN created_by,DROP COLUMN approved_by,DROP COLUMN approved_at;
ALTER TABLE vendor_bills DROP COLUMN created_by,DROP COLUMN approved_by,DROP COLUMN approved_at;
ALTER TABLE purchase_returns DROP COLUMN approved_by,DROP COLUMN approved_at;
COMMIT;

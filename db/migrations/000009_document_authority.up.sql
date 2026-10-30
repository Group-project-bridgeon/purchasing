BEGIN;

-- Existing facts remain untouched. Unknown historical creators stay NULL.
ALTER TABLE purchase_requests ADD COLUMN approved_by UUID, ADD COLUMN approved_at TIMESTAMPTZ;
ALTER TABLE purchase_orders ADD COLUMN created_by UUID, ADD COLUMN approved_by UUID, ADD COLUMN approved_at TIMESTAMPTZ;
ALTER TABLE vendor_bills ADD COLUMN created_by UUID, ADD COLUMN approved_by UUID, ADD COLUMN approved_at TIMESTAMPTZ;
ALTER TABLE purchase_returns ADD COLUMN approved_by UUID, ADD COLUMN approved_at TIMESTAMPTZ;

CREATE FUNCTION purchasing_guard_document() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE creator UUID; old_creator UUID; allowed BOOLEAN; content_new JSONB; content_old JSONB;
BEGIN
  IF TG_TABLE_NAME='purchase_requests' THEN creator:=NEW.requested_by;
  ELSE creator:=NEW.created_by; END IF;
  IF TG_OP='INSERT' THEN
    IF creator IS NULL OR NEW.status<>'draft' THEN
      RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_creator_required',MESSAGE='document creator and draft state required';
    END IF;
    RETURN NEW;
  END IF;
  IF TG_TABLE_NAME='purchase_requests' THEN old_creator:=OLD.requested_by;
  ELSE old_creator:=OLD.created_by; END IF;
  IF NEW.id IS DISTINCT FROM OLD.id OR creator IS DISTINCT FROM old_creator THEN
    RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_creator_immutable',MESSAGE='document identity is immutable';
  END IF;
  IF TG_TABLE_NAME<>'rfqs' THEN
    IF OLD.approved_by IS NOT NULL AND (NEW.approved_by IS DISTINCT FROM OLD.approved_by OR NEW.approved_at IS DISTINCT FROM OLD.approved_at) THEN
      RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_approval_immutable',MESSAGE='approval facts are immutable';
    END IF;
    IF NEW.status='approved' AND OLD.status<>'approved' THEN
      IF TG_TABLE_NAME='purchase_requests' THEN allowed:=OLD.status='submitted';
      ELSIF TG_TABLE_NAME='purchase_orders' THEN allowed:=OLD.status IN('draft','pending_approval');
      ELSIF TG_TABLE_NAME='vendor_bills' THEN allowed:=OLD.status='draft' AND OLD.match_status='matched';
      ELSE allowed:=OLD.status='draft'; END IF;
      IF NOT allowed OR creator IS NULL OR NEW.approved_by IS NULL THEN
        RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_approval_state',MESSAGE='approval state or creator is unavailable';
      END IF;
      IF creator=NEW.approved_by THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='creator cannot approve own document';
      END IF;
      NEW.approved_at:=clock_timestamp();
    ELSIF NEW.approved_by IS DISTINCT FROM OLD.approved_by OR NEW.approved_at IS DISTINCT FROM OLD.approved_at THEN
      RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_approval_transition',MESSAGE='approval metadata requires approval transition';
    END IF;
  END IF;
  IF NEW.status IS DISTINCT FROM OLD.status AND (OLD.status IN('cancelled','closed','rejected','paid','completed') OR NEW.status='draft' OR (NEW.status='pending_approval' AND OLD.status<>'draft')) THEN
    RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_terminal_document',MESSAGE='document transition is not permitted';
  END IF;
  content_new:=to_jsonb(NEW)-ARRAY['status','updated_at','approved_by','approved_at'];
  content_old:=to_jsonb(OLD)-ARRAY['status','updated_at','approved_by','approved_at'];
  -- Shipping evidence is a dedicated approved action, not a header edit.
  IF TG_TABLE_NAME='purchase_returns' AND OLD.status='approved' AND NEW.status='shipped' THEN
    content_new:=content_new-ARRAY['carrier','tracking_number'];content_old:=content_old-ARRAY['carrier','tracking_number'];
  END IF;
  -- Bill matching has its own action; it can update derived matching evidence.
  IF TG_TABLE_NAME='vendor_bills' THEN
    content_new:=content_new-ARRAY['match_status','discrepancy_reason'];content_old:=content_old-ARRAY['match_status','discrepancy_reason'];
  END IF;
  IF OLD.status<>'draft' AND content_new IS DISTINCT FROM content_old THEN
    RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_document_not_editable',MESSAGE='document content is not editable';
  END IF;
  RETURN NEW;
END $$;

CREATE TRIGGER purchasing_request_authority BEFORE INSERT OR UPDATE ON purchase_requests FOR EACH ROW EXECUTE FUNCTION purchasing_guard_document();
CREATE TRIGGER purchasing_order_authority BEFORE INSERT OR UPDATE ON purchase_orders FOR EACH ROW EXECUTE FUNCTION purchasing_guard_document();
CREATE TRIGGER purchasing_bill_authority BEFORE INSERT OR UPDATE ON vendor_bills FOR EACH ROW EXECUTE FUNCTION purchasing_guard_document();
CREATE TRIGGER purchasing_return_authority BEFORE INSERT OR UPDATE ON purchase_returns FOR EACH ROW EXECUTE FUNCTION purchasing_guard_document();
CREATE TRIGGER purchasing_rfq_authority BEFORE INSERT OR UPDATE ON rfqs FOR EACH ROW EXECUTE FUNCTION purchasing_guard_document();

CREATE FUNCTION purchasing_guard_item() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE parent_id UUID; previous_parent UUID; parent_status TEXT;
BEGIN
  IF TG_TABLE_NAME='purchase_request_items' THEN
    IF TG_OP='DELETE' THEN parent_id:=OLD.purchase_request_id; ELSE parent_id:=NEW.purchase_request_id; END IF;
    IF TG_OP='UPDATE' THEN previous_parent:=OLD.purchase_request_id; END IF;
    SELECT status INTO parent_status FROM purchase_requests WHERE id=parent_id FOR UPDATE;
  ELSE
    IF TG_OP='DELETE' THEN parent_id:=OLD.purchase_order_id; ELSE parent_id:=NEW.purchase_order_id; END IF;
    IF TG_OP='UPDATE' THEN previous_parent:=OLD.purchase_order_id; END IF;
    SELECT status INTO parent_status FROM purchase_orders WHERE id=parent_id FOR UPDATE;
  END IF;
  -- A cascading delete can observe its already-deleted parent. No public parent
  -- delete endpoint exists; this preserves existing FK cleanup semantics.
  IF NOT FOUND AND TG_OP='DELETE' THEN RETURN OLD; END IF;
  IF previous_parent IS NOT NULL AND previous_parent<>parent_id THEN
    RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_item_parent_immutable',MESSAGE='item parent is immutable';
  END IF;
  -- Existing receipt/billing workflows adjust accounting quantities only.
  IF TG_TABLE_NAME='purchase_order_items' AND TG_OP='UPDATE' AND
    (to_jsonb(NEW)-ARRAY['received_quantity','billed_quantity'])=(to_jsonb(OLD)-ARRAY['received_quantity','billed_quantity']) THEN RETURN NEW; END IF;
  IF parent_status IS DISTINCT FROM 'draft' THEN
    RAISE EXCEPTION USING ERRCODE='23514',CONSTRAINT='purchasing_item_not_editable',MESSAGE='items require an editable draft parent';
  END IF;
  IF TG_OP='DELETE' THEN RETURN OLD; END IF;
  RETURN NEW;
END $$;
CREATE TRIGGER purchasing_request_item_authority BEFORE INSERT OR UPDATE OR DELETE ON purchase_request_items FOR EACH ROW EXECUTE FUNCTION purchasing_guard_item();
CREATE TRIGGER purchasing_order_item_authority BEFORE INSERT OR UPDATE OR DELETE ON purchase_order_items FOR EACH ROW EXECUTE FUNCTION purchasing_guard_item();

COMMIT;

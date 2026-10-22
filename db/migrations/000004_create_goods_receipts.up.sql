-- Goods receipts table
CREATE TABLE goods_receipts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    grn_number VARCHAR(50) NOT NULL UNIQUE,

    purchase_order_id UUID NOT NULL,

    vendor_id UUID NOT NULL,

    received_by UUID NOT NULL,

    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    status VARCHAR(20) NOT NULL DEFAULT 'completed',

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT goods_receipts_purchase_order_fk
        FOREIGN KEY (purchase_order_id)
        REFERENCES purchase_orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT goods_receipts_status_check
        CHECK (
            status IN (
                'draft',
                'completed',
                'cancelled'
            )
        )
);

-- Goods receipt items table
CREATE TABLE goods_receipt_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    goods_receipt_id UUID NOT NULL,

    purchase_order_item_id UUID NOT NULL,

    item_id UUID NOT NULL,

    received_quantity NUMERIC(18, 3) NOT NULL,

    rejected_quantity NUMERIC(18, 3) NOT NULL DEFAULT 0,

    rejection_reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT goods_receipt_items_receipt_fk
        FOREIGN KEY (goods_receipt_id)
        REFERENCES goods_receipts(id)
        ON DELETE CASCADE,

    CONSTRAINT goods_receipt_items_po_item_fk
        FOREIGN KEY (purchase_order_item_id)
        REFERENCES purchase_order_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT goods_receipt_items_received_quantity_check
        CHECK (received_quantity >= 0),

    CONSTRAINT goods_receipt_items_rejected_quantity_check
        CHECK (rejected_quantity >= 0),

    CONSTRAINT goods_receipt_items_total_quantity_check
        CHECK (received_quantity + rejected_quantity > 0)
);

-- Indexes
CREATE INDEX idx_goods_receipts_po_id
    ON goods_receipts(purchase_order_id);

CREATE INDEX idx_goods_receipts_vendor_id
    ON goods_receipts(vendor_id);

CREATE INDEX idx_goods_receipts_received_at
    ON goods_receipts(received_at);

CREATE INDEX idx_goods_receipts_status
    ON goods_receipts(status);

CREATE INDEX idx_goods_receipt_items_grn_id
    ON goods_receipt_items(goods_receipt_id);

CREATE INDEX idx_goods_receipt_items_po_item_id
    ON goods_receipt_items(purchase_order_item_id);

CREATE INDEX idx_goods_receipt_items_item_id
    ON goods_receipt_items(item_id);

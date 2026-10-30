-- Purchase Returns table
CREATE TABLE purchase_returns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    return_number VARCHAR(50) NOT NULL UNIQUE,

    purchase_order_id UUID NOT NULL,

    goods_receipt_id UUID,

    vendor_id UUID NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'draft',

    return_date DATE NOT NULL DEFAULT CURRENT_DATE,

    total_return_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,

    reason TEXT NOT NULL,

    carrier VARCHAR(100),

    tracking_number VARCHAR(100),

    created_by UUID NOT NULL,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT purchase_returns_purchase_order_fk
        FOREIGN KEY (purchase_order_id)
        REFERENCES purchase_orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT purchase_returns_goods_receipt_fk
        FOREIGN KEY (goods_receipt_id)
        REFERENCES goods_receipts(id)
        ON DELETE SET NULL,

    CONSTRAINT purchase_returns_status_check
        CHECK (
            status IN (
                'draft',
                'approved',
                'shipped',
                'completed',
                'cancelled'
            )
        ),

    CONSTRAINT purchase_returns_total_amount_check
        CHECK (total_return_amount >= 0)
);

-- Purchase Return Items table
CREATE TABLE purchase_return_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    purchase_return_id UUID NOT NULL,

    purchase_order_item_id UUID NOT NULL,

    item_id UUID NOT NULL,

    description TEXT NOT NULL,

    return_quantity NUMERIC(18, 3) NOT NULL,

    unit_price NUMERIC(18, 2) NOT NULL,

    total_price NUMERIC(18, 2) NOT NULL,

    reason TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT purchase_return_items_return_fk
        FOREIGN KEY (purchase_return_id)
        REFERENCES purchase_returns(id)
        ON DELETE CASCADE,

    CONSTRAINT purchase_return_items_po_item_fk
        FOREIGN KEY (purchase_order_item_id)
        REFERENCES purchase_order_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT purchase_return_items_return_quantity_check
        CHECK (return_quantity > 0),

    CONSTRAINT purchase_return_items_unit_price_check
        CHECK (unit_price >= 0),

    CONSTRAINT purchase_return_items_total_price_check
        CHECK (total_price >= 0)
);

-- Indexes
CREATE INDEX idx_purchase_returns_po_id
    ON purchase_returns(purchase_order_id);

CREATE INDEX idx_purchase_returns_grn_id
    ON purchase_returns(goods_receipt_id);

CREATE INDEX idx_purchase_returns_vendor_id
    ON purchase_returns(vendor_id);

CREATE INDEX idx_purchase_returns_status
    ON purchase_returns(status);

CREATE INDEX idx_purchase_returns_return_date
    ON purchase_returns(return_date);

CREATE INDEX idx_purchase_return_items_return_id
    ON purchase_return_items(purchase_return_id);

CREATE INDEX idx_purchase_return_items_po_item_id
    ON purchase_return_items(purchase_order_item_id);

CREATE INDEX idx_purchase_return_items_item_id
    ON purchase_return_items(item_id);

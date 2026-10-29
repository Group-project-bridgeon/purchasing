-- Purchase Orders table
CREATE TABLE purchase_orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    po_number VARCHAR(50) NOT NULL UNIQUE,

    purchase_request_id UUID,

    vendor_id UUID NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'draft',

    order_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    expected_delivery_date DATE,

    currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    subtotal NUMERIC(18, 2) NOT NULL DEFAULT 0,

    tax_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,

    total_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,

    payment_terms VARCHAR(100),

    shipping_address TEXT,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT purchase_orders_purchase_request_fk
        FOREIGN KEY (purchase_request_id)
        REFERENCES purchase_requests(id)
        ON DELETE SET NULL,

    CONSTRAINT purchase_orders_status_check
        CHECK (
            status IN (
                'draft',
                'pending_approval',
                'approved',
                'issued',
                'partially_received',
                'received',
                'billed',
                'closed',
                'cancelled'
            )
        ),

    CONSTRAINT purchase_orders_subtotal_check
        CHECK (subtotal >= 0),

    CONSTRAINT purchase_orders_tax_amount_check
        CHECK (tax_amount >= 0),

    CONSTRAINT purchase_orders_total_amount_check
        CHECK (total_amount >= 0)
);

-- Purchase Order Items table
CREATE TABLE purchase_order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    purchase_order_id UUID NOT NULL,

    item_id UUID NOT NULL,

    description TEXT NOT NULL,

    quantity NUMERIC(18, 3) NOT NULL,

    unit_price NUMERIC(18, 2) NOT NULL,

    tax_rate NUMERIC(5, 2) NOT NULL DEFAULT 0,

    total_price NUMERIC(18, 2) NOT NULL,

    received_quantity NUMERIC(18, 3) NOT NULL DEFAULT 0,

    billed_quantity NUMERIC(18, 3) NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT purchase_order_items_po_fk
        FOREIGN KEY (purchase_order_id)
        REFERENCES purchase_orders(id)
        ON DELETE CASCADE,

    CONSTRAINT purchase_order_items_quantity_check
        CHECK (quantity > 0),

    CONSTRAINT purchase_order_items_unit_price_check
        CHECK (unit_price >= 0),

    CONSTRAINT purchase_order_items_tax_rate_check
        CHECK (tax_rate >= 0),

    CONSTRAINT purchase_order_items_total_price_check
        CHECK (total_price >= 0),

    CONSTRAINT purchase_order_items_received_quantity_check
        CHECK (received_quantity >= 0),

    CONSTRAINT purchase_order_items_billed_quantity_check
        CHECK (billed_quantity >= 0)
);

-- Indexes
CREATE INDEX idx_purchase_orders_vendor_id
    ON purchase_orders(vendor_id);

CREATE INDEX idx_purchase_orders_pr_id
    ON purchase_orders(purchase_request_id);

CREATE INDEX idx_purchase_orders_status
    ON purchase_orders(status);

CREATE INDEX idx_purchase_orders_order_date
    ON purchase_orders(order_date);

CREATE INDEX idx_purchase_order_items_po_id
    ON purchase_order_items(purchase_order_id);

CREATE INDEX idx_purchase_order_items_item_id
    ON purchase_order_items(item_id);

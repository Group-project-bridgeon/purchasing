-- Vendor Bills table
CREATE TABLE vendor_bills (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    internal_bill_number VARCHAR(50) NOT NULL UNIQUE,

    bill_number VARCHAR(100) NOT NULL,

    purchase_order_id UUID NOT NULL,

    vendor_id UUID NOT NULL,

    bill_date DATE NOT NULL,

    due_date DATE,

    currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    subtotal NUMERIC(18, 2) NOT NULL DEFAULT 0,

    tax_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,

    total_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,

    match_status VARCHAR(20) NOT NULL DEFAULT 'unmatched',

    status VARCHAR(20) NOT NULL DEFAULT 'draft',

    discrepancy_reason TEXT,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vendor_bills_purchase_order_fk
        FOREIGN KEY (purchase_order_id)
        REFERENCES purchase_orders(id)
        ON DELETE RESTRICT,

    CONSTRAINT vendor_bills_vendor_invoice_unique
        UNIQUE (vendor_id, bill_number),

    CONSTRAINT vendor_bills_match_status_check
        CHECK (
            match_status IN (
                'unmatched',
                'matched',
                'discrepant'
            )
        ),

    CONSTRAINT vendor_bills_status_check
        CHECK (
            status IN (
                'draft',
                'approved',
                'paid',
                'cancelled'
            )
        ),

    CONSTRAINT vendor_bills_subtotal_check
        CHECK (subtotal >= 0),

    CONSTRAINT vendor_bills_tax_amount_check
        CHECK (tax_amount >= 0),

    CONSTRAINT vendor_bills_total_amount_check
        CHECK (total_amount >= 0)
);

-- Vendor Bill Items table
CREATE TABLE vendor_bill_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    vendor_bill_id UUID NOT NULL,

    purchase_order_item_id UUID NOT NULL,

    item_id UUID NOT NULL,

    description TEXT NOT NULL,

    quantity NUMERIC(18, 3) NOT NULL,

    unit_price NUMERIC(18, 2) NOT NULL,

    tax_rate NUMERIC(5, 2) NOT NULL DEFAULT 0,

    total_price NUMERIC(18, 2) NOT NULL,

    matched_quantity NUMERIC(18, 3) NOT NULL DEFAULT 0,

    match_status VARCHAR(20) NOT NULL DEFAULT 'unmatched',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vendor_bill_items_bill_fk
        FOREIGN KEY (vendor_bill_id)
        REFERENCES vendor_bills(id)
        ON DELETE CASCADE,

    CONSTRAINT vendor_bill_items_po_item_fk
        FOREIGN KEY (purchase_order_item_id)
        REFERENCES purchase_order_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT vendor_bill_items_quantity_check
        CHECK (quantity > 0),

    CONSTRAINT vendor_bill_items_unit_price_check
        CHECK (unit_price >= 0),

    CONSTRAINT vendor_bill_items_tax_rate_check
        CHECK (tax_rate >= 0),

    CONSTRAINT vendor_bill_items_total_price_check
        CHECK (total_price >= 0),

    CONSTRAINT vendor_bill_items_matched_quantity_check
        CHECK (matched_quantity >= 0),

    CONSTRAINT vendor_bill_items_match_status_check
        CHECK (
            match_status IN (
                'unmatched',
                'matched',
                'price_variance',
                'quantity_variance'
            )
        )
);

-- Indexes
CREATE INDEX idx_vendor_bills_po_id
    ON vendor_bills(purchase_order_id);

CREATE INDEX idx_vendor_bills_vendor_id
    ON vendor_bills(vendor_id);

CREATE INDEX idx_vendor_bills_status
    ON vendor_bills(status);

CREATE INDEX idx_vendor_bills_match_status
    ON vendor_bills(match_status);

CREATE INDEX idx_vendor_bills_bill_date
    ON vendor_bills(bill_date);

CREATE INDEX idx_vendor_bill_items_bill_id
    ON vendor_bill_items(vendor_bill_id);

CREATE INDEX idx_vendor_bill_items_po_item_id
    ON vendor_bill_items(purchase_order_item_id);

CREATE INDEX idx_vendor_bill_items_item_id
    ON vendor_bill_items(item_id);

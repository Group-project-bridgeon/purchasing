-- RFQs table
CREATE TABLE rfqs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    rfq_number VARCHAR(50) NOT NULL UNIQUE,

    purchase_request_id UUID,

    title VARCHAR(255) NOT NULL,

    status VARCHAR(20) NOT NULL DEFAULT 'draft',

    submission_deadline TIMESTAMPTZ NOT NULL,

    created_by UUID NOT NULL,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT rfqs_purchase_request_fk
        FOREIGN KEY (purchase_request_id)
        REFERENCES purchase_requests(id)
        ON DELETE SET NULL,

    CONSTRAINT rfqs_status_check
        CHECK (
            status IN (
                'draft',
                'published',
                'closed',
                'cancelled'
            )
        )
);

-- RFQ Items table
CREATE TABLE rfq_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    rfq_id UUID NOT NULL,

    item_id UUID NOT NULL,

    description TEXT NOT NULL,

    target_quantity NUMERIC(18, 3) NOT NULL,

    target_delivery_date DATE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT rfq_items_rfq_fk
        FOREIGN KEY (rfq_id)
        REFERENCES rfqs(id)
        ON DELETE CASCADE,

    CONSTRAINT rfq_items_target_quantity_check
        CHECK (target_quantity > 0)
);

-- Vendor Quotations table
CREATE TABLE vendor_quotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    rfq_id UUID NOT NULL,

    vendor_id UUID NOT NULL,

    quote_reference VARCHAR(100) NOT NULL,

    quotation_date DATE NOT NULL DEFAULT CURRENT_DATE,

    valid_until DATE NOT NULL,

    currency VARCHAR(3) NOT NULL DEFAULT 'USD',

    subtotal NUMERIC(18, 2) NOT NULL DEFAULT 0,

    tax_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,

    total_amount NUMERIC(18, 2) NOT NULL DEFAULT 0,

    payment_terms VARCHAR(100),

    lead_time_days INT NOT NULL DEFAULT 0,

    status VARCHAR(20) NOT NULL DEFAULT 'submitted',

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vendor_quotations_rfq_fk
        FOREIGN KEY (rfq_id)
        REFERENCES rfqs(id)
        ON DELETE CASCADE,

    CONSTRAINT vendor_quotations_rfq_vendor_unique
        UNIQUE (rfq_id, vendor_id, quote_reference),

    CONSTRAINT vendor_quotations_status_check
        CHECK (
            status IN (
                'draft',
                'submitted',
                'awarded',
                'rejected'
            )
        ),

    CONSTRAINT vendor_quotations_subtotal_check
        CHECK (subtotal >= 0),

    CONSTRAINT vendor_quotations_tax_amount_check
        CHECK (tax_amount >= 0),

    CONSTRAINT vendor_quotations_total_amount_check
        CHECK (total_amount >= 0),

    CONSTRAINT vendor_quotations_lead_time_check
        CHECK (lead_time_days >= 0)
);

-- Vendor Quotation Items table
CREATE TABLE vendor_quotation_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    vendor_quotation_id UUID NOT NULL,

    rfq_item_id UUID NOT NULL,

    item_id UUID NOT NULL,

    description TEXT NOT NULL,

    offered_quantity NUMERIC(18, 3) NOT NULL,

    unit_price NUMERIC(18, 2) NOT NULL,

    tax_rate NUMERIC(5, 2) NOT NULL DEFAULT 0,

    total_price NUMERIC(18, 2) NOT NULL,

    lead_time_days INT DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT vendor_quotation_items_quotation_fk
        FOREIGN KEY (vendor_quotation_id)
        REFERENCES vendor_quotations(id)
        ON DELETE CASCADE,

    CONSTRAINT vendor_quotation_items_rfq_item_fk
        FOREIGN KEY (rfq_item_id)
        REFERENCES rfq_items(id)
        ON DELETE RESTRICT,

    CONSTRAINT vendor_quotation_items_offered_quantity_check
        CHECK (offered_quantity > 0),

    CONSTRAINT vendor_quotation_items_unit_price_check
        CHECK (unit_price >= 0),

    CONSTRAINT vendor_quotation_items_tax_rate_check
        CHECK (tax_rate >= 0),

    CONSTRAINT vendor_quotation_items_total_price_check
        CHECK (total_price >= 0),

    CONSTRAINT vendor_quotation_items_lead_time_check
        CHECK (lead_time_days >= 0)
);

-- Indexes
CREATE INDEX idx_rfqs_status
    ON rfqs(status);

CREATE INDEX idx_rfqs_purchase_request_id
    ON rfqs(purchase_request_id);

CREATE INDEX idx_rfqs_submission_deadline
    ON rfqs(submission_deadline);

CREATE INDEX idx_rfq_items_rfq_id
    ON rfq_items(rfq_id);

CREATE INDEX idx_vendor_quotations_rfq_id
    ON vendor_quotations(rfq_id);

CREATE INDEX idx_vendor_quotations_vendor_id
    ON vendor_quotations(vendor_id);

CREATE INDEX idx_vendor_quotations_status
    ON vendor_quotations(status);

CREATE INDEX idx_vendor_quotation_items_quotation_id
    ON vendor_quotation_items(vendor_quotation_id);

CREATE INDEX idx_vendor_quotation_items_rfq_item_id
    ON vendor_quotation_items(rfq_item_id);

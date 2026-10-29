-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Purchase requests
CREATE TABLE purchase_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    request_number VARCHAR(50) NOT NULL UNIQUE,

    requested_by UUID NOT NULL,

    department_id UUID,

    status VARCHAR(20) NOT NULL DEFAULT 'draft',

    requested_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    required_by DATE,

    purpose TEXT NOT NULL,

    notes TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT purchase_requests_status_check
        CHECK (
            status IN (
                'draft',
                'submitted',
                'approved',
                'rejected'
            )
        ),

    CONSTRAINT purchase_requests_required_by_check
        CHECK (
            required_by IS NULL
            OR required_by >= requested_at::DATE
        )
);

-- Purchase request items
CREATE TABLE purchase_request_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    purchase_request_id UUID NOT NULL,

    item_id UUID NOT NULL,

    description TEXT NOT NULL,

    quantity NUMERIC(18, 3) NOT NULL,

    estimated_unit_price NUMERIC(18, 2),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT purchase_request_items_request_fk
        FOREIGN KEY (purchase_request_id)
        REFERENCES purchase_requests(id)
        ON DELETE CASCADE,

    CONSTRAINT purchase_request_items_quantity_check
        CHECK (quantity > 0),

    CONSTRAINT purchase_request_items_price_check
        CHECK (
            estimated_unit_price IS NULL
            OR estimated_unit_price >= 0
        )
);

-- Indexes
CREATE INDEX idx_purchase_requests_requested_by
    ON purchase_requests(requested_by);

CREATE INDEX idx_purchase_requests_department_id
    ON purchase_requests(department_id);

CREATE INDEX idx_purchase_requests_status
    ON purchase_requests(status);

CREATE INDEX idx_purchase_requests_requested_at
    ON purchase_requests(requested_at);

CREATE INDEX idx_purchase_request_items_request_id
    ON purchase_request_items(purchase_request_id);

CREATE INDEX idx_purchase_request_items_item_id
    ON purchase_request_items(item_id);
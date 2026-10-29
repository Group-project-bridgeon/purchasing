--
-- PostgreSQL database dump
--

\restrict IB9bXG7F21xVUTzgcuXFMbUSCBFdd6vcbIeQAWSUJp0tSHBerLiLFZKKqJmnbQr

-- Dumped from database version 16.15 (Debian 16.15-1.pgdg13+2)
-- Dumped by pg_dump version 16.15 (Debian 16.15-1.pgdg13+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pgcrypto; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pgcrypto WITH SCHEMA public;


--
-- Name: EXTENSION pgcrypto; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION pgcrypto IS 'cryptographic functions';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: document_number_counters; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.document_number_counters (
    document_type character varying(50) NOT NULL,
    year integer NOT NULL,
    last_number bigint DEFAULT 0 NOT NULL,
    CONSTRAINT document_number_counters_last_number_check CHECK ((last_number >= 0)),
    CONSTRAINT document_number_counters_year_check CHECK ((year >= 2000))
);


ALTER TABLE public.document_number_counters OWNER TO postgres;

--
-- Name: purchase_request_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.purchase_request_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    purchase_request_id uuid NOT NULL,
    item_id uuid NOT NULL,
    description text NOT NULL,
    quantity numeric(18,3) NOT NULL,
    estimated_unit_price numeric(18,2),
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT purchase_request_items_price_check CHECK (((estimated_unit_price IS NULL) OR (estimated_unit_price >= (0)::numeric))),
    CONSTRAINT purchase_request_items_quantity_check CHECK ((quantity > (0)::numeric))
);


ALTER TABLE public.purchase_request_items OWNER TO postgres;

--
-- Name: purchase_requests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.purchase_requests (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    request_number character varying(50) NOT NULL,
    requested_by uuid NOT NULL,
    department_id uuid,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    requested_at timestamp with time zone DEFAULT now() NOT NULL,
    required_by date,
    purpose text NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT purchase_requests_required_by_check CHECK (((required_by IS NULL) OR (required_by >= (requested_at)::date))),
    CONSTRAINT purchase_requests_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'submitted'::character varying, 'approved'::character varying, 'rejected'::character varying])::text[])))
);


ALTER TABLE public.purchase_requests OWNER TO postgres;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO postgres;

--
-- Name: document_number_counters document_number_counters_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.document_number_counters
    ADD CONSTRAINT document_number_counters_pkey PRIMARY KEY (document_type, year);


--
-- Name: purchase_request_items purchase_request_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_request_items
    ADD CONSTRAINT purchase_request_items_pkey PRIMARY KEY (id);


--
-- Name: purchase_requests purchase_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_requests
    ADD CONSTRAINT purchase_requests_pkey PRIMARY KEY (id);


--
-- Name: purchase_requests purchase_requests_request_number_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_requests
    ADD CONSTRAINT purchase_requests_request_number_key UNIQUE (request_number);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: idx_purchase_request_items_item_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_purchase_request_items_item_id ON public.purchase_request_items USING btree (item_id);


--
-- Name: idx_purchase_request_items_request_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_purchase_request_items_request_id ON public.purchase_request_items USING btree (purchase_request_id);


--
-- Name: idx_purchase_requests_department_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_purchase_requests_department_id ON public.purchase_requests USING btree (department_id);


--
-- Name: idx_purchase_requests_requested_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_purchase_requests_requested_at ON public.purchase_requests USING btree (requested_at);


--
-- Name: idx_purchase_requests_requested_by; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_purchase_requests_requested_by ON public.purchase_requests USING btree (requested_by);


--
-- Name: idx_purchase_requests_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_purchase_requests_status ON public.purchase_requests USING btree (status);


--
-- Name: purchase_request_items purchase_request_items_request_fk; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.purchase_request_items
    ADD CONSTRAINT purchase_request_items_request_fk FOREIGN KEY (purchase_request_id) REFERENCES public.purchase_requests(id) ON DELETE CASCADE;


--
-- Name: purchase_orders; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.purchase_orders (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    po_number character varying(50) NOT NULL,
    purchase_request_id uuid,
    vendor_id uuid NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    order_date timestamp with time zone DEFAULT now() NOT NULL,
    expected_delivery_date date,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    subtotal numeric(18,2) DEFAULT 0 NOT NULL,
    tax_amount numeric(18,2) DEFAULT 0 NOT NULL,
    total_amount numeric(18,2) DEFAULT 0 NOT NULL,
    payment_terms character varying(100),
    shipping_address text,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT purchase_orders_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'pending_approval'::character varying, 'approved'::character varying, 'issued'::character varying, 'partially_received'::character varying, 'received'::character varying, 'billed'::character varying, 'closed'::character varying, 'cancelled'::character varying])::text[]))),
    CONSTRAINT purchase_orders_subtotal_check CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT purchase_orders_tax_amount_check CHECK ((tax_amount >= (0)::numeric)),
    CONSTRAINT purchase_orders_total_amount_check CHECK ((total_amount >= (0)::numeric))
);

ALTER TABLE public.purchase_orders OWNER TO postgres;

--
-- Name: purchase_order_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.purchase_order_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    purchase_order_id uuid NOT NULL,
    item_id uuid NOT NULL,
    description text NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit_price numeric(18,2) NOT NULL,
    tax_rate numeric(5,2) DEFAULT 0 NOT NULL,
    total_price numeric(18,2) NOT NULL,
    received_quantity numeric(18,3) DEFAULT 0 NOT NULL,
    billed_quantity numeric(18,3) DEFAULT 0 NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT purchase_order_items_quantity_check CHECK ((quantity > (0)::numeric)),
    CONSTRAINT purchase_order_items_unit_price_check CHECK ((unit_price >= (0)::numeric)),
    CONSTRAINT purchase_order_items_tax_rate_check CHECK ((tax_rate >= (0)::numeric)),
    CONSTRAINT purchase_order_items_total_price_check CHECK ((total_price >= (0)::numeric)),
    CONSTRAINT purchase_order_items_received_quantity_check CHECK ((received_quantity >= (0)::numeric)),
    CONSTRAINT purchase_order_items_billed_quantity_check CHECK ((billed_quantity >= (0)::numeric))
);

ALTER TABLE public.purchase_order_items OWNER TO postgres;

ALTER TABLE ONLY public.purchase_orders
    ADD CONSTRAINT purchase_orders_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.purchase_orders
    ADD CONSTRAINT purchase_orders_po_number_key UNIQUE (po_number);

ALTER TABLE ONLY public.purchase_order_items
    ADD CONSTRAINT purchase_order_items_pkey PRIMARY KEY (id);

CREATE INDEX idx_purchase_orders_vendor_id ON public.purchase_orders USING btree (vendor_id);
CREATE INDEX idx_purchase_orders_pr_id ON public.purchase_orders USING btree (purchase_request_id);
CREATE INDEX idx_purchase_orders_status ON public.purchase_orders USING btree (status);
CREATE INDEX idx_purchase_orders_order_date ON public.purchase_orders USING btree (order_date);
CREATE INDEX idx_purchase_order_items_po_id ON public.purchase_order_items USING btree (purchase_order_id);
CREATE INDEX idx_purchase_order_items_item_id ON public.purchase_order_items USING btree (item_id);

ALTER TABLE ONLY public.purchase_orders
    ADD CONSTRAINT purchase_orders_purchase_request_fk FOREIGN KEY (purchase_request_id) REFERENCES public.purchase_requests(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.purchase_order_items
    ADD CONSTRAINT purchase_order_items_po_fk FOREIGN KEY (purchase_order_id) REFERENCES public.purchase_orders(id) ON DELETE CASCADE;

--
-- Name: goods_receipts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.goods_receipts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    grn_number character varying(50) NOT NULL,
    purchase_order_id uuid NOT NULL,
    vendor_id uuid NOT NULL,
    received_by uuid NOT NULL,
    received_at timestamp with time zone DEFAULT now() NOT NULL,
    status character varying(20) DEFAULT 'completed'::character varying NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT goods_receipts_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'completed'::character varying, 'cancelled'::character varying])::text[])))
);

ALTER TABLE public.goods_receipts OWNER TO postgres;

--
-- Name: goods_receipt_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.goods_receipt_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    goods_receipt_id uuid NOT NULL,
    purchase_order_item_id uuid NOT NULL,
    item_id uuid NOT NULL,
    received_quantity numeric(18,3) NOT NULL,
    rejected_quantity numeric(18,3) DEFAULT 0 NOT NULL,
    rejection_reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT goods_receipt_items_received_quantity_check CHECK ((received_quantity >= (0)::numeric)),
    CONSTRAINT goods_receipt_items_rejected_quantity_check CHECK ((rejected_quantity >= (0)::numeric)),
    CONSTRAINT goods_receipt_items_total_quantity_check CHECK (((received_quantity + rejected_quantity) > (0)::numeric))
);

ALTER TABLE public.goods_receipt_items OWNER TO postgres;

ALTER TABLE ONLY public.goods_receipts
    ADD CONSTRAINT goods_receipts_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.goods_receipts
    ADD CONSTRAINT goods_receipts_grn_number_key UNIQUE (grn_number);

ALTER TABLE ONLY public.goods_receipt_items
    ADD CONSTRAINT goods_receipt_items_pkey PRIMARY KEY (id);

CREATE INDEX idx_goods_receipts_po_id ON public.goods_receipts USING btree (purchase_order_id);
CREATE INDEX idx_goods_receipts_vendor_id ON public.goods_receipts USING btree (vendor_id);
CREATE INDEX idx_goods_receipts_received_at ON public.goods_receipts USING btree (received_at);
CREATE INDEX idx_goods_receipts_status ON public.goods_receipts USING btree (status);
CREATE INDEX idx_goods_receipt_items_grn_id ON public.goods_receipt_items USING btree (goods_receipt_id);
CREATE INDEX idx_goods_receipt_items_po_item_id ON public.goods_receipt_items USING btree (purchase_order_item_id);
CREATE INDEX idx_goods_receipt_items_item_id ON public.goods_receipt_items USING btree (item_id);

ALTER TABLE ONLY public.goods_receipts
    ADD CONSTRAINT goods_receipts_purchase_order_fk FOREIGN KEY (purchase_order_id) REFERENCES public.purchase_orders(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.goods_receipt_items
    ADD CONSTRAINT goods_receipt_items_receipt_fk FOREIGN KEY (goods_receipt_id) REFERENCES public.goods_receipts(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.goods_receipt_items
    ADD CONSTRAINT goods_receipt_items_po_item_fk FOREIGN KEY (purchase_order_item_id) REFERENCES public.purchase_order_items(id) ON DELETE RESTRICT;

--
-- Name: vendor_bills; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.vendor_bills (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    internal_bill_number character varying(50) NOT NULL,
    bill_number character varying(100) NOT NULL,
    purchase_order_id uuid NOT NULL,
    vendor_id uuid NOT NULL,
    bill_date date NOT NULL,
    due_date date,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    subtotal numeric(18,2) DEFAULT 0 NOT NULL,
    tax_amount numeric(18,2) DEFAULT 0 NOT NULL,
    total_amount numeric(18,2) DEFAULT 0 NOT NULL,
    match_status character varying(20) DEFAULT 'unmatched'::character varying NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    discrepancy_reason text,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT vendor_bills_match_status_check CHECK (((match_status)::text = ANY ((ARRAY['unmatched'::character varying, 'matched'::character varying, 'discrepant'::character varying])::text[]))),
    CONSTRAINT vendor_bills_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'approved'::character varying, 'paid'::character varying, 'cancelled'::character varying])::text[]))),
    CONSTRAINT vendor_bills_subtotal_check CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT vendor_bills_tax_amount_check CHECK ((tax_amount >= (0)::numeric)),
    CONSTRAINT vendor_bills_total_amount_check CHECK ((total_amount >= (0)::numeric))
);

ALTER TABLE public.vendor_bills OWNER TO postgres;

--
-- Name: vendor_bill_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.vendor_bill_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    vendor_bill_id uuid NOT NULL,
    purchase_order_item_id uuid NOT NULL,
    item_id uuid NOT NULL,
    description text NOT NULL,
    quantity numeric(18,3) NOT NULL,
    unit_price numeric(18,2) NOT NULL,
    tax_rate numeric(5,2) DEFAULT 0 NOT NULL,
    total_price numeric(18,2) NOT NULL,
    matched_quantity numeric(18,3) DEFAULT 0 NOT NULL,
    match_status character varying(20) DEFAULT 'unmatched'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT vendor_bill_items_quantity_check CHECK ((quantity > (0)::numeric)),
    CONSTRAINT vendor_bill_items_unit_price_check CHECK ((unit_price >= (0)::numeric)),
    CONSTRAINT vendor_bill_items_tax_rate_check CHECK ((tax_rate >= (0)::numeric)),
    CONSTRAINT vendor_bill_items_total_price_check CHECK ((total_price >= (0)::numeric)),
    CONSTRAINT vendor_bill_items_matched_quantity_check CHECK ((matched_quantity >= (0)::numeric)),
    CONSTRAINT vendor_bill_items_match_status_check CHECK (((match_status)::text = ANY ((ARRAY['unmatched'::character varying, 'matched'::character varying, 'price_variance'::character varying, 'quantity_variance'::character varying])::text[])))
);

ALTER TABLE public.vendor_bill_items OWNER TO postgres;

ALTER TABLE ONLY public.vendor_bills
    ADD CONSTRAINT vendor_bills_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.vendor_bills
    ADD CONSTRAINT vendor_bills_internal_bill_number_key UNIQUE (internal_bill_number);

ALTER TABLE ONLY public.vendor_bills
    ADD CONSTRAINT vendor_bills_vendor_invoice_unique UNIQUE (vendor_id, bill_number);

ALTER TABLE ONLY public.vendor_bill_items
    ADD CONSTRAINT vendor_bill_items_pkey PRIMARY KEY (id);

CREATE INDEX idx_vendor_bills_po_id ON public.vendor_bills USING btree (purchase_order_id);
CREATE INDEX idx_vendor_bills_vendor_id ON public.vendor_bills USING btree (vendor_id);
CREATE INDEX idx_vendor_bills_status ON public.vendor_bills USING btree (status);
CREATE INDEX idx_vendor_bills_match_status ON public.vendor_bills USING btree (match_status);
CREATE INDEX idx_vendor_bills_bill_date ON public.vendor_bills USING btree (bill_date);
CREATE INDEX idx_vendor_bill_items_bill_id ON public.vendor_bill_items USING btree (vendor_bill_id);
CREATE INDEX idx_vendor_bill_items_po_item_id ON public.vendor_bill_items USING btree (purchase_order_item_id);
CREATE INDEX idx_vendor_bill_items_item_id ON public.vendor_bill_items USING btree (item_id);

ALTER TABLE ONLY public.vendor_bills
    ADD CONSTRAINT vendor_bills_purchase_order_fk FOREIGN KEY (purchase_order_id) REFERENCES public.purchase_orders(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.vendor_bill_items
    ADD CONSTRAINT vendor_bill_items_bill_fk FOREIGN KEY (vendor_bill_id) REFERENCES public.vendor_bills(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.vendor_bill_items
    ADD CONSTRAINT vendor_bill_items_po_item_fk FOREIGN KEY (purchase_order_item_id) REFERENCES public.purchase_order_items(id) ON DELETE RESTRICT;

--
-- Name: purchase_returns; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.purchase_returns (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    return_number character varying(50) NOT NULL,
    purchase_order_id uuid NOT NULL,
    goods_receipt_id uuid,
    vendor_id uuid NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    return_date date DEFAULT CURRENT_DATE NOT NULL,
    total_return_amount numeric(18,2) DEFAULT 0 NOT NULL,
    reason text NOT NULL,
    carrier character varying(100),
    tracking_number character varying(100),
    created_by uuid NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT purchase_returns_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'approved'::character varying, 'shipped'::character varying, 'completed'::character varying, 'cancelled'::character varying])::text[]))),
    CONSTRAINT purchase_returns_total_amount_check CHECK ((total_return_amount >= (0)::numeric))
);

ALTER TABLE public.purchase_returns OWNER TO postgres;

--
-- Name: purchase_return_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.purchase_return_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    purchase_return_id uuid NOT NULL,
    purchase_order_item_id uuid NOT NULL,
    item_id uuid NOT NULL,
    description text NOT NULL,
    return_quantity numeric(18,3) NOT NULL,
    unit_price numeric(18,2) NOT NULL,
    total_price numeric(18,2) NOT NULL,
    reason text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT purchase_return_items_return_quantity_check CHECK ((return_quantity > (0)::numeric)),
    CONSTRAINT purchase_return_items_unit_price_check CHECK ((unit_price >= (0)::numeric)),
    CONSTRAINT purchase_return_items_total_price_check CHECK ((total_price >= (0)::numeric))
);

ALTER TABLE public.purchase_return_items OWNER TO postgres;

ALTER TABLE ONLY public.purchase_returns
    ADD CONSTRAINT purchase_returns_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.purchase_returns
    ADD CONSTRAINT purchase_returns_return_number_key UNIQUE (return_number);

ALTER TABLE ONLY public.purchase_return_items
    ADD CONSTRAINT purchase_return_items_pkey PRIMARY KEY (id);

CREATE INDEX idx_purchase_returns_po_id ON public.purchase_returns USING btree (purchase_order_id);
CREATE INDEX idx_purchase_returns_grn_id ON public.purchase_returns USING btree (goods_receipt_id);
CREATE INDEX idx_purchase_returns_vendor_id ON public.purchase_returns USING btree (vendor_id);
CREATE INDEX idx_purchase_returns_status ON public.purchase_returns USING btree (status);
CREATE INDEX idx_purchase_returns_return_date ON public.purchase_returns USING btree (return_date);
CREATE INDEX idx_purchase_return_items_return_id ON public.purchase_return_items USING btree (purchase_return_id);
CREATE INDEX idx_purchase_return_items_po_item_id ON public.purchase_return_items USING btree (purchase_order_item_id);
CREATE INDEX idx_purchase_return_items_item_id ON public.purchase_return_items USING btree (item_id);

ALTER TABLE ONLY public.purchase_returns
    ADD CONSTRAINT purchase_returns_purchase_order_fk FOREIGN KEY (purchase_order_id) REFERENCES public.purchase_orders(id) ON DELETE RESTRICT;

ALTER TABLE ONLY public.purchase_returns
    ADD CONSTRAINT purchase_returns_goods_receipt_fk FOREIGN KEY (goods_receipt_id) REFERENCES public.goods_receipts(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.purchase_return_items
    ADD CONSTRAINT purchase_return_items_return_fk FOREIGN KEY (purchase_return_id) REFERENCES public.purchase_returns(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.purchase_return_items
    ADD CONSTRAINT purchase_return_items_po_item_fk FOREIGN KEY (purchase_order_item_id) REFERENCES public.purchase_order_items(id) ON DELETE RESTRICT;

--
-- Name: rfqs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rfqs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rfq_number character varying(50) NOT NULL,
    purchase_request_id uuid,
    title character varying(255) NOT NULL,
    status character varying(20) DEFAULT 'draft'::character varying NOT NULL,
    submission_deadline timestamp with time zone NOT NULL,
    created_by uuid NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT rfqs_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'published'::character varying, 'closed'::character varying, 'cancelled'::character varying])::text[])))
);

ALTER TABLE public.rfqs OWNER TO postgres;

--
-- Name: rfq_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rfq_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rfq_id uuid NOT NULL,
    item_id uuid NOT NULL,
    description text NOT NULL,
    target_quantity numeric(18,3) NOT NULL,
    target_delivery_date date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT rfq_items_target_quantity_check CHECK ((target_quantity > (0)::numeric))
);

ALTER TABLE public.rfq_items OWNER TO postgres;

--
-- Name: vendor_quotations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.vendor_quotations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    rfq_id uuid NOT NULL,
    vendor_id uuid NOT NULL,
    quote_reference character varying(100) NOT NULL,
    quotation_date date DEFAULT CURRENT_DATE NOT NULL,
    valid_until date NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying NOT NULL,
    subtotal numeric(18,2) DEFAULT 0 NOT NULL,
    tax_amount numeric(18,2) DEFAULT 0 NOT NULL,
    total_amount numeric(18,2) DEFAULT 0 NOT NULL,
    payment_terms character varying(100),
    lead_time_days integer DEFAULT 0 NOT NULL,
    status character varying(20) DEFAULT 'submitted'::character varying NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT vendor_quotations_status_check CHECK (((status)::text = ANY ((ARRAY['draft'::character varying, 'submitted'::character varying, 'awarded'::character varying, 'rejected'::character varying])::text[]))),
    CONSTRAINT vendor_quotations_subtotal_check CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT vendor_quotations_tax_amount_check CHECK ((tax_amount >= (0)::numeric)),
    CONSTRAINT vendor_quotations_total_amount_check CHECK ((total_amount >= (0)::numeric)),
    CONSTRAINT vendor_quotations_lead_time_check CHECK ((lead_time_days >= 0))
);

ALTER TABLE public.vendor_quotations OWNER TO postgres;

--
-- Name: vendor_quotation_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.vendor_quotation_items (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    vendor_quotation_id uuid NOT NULL,
    rfq_item_id uuid NOT NULL,
    item_id uuid NOT NULL,
    description text NOT NULL,
    offered_quantity numeric(18,3) NOT NULL,
    unit_price numeric(18,2) NOT NULL,
    tax_rate numeric(5,2) DEFAULT 0 NOT NULL,
    total_price numeric(18,2) NOT NULL,
    lead_time_days integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT vendor_quotation_items_offered_quantity_check CHECK ((offered_quantity > (0)::numeric)),
    CONSTRAINT vendor_quotation_items_unit_price_check CHECK ((unit_price >= (0)::numeric)),
    CONSTRAINT vendor_quotation_items_tax_rate_check CHECK ((tax_rate >= (0)::numeric)),
    CONSTRAINT vendor_quotation_items_total_price_check CHECK ((total_price >= (0)::numeric)),
    CONSTRAINT vendor_quotation_items_lead_time_check CHECK ((lead_time_days >= 0))
);

ALTER TABLE public.vendor_quotation_items OWNER TO postgres;

ALTER TABLE ONLY public.rfqs
    ADD CONSTRAINT rfqs_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.rfqs
    ADD CONSTRAINT rfqs_rfq_number_key UNIQUE (rfq_number);

ALTER TABLE ONLY public.rfq_items
    ADD CONSTRAINT rfq_items_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.vendor_quotations
    ADD CONSTRAINT vendor_quotations_pkey PRIMARY KEY (id);

ALTER TABLE ONLY public.vendor_quotations
    ADD CONSTRAINT vendor_quotations_rfq_vendor_unique UNIQUE (rfq_id, vendor_id, quote_reference);

ALTER TABLE ONLY public.vendor_quotation_items
    ADD CONSTRAINT vendor_quotation_items_pkey PRIMARY KEY (id);

CREATE INDEX idx_rfqs_status ON public.rfqs USING btree (status);
CREATE INDEX idx_rfqs_purchase_request_id ON public.rfqs USING btree (purchase_request_id);
CREATE INDEX idx_rfqs_submission_deadline ON public.rfqs USING btree (submission_deadline);
CREATE INDEX idx_rfq_items_rfq_id ON public.rfq_items USING btree (rfq_id);
CREATE INDEX idx_vendor_quotations_rfq_id ON public.vendor_quotations USING btree (rfq_id);
CREATE INDEX idx_vendor_quotations_vendor_id ON public.vendor_quotations USING btree (vendor_id);
CREATE INDEX idx_vendor_quotations_status ON public.vendor_quotations USING btree (status);
CREATE INDEX idx_vendor_quotation_items_quotation_id ON public.vendor_quotation_items USING btree (vendor_quotation_id);
CREATE INDEX idx_vendor_quotation_items_rfq_item_id ON public.vendor_quotation_items USING btree (rfq_item_id);

ALTER TABLE ONLY public.rfqs
    ADD CONSTRAINT rfqs_purchase_request_fk FOREIGN KEY (purchase_request_id) REFERENCES public.purchase_requests(id) ON DELETE SET NULL;

ALTER TABLE ONLY public.rfq_items
    ADD CONSTRAINT rfq_items_rfq_fk FOREIGN KEY (rfq_id) REFERENCES public.rfqs(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.vendor_quotations
    ADD CONSTRAINT vendor_quotations_rfq_fk FOREIGN KEY (rfq_id) REFERENCES public.rfqs(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.vendor_quotation_items
    ADD CONSTRAINT vendor_quotation_items_quotation_fk FOREIGN KEY (vendor_quotation_id) REFERENCES public.vendor_quotations(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.vendor_quotation_items
    ADD CONSTRAINT vendor_quotation_items_rfq_item_fk FOREIGN KEY (rfq_item_id) REFERENCES public.rfq_items(id) ON DELETE RESTRICT;

--
-- Name: outbox_events; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.outbox_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    aggregate_type character varying(50) NOT NULL,
    aggregate_id character varying(50) NOT NULL,
    event_type character varying(100) NOT NULL,
    payload jsonb NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    retry_count integer DEFAULT 0 NOT NULL,
    last_error text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    published_at timestamp with time zone,
    CONSTRAINT outbox_events_status_check CHECK (((status)::text = ANY ((ARRAY['pending'::character varying, 'published'::character varying, 'failed'::character varying])::text[])))
);

ALTER TABLE ONLY public.outbox_events
    ADD CONSTRAINT outbox_events_pkey PRIMARY KEY (id);

CREATE INDEX idx_outbox_events_pending ON public.outbox_events (created_at ASC) WHERE status = 'pending';
CREATE INDEX idx_outbox_events_aggregate ON public.outbox_events (aggregate_type, aggregate_id);

--
-- PostgreSQL database dump complete
--

\unrestrict IB9bXG7F21xVUTzgcuXFMbUSCBFdd6vcbIeQAWSUJp0tSHBerLiLFZKKqJmnbQr


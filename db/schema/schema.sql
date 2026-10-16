--
-- PostgreSQL database dump
--

\restrict cqFZORBOCPScFfPb7lyKJGsCTilcjqEJb5ipiy1FXqaNiO9AUzrRDXd0Wsk3InR

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
-- PostgreSQL database dump complete
--

\unrestrict cqFZORBOCPScFfPb7lyKJGsCTilcjqEJb5ipiy1FXqaNiO9AUzrRDXd0Wsk3InR


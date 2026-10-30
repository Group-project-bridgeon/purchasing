# ERP Purchasing Service (Procure-to-Pay / P2P)

A high-performance, enterprise-grade **Procure-to-Pay (P2P)** microservice built in Go for managing the complete procurement lifecycle from external sellers and vendors.

---

## 1. Architectural Highlights

- **Framework**: [Fiber v2](https://github.com/gofiber/fiber) (Fast HTTP engine based on Fasthttp).
- **Database Driver**: [pgx/v5](https://github.com/jackc/pgx) with connection pooling (`pgxpool.Pool`).
- **SQL Code Generation**: [sqlc](https://sqlc.dev/) (Type-safe Go code generated directly from raw PostgreSQL queries).
- **Arbitrary Precision Decimals**: [shopspring/decimal](https://github.com/shopspring/decimal) for financial amounts, tax calculations, and fulfillment quantities (no float rounding errors).
- **Architecture**: Strict Layered Clean Architecture:
  $$\text{Domain Models} \longrightarrow \text{Repositories} \longrightarrow \text{Services} \longrightarrow \text{Handlers} \longrightarrow \text{DTOs}$$
- **Atomic Transaction Management**: Unit-of-Work transaction managers ensuring ACID consistency for multi-table updates (PO creation from PR, GRN fulfillment increments, 3-way matching updates, Return rollbacks, and RFQ quote awarding).
- **Sequential Document Numbering**: Yearly auto-resetting document counter table (`document_numbers`) ensuring unique human-readable references:
  - Purchase Requests: `PR-YYYY-XXXXX`
  - RFQs: `RFQ-YYYY-XXXXX`
  - Purchase Orders: `PO-YYYY-XXXXX`
  - Goods Receipt Notes: `GRN-YYYY-XXXXX`
  - Vendor Invoices: `BILL-YYYY-XXXXX`
  - Purchase Returns: `RET-YYYY-XXXXX`

---

## 2. End-to-End P2P Workflow

```mermaid
flowchart TD
    PR[1. Purchase Request: PR-YYYY-XXXXX] -->|Approved| SOURCING{Direct PO or Sourcing?}
    
    SOURCING -->|Multi-Vendor Bidding| RFQ[2. RFQ: RFQ-YYYY-XXXXX]
    RFQ --> VQ1[Vendor Quotation A]
    RFQ --> VQ2[Vendor Quotation B]
    RFQ --> VQ3[Vendor Quotation C]
    VQ1 --> COMP[Quotation Comparison Matrix]
    VQ2 --> COMP
    VQ3 --> COMP
    COMP -->|Award Winning Bid| PO[3. Purchase Order: PO-YYYY-XXXXX]
    
    SOURCING -->|Single Vendor| PO
    
    PO -->|Status: Issued| SHIP[Seller Dispatches Goods]
    SHIP --> GRN[4. Goods Receipt Note: GRN-YYYY-XXXXX]
    
    GRN -->|Accepted Items| PO_REC[PO: partially_received / received]
    GRN -->|Damaged / Excess Items| RET[5. Purchase Return / RTV: RET-YYYY-XXXXX]
    RET -->|Completed| ADJ[Rollback PO Received Qty & Debit Note]
    
    PO_REC --> MATCH{6. 3-Way Matching Engine}
    BILL[Vendor Bill / Invoice: BILL-YYYY-XXXXX] --> MATCH
    GRN --> MATCH
    
    MATCH -->|Price & Qty Verified| APPR[Bill: matched & approved]
    APPR --> AP[Accounts Payable Payout]
    
    PO --> AN[7. Analytics Engine: OTIF, Quality %, Liability, Spend]
    GRN --> AN
    BILL --> AN
    RET --> AN
```

---

## 3. Database Schema & Migration Architecture

Migrations are located in `db/migrations/`. Startup is read-only and requires clean golang-migrate schema 9; operator adoption rules are in [platform prerequisites](docs/platform-prerequisites.md).

| Migration | Up File | Down File | Description |
|---|---|---|---|
| **000001** | `000001_create_purchase_requests.up.sql` | `000001_create_purchase_requests.down.sql` | `purchase_requests` and `purchase_request_items` |
| **000002** | `000002_create_document_number_counters.up.sql` | `000002_create_document_number_counters.down.sql` | `document_numbers` |
| **000003** | `000003_create_purchase_orders.up.sql` | `000003_create_purchase_orders.down.sql` | `purchase_orders` & `purchase_order_items` |
| **000004** | `000004_create_goods_receipts.up.sql` | `000004_create_goods_receipts.down.sql` | `goods_receipts` & `goods_receipt_items` |
| **000005** | `000005_create_vendor_bills.up.sql` | `000005_create_vendor_bills.down.sql` | `vendor_bills` & `vendor_bill_items` (3-way match) |
| **000006** | `000006_create_purchase_returns.up.sql` | `000006_create_purchase_returns.down.sql` | `purchase_returns` & `purchase_return_items` |
| **000007** | `000007_create_rfq_and_quotations.up.sql` | `000007_create_rfq_and_quotations.down.sql` | `rfqs`, `rfq_items`, `vendor_quotations`, `vendor_quotation_items` |
| **000008** | `000008_create_outbox_events.up.sql` | `000008_create_outbox_events.down.sql` | Transactional Outbox (publisher disabled) |
| **000009** | `000009_document_authority.up.sql` | `000009_document_authority.down.sql` | Immutable creator/approval metadata, actor separation and lifecycle guards |

---

## 4. REST API Reference

### 4.0 Interactive Documentation & Health Checks

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/docs` or `/swagger` | Interactive Swagger UI API explorer |
| `GET` | `/docs/openapi.json` | Full OpenAPI 3.0.3 machine-readable JSON specification |
| `GET` | `/health` | Kubernetes / Docker liveness healthcheck (`{"status":"ok"}`) |
| `GET` | `/api/v1/health` | API version healthcheck |

### 4.1 Purchase Requests (`/api/v1/purchase-requests`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/purchase-requests` | Create PR with initial line items |
| `GET` | `/api/v1/purchase-requests` | List PRs (query: `status`, `department`, `limit`, `offset`) |
| `GET` | `/api/v1/purchase-requests/:id` | Get PR details by UUID |
| `PUT` | `/api/v1/purchase-requests/:id` | Update draft PR details |
| `POST` | `/api/v1/purchase-requests/:id/submit` | Transition PR from `draft` to `submitted` |
| `POST` | `/api/v1/purchase-requests/:id/approve` | Transition PR from `submitted` to `approved` |
| `POST` | `/api/v1/purchase-requests/:id/reject` | Transition PR to `rejected` with reason |
| `POST` | `/api/v1/purchase-requests/:id/cancel` | Cancel PR |
| `POST` | `/api/v1/purchase-requests/:id/items` | Add new line item to draft PR |
| `PUT` | `/api/v1/purchase-requests/:id/items/:itemId` | Update line item in draft PR |
| `DELETE`| `/api/v1/purchase-requests/:id/items/:itemId` | Remove line item from draft PR |

---

### 4.2 Sourcing & Request for Quotations (`/api/v1/rfqs`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/rfqs` | Create an ad-hoc RFQ with line items |
| `POST` | `/api/v1/rfqs/from-request/:prId` | Auto-generate an RFQ from an approved PR |
| `GET` | `/api/v1/rfqs` | List RFQs (query: `status`, `purchase_request_id`, `limit`, `offset`) |
| `GET` | `/api/v1/rfqs/:id` | Get RFQ details with items |
| `GET` | `/api/v1/rfqs/number/:rfqNumber` | Get RFQ by document number (e.g. `RFQ-2026-000001`) |
| `PUT` | `/api/v1/rfqs/:id` | Update draft RFQ (title, deadline, notes) |
| `POST` | `/api/v1/rfqs/:id/publish` | Publish RFQ to open vendor bidding |
| `POST` | `/api/v1/rfqs/:id/close` | Close RFQ to stop accepting bids |
| `POST` | `/api/v1/rfqs/:id/cancel` | Cancel RFQ |
| `POST` | `/api/v1/rfqs/:id/quotations` | Submit vendor quotation proposal with line-item pricing |
| `GET` | `/api/v1/rfqs/:id/quotations` | List submitted vendor quotations for an RFQ |
| `GET` | `/api/v1/rfqs/:id/quotations/:quotationId`| Get quotation detail with item pricing |
| `GET` | `/api/v1/rfqs/:id/compare` | **Quotation Comparison Matrix**: evaluates lowest bid and fastest lead time |
| `POST` | `/api/v1/rfqs/:id/award/:quotationId` | **Award Bid**: awards quote, rejects competing bids, closes RFQ, auto-generates PO |

---

### 4.3 Purchase Orders (`/api/v1/purchase-orders`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/purchase-orders` | Create draft Purchase Order directly |
| `POST` | `/api/v1/purchase-orders/from-request/:prId` | Convert approved PR into draft PO |
| `GET` | `/api/v1/purchase-orders` | List POs (query: `status`, `vendor_id`, `limit`, `offset`) |
| `GET` | `/api/v1/purchase-orders/:id` | Get PO by UUID with line items and fulfillment stats |
| `GET` | `/api/v1/purchase-orders/number/:poNumber` | Get PO by number (e.g. `PO-2026-000001`) |
| `PUT` | `/api/v1/purchase-orders/:id` | Update draft PO |
| `POST` | `/api/v1/purchase-orders/:id/approve` | Approve PO (ready to be issued) |
| `POST` | `/api/v1/purchase-orders/:id/issue` | Issue PO to vendor (ready for receiving) |
| `POST` | `/api/v1/purchase-orders/:id/cancel` | Cancel unfulfilled PO |
| `POST` | `/api/v1/purchase-orders/:id/close` | Close PO |

---

### 4.4 Goods Receipt Notes / Receiving (`/api/v1/goods-receipts`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/goods-receipts` | Record physical delivery against PO (updates received quantities & PO status) |
| `GET` | `/api/v1/goods-receipts` | List GRNs (query: `vendor_id`, `limit`, `offset`) |
| `GET` | `/api/v1/goods-receipts/:id` | Get GRN detail with accepted and rejected quantities |
| `GET` | `/api/v1/goods-receipts/number/:grnNumber` | Get GRN by number (e.g. `GRN-2026-000001`) |
| `POST` | `/api/v1/goods-receipts/:id/cancel` | Cancel GRN and rollback received counts on PO items |
| `GET` | `/api/v1/purchase-orders/:id/goods-receipts` | List all delivery receipts for a specific PO |

---

### 4.5 Vendor Bills & 3-Way Matching (`/api/v1/vendor-bills`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/vendor-bills` | Record vendor invoice against PO |
| `GET` | `/api/v1/vendor-bills` | List bills (query: `po_id`, `vendor_id`, `status`, `match_status`) |
| `GET` | `/api/v1/vendor-bills/:id` | Get bill details, item matched quantities, and discrepancy reasons |
| `POST` | `/api/v1/vendor-bills/:id/match` | **3-Way Match**: validates invoice prices vs. PO & invoiced qtys vs. GRNs |
| `POST` | `/api/v1/vendor-bills/:id/approve` | Approve matched bill for AP disbursement |
| `POST` | `/api/v1/vendor-bills/:id/cancel` | Cancel bill and rollback `billed_quantity` on PO items |
| `GET` | `/api/v1/purchase-orders/:id/vendor-bills` | List all vendor invoices linked to a PO |

---

### 4.6 Purchase Returns / Debit Notes (`/api/v1/purchase-returns`)

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/purchase-returns` | Create draft return (guards against returning more than received) |
| `GET` | `/api/v1/purchase-returns` | List returns (query: `po_id`, `vendor_id`, `status`) |
| `GET` | `/api/v1/purchase-returns/:id` | Get return details and returned line items |
| `GET` | `/api/v1/purchase-returns/number/:returnNumber` | Get return by number (e.g. `RET-2026-000001`) |
| `PUT` | `/api/v1/purchase-returns/:id` | Update draft return details |
| `POST` | `/api/v1/purchase-returns/:id/approve` | Manager approval for return |
| `POST` | `/api/v1/purchase-returns/:id/ship` | Record shipping carrier and tracking number |
| `POST` | `/api/v1/purchase-returns/:id/complete` | **Finalize Return**: rollbacks PO received quantities and issues Debit Note |
| `POST` | `/api/v1/purchase-returns/:id/cancel` | Cancel draft/approved return |

---

### 4.7 Procurement Analytics & Vendor Scorecards (`/api/v1/analytics`)

| Method | Endpoint | Query Parameters | Description |
|---|---|---|---|
| `GET` | `/api/v1/analytics/vendors/:vendorId/scorecard` | `from_date`, `to_date` | Detailed vendor scorecard (OTIF %, Quality %, Return %, Billing Accuracy %, Composite Score) |
| `GET` | `/api/v1/analytics/vendors/performance` | `from_date`, `to_date`, `limit`, `offset` | Ranked vendor list sorted by composite reliability score |
| `GET` | `/api/v1/analytics/dashboard/kpis` | `from_date`, `to_date` | Executive KPIs (Committed spend, unbilled receiving liability, pending delivery exposure, PO counts) |
| `GET` | `/api/v1/analytics/spend/by-vendor` | `from_date`, `to_date`, `limit` | Vendor spend distribution and percentage of total procurement budget |

---

### 4.8 Transactional Outbox & Domain Events (administration disabled)

The routes below are historical references and are **unregistered**. No publisher starts, no log-only delivery acknowledges events, and pending records are preserved.

| Method | Endpoint | Query Parameters | Description |
|---|---|---|---|
| `GET` | `/api/v1/outbox/events` | `status`, `aggregate_type`, `limit`, `offset` | Query domain events (filter by `pending`, `published`, `failed`) |
| `GET` | `/api/v1/outbox/events/:id` | — | Retrieve single domain event payload by UUID |
| `POST` | `/api/v1/outbox/events/:id/retry` | — | Requeue a failed or stalled event for immediate background redispatch |

---

## 5. Core Business Logic & State Machines

### 5.1 Purchase Order Lifecycle
```mermaid
stateDiagram-v2
    [*] --> draft
    draft --> approved: ApprovePO()
    approved --> issued: IssuePO()
    issued --> partially_received: Partial GRN Delivery
    partially_received --> received: Full GRN Delivery
    partially_received --> partially_received: Additional Partial GRN
    received --> partially_received: Return Completed (RTV)
    received --> billed: Full 3-Way Match Approved
    billed --> closed: ClosePO()
    draft --> cancelled: CancelPO()
    issued --> cancelled: CancelPO() (if 0 received)
```

### 5.2 3-Way Matching Rules
1. **Price Verification**:
   $$\text{BillItem.UnitPrice} \le \text{POItem.UnitPrice}$$
   If invoiced unit price exceeds contracted price, the item is flagged with `match_status = 'price_variance'`.
2. **Quantity Verification**:
   $$\text{BillItem.Quantity} \le \text{POItem.ReceivedQuantity} - (\text{POItem.BilledQuantity} - \text{CurrentMatched})$$
   If invoiced quantity exceeds goods accepted in GRN, the item is flagged with `match_status = 'quantity_variance'`.
3. **Approval Guard**: A vendor bill can **only** be approved if `match_status == 'matched'` (all items 100% variance-free).

### 5.3 Vendor Composite Score Formula (0 to 100)
$$\text{CompositeScore} = (0.40 \times \text{OTIF}) + (0.35 \times \text{QualityAcceptance}) + (0.15 \times \text{BillingAccuracy}) + (0.10 \times \max(0, 100 - \text{ReturnRate}))$$

---

## 6. Project Directory Structure

```text
purchasing/
├── cmd/
│   └── server/
│       └── main.go                                  # Thin entrypoint (< 50 lines) with graceful shutdown
├── config/
│   └── routes.yaml                                  # Centralized YAML route definitions
├── db/
│   ├── migrations/                                  # PostgreSQL UP/DOWN migrations (000001 - 000009)
│   ├── queries/                                     # Raw SQL queries for sqlc
│   └── schema/                                      # Consolidated database schema
├── internal/
│   ├── api/                                         # Centralized API & routing layer (categorized by domain)
│   │   ├── analytics.go                             # Analytics & KPI endpoints
│   │   ├── goods_receipts.go                        # Goods Receipts (GRN) endpoints
│   │   ├── outbox.go                                # Transactional Outbox & Event replay endpoints
│   │   ├── purchase_orders.go                       # Purchase Orders endpoints
│   │   ├── purchase_requests.go                     # Purchase Requests endpoints
│   │   ├── purchase_returns.go                      # Purchase Returns & Debit Notes endpoints
│   │   ├── rfqs.go                                  # RFQ & Sourcing endpoints
│   │   ├── router.go                                # Master router setup & middleware mounting
│   │   └── vendor_bills.go                          # Vendor Bills & 3-Way Matching endpoints
│   ├── app/
│   │   └── container.go                             # Dependency Injection Container & Service Factory
│   ├── config/                                      # Process environment & Routes (routes.yaml) config
│   ├── database/                                    # Bounded pgxpool connection & read-only schema verifier
│   ├── db/
│   │   └── generated/                               # Type-safe Go code generated by sqlc
│   ├── docs/                                        # Embedded Swagger UI & OpenAPI 3.0.3 specification
│   ├── domain/                                      # Pure business entities & sentinel errors
│   ├── dto/                                         # Request/response data transfer objects
│   ├── handler/                                     # Fiber HTTP controllers
│   │   └── helper/                                  # Request parsers and DTO mappers
│   ├── middleware/                                  # Production middlewares (RequestID, Logger, Recovery, CORS)
│   ├── repository/                                  # Interfaces, Postgres implementations, & Tx managers
│   └── service/                                     # Core business logic engines & unit tests
├── .env.example                                     # Sanitized environment placeholders
├── .gitignore                                       # Git ignore rules
├── Dockerfile                                       # Multi-stage production container build
├── docker-compose.yml                               # Service stack with PostgreSQL healthcheck
├── Makefile                                         # Standard developer commands
├── go.mod                                           # Go module definitions
├── README.md                                        # Service technical documentation
└── sqlc.yaml                                        # sqlc code generation configuration
```

---

## 7. Configuration and authorization checkpoint

Read [platform prerequisites](docs/platform-prerequisites.md) for the complete approved route/resource/action matrix, all settings and fresh/legacy migration procedures, and [verification report](docs/platform-prerequisite-report.md) for actual evidence and limitations.

Configuration is process-environment only; startup does not read `.env`. [.env.example](.env.example) contains placeholders. Supply a service-owned database URL and read-only Auth RSA public key. JWT audience is exactly `erp-api`; protected routes require verified humans and exact GLOBAL grants. GLOBAL authority crosses record creators, but PR/PO/return/bill creator and approver must differ. Add-item routes require the parent's update grant. Outbox administration is disabled. `/health` and `/api/v1/health` are public liveness; `/ready` is bounded public database/schema readiness. Documentation/metrics are protected.

Untracking `.env` preserves its local copy and does not erase historical Git copies. Credential exposure is not asserted. No real credentials, keys or grants were changed.

## 8. Development and verification

Use **Go 1.27.3**, sqlc **1.31.1**, Docker and PowerShell **7.4+** for the owned integration fixture. The module minimum remains Go 1.27.1. Docker uses pinned `golang:1.27.3-alpine` and non-root `alpine:3.23` with CA certificates.

```powershell
# From this repository; normal DB tests skip unless dedicated fixture variables exist.
go test -p=1 -count=1 ./...
go vet -p=1 ./...
go build -p=1 ./...
# From workspace root; generates owned credentials and verifies/removes only its resources.
pwsh -NoProfile -File purchasing/scripts/test-platform-prerequisites.ps1 -GoExe go -AuthSource Authentication
```

`make run` requires explicitly injected environment plus already reviewed schema 9. `make test-all` does not start a normal database or load `.env`. Other legacy Compose/database helper targets remain separate development conveniences, **not** the integrated deployment procedure; their existing configuration has not been certified for new auth/schema requirements. Do not run the old Compose stack expecting automatic migrations.

CI covers the capitalized `Development` branch and configured PR targets; it pins Go, checks formatting/vet/build, runs the owned PostgreSQL/image fixture, focused Linux race tests and pinned sqlc regeneration. Its standalone checkout explicitly skips sibling Auth migration verification. Workflow syntax/local commands are distinct from an actual GitHub Actions run.

ERP-Platform Compose/Kong integration, real Auth token transport, reviewed existing-database adoption and normal rollout remain pending. No Purchasing-to-Inventory or Kafka integration is claimed.

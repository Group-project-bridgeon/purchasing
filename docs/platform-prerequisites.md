# Purchasing platform prerequisites â€” authorization proposal

Status: **PROPOSED â€” explicit approval required before authorization code or Auth permission migration.**

## Exact route matrix

These are actual application routes plus the proposed /ready probe. Human means verified RS256, configured issuer, audience erp-api, existing maximum 15-minute lifetime, valid identity/time claims and exact GLOBAL grants. No machine tokens, implied permissions or automatic grants. Multiple rows for one method/path require ALL displayed pairs. Outbox rows describe existing endpoints to disable, not approved or provisioned authority.

| HTTP method | Route | Resource | Action | Actor policy |
| --- | --- | --- | --- | --- |
| GET | /api/v1/analytics/vendors/:vendorId/scorecard | purchasing.analytics | read | Human; exact GLOBAL |
| GET | /api/v1/analytics/vendors/performance | purchasing.analytics | read | Human; exact GLOBAL |
| GET | /api/v1/analytics/dashboard/kpis | purchasing.analytics | read | Human; exact GLOBAL |
| GET | /api/v1/analytics/spend/by-vendor | purchasing.analytics | read | Human; exact GLOBAL |
| POST | /api/v1/goods-receipts | purchasing.goods-receipts | create | Human; exact GLOBAL |
| GET | /api/v1/goods-receipts | purchasing.goods-receipts | read | Human; exact GLOBAL |
| GET | /api/v1/goods-receipts/:id | purchasing.goods-receipts | read | Human; exact GLOBAL |
| POST | /api/v1/goods-receipts/:id/cancel | purchasing.goods-receipts | cancel | Human; exact GLOBAL |
| GET | /api/v1/purchase-orders/:id/goods-receipts | purchasing.goods-receipts | read | Human; exact GLOBAL |
| GET | /api/v1/outbox/events | purchasing.outbox | read | DISABLED; not registered; no permission definition added |
| GET | /api/v1/outbox/events/:id | purchasing.outbox | read | DISABLED; not registered; no permission definition added |
| POST | /api/v1/outbox/events/:id/retry | purchasing.outbox | retry | DISABLED; not registered; no permission definition added |
| POST | /api/v1/purchase-orders | purchasing.purchase-orders | create | Human; exact GLOBAL |
| POST | /api/v1/purchase-orders/from-request/:prId | purchasing.purchase-orders | create | Human; exact GLOBAL |
| POST | /api/v1/purchase-orders/from-request/:prId | purchasing.purchase-requests | read | Human; exact GLOBAL; ALSO REQUIRED |
| GET | /api/v1/purchase-orders | purchasing.purchase-orders | read | Human; exact GLOBAL |
| GET | /api/v1/purchase-orders/:id | purchasing.purchase-orders | read | Human; exact GLOBAL |
| PUT | /api/v1/purchase-orders/:id | purchasing.purchase-orders | update | Human; exact GLOBAL |
| POST | /api/v1/purchase-orders/:id/approve | purchasing.purchase-orders | approve | Human; exact GLOBAL |
| POST | /api/v1/purchase-orders/:id/issue | purchasing.purchase-orders | issue | Human; exact GLOBAL |
| POST | /api/v1/purchase-orders/:id/cancel | purchasing.purchase-orders | cancel | Human; exact GLOBAL |
| GET | /api/v1/purchase-orders/:id/items | purchasing.purchase-orders | read | Human; exact GLOBAL |
| POST | /api/v1/purchase-orders/:id/items | purchasing.purchase-orders | update | Human; exact GLOBAL |
| PUT | /api/v1/purchase-orders/:id/items/:itemId | purchasing.purchase-orders | update | Human; exact GLOBAL |
| DELETE | /api/v1/purchase-orders/:id/items/:itemId | purchasing.purchase-orders | delete | Human; exact GLOBAL |
| POST | /api/v1/purchase-requests | purchasing.purchase-requests | create | Human; exact GLOBAL |
| GET | /api/v1/purchase-requests | purchasing.purchase-requests | read | Human; exact GLOBAL |
| GET | /api/v1/purchase-requests/:id | purchasing.purchase-requests | read | Human; exact GLOBAL |
| PUT | /api/v1/purchase-requests/:id | purchasing.purchase-requests | update | Human; exact GLOBAL |
| POST | /api/v1/purchase-requests/:id/submit | purchasing.purchase-requests | submit | Human; exact GLOBAL |
| POST | /api/v1/purchase-requests/:id/approve | purchasing.purchase-requests | approve | Human; exact GLOBAL |
| POST | /api/v1/purchase-requests/:id/reject | purchasing.purchase-requests | reject | Human; exact GLOBAL |
| GET | /api/v1/purchase-requests/:id/items | purchasing.purchase-requests | read | Human; exact GLOBAL |
| POST | /api/v1/purchase-requests/:id/items | purchasing.purchase-requests | update | Human; exact GLOBAL |
| PUT | /api/v1/purchase-requests/:id/items/:itemId | purchasing.purchase-requests | update | Human; exact GLOBAL |
| DELETE | /api/v1/purchase-requests/:id/items/:itemId | purchasing.purchase-requests | delete | Human; exact GLOBAL |
| POST | /api/v1/purchase-returns | purchasing.purchase-returns | create | Human; exact GLOBAL |
| GET | /api/v1/purchase-returns | purchasing.purchase-returns | read | Human; exact GLOBAL |
| GET | /api/v1/purchase-returns/:id | purchasing.purchase-returns | read | Human; exact GLOBAL |
| GET | /api/v1/purchase-returns/number/:returnNumber | purchasing.purchase-returns | read | Human; exact GLOBAL |
| PUT | /api/v1/purchase-returns/:id | purchasing.purchase-returns | update | Human; exact GLOBAL |
| POST | /api/v1/purchase-returns/:id/approve | purchasing.purchase-returns | approve | Human; exact GLOBAL |
| POST | /api/v1/purchase-returns/:id/ship | purchasing.purchase-returns | ship | Human; exact GLOBAL |
| POST | /api/v1/purchase-returns/:id/complete | purchasing.purchase-returns | complete | Human; exact GLOBAL |
| POST | /api/v1/purchase-returns/:id/cancel | purchasing.purchase-returns | cancel | Human; exact GLOBAL |
| POST | /api/v1/rfqs | purchasing.rfqs | create | Human; exact GLOBAL |
| POST | /api/v1/rfqs/from-request/:prId | purchasing.rfqs | create | Human; exact GLOBAL |
| POST | /api/v1/rfqs/from-request/:prId | purchasing.purchase-requests | read | Human; exact GLOBAL; ALSO REQUIRED |
| GET | /api/v1/rfqs | purchasing.rfqs | read | Human; exact GLOBAL |
| GET | /api/v1/rfqs/:id | purchasing.rfqs | read | Human; exact GLOBAL |
| GET | /api/v1/rfqs/number/:rfqNumber | purchasing.rfqs | read | Human; exact GLOBAL |
| PUT | /api/v1/rfqs/:id | purchasing.rfqs | update | Human; exact GLOBAL |
| POST | /api/v1/rfqs/:id/publish | purchasing.rfqs | publish | Human; exact GLOBAL |
| POST | /api/v1/rfqs/:id/close | purchasing.rfqs | close | Human; exact GLOBAL |
| POST | /api/v1/rfqs/:id/cancel | purchasing.rfqs | cancel | Human; exact GLOBAL |
| POST | /api/v1/rfqs/:id/quotations | purchasing.vendor-quotations | create | Human; exact GLOBAL |
| GET | /api/v1/rfqs/:id/quotations | purchasing.vendor-quotations | read | Human; exact GLOBAL |
| GET | /api/v1/rfqs/:id/quotations/:quotationId | purchasing.vendor-quotations | read | Human; exact GLOBAL |
| GET | /api/v1/rfqs/:id/compare | purchasing.vendor-quotations | read | Human; exact GLOBAL |
| POST | /api/v1/rfqs/:id/award/:quotationId | purchasing.rfqs | award | Human; exact GLOBAL |
| POST | /api/v1/rfqs/:id/award/:quotationId | purchasing.purchase-orders | create | Human; exact GLOBAL; ALSO REQUIRED |
| POST | /api/v1/vendor-bills | purchasing.vendor-bills | create | Human; exact GLOBAL |
| GET | /api/v1/vendor-bills | purchasing.vendor-bills | read | Human; exact GLOBAL |
| GET | /api/v1/vendor-bills/:id | purchasing.vendor-bills | read | Human; exact GLOBAL |
| POST | /api/v1/vendor-bills/:id/match | purchasing.vendor-bills | match | Human; exact GLOBAL |
| POST | /api/v1/vendor-bills/:id/approve | purchasing.vendor-bills | approve | Human; exact GLOBAL |
| POST | /api/v1/vendor-bills/:id/cancel | purchasing.vendor-bills | cancel | Human; exact GLOBAL |
| GET | /api/v1/purchase-orders/:id/vendor-bills | purchasing.vendor-bills | read | Human; exact GLOBAL |
| GET | /docs | purchasing.documentation | read | Human; exact GLOBAL |
| GET | /docs/ | purchasing.documentation | read | Human; exact GLOBAL |
| GET | /swagger | purchasing.documentation | read | Human; exact GLOBAL |
| GET | /swagger/ | purchasing.documentation | read | Human; exact GLOBAL |
| GET | /docs/openapi.json | purchasing.documentation | read | Human; exact GLOBAL |
| GET | /metrics | purchasing.metrics | read | Human; exact GLOBAL |
| GET | /health | â€” | â€” | Public liveness |
| GET | /api/v1/health | â€” | â€” | Public legacy liveness |
| GET | /ready (new) | â€” | â€” | Public bounded readiness |

## Confirmed actor and access policy

For creation, requested_by, received_by and created_by are the actual verified human, including RFQ creation from a PR. Omitted fields are filled from claims; supplied canonical matching actor values are accepted for compatibility; a different actor is rejected. No delegation/on-behalf-of policy. Existing stored actor values remain historical. GLOBAL permissions permit the listed operations across creators. CONFIRMED: creator and approver must differ for purchase requests, purchase orders, purchase returns and vendor bills. Exact GLOBAL approve authority is still required. Purchase orders/vendor bills require additive immutable creator tracking. Historical unknown creators remain blocked; no guessed backfill or normal adoption is authorized. Lifecycle diagnostics record the actual actor without retroactively rewriting business creator fields. The confirmed separation rule requires additive creator/approval metadata; preserve existing facts and record the actual approver/server approval time. Final authorization/schema implementation approval was received for the concrete plan below.

Adding an item to an existing purchase request/order requires its parent update permission. Other displayed item update/delete permissions remain unchanged. Every parent/item mutation must also satisfy the parent lifecycle restriction atomically; authorization must not bypass approved/issued state. From-PR creation additionally requires purchase-request read. RFQ award additionally requires purchase-order create. Analytics and quotation comparison have the displayed exact read authority. Documentation and metrics require dedicated exact GLOBAL read grants. Outbox administration routes are unregistered, even for otherwise privileged humans; no Outbox permissions are added by this task.

## Independent implementation and risk

Preserve all eight SQL migration pairs byte-for-byte. Integrated startup only verifies exact clean golang-migrate version 9 and required relations using bounded read-only queries. It never creates, converts, forces or baselines history. Existing embedded-runner adoption remains operationally blocked until backups and actual schema/history inspection are authorized. Fresh installations use service-owned golang-migrate. Default publisher is disabled; no broker exists, so enabling publication must fail clearly and an unconfigured dispatcher must never acknowledge pending events. Keep slog, stable safe JSON errors, bounded HTTP/database/shutdown work and validated request IDs. Gate hardcoded database tests before any execution.

Verification: route/claim/spoofing tests after approval; focused configuration/error/outbox/lifecycle tests; uniquely owned PostgreSQL migration/readiness/history-refusal/disabled-outbox tests; compatible compiler vet/build; non-root image/key-readability checks; formatting, CI/config syntax and whitespace checks. No normal infrastructure, shared ERP-Platform edits or Git publication. The sole authorized index change is removing .env tracking while keeping its local file.

## Toolchain checkpoint

Module minimum remains Go 1.27.1; installed global Go 1.25.4 remains untouched. Go 1.27.3 was verified with its executable, dependency Go requirements and official Docker manifests, then pinned in go.mod toolchain, CI and golang:1.27.3-alpine. Runtime alpine:3.23 supplies CA certificates and UID 10001. sqlc is pinned to 1.31.1. Official release history: https://go.dev/doc/devel/release. Docker/runtime execution evidence is recorded separately in the prerequisite report.

## Lifecycle inspection and narrowly required fix (2026-10-29)

Verified source facts: PR/PO header UPDATE queries have status=draft predicates. PR/PO item UPDATE/DELETE use draft-status EXISTS predicates. Their item INSERTs are unconditional after a separate service-level draft check. These checks do not acquire the parent row lock before mutation, so approval racing an item edit needs stronger serialization. DELETE uses exec and can silently affect zero rows. Return and RFQ header UPDATE queries lack final draft predicates even though services check earlier.

Required plan: preserve migrations 1-8; add migration 000009 for immutable creator/approver metadata and narrow database guards. Parent content edits and PR/PO item inserts/updates/deletes must serialize against the actual parent row and reject non-draft states, including concurrent approval/issue. PO receipt/billing quantity accounting remains explicitly distinct from editable ordered quantity/price/description fields, preserving legitimate existing receiving/billing workflows. Header creator identity cannot be changed. Approval records the verified actor/server time, checks a permitted source state, rejects self-approval and unknown creator, and cannot resurrect a cancelled/terminal document through a stale status update. Existing unknown PO/bill creators stay null; no automatic adoption/backfill. Disposable tests must cover every exposed parent edit/nested mutation, non-draft states, approval/edit races and actor separation. No other matrix changes.

Auth numbering checkpoint: 000010-000012 Inventory dependencies were already pending and remain unchanged. No Party migration branch/file is present in this checkout. New 000013_purchasing_permissions defines only the approved non-Outbox pairs; no grants. Recheck and coordinate that reservation with the Party owner before publication/merge; this task performs neither and claims no external teammate confirmation.


## Configuration and operational ownership

The application reads process environment only; it never loads `.env`. The sanitized [.env.example](../.env.example) is a list of names/placeholders, not deployable credentials. An operator must inject the service-owned DATABASE_URL/PGPASSWORD and a read-only RSA public-key file. Never supply Auth signing keys. Container database hostnames/ports must match the actual deployment network. The local `sslmode=disable` example is not a production TLS policy.

| Settings | Defaults / constraints |
| --- | --- |
| DATABASE_URL, JWT_PUBLIC_KEY_PATH | Required; no credential/key defaults |
| SERVER_PORT, LISTEN_HOST | 8080, 0.0.0.0; canonical port 1–65535 and IP literal |
| JWT_ISSUER, JWT_AUDIENCE | erp-auth, exact erp-api; human RS256, required identity/time claims, maximum 15-minute lifetime |
| APP_ENV, LOG_LEVEL, LOG_FORMAT | development, info, json; integrated operators use production/json |
| HTTP_READ_TIMEOUT, HTTP_WRITE_TIMEOUT, HTTP_IDLE_TIMEOUT | 5s, 15s, 60s; maximum 1m, 1m, 5m |
| REQUEST_TIMEOUT, STARTUP_TIMEOUT, READINESS_TIMEOUT, SHUTDOWN_TIMEOUT | 5s, 15s, 1s, 10s; maximum 30s, 2m, 5s, 1m |
| HTTP_MAX_BODY_BYTES | 1048576; allowed 1024–16777216 |
| DB_MAX_CONNS, DB_MIN_CONNS | 10, 0; max 1–100, min 0–max |
| DB_OPERATION_TIMEOUT | 5s; maximum 30s; statement/lock bounds and request cancellation |
| DB_MAX_CONN_LIFETIME, DB_MAX_CONN_IDLE_TIME, DB_HEALTH_CHECK_PERIOD | 30m, 5m, 1m; maximum 24h, 1h, 5m |
| OUTBOX_PUBLISHER_ENABLED | false; true rejected until a real approved dispatcher exists |

Durations must be at least 1ms. HTTP write timeout must exceed request timeout; request timeout must cover DB operation timeout; idle connection lifetime must not exceed maximum lifetime. Readiness has its own small budget. A request context provides cooperative cancellation; it does not forcibly stop arbitrary CPU work. HTTP accepted work is drained within shutdown budget before worker/pool cleanup. Listener bind failure returns immediately with a stable diagnostic category. DB readiness checks connectivity plus exact clean schema 9 and required relations. Only liveness remains available when the database fails.

JSON slog keeps service, component, operation, reason, bounded request_id, verified actor when present, and SQLSTATE where relevant. Logs and clients exclude raw database/key errors, panic text, sensitive bodies and tokens. Request IDs are bounded to 64 safe characters or replaced. Clients receive generic errors; categories distinguish authorization, lifecycle/conflict, timeout and infrastructure failure. No public provisioning endpoints were added.

Outbox records are retained. The application never starts its worker; nil-dispatcher ProcessBatch rejects before fetching/acknowledging events. The log-only production dispatcher was removed. Explicit `true` configuration fails, rather than simulating delivery. Unit fault tests use an explicit fake dispatcher; that is not broker verification. Existing domain operations can still write pending transactional Outbox facts. Kafka delivery, normal replay and Outbox administration remain unimplemented/disabled.

## Fresh installation — separately authorized operator procedure

1. Provision a service-owned fresh database, private credentials, network and read-only public-key mount through a separately reviewed Platform change. Do not use normal credentials for tests.
2. With the application stopped, apply immutable migration files 000001–000008 and additive 000009 with the standard pinned golang-migrate runner (fixture uses migrate/migrate:v4.18.3). The history must be clean numeric version 9. Do not also run the removed embedded startup runner.
3. Coordinate Auth migration 000013 with the Party/Inventory owners before publication; apply definitions only under separate deployment authorization. No roles/users receive grants from this migration. Human grants are a distinct explicit administrative action.
4. Start only the compatible new application with its required environment/key mount; verify liveness, bounded readiness, 401/403 and explicitly authorized human workflows. This document does not authorize that normal rollout. Old writers and schema 9 are not certified for mixed-version operation.

Migration 000009 adds nullable creator/approver metadata without rewriting old business rows. All new PO/bill rows require a trusted creator. Existing PR requested_by and return created_by remain historical fields: their old pre-authentication provenance is not retroactively verified. Existing unknown PO/bill creators remain null and cannot approve. Existing approval facts are not invented or backfilled. Header and PR/PO editable-item guards acquire the real parent row, reject non-draft edits and serialize against approval. Received/billed accounting counters remain available only as the existing distinct workflow updates.

Rollback 000009 is deliberately refused if any new creator/approval metadata would be discarded. A failed 9-to-8 downgrade leaves history at target version 8 with dirty=true while the metadata/schema remains intact; do not force it clean or drop facts. Recover through a separately reviewed forward/restore procedure with verified backup. Empty-metadata downgrade/re-upgrade and populated refusal are disposable acceptance cases, not normal rollback authorization.

## Existing embedded-runner adoption — BLOCKED until reviewed

No normal database was inspected or changed in this task. Source cannot prove an existing database's actual schema/history or creator provenance. Before adoption, an authorized operator must stop competing writers, make and verify a recoverable backup, inspect the actual tracker shape/applied files/schema/constraints and compare against preserved migration SQL. The legacy filename/VARCHAR/applied_at tracker is incompatible with standard bigint-version/dirty history. New startup rejects it read-only.

Do not apply old CREATE TABLE migrations over an existing schema, rename migration history, delete tracker rows, guess a baseline, run migrate force, disable guards or guess creator IDs. Prepare a specific reviewed adoption/recovery plan from actual evidence and backups. Historical actor/provenance reconciliation requires separate review; no such backfill tool or procedure is delivered here. A normal database is never assumed fresh.

## Reproducible disposable verification

From the workspace root, with Go 1.27.3 and Docker available:

```powershell
pwsh -NoProfile -File purchasing/scripts/test-platform-prerequisites.ps1 -GoExe go -AuthSource Authentication
```

The script never reads local `.env`; generated credentials, loopback ephemeral ports, unique run IDs, exact labels, tmpfs PostgreSQL and explicit database allowlists isolate it. It preserves existing container identity/start/restart state and removes only its own labelled resources. Tests require both PURCHASING_TEST_RUN_ID and PURCHASING_TEST_DATABASE_URL plus a matching database ownership marker. They reject ordinary DATABASE_URL fallback, default ports and non-fixture names. With both dedicated variables unset, DB tests explicitly skip. Missing/invalid partial fixture configuration fails.

With explicit Auth source, the fixture checks all 41 definitions and zero role grants. Standalone Purchasing CI lacks the sibling and explicitly skips that Auth-only check. Default verification includes the non-root image, a read-only synthetic public key, anonymous 401 and readiness 503/liveness 200 after stopping only fixture PostgreSQL. `-SkipImage` is an explicit reduced check, not a successful image verification claim. The script does not certify normal Auth issuance, Kong, TLS, deployment or existing database adoption.

## Remaining integration prerequisites

ERP-Platform Compose/Kong is intentionally unchanged. A later milestone must add reviewed service-owned migration/startup wiring, DB/network settings, read-only public-key mounts, health/dependency sequencing, stop grace period, routes and isolated gateway tests. Coordinate Auth numbering and explicit human grants. Verify actual Auth-issued tokens/TLS through disposable Kong before separately authorized normal rollout. Purchasing-to-Inventory mutations, broker delivery and production readiness are not delivered. See [verification report](platform-prerequisite-report.md) for actual results and limitations.

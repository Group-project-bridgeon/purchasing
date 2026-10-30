# Purchasing platform-prerequisite implementation report

2026-10-29. Verdict: **PREREQUISITE IMPLEMENTATION AND DISPOSABLE VERIFICATION PASSED; PLATFORM INTEGRATION REMAINS PENDING.**

## Completed and approved scope

The six audited prerequisite areas are addressed: human JWT/exact GLOBAL authority, migration ownership, disabled Outbox, operational bounds/readiness/safe errors, pinned compatible toolchains, and `.env` untracking. This report describes current source and executed checks, not normal deployment success. The full [approved matrix and operational procedures](platform-prerequisites.md) remain authoritative.

The revised matrix has 68 protected routes (62 business plus six documentation/metrics), three public probes and three explicitly unregistered Outbox endpoints. Compound routes require all displayed pairs. Add-item PR/PO routes require parent **update**. Exact GLOBAL grants permit listed operations across creators; they never override lifecycle or the mandatory different creator/approver rule for PR/PO/returns/bills. Creators/receivers come from verified human claims. Omitted or canonical matching values are accepted; impersonation is rejected. No delegation, machine access, implied permissions or automatic grants were added.

Auth migration **000013_purchasing_permissions** adds exactly 41 approved definitions, with zero grants and no Outbox permission. Existing pending Auth 000010–000012 were preserved. No Party migration/branch was present locally; coordinate 13 with that owner before later publication/merge. No external coordination or numbering reservation by a teammate is claimed.

## Implementation and tradeoffs

Requests pass bounded request-ID/context middleware, verified RS256 human identity, exact endpoint grants, trusted actor binding, existing application services and service-owned repositories. Migration 000009 persists immutable nullable creator/approval facts, server approval time and source-state/actor separation. Parent/item guards serialize edits against actual parent locks; approved/issued content cannot be edited through nested routes. PR/PO receipt/billing accounting counters remain a distinct allowed internal workflow. No cross-service database query/FK was added.

The new server only reads schema history/relations; it requires clean numeric golang-migrate version 9. It performs no DDL or history conversion. Existing migration pairs 1–8 are byte-for-byte preserved. Unknown legacy PO/bill creators stay null and cannot approve; historical PR/return identity provenance is not retroactively authenticated. Normal adoption remains blocked pending separately authorized backups and actual history/schema/provenance review. A populated migration-9 downgrade refuses rather than discarding authority facts; golang-migrate then records target version 8 dirty=true. Do not force/baseline it clean.

Outbox publishing defaults off; true configuration fails because no production dispatcher exists. No worker starts and nil-dispatcher processing cannot fetch/acknowledge events. Existing pending facts remain. Log-only delivery was removed; no Kafka or production dispatcher was invented.

HTTP/body/pool/query/startup/readiness/shutdown budgets are validated; accepted HTTP work drains before dependencies close. JSON slog retains stable stage/operation/reason/request_id, actual actor and SQLSTATE where relevant without raw error/panic/body/token content. Config validation hints are static safe text, not infrastructure errors. Clients receive generic errors. Public readiness checks bounded DB connectivity and schema tracker/relations; it is not full legacy schema/provenance certification.

Go minimum **1.27.1** remains unchanged; verified **Go 1.27.3** is pinned in module toolchain, CI and `golang:1.27.3-alpine`; runtime is `alpine:3.23`, CA certificates and UID 10001. Installed global Go 1.25.4 was not replaced. Original Prometheus client 1.25.0/common 0.72.0 remain; direct JWT is compatible v5.3.1. sqlc is pinned 1.31.1.

Configuration is environment-only. See [.env.example](../.env.example) and the full settings table in the contract. The application and verification never load the normal `.env`. Its actual local file exists with unchanged size/mtime; values were never read/printed. Only its Git index entry is staged for deletion. This does **not** erase old Git copies and does not establish confirmed credential exposure.

## Executed verification

| Check | Actual result |
| --- | --- |
| Pinned Go executable and selected dependency Go requirements | PASS: Go 1.27.3 supports the preserved graph/minimum |
| `pwsh -NoProfile -File purchasing/scripts/test-platform-prerequisites.ps1 -GoExe <verified-go1.27.3> -AuthSource Authentication` | PASS, final exit 0; generated credentials, non-admin role, unique loopback/tmpfs resources only |
| Fresh Purchasing migrations 1–9, empty-metadata down/up, legacy-null preservation | PASS on PostgreSQL 17 |
| Populated down refusal and explicit dirty target marker | PASS; facts retained, no force or baseline |
| Auth migrations through 13 on separate owned DB | PASS: 41 definitions, zero role grants, zero Outbox definitions |
| Full `go test -p=1 -count=1 -timeout=240s ./...` inside owned fixture | PASS, including all gated PostgreSQL and in-process signed HTTP cases |
| PR/PO insert/update/delete/header races | PASS: workers observed waiting on actual PostgreSQL locks; edits reject after approval commits; protected facts retained |
| Creator/approver separation and immutable metadata | PASS for PR, PO, return and bill; unknown creator and missing trusted actor rejected |
| JWT/route/grant/spoofing checks | PASS: invalid claims/keys/algorithms, 401/403, compound grants, actor spoofing, protected docs/metrics and disabled Outbox |
| Final focused `go test -p=1 -count=1 ./internal/api` | PASS after adding actual registered-route coverage against the matrix |
| Real app/repository HTTP on owned PostgreSQL | PASS: trusted PR/PO creator, self-approval 403, different actor approval, edit rejection and receipt receiver identity |
| Disabled Outbox persistence | PASS: actual pending event remains pending; explicit synthetic dispatcher faults tested separately |
| Readiness failure | PASS: closed-pool in-process test and real owned-container PostgreSQL outage; readiness 503, liveness 200 |
| Error/log/config/body/request-ID/listener/shutdown checks | PASS; body limit and graceful drain use actual local HTTP sockets |
| `go vet -p=1 ./...`; `go build -p=1 ./...`; `go mod verify` | PASS, exit 0; all modules verified |
| Two isolated `sqlc v1.31.1 generate` rounds | PASS: all 10 current generated files match both rounds after CRLF normalization |
| Owned Docker image build/runtime | PASS: actual image build, UID 10001, read-only synthetic public-key mount, readiness 200 and anonymous business 401 |
| YAML parsing | PASS using existing yaml.v3 dependency: CI, public routes and sqlc config; Development push/PR targets and Go pin checked |
| PowerShell parsing/native-command shadowing check | PASS |
| Final formatting/whitespace/document/source-preservation checks | See final verification checkpoint below |

Standalone earlier Go runs intentionally skipped gated DB tests when dedicated fixture variables were absent; they were not treated as DB proof. The successful final fixture supplies both variables and the ownership marker; no DB fallback exists. Standalone Purchasing CI explicitly skips sibling Auth checks unless its source is supplied.

Actual JWT verification uses synthetic issuers in these tests; actual normal Auth issuance, Kong/TLS and normal credentials are **NOT VERIFIED**. No normal database or service was accessed/changed. Windows `-race` was not run; focused Linux race execution is configured in CI but a GitHub Actions run was not executed. PostgreSQL lock-wait tests are real concurrency evidence, not Go race-detector evidence. JWT verification is offline; token_version is structurally validated, not queried against Auth on every request. Existing issued tokens may retain their embedded grants until expiry.

## Failures found and resolved

- Native sqlc installation failed with the old Windows C compiler; used the official pinned Windows release with SHA256 checked against its release asset digest. No generated code was hand-edited.
- An initial JWT pin selected older Prometheus versions. Targeted dependency edits restored original versions and selected compatible JWT 5.3.1; module/build checks passed. A broad-file restore was rejected by automatic review; the safer targeted correction resolved it without discarding work.
- A PowerShell wrapper named Docker shadowed the native command and recursed; its own invocation was interrupted before resources were created. Renamed wrapper/native executable calls and added an AST regression check.
- The downgrade test initially expected dirty version 9; actual golang-migrate semantics use target version 8. Corrected expectation and verified the retained facts and marker on PostgreSQL.
- Three legacy assertions expected raw/no-row/silent lifecycle behavior. Updated them to the approved typed non-draft rejection and preservation assertions; the full owned suite passed.

## Gotchas addressed

| Mistake / cause | Prevention | Verification |
| --- | --- | --- |
| JWT alone permits unauthorized endpoints | Exact per-route resource/action checks; all compound pairs required | Every registered protected route covered; no-grant/wrong-grant requests reject |
| Creator spoofing/self-approval | Claims-derived actor plus persisted immutable creator and atomic comparison | HTTP spoofing/self-approval cases and all four PostgreSQL approval classes |
| Separate draft check races approval | Lock actual parent in item guard and recheck state | Observed-lock PR/PO races for all exposed mutation forms |
| Migration systems disagree about history | Read-only startup; standard runner owns fresh history; legacy adoption blocked | Fresh migration plus rolled-back legacy tracker rejection |
| Log-only delivery silently loses pending events | No default dispatcher/worker; true config rejected | Unit fault tests and actual pending-row preservation |
| Tests accidentally hit a normal database | Dedicated strict run ID/URL/marker, non-default ephemeral port, no DATABASE_URL fallback | Gate rejection tests and run-owned fixture |
| Raw errors/paths become secret-bearing diagnostics | Static categories, bounded request IDs and route templates, safe client errors | Captured canary logs/responses, SQLSTATE and startup-stage tests |
| Rollback destroys new authority facts | Refuse populated downgrade; require reviewed forward recovery | Actual refusal preserves metadata and marks dirty target |
| PowerShell helper shadows a native tool | Distinct InvokeFixtureDocker plus explicit docker.exe | AST check and successful final fixture |

## Cleanup and preservation

The successful final fixture removed its run-labelled containers, uniquely tagged image and ownership-checked temporary directory, and restored process environment. Its actual PostgreSQL stop affected only the run-owned container. All **37 pre-existing containers** kept their recorded ID/start/running/restart state. No Compose down, volume removal, normal restart or normal credentials were used. Downloaded tool binaries/build cache are tooling, not running fixtures; evidence remains outside repositories under the task review-artifacts directory.

All five repository branches/HEADs match baseline. Source hashes confirm **28 Platform, 118 Auth, 202 Catalog and 177 Inventory pre-existing source files preserved**, including existing Inventory feature work; protected backups were not enumerated. All 16 Purchasing migration files 1–8 remain unchanged. Catalog go.mod and normal/protected material remain untouched.

Purchasing was initially clean; all substantive Purchasing changes listed below belong to this task. Authentication adds only two untracked migration-13 files to its 18 baseline status entries (20 total). ERP-Platform retains five baseline entries; Catalog retains its one go.mod edit; Inventory retains 80 baseline status entries including untracked backups. Some Purchasing Git status entries are CRLF/stat-only; the list below uses substantive normalized diff plus explicit new source files. No branch was switched; no reset/stash/pull/staging beyond authorized `.env` untracking, commit, push, PR or merge occurred.

## Remaining integration and operational steps

1. Coordinate Auth migration 13 with Party/Inventory before publication; no normal grants are included.
2. Separately authorize and implement Platform Compose/Kong wiring, service-owned migration order, required environment/public-key mounts, routes, readiness/dependency sequencing and shutdown grace period.
3. Verify real Auth-issued tokens through disposable Kong/TLS. Existing embedded-runner adoption requires actual history/schema/provenance inspection and verified backups before a concrete separately approved conversion/recovery procedure.
4. Authorize normal rollout/data workflows separately. No Purchasing-to-Inventory, Kafka delivery or production deployment was started. Purchasing platform integration is not complete.

## Exact substantive/new files and purpose

`.env`: staged **index removal only**; actual local file retained unread. All other listed changes remain unstaged/untracked.

| Repository-relative path | Purpose |
| --- | --- |
| `purchasing/.dockerignore` | Exclude secrets/local material from image context |
| `purchasing/.env.example` | Sanitized placeholders derived only from configuration source |
| `purchasing/.github/workflows/ci.yml` | Pinned Go, Development/PR coverage and owned CI checks |
| `purchasing/.gitignore` | Ignore actual local environment, keys, backups and fixture material |
| `purchasing/Dockerfile` | Pinned builder/runtime, CA certificates and non-root UID 10001 |
| `purchasing/Makefile` | Remove implicit DB startup/.env load from run/test-all |
| `purchasing/README.md` | Correct operational/security/migration/toolchain claims |
| `purchasing/cmd/server/main.go` | No startup migrations; bounded startup/shutdown with safe stage diagnostics |
| `purchasing/db/migrations/000009_document_authority.down.sql` | Approved additive 000009 authority guards and data-preserving rollback refusal |
| `purchasing/db/migrations/000009_document_authority.up.sql` | Approved additive 000009 authority guards and data-preserving rollback refusal |
| `purchasing/db/queries/purchase_orders.sql` | Creator/approver SQL, conditional lifecycle updates and sqlc schema input |
| `purchasing/db/queries/purchase_requests.sql` | Creator/approver SQL, conditional lifecycle updates and sqlc schema input |
| `purchasing/db/queries/purchase_returns.sql` | Creator/approver SQL, conditional lifecycle updates and sqlc schema input |
| `purchasing/db/queries/rfqs.sql` | Creator/approver SQL, conditional lifecycle updates and sqlc schema input |
| `purchasing/db/queries/vendor_bills.sql` | Creator/approver SQL, conditional lifecycle updates and sqlc schema input |
| `purchasing/db/schema/schema.sql` | Creator/approver SQL, conditional lifecycle updates and sqlc schema input |
| `purchasing/docs/platform-prerequisite-report.md` | Approved contract, settings/adoption procedures and evidence report |
| `purchasing/docs/platform-prerequisites.md` | Approved contract, settings/adoption procedures and evidence report |
| `purchasing/go.mod` | Keep Go minimum/dependency versions; pin verified toolchain and direct JWT dependency |
| `purchasing/go.sum` | Checksums for dependency resolution, including harmless extra downloaded versions |
| `purchasing/internal/actor/actor.go` | Verified-human UUID in context; no body-derived identity |
| `purchasing/internal/api/analytics.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/goods_receipts.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/outbox.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/permissions.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/probes.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/probes_test.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/purchase_orders.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/purchase_requests.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/purchase_returns.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/rfqs.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/router.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/router_test.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/security_test.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/api/vendor_bills.go` | Exact route authority, probes and route/claim/actor coverage |
| `purchasing/internal/app/container.go` | Fail-closed wiring, readiness, listener failure and HTTP drain lifecycle |
| `purchasing/internal/app/server.go` | Fail-closed wiring, readiness, listener failure and HTTP drain lifecycle |
| `purchasing/internal/app/server_test.go` | Fail-closed wiring, readiness, listener failure and HTTP drain lifecycle |
| `purchasing/internal/config/config.go` | Environment-only validated configuration and operational limits |
| `purchasing/internal/config/config_test.go` | Environment-only validated configuration and operational limits |
| `purchasing/internal/config/operations_test.go` | Environment-only validated configuration and operational limits |
| `purchasing/internal/database/migrator.go` | Bounded pool, read-only schema ownership and isolated test target |
| `purchasing/internal/database/postgres.go` | Bounded pool, read-only schema ownership and isolated test target |
| `purchasing/internal/database/postgres_test.go` | Bounded pool, read-only schema ownership and isolated test target |
| `purchasing/internal/db/generated/models.go` | sqlc 1.31.1 bindings from reviewed SQL; no manual edits |
| `purchasing/internal/db/generated/purchase_orders.sql.go` | sqlc 1.31.1 bindings from reviewed SQL; no manual edits |
| `purchasing/internal/db/generated/purchase_requests.sql.go` | sqlc 1.31.1 bindings from reviewed SQL; no manual edits |
| `purchasing/internal/db/generated/purchase_returns.sql.go` | sqlc 1.31.1 bindings from reviewed SQL; no manual edits |
| `purchasing/internal/db/generated/rfqs.sql.go` | sqlc 1.31.1 bindings from reviewed SQL; no manual edits |
| `purchasing/internal/db/generated/vendor_bills.sql.go` | sqlc 1.31.1 bindings from reviewed SQL; no manual edits |
| `purchasing/internal/domain/purchase_order.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/domain/purchase_request.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/domain/purchase_return.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/domain/vendor_bill.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/dto/purchase_order.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/dto/purchase_request.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/dto/purchase_return.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/dto/vendor_bill.go` | Nullable creator/approval facts and compatible trusted-actor input |
| `purchasing/internal/handler/analytics_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/errors.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/goods_receipt_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/helper/purchase_order_helper.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/helper/purchase_request_helper.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/helper/purchase_return_helper.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/helper/vendor_bill_helper.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/outbox_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/purchase_order_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/purchase_request_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/purchase_return_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/rfq_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/handler/vendor_bill_handler.go` | Trusted actors, approval metadata and safe client errors |
| `purchasing/internal/integration/platform_prerequisites_test.go` | Owned PostgreSQL plus signed in-process HTTP authority/lifecycle/receipt checks |
| `purchasing/internal/logger/errors.go` | Stable stage/SQLSTATE diagnostics without raw errors |
| `purchasing/internal/logger/errors_test.go` | Stable stage/SQLSTATE diagnostics without raw errors |
| `purchasing/internal/middleware/auth.go` | Human identity, exact grants, safe diagnostics, request/body/cancellation bounds |
| `purchasing/internal/middleware/errors.go` | Human identity, exact grants, safe diagnostics, request/body/cancellation bounds |
| `purchasing/internal/middleware/logger.go` | Human identity, exact grants, safe diagnostics, request/body/cancellation bounds |
| `purchasing/internal/middleware/recovery.go` | Human identity, exact grants, safe diagnostics, request/body/cancellation bounds |
| `purchasing/internal/middleware/request_id.go` | Human identity, exact grants, safe diagnostics, request/body/cancellation bounds |
| `purchasing/internal/middleware/safety_test.go` | Human identity, exact grants, safe diagnostics, request/body/cancellation bounds |
| `purchasing/internal/repository/actor.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/platform_prerequisites_test.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/postgres_purchase_order_item_repository.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/postgres_purchase_order_repository.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/postgres_purchase_request_repository.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/postgres_purchase_request_repository_test.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/postgres_purchase_return_repository.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/postgres_vendor_bill_repository.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/purchase_request_item_repository.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/repository/purchase_request_item_repository_test.go` | Trusted creator/approver persistence, typed lifecycle failures and owned DB tests |
| `purchasing/internal/security/jwt.go` | Auth-compatible human RS256 verification and invalid-claim/key tests |
| `purchasing/internal/security/jwt_test.go` | Auth-compatible human RS256 verification and invalid-claim/key tests |
| `purchasing/internal/service/outbox_safety_test.go` | Typed non-draft rejection, disabled/unconfigured Outbox and regression assertions |
| `purchasing/internal/service/outbox_service.go` | Typed non-draft rejection, disabled/unconfigured Outbox and regression assertions |
| `purchasing/internal/service/outbox_service_test.go` | Typed non-draft rejection, disabled/unconfigured Outbox and regression assertions |
| `purchasing/internal/service/purchase_order_service.go` | Typed non-draft rejection, disabled/unconfigured Outbox and regression assertions |
| `purchasing/internal/service/purchase_request_service.go` | Typed non-draft rejection, disabled/unconfigured Outbox and regression assertions |
| `purchasing/internal/service/purchase_request_service_test.go` | Typed non-draft rejection, disabled/unconfigured Outbox and regression assertions |
| `purchasing/internal/testfixture/database.go` | Owned database gate and synthetic signing keys, never normal credentials |
| `purchasing/internal/testfixture/database_test.go` | Owned database gate and synthetic signing keys, never normal credentials |
| `purchasing/internal/testfixture/jwt.go` | Owned database gate and synthetic signing keys, never normal credentials |
| `purchasing/scripts/test-platform-prerequisites.ps1` | Uniquely owned migration/Auth/Go/image fixture and guarded cleanup |
| `Authentication/db/migrations/000013_purchasing_permissions.up.sql` | Approved definitions-only migration; no grants; down preserves established definitions |
| `Authentication/db/migrations/000013_purchasing_permissions.down.sql` | Approved definitions-only migration; no grants; down preserves established definitions |

## Final verification checkpoint

PASS: final Go formatting; working-tree and staged whitespace checks (secret deletion checked by metadata only); UTF-8/new-source whitespace and selected local documentation links; PowerShell parser/native-command regression check; CI/routes/sqlc YAML syntax. Legacy unchanged Markdown whitespace was preserved.

PASS: final baseline hash comparison, all branches/HEADs, old migrations 1–8, local .env existence/unchanged size/mtime and exact index deletion, and sibling Git status counts. The 102 substantive/new Purchasing files in the inventory match actual Git diff/untracked paths. No fixture-labelled containers, owned image tags or temporary fixture directories remain. Final Git metadata and generation evidence are saved under the external task review-artifacts directory.

The last post-runtime changes were test coverage and formatting/documentation only; no application behavior changed after the successful owned suite/image checks.

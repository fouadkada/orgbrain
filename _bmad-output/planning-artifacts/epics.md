---
stepsCompleted: [1, 2]
inputDocuments:
  - planning-artifacts/prds/prd-OrgBrain-MVP-2026-05-26/prd-mvp.md
  - planning-artifacts/prds/prd-OrgBrain-2026-05-25/prd.md
  - planning-artifacts/architecture.md
---

# OrgBrain — Epic Breakdown (Phase 1 MVP)

## Overview

This document provides the complete epic and story breakdown for OrgBrain Phase 1 MVP, decomposing requirements from the MVP PRD, parent PRD (inherited NFRs and privacy architecture), and Architecture into implementable stories. Scope is Slack-first: ingestion, knowledge graph, web UI query interface, single departure risk signal, and admin/access control.

## Requirements Inventory

### Functional Requirements

FR-1p: Slack public channel ingestion — continuous ingestion of public channels after authorization; new messages appear in Knowledge Graph within 15 minutes; DMs and private channels are a hard system-level exclusion (Slack app scope + pipeline filter + runtime assertion — not policy); Admin can exclude specific public channels.

FR-5: Knowledge Node extraction — LLM-based structured extraction from ingested Slack messages; every node carries source, timestamp, author, Confidence Score, Knowledge Owner, and sensitivity_tier tag.

FR-6: Decision versioning — append-only node updates; no overwrites; all previous versions retained with full provenance.

FR-7: Knowledge Ownership Map — per-Member domain association derived from authorship and activity patterns; pre-computed and cached; recomputed every 24 hours and within 60 minutes of a significant activity event; never computed at query time; fallback to node creator for nodes < 24 hours old.

FR-8p: Natural language query (web UI only) — RAG pipeline with SSE streaming LLM responses; <10 second end-to-end response time; first token target <2 seconds; Source Attribution and Confidence Score sent pre-generation (computed from retrieval, not LLM output); session-scoped conversational follow-up with 30-minute idle expiry.

FR-9: Confidence Threshold and Knowledge Owner fallback — tunable Admin setting; pure stateless routing function returning one of ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED; low-confidence answer never surfaced as confident; Confidence Threshold invisible to end users.

FR-10: Staleness indication — answers from content older than staleness threshold (default 6 months, Admin-configurable) are flagged; Knowledge Owner fallback offered alongside stale answers.

FR-13p: Slack engagement velocity signal — 30-day trend in message frequency and response time per Member; requires member_tenure_window captured via first_seen_at on members table from day one; Admin can mark Members as on leave to exclude those windows; signal cards show when decline started and pattern detail, not just severity; computed continuously; auto-resolves on recovery; false positive rate tracked as counter-metric from pilot day one.

FR-14: Intelligence Dashboard display — plain-language signal cards per flagged Member; severity levels Watch / Concerning / Urgent; display-only in Phase 1; Intelligence Tier Members only.

FR-15: Signal card self-explanation — sufficient context for VP-level reader to understand situation and next action without training; includes Member name, plain-language pattern description, duration, severity, and suggested next action; no jargon or raw metrics visible.

FR-16p: Slack OAuth integration management — Admin-triggered connect/disconnect; channel exclusion configuration; existing Knowledge Nodes retained on disconnect.

FR-17: Member access management — Admin-invitation-based onboarding (no SSO in Phase 1); explicit Intelligence Tier assignment per Member; ingestion gate: cannot begin until at least one non-Admin Member has activated; ingestion start date displayed to all Members on activation.

FR-18: Privacy configuration — Admin can enable/disable Engagement Signal clusters; DM exclusion and raw signal exposure limits are non-configurable hard limits; all configuration changes logged with timestamp and Admin identity; disclosure flow requires explicit confirmation, not a checkbox.

FR-TF: Trust-weighted query feedback — thumbs up/down on every query response; trust-weighted signal (not raw vote count); user_trust_profile table with trust_score in [0.05, 1.0]; query_feedback table with weighted_signal generated column; rate limits (20/hour, 3/minute, 2s minimum between submissions, one per query_id per user); anomaly flags computed async every 15 minutes (burst_submission, mono_signal, fast_signal, targeted_downvote, consensus_outlier, cover_tracks).

### NonFunctional Requirements

NFR-1: Accuracy — OrgBrain must never surface a confident answer below the Confidence Threshold; false confidence is a worse failure mode than "I don't know."

NFR-2: Latency — Query Interface <10 seconds end-to-end; first token <2 seconds via streaming; ingestion lag <15 minutes from Slack post; hard LLM timeout at 25 seconds; soft SSE heartbeat at 20 seconds to prevent client-side timeout; query latency tracked as p50/p95/p99 histogram, not average.

NFR-3: Availability — Query Interface and Intelligence Dashboard target 99.5% uptime; ingestion pipelines may tolerate higher downtime; query and ingestion workers run on separate containers and are independent deploy units.

NFR-4: Scalability — up to 150 Members and 3 years of Slack history per org without degraded query performance; minimum 3 concurrent ingestion workers; auto-scaling on queue depth, not CPU; worker count tested against burst load before pilot launch.

NFR-5: Security — OAuth tokens in dedicated secrets manager; per-tenant Postgres schema isolation; database query logging on all tables from day one; runtime Slack scope assertion on ingestion worker startup (hard fail if DM-related scope present).

NFR-6: Privacy/Compliance — three-layer DM exclusion (Slack app scope + pipeline IngestionFilter + runtime scope assertion); Intelligence Tier boundary enforced at adapter layer via sensitivity_tier node tag, not in route handlers; all privacy configuration changes durably logged; disclosure confirmation logged with Admin identity and timestamp.

NFR-7: Pre-pilot load test gate — 10 concurrent queries, p95 < 8 seconds; must pass before pilot launch; runs on release/* branches only.

### Additional Requirements

AR-1: Monorepo scaffolding — five services: api (Go + Chi v5), ai-worker (Python + FastAPI), rag (Python + FastAPI), signal-job (Python cron, no HTTP server), web (Next.js 16.2 App Router TypeScript Tailwind); docker-compose.yml for local dev; Makefile with dev/test/migrate/codegen/load-test targets.

AR-2: Data infrastructure — Postgres 16 + pgvector 0.8.x on Hetzner Volume (network-attached, persists across server reinstalls); PgBouncer in transaction-mode; SET LOCAL search_path = org_{id} at the start of every transaction; session-mode pooling explicitly rejected.

AR-3: Queue — River queue (Postgres-backed, no Redis); JobQueue interface with enqueue/dequeue/ack/nack; Postgres-backed implementation for Phase 1; interface-first so concrete implementation is swappable.

AR-4: Deployment — Hetzner CAX31 (application services) + CAX11 (Coolify v4, isolated); Traefik for SSL/routing managed by Coolify; target ~€18–20/month Phase 1 infrastructure cost.

AR-5: CI/CD — GitHub Actions pipeline: go test + pytest + oapi-codegen schema drift check + cross-tenant isolation tests on every PR; pre-pilot load test on release/* branches only; GHCR image push; Coolify webhook redeploy.

AR-6: KnowledgeStoreAdapter — sole code path for all tenant storage reads and writes; enforces sensitivity_tier at query layer before SQL executes; raises SensitivityTierViolation for unauthorized tier access; sets SET LOCAL search_path internally; no raw db.Query() calls permitted in handlers or workers.

AR-7: IngestionFilter — runs before any LLM call; enforces DM exclusion and low-signal message pre-filtering; DM-type messages rejected with logged audit event; no message reaches embedding or extraction pipeline without passing through IngestionFilter.

AR-8: FallbackRouter — pure stateless function; input is RAG pipeline result struct plus threshold config; output is enum ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED; no side effects, no DB calls, no logging; all four enum values must be covered by tests.

AR-9: OpenAPI contracts — openapi.yaml in ai-worker/ and rag/ are source of truth; Go clients generated via oapi-codegen; CI fails on any schema drift; manual HTTP client code targeting internal services is forbidden.

AR-10: Schema-per-tenant migrations — shared schema migrations applied once; tenant migrations applied to every org_{id} schema via automated runner (Goose); migration runner must be automated before second org is onboarded; migrate-all-tenants.sh script in scripts/.

AR-11: Embeddings — OpenAI text-embedding-3-small; HNSW index on embeddings column mandatory from day one (m=16, ef=64); hard top-K retrieval limit; RETRIEVAL_STRATEGY env var abstraction (pgvector | typesense | hybrid) for Phase 2 upgrade path.

AR-12: Observability from day one — Prometheus /metrics endpoints on all services; key metrics instrumented: query latency histogram (p50/p95/p99), ingestion queue depth per org, ingestion lag per org, LLM call duration, embedding call duration, confidence score distribution, fallback routing type counts, SSE connection count; structured logging with required fields (service, org_id, trace_id, level, msg, duration_ms, error) on every log line across all services.

AR-13: Ingestion start gate — org_status must be 'active' AND at least one non-Admin Member must have activated before ingestion can begin; tenant provisioning is automated (Organization insert triggers background schema creation job).

AR-14: Feedback loop integrity — cover_tracks anomaly flag detects departure-risk signal + downvote spike correlation in same 7-day window; threshold recalibration pipeline rejects batches where >30% of rows carry anomaly flags; false positive rate tracked as counter-metric.

### UX Design Requirements

No UX design document exists for Phase 1 MVP. UX requirements are captured within the functional requirements above (SSE streaming, Source Attribution pre-generation, branch-specific fallback UX, signal card content, admin disclosure confirmation flow).

### FR Coverage Map

| FR | Epic | Notes |
|---|---|---|
| FR-1p | Epic 3 | Slack ingestion pipeline |
| FR-5 | Epic 3 | Knowledge Node extraction |
| FR-6 | Epic 3 | Append-only versioning |
| FR-7 | Epic 3 | Knowledge Ownership Map |
| FR-8p | Epic 4 | NL query + SSE streaming |
| FR-9 | Epic 4 | Fallback routing (4 outcomes) |
| FR-10 | Epic 4 | Staleness indication |
| FR-13p | Epic 5 | Engagement velocity signal |
| FR-14 | Epic 5 | Dashboard display |
| FR-15 | Epic 5 | Signal card self-explanation |
| FR-16p | Epic 3 | Slack OAuth management |
| FR-17 | Epic 2 | Member auth + invitations |
| FR-18 (disclosure) | Epic 2 | Required before ingestion |
| FR-18 (signal config) | Epic 5 | Signal cluster enable/disable |
| FR-TF | Epic 4 | Trust-weighted feedback |
| AR-1 to AR-12 | Epic 1 | All foundation work |
| AR-13 | Epic 2 | Tenant provisioning + ingestion gate |
| AR-14 | Epic 5 | Feedback loop integrity |

## Epic List

### Epic 1: Development Foundation & Local Dev Environment
All five services (api/Go, ai-worker/Python, rag/Python, signal-job/Python, web/Next.js) run locally in docker-compose; CI pipeline passes on empty skeleton; core abstractions (KnowledgeStoreAdapter, IngestionFilter, JobQueue, FallbackRouter) are scaffolded with stub implementations and passing tests; shared and tenant database migrations are structured and runnable; observability skeleton (structured logging, Prometheus /metrics endpoints) is in place across all services.
**ARs covered:** AR-1, AR-2, AR-3, AR-4, AR-5, AR-6, AR-7, AR-8, AR-9, AR-10, AR-11, AR-12

### Epic 2: Organization Setup & Member Authentication
An Admin can create an OrgBrain organization, complete the required disclosure confirmation flow, invite Members, and assign Intelligence Tier access. Members can accept invitations and log in. The ingestion gate (org_status = 'active' AND at least one non-Admin Member activated) is enforced. Tenant schema is provisioned automatically on org creation.
**FRs covered:** FR-17, FR-18 (disclosure confirmation + change logging), AR-13

### Epic 3: Slack Integration & Knowledge Graph
An Admin can connect Slack via OAuth, configure channel exclusions, and OrgBrain continuously ingests public channel messages — extracting Knowledge Nodes, versioning decisions append-only, and maintaining the Knowledge Ownership Map. The Knowledge Graph is populated and internally consistent. DMs are excluded at three enforced layers. Slack OAuth connect lands in the same release increment as disclosure flow completion.
**FRs covered:** FR-16p, FR-1p, FR-5, FR-6, FR-7

### Epic 4: Query Interface
Any authenticated Member can ask a natural language question via the web UI, receive a streaming sourced answer with pre-generated Source Attribution and Confidence Score, get branch-specific fallback routing when confidence is low (ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED), see staleness warnings on old content, and provide trust-weighted feedback. All four fallback routes are handled distinctly in the UI.
**FRs covered:** FR-8p, FR-9, FR-10, FR-TF

### Epic 5: Departure Risk Intelligence
Intelligence Tier Members can view the Intelligence Dashboard showing plain-language Departure Risk Signal cards based on Slack engagement velocity, with severity levels (Watch / Concerning / Urgent), decline start date, and pattern detail. Signal auto-resolves on recovery. Admins can enable/disable signal clusters. Consent architecture is explicitly designed before any signal computation code is written.
**FRs covered:** FR-13p, FR-14, FR-15, FR-18 (signal cluster enable/disable), AR-14

---

## Epic 1: Development Foundation & Local Dev Environment

All five services (api/Go, ai-worker/Python, rag/Python, signal-job/Python, web/Next.js) run locally in docker-compose; CI pipeline passes on empty skeleton; core abstractions (KnowledgeStoreAdapter, IngestionFilter, JobQueue, FallbackRouter) are scaffolded with stub implementations and passing tests; shared and tenant database migrations are structured and runnable; observability skeleton (structured logging, Prometheus /metrics endpoints) is in place across all services.

### Story 1.1: Monorepo Scaffold & Local Dev Environment

As a **developer**,
I want all five services running locally via a single command,
So that I can begin feature work immediately without environment setup friction.

**Acceptance Criteria:**

**Given** the repository is cloned and Docker is running
**When** the developer runs `make dev`
**Then** all five services start (api, ai-worker, rag, signal-job, web) alongside Postgres and PgBouncer
**And** each service responds to its health check (`/healthz` for api; startup log for signal-job)
**And** `make test` passes with empty test suites across all services
**And** the directory structure matches the architecture spec exactly (all named files and folders present as stubs)

---

### Story 1.2: Database Schema & Migration Runner

As a **developer**,
I want shared and tenant database migrations to be structured, runnable, and automated,
So that any engineer can reset to a known schema state with a single command and the system is ready for a second tenant from day one.

**Acceptance Criteria:**

**Given** Postgres is running via docker-compose
**When** the developer runs `make migrate`
**Then** all 5 shared schema migrations apply cleanly (organizations, members, sessions, river_* tables, api_rate_limits, audit_log)
**And** `scripts/migrate-all-tenants.sh` applies tenant migrations to every active `org_{id}` schema
**And** tenant migrations include all 7 tenant schema files (knowledge_nodes, embeddings with HNSW index m=16 ef=64, ownership_map, query_sessions, departure_risk_signals, query_feedback + user_trust_profile, ingestion_events + stage_checkpoints)
**And** the members table includes `first_seen_at TIMESTAMPTZ NOT NULL` from this migration (required for FR-13p signal computation — cannot be retrofitted)
**And** running `make migrate` twice is idempotent (no errors on re-run)
**And** a fresh `org_{id}` schema can be created and fully migrated by the tenant runner

---

### Story 1.3: Core Abstractions — KnowledgeStoreAdapter & IngestionFilter

As a **developer**,
I want KnowledgeStoreAdapter and IngestionFilter scaffolded with enforced contracts and passing tests,
So that all future storage and ingestion code is written against stable, safety-enforcing interfaces from day one.

**Acceptance Criteria:**

**Given** the stub KnowledgeStoreAdapter is in place
**When** any code path attempts a storage read or write
**Then** it must go through `KnowledgeStoreAdapter.WithTenant(orgID)` — direct `db.Query()` calls in handlers or workers cause a CI lint failure
**And** `WithTenant()` sets `SET LOCAL search_path = org_{id}` inside the transaction before any SQL executes
**And** `KnowledgeStoreAdapter.query()` raises `SensitivityTierViolation` when a request attempts to access a sensitivity tier above the caller's assigned tier (stub test confirms the error type)
**And** cross-tenant isolation tests pass: write to Org A → assert unreachable from Org B (both read and write directions) in `adapter_test.go`
**And** `IngestionFilter` stub rejects any message with `channel_type = im` or `channel_type = mpim` with a logged audit event (three DM exclusion layers enumerated in code comments: Slack app scope restriction, pipeline event_type filter, runtime scope assertion)
**And** all tests pass via `make test`

---

### Story 1.4: Core Abstractions — JobQueue, FallbackRouter & OpenAPI Contracts

As a **developer**,
I want the JobQueue interface, FallbackRouter decision logic, and internal OpenAPI contracts scaffolded with passing tests,
So that ingestion workers and the RAG pipeline have well-defined, testable contracts before any feature implementation begins.

**Acceptance Criteria:**

**Given** the JobQueue interface is scaffolded
**When** the River-backed implementation is used
**Then** it exposes `enqueue`, `dequeue`, `ack`, `nack` methods matching the interface contract
**And** swapping the concrete implementation requires no changes to worker logic (interface-first verified by a stub alternate implementation in tests)

**Given** the FallbackRouter is scaffolded with the decision matrix
**When** each of the four input conditions is presented:
- `candidate_node_ids` empty → `NO_COVERAGE`
- ownership map resolves → `ROUTE_TO_OWNER`
- sensitivity filter removed all candidates → `ACCESS_FILTERED`
- chunks present but low coherence → `REPHRASE`
**Then** the correct enum value is returned for each case
**And** the FallbackRouter has no side effects, no DB calls, no logging (pure function — verified by test isolation)
**And** all four enum outcomes are covered by unit tests in `rag/tests/test_fallback.py`

**Given** the OpenAPI specs are scaffolded at `ai-worker/openapi.yaml` and `rag/openapi.yaml`
**When** `oapi-codegen` runs in CI
**Then** the generated Go client code in `api/internal/client/` matches the specs with zero diff
**And** CI fails on any schema drift between the OpenAPI specs and the generated clients

---

### Story 1.5: CI/CD Pipeline & Observability Skeleton

As a **developer**,
I want every code push automatically tested and deployable, with structured logging and metrics instrumentation in place across all services,
So that the team can ship with confidence and observability is never retrofitted.

**Acceptance Criteria:**

**Given** a pull request is opened against main
**When** CI runs
**Then** `go test ./...` and `pytest` pass
**And** `oapi-codegen` schema drift check passes (regen + diff)
**And** cross-tenant isolation tests pass (both read and write directions)
**And** Docker images for all five services build successfully and push to GHCR

**Given** a push to a `release/*` branch
**When** CI runs the release pipeline
**Then** the pre-pilot load test step is present (stubbed as skipped until pilot-ready) and the Coolify webhook trigger fires on success

**Given** any service handles a request or performs an operation
**When** it logs an event
**Then** every log line carries the required fields: `service`, `level`, `msg`, `trace_id`; and `org_id` on all tenant-scoped operations; and `duration_ms` on all external calls (stubs log zero values — actual values populated in later epics)

**Given** any service starts up
**When** it is running
**Then** a Prometheus `/metrics` endpoint responds with 200 and the following counters/histograms are registered (even if unpopulated): query latency histogram, ingestion queue depth, ingestion lag, LLM call duration, embedding call duration, confidence score distribution, fallback routing type counts, SSE connection count

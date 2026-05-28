---
stepsCompleted: [1, 2, 3, 4]
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

---

## Epic 2: Organization Setup & Member Authentication

An Admin can create an OrgBrain organization, complete the required disclosure confirmation flow, invite Members, and assign Intelligence Tier access. Members can accept invitations and log in. The ingestion gate (org_status = 'active' AND at least one non-Admin Member activated) is enforced. Tenant schema is provisioned automatically on org creation.

### Story 2.1: Admin Registration & Tenant Provisioning

As an **admin**,
I want to register an OrgBrain organization,
So that the system creates a dedicated tenant environment and I can begin configuration.

**Acceptance Criteria:**

**Given** the registration endpoint receives a valid org name and admin email
**When** the admin submits registration
**Then** an `organizations` record is created with `org_status = 'provisioning'`
**And** a background River job runs `CREATE SCHEMA org_{id}` and applies all 7 tenant migrations
**And** `org_status` flips to `'active'` only after all tenant migrations succeed
**And** the admin receives a session cookie with httpOnly flag

**Given** tenant provisioning fails (e.g., migration error)
**When** the River job encounters an error
**Then** `org_status` remains `'provisioning'` and the error is logged with `service=api` and `org_id`
**And** the admin cannot proceed past registration until status is `'active'`

---

### Story 2.2: Required Organizational Disclosure Confirmation

As an **admin**,
I want to complete the organizational disclosure confirmation,
So that employees are informed of OrgBrain's deployment before any ingestion begins.

**Acceptance Criteria:**

**Given** the admin has an active org account
**When** they access the admin panel before completing disclosure
**Then** they are redirected to the disclosure confirmation step
**And** the disclosure step presents the template notice and requires explicit confirmation (not a checkbox — a deliberate affirm action)

**Given** the admin submits the disclosure confirmation
**When** it is recorded
**Then** an append-only record is written to `audit_log` with `admin_id`, `timestamp`, and `event_type = 'disclosure_confirmed'`
**And** the confirmation is visible in the Admin panel's disclosure log viewer

**Given** disclosure has not been confirmed
**When** any ingestion-enabling action is attempted
**Then** the system blocks the action and surfaces the disclosure requirement with a clear explanation

---

### Story 2.3: Member Invitation & Account Activation

As an **admin**,
I want to invite Members to OrgBrain,
So that they can access the Query Interface and I can control who joins the organization.

**Acceptance Criteria:**

**Given** the admin sends an invitation to a valid email address
**When** the invitation is created
**Then** an `invitations` record is created with a secure random token and 7-day expiry
**And** the invitation URL (`/invite/[token]`) is surfaced to the admin for distribution

**Given** a Member visits a valid invitation URL
**When** they complete account activation
**Then** a `members` record is created with `first_seen_at = NOW()` captured at activation time (this field cannot be reconstructed retroactively — required for FR-13p signal computation)
**And** the Member receives a session cookie
**And** the org's non-Admin activation count increments

**Given** a Member is deactivated by the admin
**When** deactivation is confirmed
**Then** all sessions for that Member are immediately deleted (`DELETE FROM sessions WHERE member_id = $1`)
**And** their Knowledge Graph contributions are retained

---

### Story 2.4: Member Login & Session Management

As a **member**,
I want to log in to OrgBrain and have my session persist across browser restarts,
So that I can access the Query Interface without re-authenticating on every visit.

**Acceptance Criteria:**

**Given** a Member submits valid credentials on the login page
**When** authentication succeeds
**Then** a session record is created in the `sessions` table (shared schema) with `session_id` in an httpOnly cookie
**And** the Member is redirected to the Query Interface

**Given** a Member submits invalid credentials
**When** authentication fails
**Then** the response is a generic "invalid credentials" error (no user enumeration — same message for both bad password and unknown email)
**And** no session is created

**Given** a Member with an active session navigates to an authenticated route
**When** the auth middleware validates the session cookie
**Then** `org_id` and `member_id` are injected into the request context
**And** the request proceeds to the handler

**Given** a Member logs out
**When** logout is requested
**Then** the session record is deleted and the cookie is cleared

---

### Story 2.5: Intelligence Tier Assignment & Ingestion Gate

As an **admin**,
I want to assign Intelligence Tier access to specific Members and enforce the ingestion readiness gate,
So that departure risk signals are only visible to authorized leadership, and ingestion only begins once the org is truly ready.

**Acceptance Criteria:**

**Given** the admin views the Member list
**When** they assign Intelligence Tier to a Member
**Then** the member's `tier` is updated and the change is logged to `audit_log` with `admin_id` and timestamp
**And** the assigned Member can see the Intelligence Dashboard link in navigation

**Given** a non-Intelligence-Tier Member requests `GET /v1/signals`
**When** the request is processed
**Then** a 403 is returned (tier check happens in auth middleware, not in the handler)

**Given** disclosure is confirmed and at least one non-Admin Member has activated
**When** the ingestion gate is evaluated
**Then** `org_status = 'active'` AND the non-Admin activation count > 0 — both conditions must be true
**And** the ingestion start date is displayed to all Members on activation ("Knowledge Graph building since [date]")

**Given** disclosure is confirmed but no non-Admin Member has activated yet
**When** admin attempts to connect Slack integration
**Then** the system allows the OAuth connection to be prepared but queues ingestion jobs as blocked until the gate passes

---

## Epic 3: Slack Integration & Knowledge Graph

An Admin can connect Slack via OAuth, configure channel exclusions, and OrgBrain continuously ingests public channel messages — extracting Knowledge Nodes, versioning decisions append-only, and maintaining the Knowledge Ownership Map. The Knowledge Graph is populated and internally consistent. DMs are excluded at three enforced layers. Slack OAuth connect lands in the same release increment as disclosure flow completion.

### Story 3.1: Slack OAuth Connection & Channel Management

As an **admin**,
I want to connect OrgBrain to our Slack workspace and control which channels are ingested,
So that the Knowledge Graph is populated from the right sources and nothing is ingested that we haven't approved.

**Acceptance Criteria:**

**Given** the admin navigates to Integrations and clicks "Connect Slack"
**When** they complete the Slack OAuth flow
**Then** the Slack access token is stored encrypted in the `integrations` table (tenant schema) via `pgcrypto` with key from Coolify env vars
**And** the Slack app scope is limited to public channels only — no DM-related scope in the manifest
**And** `scope_assert.go` runs on ingestion worker startup and hard-fails if any DM-related scope is present in the stored token

**Given** the Slack connection is active
**When** the admin views channel exclusions
**Then** they can toggle specific public channels off; excluded channels are persisted and respected by the ingestion pipeline

**Given** the admin disconnects the Slack integration
**When** disconnection is confirmed
**Then** future ingestion stops immediately
**And** all existing Knowledge Nodes from that workspace are retained in the Knowledge Graph

---

### Story 3.2: Slack Webhook Receiver & Reconciliation Job

As a **system**,
I want to receive Slack events reliably and recover from webhook delivery gaps,
So that the Knowledge Graph reflects all public channel messages within 15 minutes of posting.

**Acceptance Criteria:**

**Given** Slack sends a message event to `POST /webhooks/slack`
**When** the webhook is received
**Then** the handler verifies the Slack signature and acks with HTTP 200 within 3 seconds
**And** it enqueues a single River ingestion job — no processing, no DB reads, no LLM calls in the receiver
**And** the `events.go` parser rejects any event with `channel_type = im` or `channel_type = mpim` before enqueueing (second DM exclusion layer — pipeline-level)

**Given** a duplicate Slack event arrives (same `workspace_id + channel_id + message_ts + thread_ts`)
**When** the ingestion job attempts to write
**Then** the DB-level unique constraint on `ingestion_events` prevents the duplicate (idempotency is DB-enforced, not application logic)

**Given** the webhook has missed events (e.g., downtime window)
**When** the reconciliation job runs (every 30 minutes, 1-hour lookback)
**Then** it polls `conversations.history` for all non-excluded channels and enqueues any messages not yet seen in `ingestion_events`

---

### Story 3.3: Ingestion Pipeline — Filter, Embed & Stage Checkpointing

As a **system**,
I want every Slack message to be filtered before processing and embedded with stage-level fault tolerance,
So that DMs can never reach the LLM and failed ingestion jobs recover at the failed stage rather than restarting from scratch.

**Acceptance Criteria:**

**Given** an ingestion job is dequeued by a River worker
**When** it runs the filter stage
**Then** `IngestionFilter` runs first — before any embedding or LLM call
**And** any message with `channel_type = im` or `channel_type = mpim` is rejected with an audit event logged (`event_type = 'dm_rejection'`, message metadata only — no content)
**And** low-signal messages (empty, bot-only, system events) are also filtered before embedding

**Given** a message passes the filter
**When** the embed stage runs
**Then** `ai-worker POST /embed` is called with the message batch
**And** the stage checkpoint (`ingestion_event_id, stage='embed', status='complete'`) is written before proceeding

**Given** an ingestion job fails at any stage
**When** it is re-enqueued
**Then** it resumes at the failed stage only (not from the beginning) — verified by reading the stage checkpoint
**And** after 3 consecutive failures at the same stage, the job moves to dead-letter with the error logged

---

### Story 3.4: Knowledge Node Extraction & Append-Only Storage

As a **system**,
I want embedded Slack messages to be extracted into structured Knowledge Nodes stored append-only,
So that the Knowledge Graph accumulates organizational knowledge with full decision history preserved.

**Acceptance Criteria:**

**Given** an embedded message batch reaches the extract stage
**When** `ai-worker POST /extract` is called
**Then** Anthropic LLM extraction produces structured Knowledge Nodes with: content, source reference, source date, author (if determinable), Confidence Score, Knowledge Owner (if determinable), and `sensitivity_tier` tag
**And** the stage checkpoint for `extract` is written before the write stage begins

**Given** extracted nodes are written via `KnowledgeStoreAdapter`
**When** a node representing a decision that was previously stored is updated
**Then** the new version is appended — the original is never overwritten or deleted
**And** both the original and updated state are retrievable with their respective dates and sources
**And** the HNSW index on the embeddings column is updated for each new node

**Given** the ingestion pipeline writes a node
**When** `KnowledgeStoreAdapter.write()` is called
**Then** `sensitivity_tier` is tagged at write time — never inferred at query time
**And** no raw `db.Query()` call exists in any handler or worker (CI lint enforces this)

---

### Story 3.5: Knowledge Ownership Map

As a **system**,
I want the Knowledge Ownership Map to be pre-computed and kept current,
So that the Query Interface can instantly surface Knowledge Owners without any query-time computation.

**Acceptance Criteria:**

**Given** the Knowledge Graph has been populated with nodes
**When** the ownership recompute River job runs (every 24 hours)
**Then** ownership assignments are updated in the `ownership_map` table for all nodes where authorship or activity patterns have changed
**And** every node with a determinable owner has a Knowledge Owner assigned

**Given** a significant activity event occurs (e.g., a Member authors a node that becomes a key decision)
**When** the event triggers the 60-minute recompute job
**Then** the ownership map is updated within 60 minutes for the affected domains

**Given** a node was created less than 24 hours ago and no ownership assignment exists yet
**When** the Knowledge Owner is needed (e.g., fallback routing)
**Then** the system falls back to the node's creator as the Knowledge Owner

**Given** the ingestion pipeline is running
**When** ingestion lag is checked
**Then** `GET /v1/ingestion-lag` returns the current lag per org (time since most recent successfully processed message)
**And** the Query UI displays "Knowledge Graph current as of [timestamp]" using this endpoint

---

## Epic 4: Query Interface

Any authenticated Member can ask a natural language question via the web UI, receive a streaming sourced answer with pre-generated Source Attribution and Confidence Score, get branch-specific fallback routing when confidence is low (ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED), see staleness warnings on old content, and provide trust-weighted feedback. All four fallback routes are handled distinctly in the UI.

### Story 4.1: RAG Pipeline — Query Embedding, Retrieval & Confidence Scoring

As a **system**,
I want queries to be embedded, retrieved against the Knowledge Graph, and scored for confidence before any LLM generation begins,
So that Source Attribution is always computed from evidence — never from LLM output — and fallback routing can fire before generation wastes budget.

**Acceptance Criteria:**

**Given** the RAG service receives a query request with `org_id` via `X-Org-ID` header
**When** the pipeline runs
**Then** the query is embedded via OpenAI `text-embedding-3-small`
**And** HNSW top-K retrieval runs against the tenant's `embeddings` table (search_path set via `SET LOCAL` in `rag/db.py`)
**And** the Confidence Score is computed from retrieval results (source recency, source authority, consistency across sources) — before LLM generation starts
**And** Source Attribution (source name, author, date, link) is derived from retrieved nodes — also pre-generation

**Given** retrieval returns results that include nodes above the caller's `sensitivity_tier`
**When** `KnowledgeStoreAdapter.query()` runs
**Then** nodes above the caller's tier are filtered out before any result is returned (enforcement at adapter layer, not in route handler)
**And** if filtering removes all candidates, the FallbackRouter receives `ACCESS_FILTERED`

**Given** retrieval returns an empty result set
**When** the FallbackRouter evaluates
**Then** it returns `NO_COVERAGE` (pure function — no DB calls, no logging, no side effects)

---

### Story 4.2: RAG Pipeline — LLM Streaming Generation & SSE Protocol

As a **member**,
I want to see the answer sources appear immediately and the answer stream in as it's generated,
So that I can start evaluating the answer's provenance before reading the full response.

**Acceptance Criteria:**

**Given** confidence meets the threshold and retrieval returned candidates
**When** the RAG service begins generation
**Then** a `meta` SSE event is emitted first: `{"confidence": float, "sources": [...], "routing": null}`
**And** `data` SSE events follow, one per LLM token: `{"token": "..."}`
**And** the `meta` event always arrives before the first `data` token — Source Attribution is visible to the user before the answer text starts rendering

**Given** LLM generation exceeds 25 seconds
**When** the hard timeout fires
**Then** the stream closes with an `error` SSE event: `{"type": "timeout", "routing": "REPHRASE", "suggestions": [...]}`
**And** the client renders source nodes from the last received `meta` event with a "couldn't synthesize a full answer" message
**And** an SSE heartbeat comment is sent at 20 seconds to prevent client-side connection timeout

**Given** a query response is delivered
**When** query latency is recorded
**Then** it is tracked as a p50/p95/p99 histogram (not average) in the `query_latency` Prometheus metric
**And** the Confidence Score and fallback routing type are logged at decision time, including the threshold value in effect

---

### Story 4.3: Query Session Management & Fallback Routing

As a **member**,
I want to ask follow-up questions within a session and receive a clear, branch-specific response when OrgBrain can't answer,
So that conversational context is preserved and I always know exactly what to do when confidence is too low.

**Acceptance Criteria:**

**Given** a Member submits a query
**When** a session does not yet exist
**Then** a `query_sessions` record is created with `conversation_history` as JSONB and `last_active_at = NOW()`

**Given** a Member submits a follow-up question within 30 minutes of the last query
**When** the RAG service receives the request
**Then** the conversation history from the session is included in the context window
**And** `last_active_at` is updated

**Given** a session has been idle for more than 30 minutes
**When** the River expiry job runs
**Then** the `query_sessions` row is deleted
**And** the next query from that Member starts a fresh session with no prior context

**Given** the FallbackRouter returns `ROUTE_TO_OWNER`
**When** the UI renders the fallback
**Then** the `FallbackCard` shows the Knowledge Owner's name and Slack handle with a deep-link to the most relevant Slack thread

**Given** the FallbackRouter returns `NO_COVERAGE`
**When** the UI renders the fallback
**Then** the `FallbackCard` shows a "Flag for your team" action that turns the unanswered question into a contribution signal

**Given** the FallbackRouter returns `REPHRASE`
**When** the UI renders the fallback
**Then** the `FallbackCard` shows two LLM-generated query variant suggestions as one-click re-queries

**Given** the FallbackRouter returns `ACCESS_FILTERED`
**When** the UI renders the fallback
**Then** the `FallbackCard` states that an answer exists but is not accessible at the Member's tier — no further detail is surfaced

---

### Story 4.4: Query Web UI — Streaming Interface & Staleness Warning

As a **member**,
I want a clean query interface where sources appear before the answer and stale content is flagged,
So that I can trust what I'm reading and know when to verify with a colleague.

**Acceptance Criteria:**

**Given** a Member navigates to the Query Interface
**When** the page loads
**Then** a natural language input (`QueryInput`) and submit button are rendered
**And** the authenticated route guard redirects unauthenticated users to login

**Given** a Member submits a query
**When** the SSE stream begins
**Then** `useQueryStream` hook connects to `POST /v1/query` via `EventSource`
**And** on `meta` event: `SourceAttribution` cards render immediately (source name, author, date) before any answer text
**And** on `data` events: answer tokens stream into the `QueryStream` component in real time
**And** on `error` event: the stream closes and the error state renders inline (not as a toast — user needs context)

**Given** a retrieved node is older than the staleness threshold (default 6 months)
**When** the answer renders
**Then** a `StalenessWarning` component renders inline with the answer
**And** the most recent related context is surfaced alongside the original
**And** the Knowledge Owner fallback is offered

**Given** the ingestion lag exceeds the configurable staleness banner threshold
**When** `useIngestionStatus` polls `GET /v1/ingestion-lag`
**Then** a banner renders: "Knowledge Graph last updated [time ago]" to set expectations before querying

---

### Story 4.5: Trust-Weighted Query Feedback

As a **member**,
I want to give a thumbs up or down on answers,
So that OrgBrain can calibrate its Confidence Threshold based on real answer quality signals.

**Acceptance Criteria:**

**Given** a query response is displayed
**When** the Member sees `FeedbackButtons`
**Then** thumbs up and thumbs down are available with a 2-second minimum between submissions enforced client-side and server-side

**Given** a Member submits feedback
**When** `POST /v1/feedback` is called
**Then** a `query_feedback` row is created with `weighted_signal = signal × trust_weight` (generated column)
**And** rate limits are enforced: max 20/hour, max 3/minute, one per `query_id` per user (DB unique index)
**And** submissions exceeding rate limits return 429 with a `Retry-After` header

**Given** the async anomaly detection job runs (every 15 minutes)
**When** it processes recent feedback rows
**Then** it flags: `burst_submission`, `mono_signal`, `fast_signal`, `targeted_downvote`, `consensus_outlier`
**And** it flags `cover_tracks` when a departure-risk-flagged Member has a downvote spike in the same 7-day window (without exposing the departure risk signal to any unauthorized party)

**Given** a threshold recalibration batch is processed
**When** more than 30% of rows in the batch carry anomaly flags
**Then** the batch is rejected and logged — not applied to the Confidence Threshold

---

## Epic 5: Departure Risk Intelligence

Intelligence Tier Members can view the Intelligence Dashboard showing plain-language Departure Risk Signal cards based on Slack engagement velocity, with severity levels (Watch / Concerning / Urgent), decline start date, and pattern detail. Signal auto-resolves on recovery. Admins can enable/disable signal clusters. Consent architecture is explicitly designed before any signal computation code is written.

### Story 5.1: Engagement Velocity Signal Computation

As a **system**,
I want to continuously compute Slack engagement velocity trends per Member and store the results,
So that the Intelligence Dashboard always reflects current departure risk state without manual intervention.

**Acceptance Criteria:**

**Given** `signal-job` runs (cron schedule)
**When** it processes active orgs
**Then** for each Member it computes a 30-day trend in: message frequency in public channels and average response time in public channels
**And** `first_seen_at` (member_tenure_window) is used to exclude periods before the Member joined — this field was captured at activation in Story 2.3 and cannot be reconstructed retroactively
**And** any Member period marked as on-leave by the Admin is excluded from the trend window

**Given** a Member's engagement velocity shows a sustained declining trend over 30 days
**When** the signal is evaluated
**Then** a `departure_risk_signals` row is written (or updated) with severity, `pattern_detail`, and `decline_started_at`
**And** if a signal already exists for that Member and the trend has recovered
**Then** `resolved_at` is set to NOW() and the signal is no longer surfaced on the dashboard

**Given** insufficient data exists to compute a reliable baseline (minimum window TBD during beta)
**When** the job runs for that Member
**Then** a "building baseline" state is recorded — no signal row is written, and the dashboard shows the cold-start state for that Member

---

### Story 5.2: Signal Severity Classification

As a **system**,
I want departure risk signals to carry severity levels with enough detail for a VP to understand the pattern without any training,
So that the Intelligence Dashboard is self-explanatory and actionable.

**Acceptance Criteria:**

**Given** a departure risk signal is computed for a Member
**When** severity is classified
**Then** severity is one of: `Watch` (early decline, < 14 days sustained), `Concerning` (14–28 days sustained), `Urgent` (> 28 days sustained or steep drop)
**And** `pattern_detail` is stored as a plain-language string describing the observed pattern (e.g., "Message frequency down 60% over 31 days; average response time increased from 2h to 14h")
**And** `decline_started_at` records when the declining trend was first detected

**Given** a signal row exists in `departure_risk_signals`
**When** the severity threshold changes on a subsequent computation run (trend worsened or improved)
**Then** the existing row is updated in place with the new severity and pattern detail
**And** `resolved_at` is set only when the trend fully recovers — partial improvement does not resolve the signal

**Given** the false positive rate counter-metric
**When** a signal auto-resolves within 7 days of being surfaced
**Then** it is counted as a potential false positive in the `signal_false_positive_rate` Prometheus counter (for pilot calibration)

---

### Story 5.3: Intelligence Dashboard — Signal Cards & Cold Start

As a **VP of Engineering**,
I want to see clear departure risk signal cards for flagged team Members with enough context to act,
So that I can initiate retention conversations before someone hands in a resignation.

**Acceptance Criteria:**

**Given** an Intelligence Tier Member navigates to the Intelligence Dashboard
**When** the page loads
**Then** `GET /v1/signals` is called; the API enforces Intelligence Tier gate at middleware (not in the handler)
**And** `useSignalCards` polls every 60 seconds to reflect current signal state

**Given** active signals exist
**When** they render
**Then** each `SignalCard` displays: Member name, severity badge (Watch / Concerning / Urgent), plain-language pattern description, decline start date, and a suggested next action (e.g., "Consider a 1:1 check-in via their Senior EM")
**And** a `SignalSparkline` (Recharts) renders the 30-day engagement trend inline on each card
**And** no raw metrics, numerical scores, or model internals are visible at any point

**Given** no signals exist yet (org is in cold-start period)
**When** the dashboard loads
**Then** `BaselineState` renders: "OrgBrain is building your team's engagement baseline. Signals will appear once enough data has been collected."
**And** no empty state is shown that could be misread as "no risk detected"

**Given** a signal has `resolved_at` set
**When** the dashboard loads
**Then** the resolved signal does not appear — the dashboard shows only active signals

---

### Story 5.4: Admin Signal Configuration & Privacy Audit Log

As an **admin**,
I want to enable or disable engagement signal clusters and see a complete log of all privacy configuration changes,
So that our organization controls what is monitored and we have an auditable record for compliance purposes.

**Acceptance Criteria:**

**Given** the admin opens Privacy Configuration in the admin panel
**When** the page loads
**Then** each Engagement Signal cluster (Phase 1: Slack engagement velocity only) is shown with a toggle and plain-language description of what it measures
**And** the full disclosure log is visible: every `audit_log` entry with `event_type` in (`disclosure_confirmed`, `signal_cluster_toggled`, `privacy_config_changed`), with admin identity and timestamp

**Given** the admin toggles a signal cluster off
**When** the change is saved
**Then** an append-only `audit_log` row is written: `admin_id`, `timestamp`, `event_type = 'signal_cluster_toggled'`, `cluster_id`, `new_state`
**And** `signal-job` respects the disabled cluster on the next computation run — no signals are written for disabled clusters

**Given** the admin re-enables a previously disabled cluster
**When** `signal-job` next runs
**Then** computation resumes for that cluster from the current date — no retroactive signals are generated for the disabled period

**Given** the feedback loop integrity check runs (AR-14)
**When** a threshold recalibration batch has > 30% anomaly-flagged rows
**Then** the batch is rejected, an `audit_log` entry is written with `event_type = 'recalibration_batch_rejected'`, and the current threshold is unchanged

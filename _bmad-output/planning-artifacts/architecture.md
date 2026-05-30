---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
lastStep: 8
status: 'complete'
completedAt: '2026-05-26'
inputDocuments:
  - planning-artifacts/briefs/brief-OrgBrain-2026-05-22/brief.md
  - planning-artifacts/briefs/brief-OrgBrain-2026-05-22/addendum.md
  - planning-artifacts/prds/prd-OrgBrain-2026-05-25/prd.md
  - planning-artifacts/prds/prd-OrgBrain-MVP-2026-05-26/prd-mvp.md
workflowType: 'architecture'
project_name: 'OrgBrain'
user_name: 'fouad'
date: '2026-05-26'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

---

## Project Context Analysis

### Requirements Overview

**Functional Requirements (Phase 1 MVP):**

- **FR-1p — Slack ingestion:** Continuous ingestion of public Slack channels post-authorization. New messages appear in Knowledge Graph within 15 minutes. Admin can exclude specific channels. DMs and private channels are a hard system-level exclusion enforced at the Slack app scope level and in the ingestion pipeline — not just policy.
- **FR-5 — Knowledge Node extraction:** LLM-based structured extraction from ingested Slack messages; every node carries source, timestamp, author, Confidence Score, Knowledge Owner, and `sensitivity_tier` tag.
- **FR-6 — Decision versioning:** Append-only node updates; no overwrites. Previous versions retained with full provenance.
- **FR-7 — Knowledge Ownership Map:** Per-Member domain association derived from authorship and activity patterns. Pre-computed and cached; recomputed every 24 hours and within 60 minutes of a significant activity event. Never computed at query time. Fallback to node creator for nodes < 24 hours old.
- **FR-8p — Natural language query (web UI only):** RAG pipeline with streaming LLM responses (SSE or WebSocket); <10 second response time with first token target <2 seconds; Source Attribution and Confidence Score sent pre-generation (computed from retrieval, not LLM output); session-scoped conversational follow-up (30-minute idle expiry).
- **FR-9 — Confidence Threshold + Knowledge Owner fallback:** Tunable parameter exposed as a first-class Admin setting. Routing logic is a pure stateless function returning one of: `ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED`. Never surface a low-confidence answer as confident. Confidence threshold is invisible to end users.
- **FR-10 — Staleness indication:** Answers from content older than threshold (default 6 months, Admin-configurable) flagged; Knowledge Owner fallback offered alongside.
- **FR-13p — Slack engagement velocity signal:** 30-day trend in message frequency and response time per Member. Signal computation requires `member_tenure_window` (captured via `first_seen_at` on members table from day one). Admin can mark Members as on leave to exclude those windows from computation. Signal cards show when the decline started and the pattern detail — not just severity. Computed continuously; auto-resolves on recovery. False positive rate tracked as a counter-metric from pilot day one.
- **FR-14/FR-15 — Intelligence Dashboard:** Plain-language signal cards; severity levels (Watch / Concerning / Urgent); display-only in Phase 1 (no "Mark as Reviewed" action logging until Phase 2).
- **FR-16p — Slack OAuth integration management:** Admin-triggered connect/disconnect; channel exclusion; existing nodes retained on disconnect.
- **FR-17 — Member access management:** Admin-invitation-based onboarding (no SSO in Phase 1); explicit Intelligence Tier assignment per Member. Ingestion cannot begin until at least one non-Admin Member has activated. Ingestion start date displayed to all Members on activation.
- **FR-18 — Privacy configuration:** Admin can enable/disable Engagement Signal clusters. Hard limits (DM exclusion, no raw signal exposure) are non-configurable. All changes logged with timestamp and Admin identity. Disclosure flow requires explicit confirmation — not a checkbox.

**User Feedback System:**

- Thumbs up/down on every query response. Trust-weighted signal — not a raw vote counter.
- `query_feedback` table with `weighted_signal = signal × trust_weight` (generated column).
- `user_trust_profile` table with `trust_score` in [0.05, 1.0] — floor of 0.05 so rogue signals are discounted but auditable, never silently dropped.
- Rate limits: max 20/hour, max 3/minute, min 2s between submissions, one per `query_id` per user (DB unique index).
- Anomaly flags (async, every 15 min): `burst_submission`, `mono_signal`, `fast_signal`, `targeted_downvote`, `consensus_outlier`, `cover_tracks` (departure-risk flag + downvote spike in same 7-day window).
- Aggregate score = `SUM(weighted_signal) / COUNT(*)` where `suppressed = FALSE`. Threshold recalibration rejects batches where >30% of rows carry anomaly flags.

**Success Metrics (revised):**

- **SM-P1:** First paying pilot within 90 days of launch.
- **SM-P2:** At least one Departure Risk Signal acted on during pilot, surfaced 30+ days before escalation. (Leading indicator: VP opens signal card and initiates 1:1 within 7 days of surfacing.)
- **SM-P3:** Knowledge Owner fallback accuracy above 80% in first 30 days. Measured via thumbs up/down on fallback responses.
- **SM-P4a (replaces SM-P4):** At least 60% of ICs who complete a first query report the answer was "useful" via immediate post-query yes/no prompt. Validates first-session trust.
- **SM-P4b:** At least 30% of ICs who answered "yes" in SM-P4a return to submit a second query within 14 days. Validates trust-to-habit conversion.
- **SM-P5:** 50%+ of queries answered without a follow-up Slack ping within first 60 days.

**Non-Functional Requirements:**

- **Latency:** Query Interface <10 seconds end-to-end; first token <2 seconds via streaming; ingestion lag <15 minutes from Slack post. Query latency instrumented as p50/p95/p99 histogram — not average. Hard LLM timeout at 25 seconds; soft SSE heartbeat at 20 seconds to prevent client-side WebSocket timeout.
- **Availability:** Query Interface and Intelligence Dashboard target 99.5% uptime. Ingestion pipelines may tolerate higher downtime. Query and ingestion workers run on separate compute (separate containers, independent deploy units).
- **Scalability:** Up to 150 Members and 3 years of Slack history per org without degraded query performance. Minimum 3 concurrent ingestion workers; auto-scaling on queue depth, not CPU. Worker count tested against burst load before pilot launch.
- **Security:** OAuth tokens in dedicated secrets manager. Per-tenant Postgres schema isolation. Database query logging on all tables from day one. Runtime Slack scope assertion on ingestion worker startup.
- **Privacy/Compliance:** Three-layer DM exclusion: Slack app scope restriction + pipeline-level message type filter + runtime scope assertion. Intelligence Tier boundary enforced at data model level via `sensitivity_tier` node tag — enforced at adapter layer, not route handler. All privacy configuration changes durably logged.
- **Accuracy:** Confidence Threshold is a first-class Admin-configurable setting. Answer quality logging (confidence score, fallback type, trust-weighted user feedback) required from day one for empirical calibration.

**Scale & Complexity:**

- Primary domain: Full-stack web application + async AI/ML backend pipeline
- Complexity level: High
- Key scaling dimensions: per-org schema isolation, LLM extraction throughput, vector retrieval at 3 years of Slack history, time-series analytics per Member

### Technical Constraints & Dependencies

- **Slack API:** OAuth-based authorization; Events API webhook for public channel messages. Webhook receiver must ack within 3 seconds (Slack requirement) — thin ack-and-enqueue only, no processing in the receiver. Reconciliation job polls Slack `conversations.history` API on a configurable interval (default: every 30 minutes, 1-hour lookback window) to recover webhook gaps. Idempotency key: `(workspace_id, channel_id, message_ts, thread_ts)` — `thread_ts` is `NULL` for top-level messages, non-null for replies. DB-level unique constraint, not application logic.
- **LLM dependency:** Node extraction (FR-5) and query answering (FR-8p) both require LLM calls. Extraction is async and parallelizable. Query generation uses streaming API. Hard timeout 25 seconds with SSE heartbeat at 20 seconds. Graceful fallback: stream closes with structured error event; client renders source nodes with "couldn't synthesize a full answer" message.
- **Connection pooling:** PgBouncer in transaction-mode. `SET LOCAL search_path = org_{id}` at the start of every transaction. Middleware assertion rejects any query attempted outside an active tenant transaction. Session-mode pooling is explicitly rejected — it permits `search_path` leakage across requests.
- **No SSO in Phase 1:** Admin-invitation-based Member onboarding only.
- **Single cloud region (A-8):** No data residency options in Phase 1.
- **Confidence Threshold calibration:** Admin-configurable from day one. Trust-weighted thumbs up/down is the calibration feedback loop. Threshold value logged at every routing decision for replay and diagnosis.
- **Migration strategy:** Schema-per-tenant requires migrations to enumerate and apply to all tenant schemas. Must be automated before second org is onboarded. Managed via Alembic/Flyway with a tenant migration runner.
- **Deployment topology:** Managed containers (Fly.io or Railway). Separate container definitions for: web API, ingestion worker, query worker. Independent scaling and deployment. No Kubernetes in Phase 1.

### Architectural Decisions Made

- **Storage:** Postgres + pgvector. Schema-per-tenant: each org gets its own Postgres schema named `org_{id}`. All schemas structurally identical. `KnowledgeStoreAdapter` sets `SET LOCAL search_path = org_{id}` at transaction start — the only code path through which any storage read or write can occur. HNSW index on embeddings column, mandatory from day one. Hard top-K retrieval limit. Pre-computed Knowledge Owner cache — never computed at query time. Tenant deletion = `DROP SCHEMA org_{id} CASCADE`.
- **Ingestion topology:** Slack webhook → `JobQueue` (Postgres-backed, pgmq or pg-boss for Phase 1) → monolithic worker with per-stage checkpointing (filter → embed → extract → write). Stage checkpoint table: `(ingestion_event_id, stage, status, attempt, error)`. On failure, re-enqueue at failed stage only, max 3 retries with exponential backoff, then dead-letter. Thin webhook receiver (ack-and-enqueue only). Reconciliation job for webhook gap recovery. At-least-once delivery.
- **Queue abstraction:** `JobQueue` interface with `enqueue`, `dequeue`, `ack`, `nack` methods. Postgres-backed implementation for Phase 1. Interface-first design makes the concrete implementation swappable without touching worker logic.
- **Tenant onboarding:** Automated. `Organization` record insert triggers `org_status = 'provisioning'`. Background job runs `CREATE SCHEMA org_{id}`, applies tenant migrations, flips to `org_status = 'active'`. Ingestion cannot begin until `org_status = 'active'` AND at least one non-Admin Member has activated.
- **Fallback routing:** Pure stateless function. Input: RAG pipeline result struct `{ confidence, retrieved_chunks, candidate_node_ids, graph_coverage_signal, matched_sensitivity_tiers }` + current threshold config. Output: enum `ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED`. Decision tree: (1) candidate_node_ids empty → NO_COVERAGE; (2) ownership map resolves → ROUTE_TO_OWNER; (3) sensitivity filter removed candidates → ACCESS_FILTERED; (4) chunks present but low coherence → REPHRASE; (5) no chunks despite nodes → NO_COVERAGE (sparse).
- **Key abstractions:** `KnowledgeStoreAdapter` (sole code path for all storage reads/writes; enforces `sensitivity_tier` at query layer before SQL executes; raises `SensitivityTierViolation` for unauthorized tier access), `IngestionFilter` (enforces DM exclusion and low-signal message pre-filtering before LLM extraction), `JobQueue` interface (queue-agnostic worker logic).

### Cross-Cutting Concerns Identified

- **Multi-tenancy isolation:** Structural via Postgres schema-per-tenant + PgBouncer transaction-mode + `SET LOCAL`. CI includes symmetric cross-tenant isolation tests: (a) write to Org A, verify unreachable from Org B via all retrieval paths; (b) ingestion worker processing Org B event cannot write into Org A's schema. Both tests run on every deploy.
- **LLM latency budget:** 10-second query SLA decomposed: embedding (~0.5s) + HNSW retrieval (~0.5–1s) + LLM generation streaming (~3–8s) + Source Attribution pre-generation (~0s, computed from retrieval) = budget met at p50. p95/p99 require hard timeout at 25s with source-node fallback. Query latency tracked as histogram (p50/p95/p99), not average.
- **Intelligence Tier data boundary:** `sensitivity_tier` tag on Knowledge Nodes used in Engagement Signal computation. `KnowledgeStoreAdapter.query()` enforces tier-based filtering before SQL executes. Base Tier users cannot reconstruct departure risk signals via crafted queries.
- **Privacy hard limits:** Three layers: Slack app scope restriction (manifest review gate in source control) + pipeline `IngestionFilter` (message type check before any processing) + runtime scope assertion (ingestion worker startup fails hard if DM-related scope present).
- **Audit/compliance logging:** Privacy config changes, Admin disclosure confirmation, ingestion start gate events, Member deactivation — durable, append-only. Disclosure log surfaced in Admin panel UI. Ingestion start date displayed to all Members on activation.
- **Background job / query isolation:** Separate container deployments. Engagement signal computation and Knowledge Ownership Map recomputation are background jobs with no shared compute with query workers.
- **Ingestion health visibility:** Ingestion lag exposed as internal API endpoint consumed by web UI. "Knowledge Graph current as of [timestamp]" always visible in query UI. Staleness banner when lag exceeds configurable threshold. Alert fires when any org's queue lag exceeds 30 minutes.
- **Signal cold start:** Dashboard shows "building baseline" state before sufficient data exists (minimum window TBD during beta). Admin can mark Members as on leave to exclude windows from signal computation. Signal cards show pattern detail and start date — not just severity.
- **Departure risk signal integrity:** `member_tenure_window` captured via `first_seen_at` on members table from day one. Handles member churn, mid-history joins, and offboarding. Cannot be reconstructed retroactively from Slack API.
- **Feedback loop integrity:** Trust-weighted feedback system. `cover_tracks` anomaly flag detects departure-risk + downvote-spike correlation without exposing either signal to unauthorized parties. Threshold recalibration pipeline rejects batches with >30% anomaly-flagged rows.
- **"I don't know" UX:** Branch-specific fallback — never stacked. `ROUTE_TO_OWNER` surfaces Knowledge Owner with deep-link to most relevant Slack thread and estimated response time. `NO_COVERAGE` surfaces "Flag for your team" action (turns failure into contribution). `REPHRASE` surfaces two LLM-generated query variants as one-click re-queries. Confidence threshold and internal scores are invisible to end users at all times.

---

## Starter Template Evaluation

### Primary Technology Domain

Multi-service full-stack application: Go infrastructure layer + Python AI layer + Next.js frontend. Async AI/ML backend pipeline with real-time streaming query interface.

### Service Architecture (Final)

| Service | Language | Responsibility |
|---|---|---|
| **`api`** | Go + Chi v5 | HTTP API, SSE streaming, Slack webhook receiver (<3s ack), River queue consumer, ingestion pipeline orchestration, KnowledgeStoreAdapter, auth, tenant management, admin panel |
| **`ai-worker`** | Python (FastAPI) | Embedding (OpenAI text-embedding-3-small) + LLM extraction (Anthropic) — called by `api` during ingestion as one batched call |
| **`rag`** | Python (FastAPI) | Full RAG pipeline: embed query → retrieve → generate → confidence score → fallback routing — one call from `api`, streamed response back |
| **`web`** | Next.js 16.2 | Query UI, admin panel, Intelligence Dashboard |
| **`signal-job`** | Python (cron) | Engagement velocity signal computation — scheduled job, no server, no HTTP port |

**Infrastructure:** Postgres 16 + pgvector 0.8.x on Hetzner Volume · River queue (in Postgres, no Redis) · Coolify v4 on CAX11 managing CAX31 · Traefik (managed by Coolify) · OpenAI embedding API

### Rationale for Split

Go owns the performance-critical infrastructure layer: Slack webhook ack has a hard <3s Slack requirement; Go goroutines handle concurrent queue processing with ~30MB memory footprint; River uses Postgres (already present) for durable queuing — no Redis dependency.

Python owns the AI layer: `text-embedding-3-small` and Anthropic SDK are Python-first; LangChain/LlamaIndex available where needed; async FastAPI with `uvicorn --workers N` (multi-process, bypasses GIL) for concurrent query handling.

Internal communication: one HTTP call from `api` to `ai-worker` per ingestion batch; one HTTP call from `api` to `rag` per user query. The Go API is a proxy for AI work, not an orchestrator — no multi-hop chains. All inter-service APIs defined as OpenAPI specs; Go clients generated via `oapi-codegen`; CI fails on schema drift.

### Deployment

**Hetzner CAX31** (8 vCPU ARM, 16GB RAM, 160GB NVMe, 20TB traffic, ~€13–15/month) running all application containers.
**Hetzner CAX11** (~€3.79/month) running Coolify v4 — isolated from application services so Coolify maintenance doesn't cause downtime.
**Postgres data on a Hetzner Volume** (network-attached, persists across server reinstalls — most important single resilience decision for Phase 1).
**Total Phase 1 infrastructure cost: ~€18–20/month.**

Coolify v4 manages: Git-push deploys, SSL via Traefik/Let's Encrypt, health checks, container restarts, `signal-job` as a cron job.

Pre-pilot load test gate: 10 concurrent queries, p95 < 8s. Must pass before pilot launch.

### Embedding and Retrieval

**Embedding model:** OpenAI `text-embedding-3-small` ($0.02/million tokens). Eliminates Ollama CPU contention on shared host. Negligible cost at Phase 1 volume (one pilot org).

**Retrieval:** pgvector HNSW only in Phase 1. Sufficient at Phase 1 scale (≤5M nodes per org); retrieval is not the latency bottleneck (LLM generation is). `RETRIEVAL_STRATEGY` env var in RAG service (`pgvector` | `typesense` | `hybrid`) — Phase 2 flip is a config change, not a code change.

**Phase 2 retrieval upgrade path:** Self-hosted Typesense (same Hetzner box, separate collection per tenant, fallback-to-pgvector if unavailable) when pilot data shows retrieval quality is the primary cause of below-threshold confidence scores, or when Knowledge Graph exceeds 5M nodes per org.

### Initialization Commands

```bash
# Frontend
npx create-next-app@latest web \
  --typescript --tailwind --eslint \
  --app --src-dir --import-alias "@/*"

# Go API
mkdir -p api/cmd/api api/internal/{handler,store,rag,auth}
cd api && go mod init github.com/yourorg/orgbrain/api
go get github.com/go-chi/chi/v5
go get github.com/jackc/pgx/v5
go get github.com/pgvector/pgvector-go/pgx
go get github.com/riverqueue/river
go get github.com/anthropics/anthropic-sdk-go

# Python AI services — each service uses app/ as its top-level package
# uvicorn entry point: uvicorn app.main:app
cd ai-worker && python -m venv .venv && mkdir -p app
pip install fastapi uvicorn anthropic openai pydantic structlog
cd ../rag && python -m venv .venv && mkdir -p app
pip install fastapi uvicorn anthropic openai pgvector asyncpg pydantic structlog
cd ../signal-job && python -m venv .venv && mkdir -p app
pip install psycopg2-binary pydantic structlog

# Migrations
go install github.com/pressly/goose/v3/cmd/goose@latest

# OpenAPI codegen
go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
```

### Monorepo Structure

```
orgbrain/
├── api/                        # Go: Chi v5
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── handler/            # route handlers
│   │   ├── store/              # KnowledgeStoreAdapter
│   │   ├── queue/              # River job definitions
│   │   └── auth/               # member auth, tier enforcement
│   ├── internal/client/        # oapi-codegen output (ai-worker + rag clients — do not edit manually)
│   └── Dockerfile
├── ai-worker/                  # Python: FastAPI
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py             # FastAPI app instance (entry: app.main:app)
│   │   ├── embed.py            # OpenAI embedding
│   │   └── extract.py          # Anthropic LLM extraction
│   ├── openapi.yaml            # contract consumed by Go codegen
│   ├── pyproject.toml
│   └── Dockerfile
├── rag/                        # Python: FastAPI
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py             # FastAPI app instance (entry: app.main:app)
│   │   ├── pipeline.py         # embed → retrieve → generate → score
│   │   └── fallback.py         # ROUTE_TO_OWNER|NO_COVERAGE|REPHRASE|ACCESS_FILTERED
│   ├── openapi.yaml            # contract consumed by Go codegen
│   ├── pyproject.toml
│   └── Dockerfile
├── signal-job/                 # Python: cron script
│   ├── app/
│   │   ├── __init__.py
│   │   └── main.py
│   ├── pyproject.toml
│   └── Dockerfile
├── web/                        # Next.js 16.2
│   └── Dockerfile
├── migrations/
│   ├── shared/                 # public schema: orgs, members, job_queue
│   └── tenant/                 # per-tenant schema: knowledge_nodes, embeddings
├── docker-compose.yml          # local dev
└── Makefile
```

### Architectural Decisions Provided by This Stack

| Area | Decision |
|---|---|
| API language | Go + Chi v5 |
| AI language | Python + FastAPI |
| Frontend | Next.js 16.2, App Router, TypeScript, Tailwind |
| Job queue | River (Postgres-backed, no Redis) |
| DB driver | pgx v5 + pgvector-go |
| Migrations | Goose (per-schema tenant runner) |
| Embeddings | OpenAI text-embedding-3-small |
| LLM | Anthropic SDK (Go + Python) |
| Retrieval | pgvector HNSW (Phase 1); Typesense hybrid (Phase 2) |
| Internal API contracts | OpenAPI specs + oapi-codegen (CI-enforced) |
| Deployment | Hetzner CAX31 + CAX11 + Coolify v4 |
| SSL/routing | Traefik (managed by Coolify) |

**Note:** Project initialization and monorepo scaffolding is the first implementation story.

---

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
- Member auth: server-side sessions in Postgres `sessions` table (shared schema)
- Query session storage: Postgres-backed `query_sessions` table (JSONB conversation history, 30-min idle expiry via background job)
- Secrets management: Coolify encrypted env vars (Phase 1); Infisical for Phase 2 if rotation/audit required
- API versioning: `/v1/` prefix on all public routes; `/webhooks/slack` unversioned; `/healthz` unversioned
- Streaming protocol: SSE (`text/event-stream`) — unidirectional, `net/http` native in Go, no WebSocket complexity
- Error format: RFC 7807 Problem Details (`application/problem+json`) across all Go API endpoints
- Rate limiting: per-Member sliding window in Postgres (30 queries/hour on query endpoint; 100 req/min on admin endpoints); no Redis

**Important Decisions (Shape Architecture):**
- Frontend state: TanStack Query v5 for server state; local `useState` for ephemeral UI state; no Zustand/Redux
- UI components: shadcn/ui with Radix primitives (copy-pasted, not a dependency); Tailwind-native
- Charts: Recharts for Intelligence Dashboard sparklines
- Structured logging: `slog` (Go stdlib) + `structlog` (Python), JSON output, stdout; Coolify log viewer for Phase 1; Grafana Loki for Phase 2
- Metrics: Prometheus endpoints on all services (`/metrics`); key metrics instrumented from day one; Grafana for Phase 2 dashboards
- CI/CD: GitHub Actions → GHCR → Coolify webhook redeploy

**Deferred Decisions (Post-MVP):**
- JWT for external API consumers (Phase 2, when mobile or third-party integrations are added)
- Redis (not needed; River + Postgres covers queue and rate limiting)
- Grafana Loki log aggregation (Phase 2)
- SSO/SCIM (V2 per PRD)
- Infisical secrets manager (Phase 2 if audit trail on secrets required)

### Authentication & Session Management

- **Member auth:** Session cookies backed by Postgres `sessions` table (shared schema). Session ID in httpOnly cookie. `DELETE FROM sessions WHERE member_id = $1` on deactivation — instant revocation.
- **Query sessions:** `query_sessions` table (tenant schema) — `(id, member_id, conversation_history JSONB, last_active_at TIMESTAMPTZ)`. Background River job expires rows where `last_active_at < NOW() - INTERVAL '30 minutes'`. Survives API restarts.
- **Slack OAuth tokens:** Stored in Postgres `integrations` table (tenant schema), encrypted at rest via Postgres `pgcrypto` with a key from Coolify env vars.
- **Secrets:** Coolify encrypted env var store. Never in source control. `.env.example` in repo documents required vars without values.

### API & Communication Patterns

- **Versioning:** All public routes at `/v1/`. Slack webhook at `/webhooks/slack`. Health at `/healthz`.
- **Streaming:** SSE (`Content-Type: text/event-stream`). Three event types: `data` (token), `meta` (Source Attribution + Confidence Score, sent before first token), `error` (timeout or fallback, with structured payload).
- **Error format:** RFC 7807. `type` is a URI slug (e.g., `https://orgbrain.io/errors/confidence-threshold`). `routing` field carries the fallback enum when applicable.
- **Rate limiting:** Sliding window counter in Postgres. Go middleware increments an `api_rate_limits` table row per member per window. `SKIP LOCKED` avoids contention. No Redis.
- **Internal service communication:** JSON over HTTP on Docker internal network. One call per operation (Go → `ai-worker` for ingestion batch; Go → `rag` for query). OpenAPI specs are the contract; `oapi-codegen` generates Go clients; CI fails on schema drift.

### Frontend Architecture

- **State:** TanStack Query v5 for all server state (queries, mutations, cache). `useState` for UI-only state (modal open, form inputs). No global state store.
- **Components:** shadcn/ui primitives (Button, Card, Table, Badge, Dialog, Sheet). Components copied into `web/src/components/ui/` — no runtime dependency.
- **Charts:** Recharts for 30-day sparklines on Intelligence Dashboard signal cards.
- **SSE consumption:** Custom `useQueryStream` hook wrapping the browser `EventSource` API. Handles `data`, `meta`, and `error` event types. Displays Source Attribution immediately on `meta` event, before LLM tokens arrive.

### Observability & CI/CD

- **Logging:** `slog` (Go, JSON, stdout). `structlog` (Python, JSON, stdout). Fields on every log line: `service`, `org_id` (where applicable), `trace_id`, `level`, `msg`. No `org_id` in logs for shared-schema operations.
- **Metrics (day-one instrumentation):** query latency histogram (p50/p95/p99), ingestion queue depth per org, ingestion lag per org, LLM call duration, embedding call duration, confidence score distribution, fallback routing type counts, SSE connection count.
- **CI pipeline (GitHub Actions):**
  1. `go test ./...` + `pytest`
  2. `oapi-codegen` regen + diff check
  3. Cross-tenant isolation tests (read + write directions)
  4. Pre-pilot load test gate (10 concurrent queries, p95 < 8s) — runs on `release/*` branches only
  5. Docker build + push to GHCR
  6. Coolify webhook trigger → rolling redeploy

---

## Implementation Patterns & Consistency Rules

### Naming Conventions

| Layer | Convention | Example |
|---|---|---|
| Database columns | `snake_case` | `created_at`, `sensitivity_tier`, `org_id` |
| Database tables | `snake_case` | `knowledge_nodes`, `query_sessions`, `user_trust_profile` |
| Go types | `PascalCase` | `KnowledgeStoreAdapter`, `IngestionFilter`, `FallbackRoute` |
| Go functions/methods | `PascalCase` (exported), `camelCase` (unexported) | `SetTenantContext()`, `renderError()` |
| Go packages | `lowercase` single word | `store`, `handler`, `queue`, `auth` |
| Python classes | `PascalCase` | `EmbedRequest`, `ExtractionResult`, `FallbackRouter` |
| Python functions | `snake_case` | `embed_batch()`, `compute_signal()`, `route_fallback()` |
| Python modules | `snake_case` | `embed.py`, `extract.py`, `fallback.py` |
| Next.js components | `PascalCase` | `SignalCard`, `QueryStream`, `ConfidenceScore` |
| Next.js hooks | `camelCase` prefixed `use` | `useQueryStream`, `useSignalCards`, `useOrgState` |
| Environment variables | `{SERVICE}_{RESOURCE}_{PROPERTY}` | `API_DB_URL`, `RAG_OPENAI_API_KEY`, `SIGNAL_JOB_ANTHROPIC_KEY` |

### API Response Formats

**Success (Go API):**
```json
{ "data": { ... } }
```

**Error (RFC 7807 Problem Details):**
```json
{
  "type": "https://orgbrain.io/errors/confidence-threshold",
  "title": "Query confidence below threshold",
  "status": 422,
  "detail": "No reliable answer found for this query.",
  "routing": "ROUTE_TO_OWNER",
  "knowledge_owner": { "member_id": "m_123", "slack_handle": "@alice" }
}
```

**SSE event types (query streaming):**
```
event: meta
data: {"confidence": 0.72, "sources": [...], "routing": null}

event: data
data: {"token": "The deployment process..."}

event: error
data: {"type": "timeout", "routing": "REPHRASE", "suggestions": [...]}
```

The `meta` event is always sent before the first `data` token. Source Attribution is visible to the user before the answer text starts rendering. The `error` event closes the stream; the client renders source nodes from the last received `meta` event with a "couldn't synthesize a full answer" message.

### Tenant Context Passing

**Go API — every handler that touches tenant data:**
```go
func (h *Handler) QueryKnowledge(w http.ResponseWriter, r *http.Request) {
    orgID := auth.OrgIDFromContext(r.Context())
    // KnowledgeStoreAdapter.WithTenant(orgID) sets SET LOCAL search_path internally
    results, err := h.store.WithTenant(orgID).Query(r.Context(), queryParams)
    ...
}
```

**Python services — tenant identity via header:**
```python
# X-Org-ID header injected by Go API on every internal call
org_id = request.headers.get("X-Org-ID")
# Python services are stateless re: tenant — they receive org_id per request
```

**SQL — no `org_id` in WHERE clauses for tenant data:**
```sql
-- CORRECT: search_path already set to org_{id}
SELECT * FROM knowledge_nodes WHERE sensitivity_tier <= $1

-- WRONG: never filter by org_id in tenant schema queries
SELECT * FROM knowledge_nodes WHERE org_id = $1
```

### Error Handling

**Go — explicit returns, no panic:**
```go
func (s *Store) Query(ctx context.Context, params QueryParams) ([]Node, error) {
    rows, err := s.db.Query(ctx, sql, params.Embedding, params.TopK)
    if err != nil {
        return nil, fmt.Errorf("store.Query: %w", err)
    }
    ...
}

// In handlers: renderError() wraps RFC 7807
func renderError(w http.ResponseWriter, status int, errType, detail string, extra map[string]any) {
    ...
}
```

**Python — typed exception hierarchy:**
```python
class OrgBrainError(Exception): pass
class SensitivityTierViolation(OrgBrainError): pass
class EmbeddingTimeout(OrgBrainError): pass
class LLMGenerationError(OrgBrainError): pass

# FastAPI exception handler maps to HTTP status codes
@app.exception_handler(SensitivityTierViolation)
async def sensitivity_handler(request, exc):
    return JSONResponse(status_code=403, content={"error": str(exc)})
```

**Frontend — React error boundaries per major section:**
```tsx
// QueryInterface, IntelligenceDashboard, AdminPanel each have their own ErrorBoundary
// SSE error events rendered inline (not as toasts) — user needs context to act
```

### Testing Patterns

**Go — co-located `_test.go`, real Postgres via testcontainers:**
```
api/internal/store/
  knowledge_store.go
  knowledge_store_test.go   ← tests the real adapter against real Postgres
api/internal/handler/
  query_handler.go
  query_handler_test.go     ← tests with httptest + stub store
```

**Python — `tests/` directory at service root, pytest:**
```
ai-worker/
  tests/
    test_embed.py           ← unit: mock OpenAI client
    test_extract.py         ← unit: mock Anthropic client
    test_integration.py     ← integration: real FastAPI test client
rag/
  tests/
    test_pipeline.py
    test_fallback.py
```

**CI-mandatory tests (all PRs):**
- Unit + integration tests (`go test ./...` + `pytest`)
- OpenAPI schema drift check (`oapi-codegen` regen + diff)
- Cross-tenant isolation: write to Org A → assert unreachable from Org B (both read and write directions)

**No E2E tests in Phase 1.** Pre-pilot load test (10 concurrent queries, p95 < 8s) runs on `release/*` branches only.

### Structured Log Fields

Every log line across all services must carry these fields:

| Field | Type | Notes |
|---|---|---|
| `service` | string | `api`, `ai-worker`, `rag`, `signal-job` |
| `level` | string | `debug`, `info`, `warn`, `error` |
| `msg` | string | Human-readable event description |
| `trace_id` | string | Propagated via `X-Trace-ID` header on internal calls |
| `org_id` | string | Present for all tenant-scoped operations; omit on shared-schema ops |
| `duration_ms` | int | For all external calls (DB, LLM, embedding API) |
| `error` | string | Structured error string when `level=error` |

### AI Agent Enforcement Rules

These rules are mandatory for any AI agent implementing OrgBrain. They are non-negotiable and must not be relaxed without explicit architectural decision:

1. **`KnowledgeStoreAdapter` is the sole storage path.** No raw `db.Query()` calls in handlers or workers. Every storage read/write goes through `KnowledgeStoreAdapter`. Violations are a blocking PR failure.

2. **`SET LOCAL search_path` is set inside every transaction, never outside.** Session-mode `SET search_path` is forbidden — it persists across PgBouncer transaction boundaries and causes tenant data leakage.

3. **`IngestionFilter` runs before any LLM call.** No message reaches the embedding or extraction pipeline without passing through `IngestionFilter`. DM-type messages must be rejected at filter time with a logged audit event.

4. **`sensitivity_tier` enforcement happens at `KnowledgeStoreAdapter.query()`, not in route handlers.** Route handlers must not conditionally filter sensitivity tiers — the adapter enforces it unconditionally before SQL executes.

5. **`ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED` is a pure function.** The fallback router takes a result struct and threshold config; it has no side effects, no DB calls, no logging. Tests must cover all four enum values.

6. **Confidence Score and fallback routing type are always logged at decision time.** The threshold value in effect at the time of the decision must also be logged — for replay and calibration audit.

7. **No `org_id` in SQL `WHERE` clauses within tenant schemas.** Tenant isolation is via `search_path`, not column filtering. A CI lint rule must flag `WHERE org_id =` patterns in files under `internal/store/`.

8. **Ingestion webhook receiver does one thing: ack and enqueue.** No processing, no DB reads, no LLM calls. Response must be sent within 3 seconds of request receipt. Any additional logic is a River job.

9. **OpenAPI specs are the Go↔Python contract boundary.** Go clients are generated via `oapi-codegen` from the Python services' `openapi.yaml`. Manual HTTP client code targeting internal services is forbidden. CI fails on any schema drift.

---

## Project Structure & Boundaries

### Requirements → Component Mapping

| FR | Functional Area | Service(s) | Key Files |
|---|---|---|---|
| FR-1p | Slack ingestion + webhook | `api` | `slack/webhook_handler.go`, `slack/events.go`, `queue/ingestion_job.go`, `slack/reconcile.go` |
| FR-5 | Knowledge Node extraction | `ai-worker`, `api` | `app/extract.py`, `store/knowledge_store.go` |
| FR-6 | Decision versioning (append-only) | `api` | `store/knowledge_store.go`, `migrations/tenant/001_knowledge_nodes.sql` |
| FR-7 | Knowledge Ownership Map | `api` | `ownership/ownership.go`, `queue/ownership_job.go`, `store/ownership_store.go` |
| FR-8p | Natural language query (SSE) | `api`, `rag` | `handler/query_handler.go`, `rag/app/pipeline.py`, `web/query/` |
| FR-9 | Confidence threshold + fallback routing | `rag`, `api` | `rag/app/fallback.py`, `rag/app/confidence.py`, `handler/query_handler.go` |
| FR-10 | Staleness indication | `rag`, `api` | `rag/app/staleness.py`, `handler/health_handler.go` |
| FR-13p | Engagement velocity signal | `signal-job` | `signal-job/app/compute.py`, `signal-job/app/main.py` |
| FR-14/15 | Intelligence Dashboard | `web`, `api` | `web/dashboard/`, `handler/signal_handler.go`, `store/signal_store.go` |
| FR-16p | Slack OAuth integration | `api` | `slack/oauth.go`, `handler/admin_handler.go` |
| FR-17 | Member access management | `api` | `auth/session.go`, `handler/member_handler.go`, `migrations/shared/002_members.sql` |
| FR-18 | Privacy configuration | `api` | `handler/admin_handler.go`, `store/audit_store.go` |
| — | Trust-weighted feedback | `api` | `handler/feedback_handler.go`, `store/feedback_store.go` |
| — | Ingestion health visibility | `api`, `web` | `handler/health_handler.go`, `web/hooks/useIngestionStatus.ts` |

### Complete Project Directory Structure

```
orgbrain/
├── .github/
│   └── workflows/
│       ├── ci.yml                      # go test + pytest + schema drift + isolation tests
│       └── release.yml                 # load test gate + GHCR build + Coolify webhook
│
├── api/                                # Go + Chi v5
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                 # server boot, middleware chain, graceful shutdown
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── middleware.go           # session cookie validation, OrgID injection into context
│   │   │   ├── session.go              # Postgres-backed sessions (shared schema)
│   │   │   ├── session_test.go
│   │   │   └── tier.go                 # Intelligence Tier check per member
│   │   ├── handler/
│   │   │   ├── query_handler.go        # FR-8p: SSE streaming query endpoint
│   │   │   ├── query_handler_test.go
│   │   │   ├── signal_handler.go       # FR-14/15: Intelligence Dashboard read endpoints
│   │   │   ├── signal_handler_test.go
│   │   │   ├── admin_handler.go        # FR-16p/17/18: Slack OAuth, member mgmt, privacy config
│   │   │   ├── admin_handler_test.go
│   │   │   ├── webhook_handler.go      # FR-1p: Slack Events API — ack + enqueue only (<3s)
│   │   │   ├── webhook_handler_test.go
│   │   │   ├── feedback_handler.go     # thumbs up/down, rate limiting enforcement
│   │   │   ├── feedback_handler_test.go
│   │   │   └── health_handler.go       # /healthz + /v1/ingestion-lag
│   │   ├── slack/
│   │   │   ├── oauth.go                # FR-16p: Slack OAuth flow, token storage
│   │   │   ├── events.go               # Slack event type parsing, DM type rejection
│   │   │   ├── reconcile.go            # FR-1p gap recovery: conversations.history poll
│   │   │   └── scope_assert.go         # startup assertion: fails hard if DM scope present
│   │   ├── queue/
│   │   │   ├── ingestion_job.go        # River job: per-stage checkpoint (filter→embed→extract→write)
│   │   │   ├── ingestion_job_test.go
│   │   │   ├── ownership_job.go        # River job: Knowledge Ownership Map recompute
│   │   │   ├── expiry_job.go           # River job: query_sessions expiry (30-min idle)
│   │   │   └── worker.go               # River worker setup, concurrency, dead-letter config
│   │   ├── store/
│   │   │   ├── adapter.go              # KnowledgeStoreAdapter: SetTenantContext, SET LOCAL, sensitivity_tier gate
│   │   │   ├── adapter_test.go         # cross-tenant isolation tests (read + write directions)
│   │   │   ├── knowledge_store.go      # knowledge_nodes: append-only write, vector search
│   │   │   ├── knowledge_store_test.go
│   │   │   ├── session_store.go        # query_sessions CRUD
│   │   │   ├── ownership_store.go      # ownership_map read/write + 24h cache
│   │   │   ├── signal_store.go         # departure_risk_signals read
│   │   │   ├── feedback_store.go       # query_feedback + user_trust_profile + anomaly flags
│   │   │   ├── audit_store.go          # FR-18: append-only audit_log writes
│   │   │   ├── rate_limit_store.go     # sliding window rate limit (SKIP LOCKED)
│   │   │   └── tenant.go               # tenant provisioning, schema creation, migration runner
│   │   ├── filter/
│   │   │   ├── ingestion_filter.go     # IngestionFilter: DM exclusion + low-signal pre-filter
│   │   │   └── ingestion_filter_test.go
│   │   ├── ownership/
│   │   │   ├── ownership.go            # Knowledge Ownership Map computation
│   │   │   └── ownership_test.go
│   │   └── client/                     # oapi-codegen generated — do not edit manually
│   │       ├── ai_worker_client.go     # generated Go client for ai-worker
│   │       └── rag_client.go           # generated Go client for rag
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
│
├── ai-worker/                          # Python + FastAPI
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py                     # FastAPI app instance (entry: app.main:app): POST /embed, POST /extract
│   │   ├── embed.py                    # OpenAI text-embedding-3-small batching
│   │   ├── extract.py                  # Anthropic: Knowledge Node extraction from Slack message
│   │   └── models.py                   # Pydantic request/response models
│   ├── openapi.yaml                    # Contract — source of truth for Go oapi-codegen
│   ├── tests/
│   │   ├── test_embed.py               # unit: mock OpenAI client
│   │   ├── test_extract.py             # unit: mock Anthropic client
│   │   └── test_integration.py         # FastAPI test client end-to-end
│   ├── requirements.txt
│   └── Dockerfile
│
├── rag/                                # Python + FastAPI
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py                     # FastAPI app instance (entry: app.main:app): POST /query (SSE streaming response)
│   │   ├── pipeline.py                 # embed → retrieve → generate → confidence score
│   │   ├── fallback.py                 # Pure fn: FallbackRouter → ROUTE_TO_OWNER|NO_COVERAGE|REPHRASE|ACCESS_FILTERED
│   │   ├── retrieval.py                # pgvector HNSW top-K; RETRIEVAL_STRATEGY env var abstraction
│   │   ├── confidence.py               # Confidence Score computed from retrieval (pre-generation)
│   │   ├── staleness.py                # FR-10: node age check, staleness threshold logic
│   │   ├── models.py                   # Pydantic models
│   │   └── db.py                       # asyncpg pool; SET LOCAL search_path per call
│   ├── openapi.yaml                    # Contract — source of truth for Go oapi-codegen
│   ├── tests/
│   │   ├── __init__.py
│   │   ├── test_pipeline.py
│   │   ├── test_fallback.py            # covers all four enum values
│   │   ├── test_retrieval.py
│   │   └── test_confidence.py
│   ├── requirements.txt
│   └── Dockerfile
│
├── signal-job/                         # Python cron — no HTTP server, no open port
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py                     # entry: load active orgs → compute per-member → write signals
│   │   ├── compute.py                  # 30-day engagement velocity: message freq + response time trend
│   │   ├── db.py                       # psycopg2 pool; SET LOCAL search_path per org
│   │   └── models.py                   # signal result dataclasses
│   ├── tests/
│   │   ├── __init__.py
│   │   ├── test_compute.py             # unit: signal math with synthetic time-series data
│   │   └── test_cold_start.py          # edge: "building baseline" state when data < minimum window
│   ├── requirements.txt
│   └── Dockerfile
│
├── web/                                # Next.js 16.2, App Router, TypeScript, Tailwind
│   ├── src/
│   │   ├── app/
│   │   │   ├── layout.tsx              # root layout, global CSS, providers
│   │   │   ├── globals.css
│   │   │   ├── (auth)/
│   │   │   │   ├── login/page.tsx      # Member login form
│   │   │   │   └── invite/[token]/page.tsx  # Admin invitation acceptance
│   │   │   └── (app)/                  # Authenticated route group
│   │   │       ├── layout.tsx          # Auth guard, org context provider
│   │   │       ├── query/
│   │   │       │   └── page.tsx        # FR-8p: Query Interface
│   │   │       ├── dashboard/
│   │   │       │   └── page.tsx        # FR-14/15: Intelligence Dashboard (Intelligence Tier only)
│   │   │       └── admin/
│   │   │           ├── page.tsx        # FR-17/18: Member list + privacy config
│   │   │           ├── integrations/page.tsx   # FR-16p: Slack OAuth + channel exclusions
│   │   │           └── audit/page.tsx          # Disclosure log viewer
│   │   ├── components/
│   │   │   ├── ui/                     # shadcn/ui copies (Button, Card, Table, Badge, Dialog, Sheet)
│   │   │   ├── query/
│   │   │   │   ├── QueryInput.tsx      # Natural language input + submit button
│   │   │   │   ├── QueryStream.tsx     # SSE consumer: renders meta/data/error events
│   │   │   │   ├── SourceAttribution.tsx   # Source cards shown immediately on meta event
│   │   │   │   ├── ConfidenceBadge.tsx
│   │   │   │   ├── FallbackCard.tsx    # Branch-specific: ROUTE_TO_OWNER / NO_COVERAGE / REPHRASE
│   │   │   │   ├── FeedbackButtons.tsx # Thumbs up/down + 2s cooldown enforcement
│   │   │   │   └── StalenessWarning.tsx  # FR-10: staleness flag inline with answer
│   │   │   ├── dashboard/
│   │   │   │   ├── SignalCard.tsx      # Severity badge + decline start date + pattern detail
│   │   │   │   ├── SignalSparkline.tsx # Recharts 30-day engagement sparkline
│   │   │   │   ├── SeverityBadge.tsx  # Watch / Concerning / Urgent
│   │   │   │   └── BaselineState.tsx  # "Building baseline" empty state (cold start)
│   │   │   └── admin/
│   │   │       ├── MemberTable.tsx    # Member list + tier badge + deactivate action
│   │   │       ├── SlackConnectButton.tsx
│   │   │       ├── ChannelExclusions.tsx
│   │   │       └── PrivacyConfigPanel.tsx  # Signal cluster toggles + disclosure log
│   │   ├── hooks/
│   │   │   ├── useQueryStream.ts      # EventSource wrapper: meta/data/error, reconnect logic
│   │   │   ├── useSignalCards.ts      # TanStack Query: fetch + 60s poll for signal cards
│   │   │   └── useIngestionStatus.ts  # Ingestion lag polling → staleness banner trigger
│   │   ├── lib/
│   │   │   ├── api.ts                 # Typed fetch wrappers for all /v1/ endpoints
│   │   │   └── utils.ts               # cn(), date formatting, confidence display helpers
│   │   └── types/
│   │       └── api.ts                 # API response types
│   ├── next.config.js
│   ├── tailwind.config.js
│   ├── tsconfig.json
│   ├── package.json
│   ├── .env.local
│   ├── .env.example
│   └── Dockerfile
│
├── migrations/
│   ├── shared/                        # public schema: applied once on first deploy
│   │   ├── 001_organizations.sql      # organizations table, org_status enum, integrations
│   │   ├── 002_members.sql            # members, sessions, invitations
│   │   ├── 003_job_queue.sql          # River queue tables
│   │   ├── 004_rate_limits.sql        # api_rate_limits (sliding window)
│   │   └── 005_audit_log.sql          # shared audit_log (privacy config events)
│   └── tenant/                        # per-org schema: applied to every org_{id} by tenant runner
│       ├── 001_knowledge_nodes.sql    # knowledge_nodes: append-only, sensitivity_tier, versioning
│       ├── 002_embeddings.sql         # embeddings column + HNSW index (m=16 ef=64)
│       ├── 003_ownership_map.sql      # ownership_map cache table
│       ├── 004_query_sessions.sql     # query_sessions (JSONB conversation_history, last_active_at)
│       ├── 005_signals.sql            # departure_risk_signals (severity, pattern_detail, resolved_at)
│       ├── 006_feedback.sql           # query_feedback + user_trust_profile (trust_score [0.05, 1.0])
│       └── 007_ingestion_events.sql   # ingestion_events + stage_checkpoints (idempotency key)
│
├── openapi/                           # Source of truth for Go↔Python contracts
│   ├── ai-worker.yaml
│   └── rag.yaml
│
├── scripts/
│   ├── migrate-all-tenants.sh         # applies tenant migrations to all active org_{id} schemas
│   ├── load-test.sh                   # k6: 10 concurrent queries, assert p95 < 8s
│   └── isolation-test.sh              # CI: Org A write → assert Org B cannot read (both directions)
│
├── docker-compose.yml                 # local dev: all services + Postgres + PgBouncer
├── Makefile                           # make dev | test | migrate | codegen | load-test
├── .env.example                       # all env vars across all services (no values)
└── .gitignore
```

### Architectural Boundaries

**External API Boundaries:**

| Boundary | Path | Auth | Notes |
|---|---|---|---|
| Query endpoint | `POST /v1/query` (SSE) | Session cookie | Member-tier; streams `meta` → `data` → `error` |
| Intelligence Dashboard | `GET /v1/signals` | Session cookie | Intelligence Tier only |
| Admin endpoints | `/v1/admin/*` | Session cookie | Admin role only |
| Member endpoints | `/v1/members/*` | Session cookie | Self-service + Admin |
| Slack webhook | `POST /webhooks/slack` | Slack signature verification | Unversioned; ack <3s |
| Health + lag | `GET /healthz`, `GET /v1/ingestion-lag` | None | Consumed by web UI + Coolify |

**Internal Service Communication:**

```
Slack Events API
       │ POST /webhooks/slack
       ▼
   api (Go) ──── River enqueue ──── Postgres (job queue)
       │                                   │
       │                         River worker (in api process)
       │                                   │
       │              ┌────────────────────┴────────────────────┐
       │       POST /embed                             POST /extract
       │   ai-worker (Python)                      ai-worker (Python)
       │              └────────────────────┬────────────────────┘
       │                                   │ writes via KnowledgeStoreAdapter
       │                          Postgres (org_{id} schema)
       │
       │ POST /query (SSE)
       ▼
   rag (Python) ──── asyncpg ──── Postgres (org_{id} schema + pgvector)
       │ SSE stream back to api
       ▼
   api (Go) ──── SSE to browser

signal-job (Python cron, no HTTP)
       └──── psycopg2 ──── Postgres (all active org_{id} schemas)
                                   └── writes departure_risk_signals
```

**Data Boundaries:**

| Scope | Schema | Tables | Access Path |
|---|---|---|---|
| Platform-wide | `public` | `organizations`, `members`, `sessions`, `invitations`, `river_*`, `api_rate_limits`, `audit_log` | Direct pgx/psycopg2 |
| Per-tenant | `org_{id}` | `knowledge_nodes`, `embeddings`, `ownership_map`, `query_sessions`, `departure_risk_signals`, `query_feedback`, `user_trust_profile`, `ingestion_events`, `stage_checkpoints` | Only via `KnowledgeStoreAdapter.WithTenant(orgID)` |

**External Integrations:**

| Integration | Direction | Location | Notes |
|---|---|---|---|
| Slack Events API | Inbound | `api/internal/slack/` | Webhook + OAuth + reconciliation |
| OpenAI Embeddings | Outbound | `ai-worker/embed.py`, `rag/pipeline.py` | `text-embedding-3-small`; batched |
| Anthropic LLM | Outbound | `ai-worker/extract.py`, `rag/pipeline.py` | Streaming for query; batch for extraction |
| Coolify v4 | Deployment | `.github/workflows/release.yml` | Webhook redeploy trigger |
| GHCR | Image registry | `.github/workflows/release.yml` | Docker image push per service |

---

## Architecture Validation Results

### Coherence Validation

**Decision Compatibility:** All technology choices are version-compatible and conflict-free. River is Postgres-native — no Redis introduced anywhere (rate limiting, sessions, queue all Postgres-backed, consistent). PgBouncer transaction-mode paired correctly with `SET LOCAL` — session-mode explicitly rejected and documented as the failure mode. FastAPI auto-generates OpenAPI 3.x specs consumed directly by `oapi-codegen` — no manual schema maintenance. SSE is native to Go's `net/http` — no WebSocket library dependency. `asyncpg` (async) in `rag/app/db.py` and `psycopg2` (sync) in `signal-job/app/db.py` is intentional — `signal-job` is a blocking cron script, not an async server.

**Pattern Consistency:** Naming conventions cover all layers comprehensively. `SET LOCAL search_path` appears in all three code paths that touch tenant schemas: `adapter.go` (Go), `rag/app/db.py` (Python), `signal-job/app/db.py` (Python). RFC 7807 error format applied only at the Go API boundary — internal Python services return plain JSON (correct: they are not client-facing). SSE event types (`meta`/`data`/`error`) defined in patterns and consumed in `useQueryStream.ts` — consistent end-to-end.

**Structure Alignment:** Every architectural abstraction has a named home. Shared and tenant migrations are structurally separated. `openapi/` at monorepo root is the single source of truth, consumed by generated code in `api/internal/client/`. `scripts/isolation-test.sh` enforces the cross-tenant CI requirement named in patterns — structure matches rules.

### Requirements Coverage Validation

All MVP FRs covered. All NFRs architecturally addressed.

| FR | Coverage |
|---|---|
| FR-1p Slack ingestion | `api/internal/slack/` + `queue/ingestion_job.go`; reconciliation for gap recovery; idempotency key prevents duplicates |
| FR-5 Knowledge Node extraction | `ai-worker/extract.py`; `IngestionFilter` gates before any LLM call; `sensitivity_tier` tagged at write time |
| FR-6 Decision versioning | Append-only `knowledge_store.go`; previous versions retained with full provenance |
| FR-7 Knowledge Ownership Map | Pre-computed in `ownership/ownership.go`; `first_seen_at` captured from day one for `member_tenure_window` |
| FR-8p NL query (SSE) | `meta` event pre-generation; 25s hard timeout; SSE heartbeat at 20s; p50/p95/p99 histogram |
| FR-9 Confidence + fallback | Stateless `FallbackRouter`; all four enum values required in tests; threshold logged at every decision |
| FR-10 Staleness indication | `staleness.py` + `StalenessWarning.tsx`; Admin-configurable threshold |
| FR-13p Engagement velocity | `compute.py`; cold start detection in `test_cold_start.py`; leave period Admin-markable |
| FR-14/15 Intelligence Dashboard | `SignalCard.tsx` shows pattern detail + start date; `BaselineState.tsx` for cold start |
| FR-16p Slack OAuth | `slack/oauth.go`; tokens encrypted via pgcrypto |
| FR-17 Member access | Invitation-based; instant revocation via `DELETE FROM sessions WHERE member_id = $1` |
| FR-18 Privacy config | Disclosure confirmation gate; append-only `audit_store.go`; log surfaced in Admin UI |

### Implementation Readiness Validation

All critical decisions carry specific versions. All files are named — no placeholders. The 9 mandatory enforcement rules are specific, testable, and blockable at CI/PR review. `adapter_test.go` (cross-tenant isolation) and `test_cold_start.py` (signal cold start) are explicitly named files.

### Gap Analysis

**Critical gaps:** none.

**Important gaps:** none.

**Nice-to-have (post-pilot):**

1. `openapi-typescript` codegen for `web/types/api.ts` — currently noted as manually synced; safe to defer until API surface stabilizes post-pilot.
2. `deploy/pgbouncer.ini.example` — PgBouncer configuration is a deployment artifact not shown in the project tree; worth adding for reproducibility.
3. SQL lint step in `ci.yml` — Rule 7 specifies a CI check for `WHERE org_id =` patterns in `internal/store/`; low risk for Phase 1 (solo engineer), worth adding before second engineer joins.

### Architecture Completeness Checklist

**Requirements Analysis**
- [x] Project context thoroughly analyzed
- [x] Scale and complexity assessed
- [x] Technical constraints identified
- [x] Cross-cutting concerns mapped

**Architectural Decisions**
- [x] Critical decisions documented with versions
- [x] Technology stack fully specified
- [x] Integration patterns defined
- [x] Performance considerations addressed

**Implementation Patterns**
- [x] Naming conventions established
- [x] Structure patterns defined
- [x] Communication patterns specified
- [x] Process patterns documented

**Project Structure**
- [x] Complete directory structure defined
- [x] Component boundaries established
- [x] Integration points mapped
- [x] Requirements to structure mapping complete

### Architecture Readiness Assessment

**Overall Status: READY FOR IMPLEMENTATION**

**Confidence Level:** High

**Key Strengths:**
- Three-layer DM exclusion with startup hard-fail — structurally impossible to accidentally ingest DMs
- `KnowledgeStoreAdapter` as sole storage path — enforcement auditable at PR review, not dependent on discipline
- River on Postgres eliminates Redis — one fewer infra dependency at zero additional cost
- Stateless `FallbackRouter` with four required test paths — cannot be partially implemented
- Pre-generation `meta` SSE event — Source Attribution visible before LLM output starts
- `cover_tracks` anomaly flag — departure risk signal integrity under adversarial conditions
- `RETRIEVAL_STRATEGY` env var abstraction — Phase 2 Typesense upgrade is a config change, not a refactor
- €18–20/month Phase 1 infra — sustainable through pilot without fundraising

**Areas for Future Enhancement:**
- `openapi-typescript` for TypeScript type generation (vs. manual sync)
- Grafana Loki for log aggregation (Phase 2)
- Typesense hybrid retrieval when pilot data shows retrieval quality as primary confidence bottleneck (Phase 2)
- JWT + external API consumer support (Phase 2)
- Infisical for secrets rotation audit trail (Phase 2)
- SSO/SCIM (V2 per PRD)

### Implementation Handoff

**AI Agent Guidelines:**
- Follow all architectural decisions exactly as documented
- Use implementation patterns consistently across all components
- Respect the 9 mandatory enforcement rules — they are blocking, not advisory
- `KnowledgeStoreAdapter`, `IngestionFilter`, and `FallbackRouter` are the three critical abstractions — implement them before any feature work
- Refer to this document for all architectural questions; do not make technology substitutions without an ADR

**First Implementation Priority:**

Story 1 — Monorepo scaffolding and local dev environment. Create the directory structure from the Project Structure section. Get `docker-compose.yml` running with all five services (api, ai-worker, rag, signal-job, web) + Postgres + PgBouncer. Verify `make test` passes on an empty skeleton before any feature work begins.

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Status

OrgBrain is currently in the **planning phase**. No application code exists yet. The repository contains planning artifacts (PRD, architecture, epics/stories) and the BMad workflow tooling used to produce them. Implementation begins with Story 1.1 (monorepo scaffolding).

## Planning Artifacts

All planning documents live in `_bmad-output/planning-artifacts/`:

| Document | Path | Purpose |
|---|---|---|
| Architecture | `architecture.md` | Tech stack, service topology, naming conventions, 9 mandatory enforcement rules — read this before writing any code |
| MVP PRD | `prds/prd-OrgBrain-MVP-2026-05-26/prd-mvp.md` | Phase 1 scope authority |
| Full PRD | `prds/prd-OrgBrain-2026-05-25/prd.md` | Personas, NFRs, privacy architecture (inherited by MVP PRD) |
| Epics & Stories | `epics.md` | All implementation stories with acceptance criteria |

**Start here for any implementation work:** read `architecture.md` fully. It contains the directory structure, service boundaries, naming conventions, and 9 non-negotiable enforcement rules that all code must satisfy.

## Planned Architecture

Five-service monorepo (none scaffolded yet):

```
api/          Go + Chi v5       — HTTP API, SSE, Slack webhook, River queue worker
ai-worker/    Python + FastAPI  — Embedding (OpenAI) + extraction (Anthropic)
rag/          Python + FastAPI  — RAG pipeline: embed → retrieve → generate → score
signal-job/   Python cron       — Engagement velocity signal computation (no HTTP port)
web/          Next.js 16.2      — Query UI, Intelligence Dashboard, Admin panel
migrations/   SQL               — shared/ (once) + tenant/ (per org_{id} schema)
```

Infrastructure: Postgres 16 + pgvector · River queue (no Redis) · PgBouncer transaction-mode · Hetzner + Coolify v4

## Commands (post-scaffolding)

These will exist after Story 1.1 is implemented:

```bash
make dev          # start all services + Postgres + PgBouncer via docker-compose
make test         # go test ./... + pytest across all services
make migrate      # apply shared + tenant migrations via Goose
make codegen      # regenerate Go clients from ai-worker/openapi.yaml and rag/openapi.yaml
make load-test    # k6: 10 concurrent queries, assert p95 < 8s (release/* only)
```

## Critical Architectural Rules

These are non-negotiable (from `architecture.md` §"AI Agent Enforcement Rules"):

1. **`KnowledgeStoreAdapter` is the sole storage path.** No raw `db.Query()` calls in handlers or workers. CI lint must flag violations.
2. **`SET LOCAL search_path = org_{id}` inside every transaction, never outside.** Session-mode `search_path` causes tenant data leakage through PgBouncer.
3. **`IngestionFilter` runs before any LLM call.** DM-type messages (`channel_type = im|mpim`) must be rejected with a logged audit event.
4. **`sensitivity_tier` enforcement at `KnowledgeStoreAdapter.query()`, not in route handlers.**
5. **`FallbackRouter` is a pure function.** No side effects, no DB calls, no logging. All four outcomes (`ROUTE_TO_OWNER | NO_COVERAGE | REPHRASE | ACCESS_FILTERED`) must have unit tests.
6. **Confidence Score and fallback routing type always logged at decision time**, including the threshold value in effect.
7. **No `org_id` in SQL `WHERE` clauses within tenant schemas.** Isolation is via `search_path`, not column filtering.
8. **Slack webhook receiver does one thing: ack and enqueue.** Must respond within 3 seconds. No processing logic.
9. **OpenAPI specs are the Go↔Python contract boundary.** Go clients generated via `oapi-codegen`. Manual HTTP client code targeting internal services is forbidden. CI fails on schema drift.

## Multi-Tenancy Pattern

Every tenant gets its own Postgres schema `org_{id}`. The pattern throughout the codebase:

```go
// Go: KnowledgeStoreAdapter sets search_path per transaction
results, err := h.store.WithTenant(orgID).Query(ctx, params)
```

```python
# Python services: receive org_id via X-Org-ID header, set search_path per call
org_id = request.headers.get("X-Org-ID")
```

```sql
-- SQL: no org_id in WHERE clauses inside tenant schemas
SELECT * FROM knowledge_nodes WHERE sensitivity_tier <= $1  -- CORRECT
SELECT * FROM knowledge_nodes WHERE org_id = $1            -- WRONG
```

## SSE Streaming Protocol

The query endpoint streams three event types in order:

```
event: meta   → {"confidence": float, "sources": [...], "routing": null}  ← always first
event: data   → {"token": "..."}                                           ← LLM tokens
event: error  → {"type": "timeout", "routing": "REPHRASE", ...}           ← closes stream
```

Source Attribution must be visible to the user before any answer text renders (`meta` arrives before first `data`). Hard LLM timeout at 25s; SSE heartbeat at 20s.

## Project Rules

See [`docs/rules.md`](docs/rules.md) for project-specific rules. Key items:

- **Never add `Co-Authored-By` trailers to commits.**

## BMad Workflow

Planning artifacts were produced using the BMad Method CLI installed in `.claude/skills/`. To continue planning work (e.g., finish epics/stories, run sprint planning):

```
/bmad-help          — shows current workflow state and next recommended step
/bmad-create-story  — prepare next story for implementation
/bmad-sprint-planning — produce sprint plan from epics
```

BMad config: `_bmad/bmm/config.yaml` (paths) · `_bmad/core/config.yaml` (project metadata)

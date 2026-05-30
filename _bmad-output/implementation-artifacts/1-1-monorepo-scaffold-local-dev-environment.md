# Story 1.1: Monorepo Scaffold & Local Dev Environment

Status: ready-for-dev

## Story

As a **developer**,
I want all five services running locally via a single command,
so that I can begin feature work immediately without environment setup friction.

## Acceptance Criteria

1. **Given** the repository is cloned and Docker is running  
   **When** the developer runs `make dev`  
   **Then** all five services start (api, ai-worker, rag, signal-job, web) alongside Postgres 16 and PgBouncer  
   **And** each service responds to its health check (`GET /healthz` → 200 for api, ai-worker, rag; startup log for signal-job; Next.js dev server port 3000 for web)  

2. **Given** all services are running  
   **When** the developer runs `make test`  
   **Then** `go test ./...` passes with empty/stub test suites in `api/`  
   **And** `pytest` passes with empty/stub test suites in `ai-worker/`, `rag/`, and `signal-job/`  

3. **Given** the repository is scaffolded  
   **When** the directory structure is inspected  
   **Then** every file and folder listed in the architecture spec (§"Complete Project Directory Structure") exists as a stub — no missing paths  

4. **Given** `make codegen` runs  
   **When** `oapi-codegen` is invoked against `openapi/ai-worker.yaml` and `openapi/rag.yaml`  
   **Then** it exits 0 (even against minimal stub specs)  

5. **Given** `make migrate` runs  
   **When** Goose applies shared migrations  
   **Then** it exits 0 against the Postgres instance (even if migration files are empty stubs)

## Tasks / Subtasks

- [ ] Task 1: Initialize root structure (AC: 3)
  - [ ] Create top-level files: `docker-compose.yml`, `Makefile`, `.env.example`, `.gitignore`
  - [ ] Create `.github/workflows/ci.yml` and `release.yml` (stubbed — jobs echo + exit 0)
  - [ ] Create `scripts/migrate-all-tenants.sh`, `load-test.sh`, `isolation-test.sh` (stubbed with comments)
  - [ ] Create `openapi/ai-worker.yaml` and `openapi/rag.yaml` (minimal valid OpenAPI 3.0 stubs with `/healthz` path)
  - [ ] Create `migrations/shared/` with 5 stub `.sql` files (named per spec, empty `-- placeholder` body)
  - [ ] Create `migrations/tenant/` with 7 stub `.sql` files (named per spec, empty `-- placeholder` body)

- [ ] Task 2: Scaffold Go `api` service (AC: 1, 2, 3, 4)
  - [ ] Run `go mod init github.com/orgbrain/orgbrain/api` in `api/`
  - [ ] Install deps: `go get github.com/go-chi/chi/v5 github.com/jackc/pgx/v5 github.com/pgvector/pgvector-go/pgx github.com/riverqueue/river github.com/anthropics/anthropic-sdk-go`
  - [ ] Create `api/cmd/api/main.go` — Chi router, `/healthz` → 200 `{"status":"ok"}`, graceful shutdown stub
  - [ ] Create all `internal/` package directories with stub `.go` files: `auth/`, `handler/`, `slack/`, `queue/`, `store/`, `filter/`, `ownership/`, `client/`
  - [ ] Create named stub files per spec (e.g. `handler/query_handler.go`, `store/adapter.go`) — each file has package declaration + one TODO comment
  - [ ] Create `api/Dockerfile` (multi-stage: `golang:1.23-alpine` builder + `alpine` runner)
  - [ ] Verify `go test ./...` passes

- [ ] Task 3: Scaffold Python `ai-worker` service (AC: 1, 2, 3)
  - [ ] Create `ai-worker/app/__init__.py`, `app/main.py` (FastAPI app, `GET /healthz` → `{"status":"ok"}`), `app/embed.py`, `app/extract.py`, `app/models.py` as stubs
  - [ ] Create `ai-worker/requirements.txt`: `fastapi`, `uvicorn[standard]`, `anthropic`, `openai`, `pydantic`, `structlog` (required for logging stubs in app/main.py)
  - [ ] Create `ai-worker/tests/__init__.py`, `test_embed.py`, `test_extract.py`, `test_integration.py` (all with single `pass` test)
  - [ ] Create `ai-worker/Dockerfile` (`python:3.12-slim`, `uvicorn app.main:app --host 0.0.0.0 --port 8001`)
  - [ ] Create `ai-worker/pyproject.toml` with `[tool.pytest.ini_options]`

- [ ] Task 4: Scaffold Python `rag` service (AC: 1, 2, 3)
  - [ ] Create `rag/app/__init__.py`, `main.py` (FastAPI app, `GET /healthz` → `{"status":"ok"}`), `pipeline.py`, `fallback.py`, `retrieval.py`, `confidence.py`, `staleness.py`, `models.py`, `db.py` as stubs
  - [ ] Create `rag/requirements.txt`: `fastapi`, `uvicorn[standard]`, `anthropic`, `openai`, `pgvector`, `asyncpg`, `pydantic`, `structlog` (required for logging stubs in app/main.py)
  - [ ] Create `rag/tests/__init__.py`, `test_pipeline.py`, `test_fallback.py`, `test_retrieval.py`, `test_confidence.py` (all with single `pass` test)
  - [ ] Create `rag/Dockerfile` (`python:3.12-slim`, `uvicorn app.main:app --host 0.0.0.0 --port 8002`)
  - [ ] Create `rag/pyproject.toml`

- [ ] Task 5: Scaffold Python `signal-job` service (AC: 1, 2, 3)
  - [ ] Create `signal-job/app/__init__.py`, `compute.py`, `db.py`, `models.py` as stubs
  - [ ] Create `signal-job/app/main.py` with a local-dev sleep loop so docker-compose doesn't restart-loop:
    ```python
    import time, structlog
    logger = structlog.get_logger()
    if __name__ == "__main__":
        logger.info("signal-job started", service="signal-job")
        while True:   # local-dev only: keeps container alive; production uses Coolify cron
            time.sleep(60)
    ```
  - [ ] Create `signal-job/requirements.txt`: `psycopg2-binary>=2.9.0`, `pydantic>=2.0.0`, `structlog>=24.0.0` — do NOT add `asyncpg`; signal-job is synchronous (blocking cron script, not async server)
  - [ ] Create `signal-job/tests/__init__.py`, `test_compute.py`, `test_cold_start.py` (single `pass` tests)
  - [ ] Create `signal-job/Dockerfile` (`python:3.12-slim`, `CMD ["python", "-m", "app.main"]`)
  - [ ] Create `signal-job/pyproject.toml`

- [ ] Task 6: Scaffold Next.js `web` service (AC: 1, 3)
  - [ ] Bootstrap with `npx create-next-app@latest web --typescript --tailwind --eslint --app --src-dir --import-alias "@/*"`
  - [ ] Create component stubs under `src/components/ui/` (Button, Card, Table, Badge, Dialog, Sheet — empty stub files; shadcn copies land here in later stories), `src/components/query/`, `dashboard/`, `admin/` per spec (empty `export default function X() { return null }`)
  - [ ] Create hook stubs `src/hooks/useQueryStream.ts`, `useSignalCards.ts`, `useIngestionStatus.ts`
  - [ ] Create `src/lib/api.ts`, `utils.ts`, `src/types/api.ts` as stubs
  - [ ] Create all page and layout stubs: `app/(auth)/login/page.tsx`, `app/(auth)/invite/[token]/page.tsx`, `app/(app)/layout.tsx` (auth guard + org context provider — required by all authenticated pages), `app/(app)/query/page.tsx`, `app/(app)/dashboard/page.tsx`, `app/(app)/admin/page.tsx`, `app/(app)/admin/integrations/page.tsx`, `app/(app)/admin/audit/page.tsx`
  - [ ] Create `web/.env.local` with `NEXT_PUBLIC_API_URL=http://localhost:8080` — Next.js only exposes env vars to the browser when prefixed with `NEXT_PUBLIC_`; the API URL must use this prefix or client-side fetches will fail in all later stories
  - [ ] Create `web/Dockerfile`

- [ ] Task 7: docker-compose.yml (AC: 1)
  - [ ] Define services: `postgres` (postgres:16-alpine, port 5432), `pgbouncer` (edoburu/pgbouncer:latest, port 6432, transaction-mode), `api` (port 8080), `ai-worker` (port 8001), `rag` (port 8002), `signal-job` (no port, `restart: "no"`), `web` (port 3000)
  - [ ] Add named volume for Postgres data (`pgdata`)
  - [ ] Set Postgres init env vars on the `postgres` service: `POSTGRES_USER=orgbrain`, `POSTGRES_PASSWORD=orgbrain`, `POSTGRES_DB=orgbrain` — without these the database does not exist and all connections fail
  - [ ] Set PgBouncer env vars on the `pgbouncer` service: `DB_USER=orgbrain`, `DB_PASSWORD=orgbrain`, `DB_HOST=postgres`, `DB_PORT=5432`, `DB_NAME=orgbrain`, `POOL_MODE=transaction`, `IGNORE_STARTUP_PARAMETERS=extra_float_digits` — `edoburu/pgbouncer` is configured via env vars, not a mounted ini; `POOL_MODE=transaction` is mandatory
  - [ ] Set `DATABASE_URL` env vars pointing through PgBouncer (port 6432) for api/rag; `DIRECT_DATABASE_URL` pointing directly to Postgres (port 5432) for signal-job (psycopg2 sync) and migrations
  - [ ] Healthcheck on `postgres` service so dependent services wait for it

- [ ] Task 8: Makefile targets (AC: 1, 2, 4, 5)
  - [ ] `make dev` — runs `docker-compose up --build`
  - [ ] `make test` — runs `go test ./...` in `api/` + `pytest` in `ai-worker/ rag/ signal-job/`
  - [ ] `make migrate` — runs `goose -dir migrations/shared postgres "$(DIRECT_DATABASE_URL)" up`
  - [ ] `make codegen` — runs `oapi-codegen` against `openapi/ai-worker.yaml` → `api/internal/client/ai_worker_client.go` and `openapi/rag.yaml` → `api/internal/client/rag_client.go`
  - [ ] `make load-test` — stub that echoes "load-test: not yet enabled" and exits 0

- [ ] Task 9: Verify end-to-end (AC: 1–5)
  - [ ] `make dev` starts cleanly, no crash loops
  - [ ] `curl localhost:8080/healthz`, `curl localhost:8001/healthz`, `curl localhost:8002/healthz` all return 200
  - [ ] `make test` exits 0
  - [ ] `make codegen` exits 0

## Dev Notes

### Critical Architecture Rules (Non-Negotiable)
This story creates stubs. Future stories will flesh them out. Even in stub form, these rules must be visible as scaffolding structure:

1. `KnowledgeStoreAdapter` (`api/internal/store/adapter.go`) is the **sole storage path** — stub it with the interface and a comment: `// ALL tenant storage reads and writes go through this adapter. No raw db.Query() in handlers or workers.`
2. `IngestionFilter` (`api/internal/filter/ingestion_filter.go`) must be stubbed with a comment: `// Must run before any LLM call. Rejects DMs (channel_type=im|mpim) with audit log event.`
3. `FallbackRouter` (`rag/app/fallback.py`) must be stubbed with the four enum constants: `ROUTE_TO_OWNER`, `NO_COVERAGE`, `REPHRASE`, `ACCESS_FILTERED`
4. `rag/app/db.py` must include a comment: `// SET LOCAL search_path = org_{id} per transaction — never session-level.`

### Service Ports & Internal Network
| Service | External Port | Internal Host (docker network) |
|---|---|---|
| api (Go) | 8080 | `api:8080` |
| ai-worker (Python) | 8001 | `ai-worker:8001` |
| rag (Python) | 8002 | `rag:8002` |
| signal-job (Python) | none | none |
| web (Next.js) | 3000 | `web:3000` |
| postgres | 5432 (direct) | `postgres:5432` |
| pgbouncer | 6432 | `pgbouncer:6432` |

**Critical PgBouncer config:** Must be `pool_mode=transaction`. This is required — session mode causes tenant `search_path` leakage. The `pgbouncer.ini` must include `pool_mode = transaction` and `ignore_startup_parameters = extra_float_digits`.

### Environment Variables Pattern
All env vars follow `{SERVICE}_{RESOURCE}_{PROPERTY}` convention. Key vars for `.env.example`:
```
# Database (shared)
DATABASE_URL=postgres://orgbrain:orgbrain@pgbouncer:6432/orgbrain
DIRECT_DATABASE_URL=postgres://orgbrain:orgbrain@postgres:5432/orgbrain

# API
API_SESSION_SECRET=change-me-in-production
API_PORT=8080

# ai-worker
AI_WORKER_PORT=8001
AI_WORKER_OPENAI_API_KEY=
AI_WORKER_ANTHROPIC_API_KEY=

# rag
RAG_PORT=8002
RAG_OPENAI_API_KEY=
RAG_ANTHROPIC_API_KEY=
RAG_RETRIEVAL_STRATEGY=pgvector

# signal-job
SIGNAL_JOB_DATABASE_URL=postgres://orgbrain:orgbrain@postgres:5432/orgbrain
SIGNAL_JOB_ANTHROPIC_KEY=
```

### Go Module Layout
- Module path: `github.com/orgbrain/orgbrain/api` (adjust to actual GitHub org if different)
- Go version: **1.23** minimum (required for `slog` stdlib structured logging)
- Key dependencies and versions:
  - `github.com/go-chi/chi/v5` v5.x (latest)
  - `github.com/jackc/pgx/v5` v5.x
  - `github.com/pgvector/pgvector-go/pgx` (matches pgvector 0.8.x Postgres extension)
  - `github.com/riverqueue/river` v0.x (Postgres-native queue, no Redis)
  - `github.com/anthropics/anthropic-sdk-go`
  - `github.com/pressly/goose/v3` (migration runner)

### Python Service Layout
All three Python services (`ai-worker`, `rag`, `signal-job`) use `app/` as the top-level package. Uvicorn entry point: `uvicorn app.main:app`. **Do NOT put `main.py` at the service root** — it must be `app/main.py`.

The architecture's "Complete Project Directory Structure" section shows `ai-worker` files at the root without an `app/` folder, but this contradicts the "Initialization Commands" section which explicitly runs `mkdir -p app` for `ai-worker` and states `uvicorn app.main:app` as the entry point. The `app/` layout is correct for all three Python services.

### Python Requirements
Use `requirements.txt` (not `pyproject.toml` dependencies) for simplicity in Phase 1. Pin to compatible versions:
- `fastapi>=0.115.0`
- `uvicorn[standard]>=0.30.0`
- `pydantic>=2.0.0` (Pydantic v2 — do NOT use v1 syntax)
- `anthropic>=0.34.0`
- `openai>=1.50.0`
- `asyncpg>=0.29.0`
- `pgvector>=0.3.0`
- `psycopg2-binary>=2.9.0` + `structlog>=24.0.0` (signal-job only — sync driver; do NOT add asyncpg to signal-job)

### Logging Stubs
Even in stub files, include the logging pattern comment. Every log line across all services must carry: `service`, `level`, `msg`, `trace_id`, `org_id` (tenant ops), `duration_ms` (external calls).

**Go** (`api/cmd/api/main.go`):
```go
import "log/slog"
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
// All log calls: logger.Info("msg", "service", "api", "trace_id", traceID, ...)
```

**Python** — import `structlog` in each service's `main.py`:
```python
import structlog
logger = structlog.get_logger()
# All log calls: logger.info("msg", service="ai-worker", trace_id=..., ...)
```
Add `structlog>=24.0.0` to Python requirements.

### Makefile: test target
**CRITICAL: Makefile recipe lines MUST use hard TAB characters (`\t`), not spaces.** Using spaces produces `*** missing separator. Stop.` Make will not hint at the cause. Every indented line in the Makefile must be a real tab.

```makefile
test:
	cd api && go test ./...
	cd ai-worker && python -m pytest tests/
	cd rag && python -m pytest tests/
	cd signal-job && python -m pytest tests/
```
(The indentation above is a TAB — verify your editor inserts tabs in `.makefile` context.)

### OpenAPI Stub Format
Both `openapi/ai-worker.yaml` and `openapi/rag.yaml` must be valid OpenAPI 3.0 documents for `oapi-codegen` to succeed:
```yaml
openapi: "3.0.3"
info:
  title: ai-worker
  version: "0.1.0"
paths:
  /healthz:
    get:
      operationId: healthz
      responses:
        "200":
          description: OK
```

### oapi-codegen Configuration
`make codegen` requires `oapi-codegen` config files. Create `api/oapi-codegen-ai-worker.yaml` and `api/oapi-codegen-rag.yaml`. Use the **v2 config format** (installed via `go install github.com/oapi-codegen/oapi-codegen/v2/...`). The v1 key `chi-server` does not exist in v2 — omit it entirely:
```yaml
# api/oapi-codegen-ai-worker.yaml
package: client
generate:
  models: true
  client: true
output: internal/client/ai_worker_client.go
```
```yaml
# api/oapi-codegen-rag.yaml
package: client
generate:
  models: true
  client: true
output: internal/client/rag_client.go
```

### Migration Files
Migration files use Goose format. Even as stubs, they must include the Goose up/down markers:
```sql
-- +goose Up
-- placeholder

-- +goose Down
-- placeholder
```

### GitHub Actions Stub
The CI workflow `ci.yml` should have a stub job that echoes the future steps as comments, so the CI badge turns green immediately:
```yaml
name: CI
on: [push, pull_request]
jobs:
  ci:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: echo "CI stub — tests added in Story 1.5"
```

### What NOT to implement in this story
- No actual database schema (Story 1.2)
- No KnowledgeStoreAdapter logic (Story 1.3)
- No FallbackRouter logic (Story 1.4)
- No CI pipeline tests (Story 1.5)
- No authentication middleware
- No business logic anywhere

This story is 100% scaffolding: create every file in the spec, wire up docker-compose so `make dev` starts all services, `make test` passes on empty suites.

### Project Structure Notes

The architecture document (`architecture.md` §"Complete Project Directory Structure") is the authoritative source. Every named file in that tree must exist after this story. Key paths to verify are present (not exhaustive):

**api/**
- `cmd/api/main.go`
- `internal/auth/middleware.go`, `session.go`, `session_test.go`, `tier.go`
- `internal/handler/query_handler.go`, `query_handler_test.go`, `signal_handler.go`, `signal_handler_test.go`, `admin_handler.go`, `admin_handler_test.go`, `webhook_handler.go`, `webhook_handler_test.go`, `feedback_handler.go`, `feedback_handler_test.go`, `health_handler.go`
- `internal/slack/oauth.go`, `events.go`, `reconcile.go`, `scope_assert.go`
- `internal/queue/ingestion_job.go`, `ingestion_job_test.go`, `ownership_job.go`, `expiry_job.go`, `worker.go`
- `internal/store/adapter.go`, `adapter_test.go`, `knowledge_store.go`, `knowledge_store_test.go`, `session_store.go`, `ownership_store.go`, `signal_store.go`, `feedback_store.go`, `audit_store.go`, `rate_limit_store.go`, `tenant.go`
- `internal/filter/ingestion_filter.go`, `ingestion_filter_test.go`
- `internal/ownership/ownership.go`, `ownership_test.go`
- `internal/client/ai_worker_client.go`, `rag_client.go` (oapi-codegen generated stubs)
- `Dockerfile`, `go.mod`, `go.sum` (auto-generated by `go get` — do NOT create manually)

**ai-worker/**
- `app/__init__.py`, `app/main.py`, `app/embed.py`, `app/extract.py`, `app/models.py`
- `openapi.yaml` (also duplicated as `openapi/ai-worker.yaml` at monorepo root)
- `tests/__init__.py`, `tests/test_embed.py`, `test_extract.py`, `test_integration.py`
- `requirements.txt`, `Dockerfile`

**rag/**
- `app/__init__.py`, `app/main.py`, `app/pipeline.py`, `app/fallback.py`, `app/retrieval.py`, `app/confidence.py`, `app/staleness.py`, `app/models.py`, `app/db.py`
- `openapi.yaml` (also duplicated as `openapi/rag.yaml` at monorepo root)
- `tests/__init__.py`, `tests/test_pipeline.py`, `tests/test_fallback.py`, `tests/test_retrieval.py`, `tests/test_confidence.py`
- `requirements.txt`, `Dockerfile`

**signal-job/**
- `app/__init__.py`, `main.py`, `compute.py`, `db.py`, `models.py`
- `tests/test_compute.py`, `test_cold_start.py`
- `requirements.txt`, `Dockerfile`

**web/** — generated by `create-next-app`, then add component/hook/page stubs per spec.

**migrations/**
- `shared/001_organizations.sql` through `005_audit_log.sql`
- `tenant/001_knowledge_nodes.sql` through `007_ingestion_events.sql`

**root**
- `docker-compose.yml`, `Makefile`, `.env.example`, `.gitignore`
- `.github/workflows/ci.yml`, `release.yml`
- `openapi/ai-worker.yaml`, `openapi/rag.yaml`
- `scripts/migrate-all-tenants.sh`, `load-test.sh`, `isolation-test.sh`

### References

- Architecture §"Monorepo Structure": exact directory layout authority — [Source: `_bmad-output/planning-artifacts/architecture.md#Monorepo Structure`]
- Architecture §"Complete Project Directory Structure": full file list — [Source: `_bmad-output/planning-artifacts/architecture.md#Complete Project Directory Structure`]
- Architecture §"Initialization Commands": exact CLI commands for scaffolding — [Source: `_bmad-output/planning-artifacts/architecture.md#Initialization Commands`]
- Architecture §"AI Agent Enforcement Rules": 9 non-negotiable rules, rules 1–3 directly affect file naming and structure — [Source: `_bmad-output/planning-artifacts/architecture.md#AI Agent Enforcement Rules`]
- Architecture §"Naming Conventions": env var naming, package naming — [Source: `_bmad-output/planning-artifacts/architecture.md#Naming Conventions`]
- Epics Story 1.1 acceptance criteria — [Source: `_bmad-output/planning-artifacts/epics.md#Story 1.1`]

## Dev Agent Record

### Agent Model Used

claude-sonnet-4-6 (story context created 2026-05-30)

### Debug Log References

### Completion Notes List

### File List

# OrgBrain

AI-powered organizational memory for remote-first tech companies. Captures knowledge from where it actually lives — Slack, meeting transcripts, documents — and makes it queryable in plain language. Surfaces departure risk signals for leadership before they become resignations.

## What it does

**Base Tier (all Members)**
- Continuous ingestion of public Slack channels into a structured Knowledge Graph
- Natural language query interface: ask anything, get a sourced, confidence-scored answer in under 10 seconds
- Knowledge Owner fallback when confidence is low — routes to the right person, never surfaces a guess as fact

**Intelligence Tier (leadership)**
- Departure risk signals derived from Slack engagement velocity trends
- Plain-language signal cards with severity levels (Watch / Concerning / Urgent)
- 30+ days of lead time before a likely resignation surfaces as a signal

## Architecture

Five-service monorepo:

| Service | Stack | Responsibility |
|---|---|---|
| `api` | Go + Chi v5 | HTTP API, SSE streaming, Slack webhook, River queue, auth |
| `ai-worker` | Python + FastAPI | Embedding (OpenAI) + Knowledge Node extraction (Anthropic) |
| `rag` | Python + FastAPI | RAG pipeline: embed → retrieve → generate → confidence score |
| `signal-job` | Python cron | Engagement velocity signal computation |
| `web` | Next.js 16.2 | Query UI, Intelligence Dashboard, Admin panel |

**Infrastructure:** Postgres 16 + pgvector · River queue (no Redis) · PgBouncer · Hetzner + Coolify v4 · ~€18–20/month

## Status

Planning complete. Implementation starting.

| Artifact | Status |
|---|---|
| Product Brief | Done |
| PRD (full V1) | Done |
| MVP PRD (Phase 1) | Done |
| Architecture | Done |
| Epics & Stories | In progress |

## Project layout

```
_bmad-output/planning-artifacts/   # PRD, architecture, epics & stories
docs/                              # Domain research and project knowledge
api/                               # Go API service (implementation — coming)
ai-worker/                         # Python embedding + extraction (coming)
rag/                               # Python RAG pipeline (coming)
signal-job/                        # Python signal computation (coming)
web/                               # Next.js frontend (coming)
migrations/                        # Shared + per-tenant Postgres migrations (coming)
```

## Privacy

OrgBrain never ingests Slack DMs or private channels — in any configuration, in any tier, ever. This is a hard system constraint enforced at three independent layers, not a policy setting.

Deploying OrgBrain requires the organization to disclose its use to employees before ingestion begins.

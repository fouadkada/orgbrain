---
title: OrgBrain MVP PRD — Phase 1
status: final
created: 2026-05-26
updated: 2026-05-26
---

# PRD: OrgBrain — Phase 1 MVP

> **This is a scoped subset of the full V1 PRD.** Personas, glossary, privacy architecture, and NFRs are inherited from the parent document: `prds/prd-OrgBrain-2026-05-25/prd.md`. This document specifies only what is built in Phase 1 and why.

---

## 0. Purpose and Framing

Phase 1 is the minimum build needed to close a first paying pilot and validate OrgBrain's core thesis: that passive ingestion of organizational communication produces a knowledge graph accurate enough to answer team questions and surface departure risk signals a VP will act on.

**The bet:** Slack is the richest single source of organizational knowledge in a remote-first tech company. Starting with Slack only is an honest, credible Phase 1 — not a limitation. It is framed to pilot customers as "Slack-first, with additional ingestion channels in Phase 2."

**What Phase 1 proves:**
1. Ingestion and Knowledge Node extraction work reliably on real organizational Slack data
2. The Query Interface produces answers accurate enough that ICs use it instead of pinging a colleague
3. The departure risk signal is trustworthy enough that a VP pays for it and acts on it

---

## 1. Vision (Phase 1)

Same thesis as the full PRD. One channel, one signal, two surfaces. A VP pays because the departure risk signal is real — not because the product is complete.

---

## 2. Target Users (Phase 1)

Inherits all personas from §2.1 of the parent PRD. Phase 1 delivers value to three of the four:

| Persona | Phase 1 value |
|---|---|
| **Buyer: VP of Engineering** | Pays for the Slack engagement departure risk signal |
| **IC** | Gets answers to questions via web UI, without waiting for a Slack reply |
| **New Hire** | Same web UI query access as IC; no Pinned Content landing experience yet (Phase 2) |

### Phase 1 User Journeys

**UJ-1. IC gets unblocked (web UI only).**

- **Entry state:** Authenticated, OrgBrain web UI.
- **Path:** Navigates to OrgBrain web UI. Types natural language question. OrgBrain returns answer with Source Attribution, Confidence Score, and relevant context. If below Confidence Threshold: declines, surfaces Knowledge Owner.
- **Climax:** Engineer has a sourced answer or a named person to ping.
- **Edge case:** Stale answer flagged; Knowledge Owner offered.

**UJ-4. VP sees a Slack engagement signal and acts.**

- **Entry state:** Authenticated, OrgBrain web UI, Intelligence Tier.
- **Path:** Opens Intelligence Dashboard. Sees plain-language signal cards based on Slack engagement velocity trend. Identifies a flagged Member, escalates to Senior EM.
- **Climax:** Retention conversation happens with 30–90 days of lead time.
- **Edge case:** False positive — signal resolves automatically as Slack activity normalizes.

---

## 3. Glossary

Inherits all terms from §3 of the parent PRD. No new terms introduced in Phase 1.

---

## 4. Phase 1 Features

### 4.1 Slack Ingestion

Identical to FR-1 in the parent PRD.

The system ingests all messages from connected Slack public channels continuously after authorization. Private channels and DMs are never ingested in any configuration.

**Consequences:**
- New messages appear in the Knowledge Graph within 15 minutes of posting. [ASSUMPTION: A-1]
- The Admin can exclude specific public channels from ingestion.

**Out of scope in Phase 1:** Google Meet transcripts, Google Drive, Historical Import. These are Phase 2.

---

### 4.2 Knowledge Graph & Memory

Identical to FR-5, FR-6, FR-7 in the parent PRD. All three are required in Phase 1 — the Knowledge Ownership Map is needed for the Knowledge Owner fallback (FR-9) and the departure risk signal (FR-13p).

---

### 4.3 Query Interface — Web UI Only

**FR-8p: Natural language query (web UI)**

Any authenticated Member can submit a natural language question via the OrgBrain web UI and receive a response.

**Consequences:**
- Response includes: direct answer, Source Attribution, Confidence Score, and relevant related context.
- Response time under 10 seconds. [ASSUMPTION: A-3]
- Follow-up questions supported within a session. Session expires after 30 minutes of inactivity. [ASSUMPTION: A-4]

**Phase 1 scope note:** The Slack DM bot surface is Phase 2. Web UI is the only query surface in Phase 1.

**FR-9: Confidence Threshold and Knowledge Owner fallback** — unchanged from parent PRD.

**FR-10: Staleness indication** — unchanged from parent PRD.

---

### 4.4 Intelligence Dashboard — Single Signal

**FR-13p: Slack engagement velocity signal**

The system computes a departure risk indicator based on one Engagement Signal cluster: Slack message frequency and response time trend.

**Signal definition:**
- **Engagement velocity** — 30-day trend in message frequency and response time in Slack public channels per Member.
- A Departure Risk Signal is generated when engagement velocity shows a sustained declining trend over 30 days (consistent with the 30-day window and trend logic in FR-13 of the parent PRD).

**Consequences:**
- Signal is computed continuously; dashboard reflects current state.
- If engagement velocity recovers, the signal resolves automatically.

**Phase 1 scope note:** Only one of the four Engagement Signal clusters from the full PRD (FR-13) is implemented in Phase 1. The remaining three (meeting participation, calendar activity, future-orientation) are Phase 2. The framing to pilot customers: "Slack engagement is our first signal; additional signals from meeting participation, calendar activity, and language patterns are in active development."

**FR-14: Intelligence Dashboard display** — unchanged from parent PRD. Plain-language signal cards, no raw metrics, severity levels (Watch / Concerning / Urgent).

**FR-15: Signal card self-explanation** — unchanged from parent PRD.

---

### 4.5 Admin & Access Control — Slack Only

**FR-16p: Slack integration management**

An Admin can connect and disconnect the Slack integration.

**Consequences:**
- Connection requires Slack OAuth authorization.
- Admin can exclude specific public channels from ingestion.
- Disconnecting stops future ingestion; existing Knowledge Nodes are retained.

**Phase 1 scope note:** Google Workspace integration management is Phase 2.

**FR-17: Member access management** — unchanged from parent PRD.

**FR-18: Privacy configuration** — unchanged from parent PRD.

---

## 5. What Phase 1 Does Not Include

- Slack DM bot surface (web UI only)
- Google Meet transcript ingestion
- Google Drive document ingestion
- Historical Import
- Content Pinning and new Member landing experience
- Full 4-cluster Intelligence Dashboard (only Slack engagement velocity)
- Google Calendar integration

*All items above are Phase 2. Full V1 out-of-scope items are in §6.3 of the parent PRD.*

---

## 6. Success Metrics (Phase 1)

**Primary**

- **SM-P1: First paying pilot** — At least one VP of Engineering signs a pilot agreement within 90 days of Phase 1 launch.
- **SM-P2: Signal lead time** — At least one Departure Risk Signal acted on during the pilot, surfaced 30+ days before the situation escalated. Validates FR-13p, FR-14.
- **SM-P3: Query accuracy** — Knowledge Owner fallback accuracy above 80% in the first 30 days of a pilot (OrgBrain routes to the right person when it doesn't know). Validates FR-9.

**Secondary**

- **SM-P4: IC adoption** — At least 40% of pilot org ICs submit at least one query per week within the first 30 days. Validates FR-8p.
- **SM-P5: Query deflection** — 50%+ of queries answered without a follow-up Slack ping to a colleague within first 60 days. Validates FR-8p, FR-9.

**Counter-metrics (do not optimize)**

- **SM-C1: Signal gaming** — same as parent PRD. One-signal MVP is more gameable; watch for ICs artificially inflating Slack activity once aware of the signal.
- **SM-C2: Answer rate vs. accuracy** — same as parent PRD.

---

## 7. Open Questions (Phase 1 Specific)

1. **Pilot pricing:** What does the Phase 1 pilot cost? A fixed monthly fee (e.g., $500–$2,000/month flat for the org) reduces friction for a first pilot and avoids the per-seat negotiation before the product has proven itself.
2. **Minimum data volume for signal reliability:** How many weeks of Slack data does OrgBrain need before the engagement velocity signal is trustworthy enough to surface? This sets the "activation delay" expectation for pilot customers.
3. **Signal threshold calibration:** The 30-day window and "sustained declining trend" definition need calibration against real data. This is the first engineering uncertainty to resolve in Phase 1 implementation.

---

## 8. Assumptions (Phase 1)

Inherits A-1 through A-8 from the parent PRD. Phase 1 specific:

- **A-P1:** Pilot customers accept Slack-only as a credible Phase 1 scope when framed as "Slack-first, additional channels in Phase 2." Validated by founder's read of buyer behavior — confirm with first sales conversations.
- **A-P2:** One departure risk signal (Slack engagement velocity) is sufficient to demonstrate value and close a pilot. If a VP requires multiple corroborating signals before trusting the output, Phase 2 acceleration is needed.

---

## 9. Privacy & Disclosure Architecture

Inherits fully from §10 of the parent PRD. No changes for Phase 1. The hard limits, organizational disclosure model, and data governance rules apply identically — the reduced scope does not reduce the privacy obligations.

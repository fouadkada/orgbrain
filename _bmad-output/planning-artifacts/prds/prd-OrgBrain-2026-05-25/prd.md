---
title: OrgBrain PRD
status: final
created: 2026-05-25
updated: 2026-05-26
---

# PRD: OrgBrain

## 0. Document Purpose

This PRD defines the V1 product requirements for OrgBrain, an AI-powered organizational memory system. Written for the founder (who is also the architect and lead engineer), and serves as the authoritative input for downstream architecture, epics, and implementation work. Structure: Glossary-anchored vocabulary (§3), features grouped by capability cluster with globally numbered FRs (§4), assumptions tagged inline and indexed in §9. Privacy architecture is treated as a first-class section (§10) given the sensitivity of the data OrgBrain handles.

Inputs this PRD synthesizes:
- Product brief: `planning-artifacts/briefs/brief-OrgBrain-2026-05-22/brief.md`
- Brief addendum: `planning-artifacts/briefs/brief-OrgBrain-2026-05-22/addendum.md`
- PRD coaching session: 2026-05-25

Related documents:
- **Phase 1 MVP PRD:** `prds/prd-OrgBrain-MVP-2026-05-26/prd-mvp.md` — Slack-first subset for first pilot. Inherits personas, glossary, and privacy architecture from this document.

---

## 1. Vision

Every organization leaks knowledge constantly. Decisions made in Slack threads nobody saved. A meeting recorded nowhere. A key engineer who leaves and takes five years of context with them. The company's real operating knowledge — how things actually work, why they were built that way, who actually owns what — lives nowhere, belongs to no one, and quietly evaporates.

OrgBrain is an AI-powered organizational memory system that captures knowledge from where it actually lives — Slack, meeting transcripts, documents — and makes it queryable by anyone on the team in plain language. No one has to decide to document anything. When someone asks "what did we decide about the API versioning?" or "who owns the auth work?" or "what blockers are open?" — OrgBrain answers in seconds, with sources.

For leadership, OrgBrain goes further: it surfaces the signals hiding inside the organization's behavioral patterns — who is becoming disengaged, who holds critical knowledge alone, where the org is quietly losing people before it knows it. The knowledge layer makes the organization more productive. The intelligence layer makes it more resilient.

V1 ships two tiers: **Base Tier** (knowledge graph + query interface for all employees) and **Intelligence Tier** (departure risk dashboard for leadership). Target market: remote-first tech companies, 30–150 employees.

---

## 2. Target Users

### 2.1 Personas

**Buyer: VP of Engineering** (or CEO at smaller companies)
Approves, installs, and is a primary beneficiary of the Intelligence Tier. Currently relies on Senior EMs escalating retention concerns and People Business Partners running manual 1:1s to detect flight risk — an expensive, slow, and inconsistent process. Wants to see signals before someone resigns, not after. Decision timeline: days, not quarters.

**Individual Contributor (IC)**
Remote-first engineer who gets blocked waiting hours for answers in async environments. Will use OrgBrain if it's already in their workflow and frictionless; will not seek it out or configure it. Core value: get unblocked without taxing a senior colleague. Engineering Managers are included in this persona for V1 — their V1 value is identical to ICs. They re-emerge as a distinct persona in V2 when proactive pre-meeting briefs are introduced.

**New Hire**
Joining a remote company with no ambient context, afraid to ask questions that reveal what they don't know. Currently pings the most approachable team member or their onboarding buddy — creating interrupt load on the team's most accessible people. Core value: always-available knowledge access with no social cost.

### 2.2 Jobs To Be Done

- Get an answer without waiting hours for a Slack reply
- Understand why a system was built a certain way without finding the person who built it
- Onboard without becoming a drain on the team's most approachable members
- Detect who is likely to leave before they hand in a resignation

### 2.3 Non-Users (V1)

- Organizations outside the 30–150 employee range
- Non-technical teams (primary V1 validation is in engineering orgs)
- Microsoft 365 organizations (V2)
- Organizations with strict data residency requirements outside standard cloud regions

### 2.4 Key User Journeys

**UJ-1. IC gets unblocked without interrupting a colleague.**

- **Persona + context:** Remote engineer, blocked on a past decision or system design question, async environment, answer could be hours away via Slack.
- **Entry state:** Authenticated, in Slack or OrgBrain web UI.
- **Path:** Opens Slack → OrgBrain DM bot, or navigates to OrgBrain web UI. Types natural language question. OrgBrain returns an answer with source links, Confidence Score, and all relevant related context. If confidence is below Confidence Threshold: declines to answer, surfaces the Knowledge Owner most likely to know.
- **Climax:** Engineer has a sourced answer or knows exactly who to ping. No senior engineer was interrupted.
- **Resolution:** Engineer resumes work. Question is logged to the Knowledge Graph as a signal of what the team is asking about.
- **Edge case:** Answer exists but is stale (source is 8+ months old, newer decisions exist). OrgBrain flags staleness, surfaces the most recent related context, and offers the Knowledge Owner fallback.

**UJ-2. New hire gets oriented without taxing the team.**

- **Persona + context:** First week at a remote company. Doesn't know what they don't know. Onboarding buddy exists but is a single point of failure.
- **Entry state:** Authenticated, in Slack or OrgBrain web UI.
- **Path:** Landing experience shows Pinned Content — team structure, key decisions, processes, company glossary, norms. New hire reads and browses. Drops into free-form query for everything else (same interface as UJ-1).
- **Climax:** New hire gets oriented without scheduling a meeting or sending a Slack message. Onboarding buddy's time is protected.
- **Resolution:** New hire is productive faster; team's most approachable members are not interrupt-taxed.
- **Edge case:** Pinned content is outdated. OrgBrain surfaces the pin's creation date; any Member can update or remove their own pins.

**UJ-3 (EM pre-meeting brief) — deferred to V2. See §6.2.**

**UJ-4. VP acts on a departure risk signal before it's too late.**

- **Persona + context:** VP of Engineering, 45-person remote company. Doesn't know someone is disengaging until they resign.
- **Entry state:** Authenticated, OrgBrain web UI, Intelligence Tier.
- **Path:** Opens Intelligence Dashboard. Sees plain-language signal cards — "Engagement declining over 30 days across 3 signals" — with severity level (Watch / Concerning / Urgent). No raw metrics, no scores. VP identifies a flagged Member, escalates to the responsible Senior EM.
- **Climax:** Retention conversation happens while there's still time to act. Signal surfaced 60–90 days before a likely resignation.
- **Resolution:** VP has visibility they previously didn't have. Whether they act is their call — OrgBrain's job ends at surfacing the signal.
- **Edge case:** Signal is a false positive (personal situation, now resolved). Dashboard resolves the signal automatically as Engagement Signals normalize.

---

## 3. Glossary

- **Knowledge Graph** — The structured store of organizational knowledge extracted from ingested sources. Contains Knowledge Nodes, relationships between them, and temporal metadata.
- **Knowledge Node** — A discrete unit of knowledge: a decision, a fact, an ownership assignment, a commitment, or a process. Each Node carries a source, a timestamp, a Confidence Score, and a Knowledge Owner where determinable.
- **Knowledge Owner** — The Member OrgBrain identifies as most associated with a given Knowledge Node — based on authorship, decision-making participation, or domain activity patterns. Surfaced as a fallback when the Query Interface cannot answer with sufficient confidence.
- **Confidence Score** — A signal displayed to the querying Member indicating how reliable OrgBrain's answer is, based on source recency, source authority, and consistency across sources.
- **Confidence Threshold** — The minimum Confidence Score required for OrgBrain to surface an answer. Below this threshold, OrgBrain declines and routes to the Knowledge Owner instead.
- **Source Attribution** — The provenance chain attached to every answer: the original source document or message, its author, and its date.
- **Pinned Content** — Knowledge Nodes or documents explicitly surfaced by any Member for visibility — shown prominently in the landing experience and new Member onboarding flow.
- **Historical Import** — A one-time, Admin-triggered ingestion of existing organizational data (Slack history, Drive documents, Meet transcripts) from before OrgBrain's installation date.
- **Engagement Signal** — A behavioral indicator derived from a Member's activity patterns across connected sources (Slack, Meet, Drive) that contributes to Departure Risk Signal computation.
- **Departure Risk Signal** — A composite, plain-language indicator surfaced on the Intelligence Dashboard when a Member's Engagement Signals show a sustained declining trend across multiple signal clusters.
- **Admin** — A Member with organization-level configuration privileges: managing integrations, triggering Historical Import, configuring privacy settings, and managing Member access.
- **Member** — Any authenticated user of OrgBrain within an organization. All Members have access to the Base Tier.
- **Intelligence Tier** — The paid upgrade layer providing access to the Intelligence Dashboard and Departure Risk Signals. Scoped to designated leadership Members.
- **Base Tier** — The foundational product layer: Knowledge Graph, Query Interface, and Content Pinning. Available to all Members.
- **Knowledge Ownership Map** — The internal data structure tracking which Members are most associated with which knowledge domains, derived from authorship and activity patterns. Powers Knowledge Owner fallback and Engagement Signal computation.

---

## 4. Features

### 4.1 Knowledge Ingestion

**Description:** OrgBrain continuously ingests content from connected sources and extracts structured Knowledge Nodes from it. Ingestion is passive — no Member needs to decide to document anything. V1 sources: Slack (public channels only), Google Meet (transcripts), Google Drive. A one-time Historical Import gives the organization a populated Knowledge Graph on day one rather than building from installation forward. Realizes UJ-1, UJ-2, UJ-4.

**Functional Requirements:**

#### FR-1: Slack public channel ingestion

The system ingests all messages from connected Slack public channels continuously after authorization. Private channels and direct messages are not ingested in any configuration.

**Consequences:**
- New messages in public channels appear in the Knowledge Graph within 15 minutes of posting. [ASSUMPTION: A-1]
- Private channels and DMs are never ingested regardless of Admin configuration.
- The Admin can exclude specific public channels from ingestion.

**Out of Scope:** Private channels, DMs, Slack Connect (external org) channels.

#### FR-2: Google Meet transcript ingestion

The system ingests transcripts from Google Meet meetings where at least one participant is a Member, after the meeting ends.

**Consequences:**
- Transcripts are processed and Knowledge Nodes extracted within 60 minutes of meeting end. [ASSUMPTION: A-2]
- Meetings with transcription disabled produce no ingestion.
- The Admin can exclude specific participants or meeting types from transcript ingestion.

#### FR-3: Google Drive document ingestion

The system ingests documents from connected Google Drive that are shared with at least one Member.

**Consequences:**
- Documents are processed on first connection and re-processed within 2 hours of a document update being detected.
- Only documents with at least viewer-level sharing with a Member are ingested.
- The Admin can exclude specific folders or documents from ingestion.

#### FR-4: Historical Import

An Admin can trigger a one-time Historical Import of pre-installation organizational data. The Admin reviews a summary of what will be ingested before confirming.

**Consequences:**
- Import scope is configurable: date range, source types, channel and folder inclusion/exclusion.
- Admin sees a pre-import summary (volume, sources, date range) before confirming. Import does not begin until confirmed.
- Import runs as a background job; Admin is notified on completion.
- Historical Import can be triggered once per source type; re-triggering requires explicit Admin confirmation.

**Notes:** [NOTE FOR PM: The review-before-confirm step is the organizational consent moment for historical data. The UX must make this legible — what is being ingested, from when, and who can see it. This must not feel like a rubber-stamp.]

---

### 4.2 Knowledge Graph & Memory

**Description:** The Knowledge Graph is OrgBrain's core data structure — the organized, versioned, queryable store of everything OrgBrain has learned. It preserves the history of decisions, not just their current state: when a decision is reversed, the original reasoning is retained alongside the new direction. The Knowledge Ownership Map tracks who is most associated with which knowledge domains, enabling the Knowledge Owner fallback in the Query Interface and Departure Risk Signal computation in the Intelligence Tier. Realizes UJ-1, UJ-2, UJ-4.

**Functional Requirements:**

#### FR-5: Knowledge Node extraction and storage

The system extracts structured Knowledge Nodes from ingested content and stores them with source, timestamp, author, and Confidence Score.

**Consequences:**
- Every Knowledge Node carries: content, source reference, source date, author (if determinable), Confidence Score, and Knowledge Owner (if determinable).
- Nodes are updated when source content changes; all previous versions are retained.

#### FR-6: Decision versioning

When a Knowledge Node representing a decision is subsequently reversed or updated, both the original and updated state are retained with their respective dates and sources.

**Consequences:**
- Querying a reversed decision surfaces both the original reasoning and the current state, in chronological order.
- No Knowledge Node is deleted or overwritten on update — only appended.

#### FR-7: Knowledge Ownership Map

The system maintains a map of which Members are most associated with which knowledge domains, derived from authorship, decision participation, and activity patterns across ingested sources.

**Consequences:**
- Every Knowledge Node with a determinable owner has a Knowledge Owner assigned.
- Ownership assignments are recomputed every 24 hours and within 60 minutes of a significant activity event (e.g., authoring a document, leading a decision in a meeting transcript).
- The Knowledge Ownership Map drives FR-9 (Knowledge Owner fallback) and FR-13 (Departure Risk Signal computation).

---

### 4.3 Query Interface

**Description:** Any Member can ask OrgBrain a natural language question and receive a sourced, confidence-scored answer. Available on two surfaces: the OrgBrain Slack DM bot (private, per-Member) and the OrgBrain web UI. When OrgBrain cannot meet the Confidence Threshold, it declines and routes to the most relevant Knowledge Owner — it never surfaces a low-confidence answer as a confident one. Realizes UJ-1, UJ-2.

**Functional Requirements:**

#### FR-8: Natural language query

Any authenticated Member can submit a natural language question to OrgBrain via Slack DM bot or web UI and receive a response. All Slack interactions occur privately via DM — there is no shared OrgBrain channel.

**Consequences:**
- Response includes: direct answer, Source Attribution (source name, author, date, link), Confidence Score, and relevant related context from the Knowledge Graph.
- Response time under 10 seconds for standard queries. [ASSUMPTION: A-3]
- Follow-up questions within the same session are supported; conversational context is maintained within the session. A session expires after 30 minutes of inactivity; a new query after expiry starts without prior session context. [ASSUMPTION: A-4]

#### FR-9: Confidence Threshold and Knowledge Owner fallback

When OrgBrain's answer confidence falls below the Confidence Threshold, OrgBrain declines to answer and surfaces the Knowledge Owner most likely to have the information.

**Consequences:**
- Response states clearly that OrgBrain does not have a reliable answer.
- Response includes the Knowledge Owner's name and Slack handle (if available) for the relevant domain.
- If no Knowledge Owner can be determined, response states this and suggests the Admin.
- A low-confidence answer is never surfaced as a confident answer.

#### FR-10: Staleness indication

When OrgBrain surfaces an answer from content older than the staleness threshold, it flags the answer as potentially stale and surfaces the most recent related context alongside the original. [ASSUMPTION: A-5 — default 6-month threshold, Admin-configurable]

**Consequences:**
- Staleness flag is visually distinct in both Slack and web UI responses.
- Knowledge Owner fallback is offered alongside stale answers.

---

### 4.4 Content Pinning

**Description:** Any Member can pin Knowledge Nodes or documents to make them prominently visible in the OrgBrain landing experience and new Member onboarding flow. Pinning is the organization's deliberate curation layer on top of passive ingestion — no special role or approval required. Realizes UJ-2.

**Functional Requirements:**

#### FR-11: Pin creation

Any authenticated Member can pin a Knowledge Node or document.

**Consequences:**
- Pinned content appears in the OrgBrain web UI landing experience and in the OrgBrain Slack DM bot welcome message sent to new Members on first activation.
- Each pin displays the pinning Member's name and creation date.
- No approval workflow is required to create a pin.

#### FR-12: Pin management

Any Member can remove or update their own pins. Admins can remove any pin.

**Consequences:**
- Removing a pin removes it from the landing experience immediately.
- The underlying Knowledge Node is not affected by pin removal.

---

### 4.5 Intelligence Dashboard

**Description:** The Intelligence Dashboard is the Intelligence Tier's primary surface — accessible only to Members with Intelligence Tier access (typically VP of Engineering and above). It surfaces Departure Risk Signals in plain language, designed to be self-explanatory without training. No raw metrics, no numerical scores — interpreted signals with clear severity levels. Web UI only in V1. Realizes UJ-4.

**Functional Requirements:**

#### FR-13: Departure Risk Signal computation

The system continuously computes Departure Risk Signals for each Member based on four Engagement Signal clusters derived from V1 data sources.

**Engagement Signal clusters:**
1. **Engagement velocity** — message frequency and response time trend in Slack public channels
2. **Meeting participation** — speaking time trend in Google Meet transcripts
3. **Calendar activity** — meeting acceptance rate and absence pattern trends from Google Calendar [ASSUMPTION: A-6]
4. **Future-orientation** — trend in forward-looking language in Slack and Meet transcripts

**Consequences:**
- A Departure Risk Signal is generated when a Member shows a sustained declining trend across 3 or more Engagement Signal clusters over a 30-day window.
- Signals are computed continuously; the dashboard reflects current state.
- If signals recover, the Departure Risk Signal resolves automatically.

**Out of Scope (V1):** Engineering productivity metrics (PRs, commits, lines of code) — V2/V3 via GitHub/GitLab integration.

#### FR-14: Intelligence Dashboard display

Intelligence Tier Members can view the Intelligence Dashboard in the OrgBrain web UI.

**Consequences:**
- Dashboard displays plain-language signal cards per flagged Member: who, what pattern observed, duration, and severity level (Watch / Concerning / Urgent).
- No raw metrics or numerical scores are visible to the viewer.
- Severity levels are determined by number of declining signal clusters and trend duration.
- Dashboard access is scoped exclusively to Intelligence Tier Members.

#### FR-15: Signal card self-explanation

Each signal card contains sufficient context for a VP-level reader to understand the situation and next action without training or documentation.

**Consequences:**
- Signal card includes: Member name, plain-language pattern description, duration, severity, and a suggested next action (e.g., "Consider a 1:1 check-in via your Senior EM").
- No jargon, model explanation, or raw data is visible by default.

---

### 4.6 Admin & Access Control

**Description:** An Admin manages OrgBrain's organization-level configuration: connecting integrations, triggering Historical Import, configuring privacy settings, and managing Member access and Tier assignments. The Admin role is independent of Intelligence Tier access — being an Admin does not grant access to the Intelligence Dashboard. Underlies all journeys.

**Functional Requirements:**

#### FR-16: Integration management

An Admin can connect, configure, and disconnect V1 data source integrations (Slack, Google Workspace).

**Consequences:**
- Connection requires OAuth authorization by an account with sufficient permissions in the source system.
- Admin can exclude specific channels, folders, documents, or meeting types from ingestion.
- Disconnecting an integration stops future ingestion but does not delete existing Knowledge Nodes from that source.

#### FR-17: Member access management

An Admin can invite Members, assign Intelligence Tier access, and deactivate Members.

**Consequences:**
- Members access OrgBrain only after Admin invitation. [ASSUMPTION: A-7 — SSO is V2]
- Intelligence Tier access is explicitly assigned per Member — not granted by default.
- Deactivated Members lose access immediately; their Knowledge Graph contributions are retained.

#### FR-18: Privacy configuration

An Admin can configure organization-level privacy settings within defined bounds.

**Consequences:**
- Admin can enable or disable specific Engagement Signal clusters from contributing to Departure Risk Signal computation.
- Admin cannot enable DM or private channel ingestion — this is a hard system limit, not a configurable setting.
- All privacy configuration changes are logged with timestamp and Admin identity.

---

## 5. Non-Goals (Explicit)

- **OrgBrain is not an individual productivity monitoring tool.** Engagement Signals feed Departure Risk Signal computation for leadership visibility only — not for performance reviews, compensation decisions, or any reporting visible to direct managers or peers.
- **OrgBrain does not replace documentation tools.** It captures what wasn't documented; it does not generate wikis or structured documentation outputs.
- **OrgBrain does not ingest DMs or private Slack channels** — in any configuration, in any tier, ever.
- **OrgBrain does not make retention decisions.** It surfaces signals; humans decide what to do.
- **OrgBrain does not surface confident answers below the Confidence Threshold.** It does not hallucinate.

*V1 tactical scope boundaries are in §6.2.*

---

## 6. MVP Scope

> **A dedicated Phase 1 MVP PRD exists:** `prds/prd-OrgBrain-MVP-2026-05-26/prd-mvp.md`. It specifies the Slack-first subset built to close a first paying pilot and validate the core thesis before expanding ingestion sources. The scope below reflects the full V1 target.

### 6.1 Phase 1 — MVP (build first)

Slack-only ingestion, web UI query interface, one departure risk signal. Full spec in the MVP PRD.

- Slack public channel ingestion
- Knowledge Graph + Query Interface (web UI only)
- One departure risk signal: Slack engagement velocity trend
- Basic Admin panel: Slack OAuth connection + Member access

### 6.2 Phase 2 — Full V1 additions

- Google Meet transcript ingestion
- Google Drive ingestion
- Slack DM bot surface
- Full 4-cluster Intelligence Dashboard (all Engagement Signal clusters)
- Content Pinning + new Member landing experience
- Admin-triggered Historical Import with review-before-confirm UX

### 6.3 Out of Scope for V1

- **Proactive pre-meeting briefs** (1:1 and planning) — V2 [NOTE FOR PM: This is the EM persona's primary daily value driver. V2 prioritization should be high once Phase 2 validates knowledge graph quality.]
- **Microsoft 365 integration** (Teams, OneDrive, Outlook) — V2 (maybe)
- **Email ingestion** — deferred (privacy complexity)
- **Slack DM and private channel ingestion** — not planned
- **SSO / SCIM provisioning** — V2
- **Engineering productivity metrics** (GitHub/GitLab) — V2/V3
- **SOC 2 certification** — planned post-launch
- **Full executive intelligence suite** (shadow org chart, decision velocity, strategy-execution gap) — V2+
- **Portable professional memory** (employee-owned knowledge graph) — long-term vision

---

## 7. Success Metrics

**Primary**

- **SM-1: Time-to-first-answer** — Median time from question submission to answer received, target < 30 seconds end-to-end. Validates FR-8.
- **SM-2: New hire productivity ramp** — New hires reach working productivity in under 4 weeks (vs. 3–6 month industry baseline). Validates FR-8, FR-11, FR-12.
- **SM-3: Departure risk signal lead time** — At least one Departure Risk Signal acted on within 90 days of deployment, with signal surfaced 30+ days before the situation escalated. Validates FR-13, FR-14.

**Secondary**

- **SM-4: Query deflection rate** — Percentage of IC questions answered by OrgBrain without a follow-up Slack ping to a colleague. Target: 60%+ within first 90 days. Validates FR-8, FR-9.
- **SM-5: Knowledge Owner fallback accuracy** — When OrgBrain routes to a Knowledge Owner, the Knowledge Owner confirms they hold the relevant knowledge. Target: 80%+. Validates FR-7, FR-9.
- **SM-6: Net Revenue Retention** — NRR above 120% at 12 months. Validates product-market fit across both tiers.

**Counter-metrics (do not optimize)**

- **SM-C1: Signal gaming rate.** If Members begin inflating Slack activity or joining meetings without participating, signal quality degrades. Optimize for detection accuracy, not signal coverage.
- **SM-C2: Answer rate.** Do not optimize answer rate at the expense of accuracy. A high answer rate with low Knowledge Owner fallback accuracy is worse than a lower answer rate with high accuracy.

---

## 8. Open Questions

1. **Pricing specifics:** Per-seat or per-org flat pricing for Base and Intelligence tiers? Trial or freemium strategy for initial customer acquisition? (Business model — not blocking V1 requirements.)
2. **Confidence Threshold value:** The specific threshold below which OrgBrain declines to answer requires empirical calibration during beta. Must be implemented as a tunable parameter, not hardcoded.
3. **Google Calendar access scope:** FR-13 requires Google Calendar read access for meeting acceptance rate signals (Engagement Signal cluster 3). Does this fall under the existing Google Workspace OAuth scope or require a separate authorization step?
4. ~~**Slack bot channel model**~~ — **Resolved:** DM bot per Member. All interactions are private. No shared channel. (Decided 2026-05-26)
5. **Intelligence Tier seat count:** How many Intelligence Tier seats does a typical 30–150 person company need? Affects pricing model design and packaging.
6. **Anti-champion sales motion:** The person at a target company whose influence depends on information asymmetry will resist OrgBrain actively. Sales motion to handle or route around this person should be defined before go-to-market.

---

## 9. Assumptions Index

- **A-1** (FR-1): 15-minute ingestion lag for Slack messages is acceptable in V1. Real-time ingestion is not required.
- **A-2** (FR-2): Meet transcripts are processed within 60 minutes of meeting end.
- **A-3** (FR-8): Query response time under 10 seconds for standard queries.
- **A-4** (FR-8): Conversational follow-up questions are supported within a session; context is maintained within the session.
- **A-5** (FR-10): Default staleness threshold is 6 months. Configurable by Admin.
- **A-6** (FR-13): Google Calendar read access is required in V1 for meeting acceptance rate signals, even though calendar-triggered briefs are V2.
- **A-7** (FR-17): SSO/SCIM provisioning is V2; V1 uses Admin-invitation-based Member onboarding.
- **A-8** (§10.3): OrgBrain operates from a single cloud region in V1. Data residency options are V2.

---

## 10. Privacy & Disclosure Architecture

OrgBrain operates at the intersection of organizational productivity and behavioral monitoring. This section defines the non-negotiable privacy boundaries, the organizational disclosure model, and the constraints that must survive every product and engineering decision downstream.

### 10.1 Hard Limits (Non-Configurable)

- Slack DMs and private channels are never ingested. This is a hard system constraint, not an Admin setting.
- Departure Risk Signals are visible only to Intelligence Tier Members. They are never visible to the assessed Member, their direct manager, or peers.
- Engagement Signals are never exposed as raw data to any user in any tier.
- OrgBrain does not produce individual performance reports. Signal computation is departure risk detection only.

### 10.2 Organizational Disclosure Model

Deploying OrgBrain requires the organization (not OrgBrain) to disclose its use to employees. OrgBrain must make this operationally easy:

- The Admin onboarding flow includes a required step: confirmation that the organization has communicated OrgBrain's deployment to Members — including what is monitored (public Slack channels, Meet transcripts, Drive documents) and what is not (DMs, private channels).
- OrgBrain provides a template disclosure notice (plain English) that organizations can adapt.
- The disclosure confirmation is logged with Admin identity and timestamp.

**Why this matters:** In GDPR-covered jurisdictions (EU, UK) and several US states, passive behavioral monitoring of employees requires disclosure. OrgBrain cannot make a non-compliant deployment compliant by itself — but it can make non-compliance operationally harder than compliance.

### 10.3 Data Governance

- All organizational data is scoped to the organization's OrgBrain tenant. No cross-tenant data access in any configuration.
- Knowledge Nodes retain source provenance and are processed and stored in OrgBrain's infrastructure. [ASSUMPTION: A-8]
- Members whose accounts are deactivated have their Engagement Signals excluded from future computation. Their Knowledge Graph contributions are retained as part of the organizational record.
- Admins can request full data deletion for their organization. Deletion is permanent and logged.

---

## 11. Cross-Cutting NFRs

- **Accuracy:** OrgBrain must never surface a confident answer below the Confidence Threshold. False confidence is a worse failure mode than "I don't know."
- **Latency:** Query Interface responses under 10 seconds. Historical Import and background ingestion jobs have no latency SLA but must not degrade query performance during execution.
- **Security:** OAuth tokens stored in a dedicated secrets manager, never in the Knowledge Graph store. All organizational data scoped to the org's tenant with no cross-tenant read path. No OrgBrain employee has read access to an organization's Knowledge Graph content without explicit customer authorization.
- **Availability:** Query Interface and Intelligence Dashboard target 99.5% uptime. Ingestion pipelines may tolerate higher downtime without impacting user-facing availability.
- **Scalability:** V1 must handle organizations up to 150 Members and up to 3 years of historical Slack, Drive, and Meet data without degraded query performance.

---

## 12. Monetization

**Base Tier:** Knowledge Graph + Query Interface + Content Pinning. Available to all Members. [ASSUMPTION: flat per-org pricing with a seat cap; specific price TBD — see Open Question 1]

**Intelligence Tier:** Adds Intelligence Dashboard and Departure Risk Signals. Priced as an upgrade above Base. Scoped to designated leadership Members (typically 2–5 seats per organization).

**Land-and-expand motion:** Sell Base to the engineering org. Intelligence Tier upsell when the VP wants the departure risk dashboard — typically 30–90 days post-deployment once the Knowledge Graph has proven its query value.

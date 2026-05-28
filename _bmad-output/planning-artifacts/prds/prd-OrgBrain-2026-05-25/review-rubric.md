# PRD Quality Review — OrgBrain

## Overall verdict

This is a strong PRD. The thesis is clear and honest (passive ingestion as the core differentiator, two-tier value architecture), scope decisions are explicit and defensible, and the Privacy & Disclosure section is unusually substantive — it names the real constraints rather than gesturing at them. The main risks before handoff: UJ-3 numbering gap will confuse downstream readers, and Open Question 4 (Slack bot channel model) is architecture-blocking and should be resolved before epics. FR-3 and FR-7 each need one measurable bound added.

---

## Decision-readiness — strong

Trade-offs are named as trade-offs. The manager layer deferral (V2), the DM exclusion (hard limit, not configurable), and the Microsoft 365 decision (V2 maybe) are all stated as decisions with rationale, not buried as "considerations." Open Questions are genuinely open — pricing and confidence threshold value are correctly held as beta-calibration items rather than answered in the next paragraph. The NOTE FOR PM callouts in FR-4 and §6.2 are at real tensions.

### Findings
- **high** Open Question 4 unresolved before architecture (§8) — The Slack bot channel model (shared channel vs. per-Member DM bot) is listed as an open question, but it is architecture-blocking: it affects how Member queries are routed, how Pinned Content is delivered to new Members, and how the Slack bot surface is implemented. This needs resolution before the architecture pass, not after. *Fix:* Escalate to phase-blocker; resolve in this session or mark as the first decision in the architecture document.

---

## Substance over theater — strong

Vision is specific to OrgBrain — the passive ingestion thesis and the "nobody decided to document" framing are not portable to another product in this category. Personas (4) each drove at least one concrete decision: IC drove the Slack-native zero-friction requirement (FR-8), New Hire drove Pinned Content (FR-11), VP drove the Intelligence Dashboard (FR-13), and the Buyer persona grounded the go-to-market in a concrete escalation path (Senior EM → VP). NFRs in §11 carry product-specific thresholds (150 Members, 3 years of data, 99.5% uptime, 10-second query response) rather than generic boilerplate.

### Findings
- **low** Security NFR is partially boilerplate (§11) — "All organizational data encrypted at rest and in transit" is standard. The tenant isolation guarantee ("No OrgBrain employee has read access without explicit customer authorization") is product-specific and strong. The encryption line adds little without specifying the threat model (multi-tenant isolation, OAuth token storage). *Fix:* Either add one OrgBrain-specific bound (e.g., "OAuth tokens stored in a dedicated secrets manager, never in the Knowledge Graph store") or drop the generic encryption line and keep only the tenant isolation statement.

---

## Strategic coherence — strong

The PRD bets on a specific thesis: most organizational knowledge is never documented, and the only way to capture it is passive ingestion. Every feature cluster follows from that bet — ingestion is the foundation, the Knowledge Graph is the memory, the Query Interface is the value delivery, and the Intelligence Dashboard is the premium monetization layer. Success Metrics validate the thesis (SM-1 query speed, SM-2 new hire ramp, SM-3 departure signal lead time) rather than measuring activity. Counter-metrics are present and non-trivial — SM-C1 (signal gaming) and SM-C2 (answer rate vs. accuracy) name real optimization failure modes.

### Findings
None.

---

## Done-ness clarity — adequate

Most FRs have at least one testable consequence. FR-1 (15 minutes), FR-4 (background job + notification), FR-13 (3+ clusters, 30-day window), FR-15 (named card elements), FR-18 (logged with timestamp and identity) are all verifiable. Two FRs need bounds added.

### Findings
- **medium** FR-3 has no re-processing latency or frequency bound (§4.1) — "re-processed when a document is updated" is not testable. A document could be "re-processed" in 5 seconds or 24 hours; an engineer has no definition of done. *Fix:* Add a consequence: "Updated documents are re-processed and Knowledge Nodes updated within [X hours] of the change being detected." Align with A-2 (60-minute Meet transcript bound) for consistency or set a separate bound.
- **medium** FR-7 "continuously" is not testable (§4.2) — "Ownership is updated continuously as activity patterns evolve" has no measurable bound. *Fix:* Replace "continuously" with a specific cadence: e.g., "Ownership assignments are recomputed at least every 24 hours and within 60 minutes of a significant activity pattern change."
- **low** FR-8 "session" is undefined (§4.3) — "Conversational context is maintained within the session" — what terminates a session? Time-based idle timeout? Browser close? Explicit end? *Fix:* Add to Glossary or add one consequence: e.g., "A session expires after 30 minutes of inactivity; a new query after expiry starts without prior context." [ASSUMPTION: A-4 acknowledges this is inferred — needs confirmation.]

---

## Scope honesty — strong

Non-Goals section does real work — the "not a productivity monitoring tool" and "does not ingest DMs in any configuration, in any tier" statements are load-bearing constraints that will prevent scope creep at the story level. V2+ parking lot is detailed and credibly deferred. Assumptions are indexed and all appear inline. NOTE FOR PM callouts are at genuine tensions (FR-4 consent UX, §6.2 EM V2 timing).

### Findings
- **medium** EM persona V1 value is essentially the same as IC (§2.1) — The Manager/EM is listed as a persona, but their primary value (pre-meeting briefs) is V2. Their V1 value ("query interface") is identical to the IC's. If a reader asks "why is EM a separate persona for V1?" the PRD doesn't have a clear answer. *Fix:* Either (a) add one V1-specific EM behavior that differentiates them from IC (e.g., EMs use the query interface to prepare for 1:1s even without the automated brief — they query "what did Alex mention blocking them last week?"), or (b) collapse EM into IC for V1 and note EM re-emerges as distinct in V2.

---

## Downstream usability — adequate

Glossary has 15 defined terms used consistently throughout. FR/SM IDs are contiguous. UJ-persona linkage is correct for UJ-1, UJ-2, and UJ-4.

### Findings
- **medium** UJ-3 is missing from the numbering (§2.4) — UJ-3a and UJ-3b (EM pre-meeting briefs) were deferred to V2 during Discovery, but the numbering jumps from UJ-2 to UJ-4. Downstream readers (architect, story creator) will ask what UJ-3 is. *Fix:* Either renumber UJ-4 → UJ-3, or add a one-line placeholder: "UJ-3 (EM pre-meeting brief) — deferred to V2; see §6.2."
- **low** "Intelligence tier" case drift — §1 uses "Intelligence tier" (lowercase t); §2.1, §4.5, §10.1 use "Intelligence Tier" (capitalized). Glossary defines "Intelligence Tier" (capitalized). *Fix:* Find and replace to match Glossary definition throughout.

---

## Shape fit — strong

Multi-stakeholder B2B product with distinct UX per persona — Personas + Journeys entry was the right choice and the PRD uses it well. This is a chain-top PRD (feeds architecture → epics → stories), and downstream usability is appropriately treated as load-bearing. Length is right for a launch-intent product at pre-seed stage.

### Findings
None.

---

## Mechanical notes

- **Glossary drift:** "Intelligence Tier" vs. "Intelligence tier" — see downstream usability above. All other Glossary terms are capitalized consistently.
- **ID continuity:** FR-1 through FR-18 are contiguous ✓. SM-1 through SM-6 and SM-C1/SM-C2 are present ✓. UJ-3 gap — see finding above.
- **Assumptions Index roundtrip:** All 7 inline `[ASSUMPTION]` tags (A-1 through A-7) appear in §9 ✓. One additional assumption appears inline in §10.3 ("single cloud region in V1") — not indexed in §9. *Fix:* Add as A-8.
- **UJ persona linkage:** UJ-1 → IC ✓, UJ-2 → New Hire ✓, UJ-4 → VP of Engineering ✓.

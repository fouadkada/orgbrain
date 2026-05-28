---
title: OrgBrain Product Brief
status: draft
created: 2026-05-22
updated: 2026-05-22
---

# Product Brief: OrgBrain

## Executive Summary

Every organization leaks knowledge constantly. Decisions get made in Slack DMs nobody saved. A meeting goes unrecorded and the person who missed it spends three days reconstructing what was agreed. A key engineer leaves and takes five years of context with them. The company's tribal knowledge — the real operating system of how things actually work — lives nowhere, belongs to no one, and quietly evaporates.

OrgBrain is an AI-powered organizational memory that runs silently in the background, ingesting knowledge from where it actually lives: Slack channels, meeting transcripts, emails, documents. No one has to decide to document anything. When someone asks "why did we make this decision?" or "who owns this?" or "what did we agree in the March planning meeting?" — OrgBrain answers in seconds, with sources.

For leadership, OrgBrain goes further: it surfaces the signals hidden inside the organization's knowledge patterns — who is the single point of failure for critical systems, which teams are moving slowly, where strategy and execution have quietly diverged. It turns what was invisible organizational risk into a visible, manageable dashboard.

## The Problem

In every company, there is a gap between what the organization officially knows and what it actually knows. Official knowledge lives in wikis, docs, and project management tools — but only if someone decided to write it down. Actual knowledge lives in a Slack thread from eight months ago, in a decision made during a meeting nobody recorded, in the head of the one engineer who built that part of the system.

This gap has a concrete cost:

- **New hires take 3-6 months to reach productivity** because they are constantly hunting for context that nobody captured
- **Decisions get relitigated** because nobody can find the record of why a choice was made
- **Key people become single points of failure** — the organization's resilience is quietly held hostage to whether three specific people show up tomorrow
- **Executives make strategy based on assumption** because they cannot see where attention is actually going versus where they think it is

Existing tools do not solve this. Notion, Confluence, and Guru require humans to deliberately document things — and most valuable organizational knowledge never makes it to a wiki because nobody had the time or thought to do it. Enterprise search tools like Glean surface what was already written; they cannot capture what was never a document. Meeting tools like Fireflies capture individual meetings but build no cross-meeting, org-level picture.

The knowledge management market has assumed that documentation is a behavior that can be trained into organizations. OrgBrain assumes it cannot — and builds a system that captures knowledge whether or not anyone chooses to share it.

## The Solution

OrgBrain is installed once, by an IT administrator, and then operates silently. It connects to the organization's Slack workspace, calendar, meeting transcripts, and document repositories. From that point, it ingests continuously — not because anyone asked it to, but because it is always watching.

For organizations with existing history, OrgBrain provides an initial import — so the system is knowledgeable on day one, not just from the moment of installation forward.

What any employee gets: a conversational interface they can ask anything. *"What did we decide about the API versioning strategy?" "Who owns the billing integration?" "Why did we deprecate the mobile app?" "What does OODA mean in this company's context?"* Every answer comes with provenance — the source, the date, the confidence level. Knowledge with a chain of custody, not a magic answer from nowhere.

What managers and team leads get: a live picture of their team's knowledge health — who holds critical knowledge alone, which domains have gone stale, where onboarding gaps actually are. The pre-meeting brief, delivered 15 minutes before every calendar event, surfaces everything OrgBrain knows about who is in the room and what decisions involve them.

What executives get: organizational intelligence that did not previously exist. Which decisions are taking longest and where they are getting stuck. Whether the teams actually working on AI represent the 40% priority leadership announced, or the 4% the data reveals. Who the real influencers are in the organization versus who appears on the org chart. And — the signal executives pay most attention to — who is likely to leave before they do, based on behavioral patterns visible months before a resignation lands.

## What Makes This Different

**Passive ingestion is the core.** Every competitor assumes that someone will decide to document something. OrgBrain captures what nobody decided to document — which is most of what actually matters.

**Connects to everything you already use.** Confluence is built for Atlassian shops. Notion AI works best inside Notion. OrgBrain has no native document format to protect, so it connects to everything — Slack, Google Workspace, Zoom, Notion, Jira — without asking the organization to change how it works.

**Knowledge has a history, not just a timestamp.** When a decision is reversed, OrgBrain does not overwrite it. It keeps the original reasoning alongside the new direction — so when someone asks "why did we change our pricing model?" they get the full story, not just the current state. Think of it as version control for organizational memory.

**No one else has built the executive layer.** No scaled commercial product combines passive ingestion with org-health intelligence signals — bus factor, departure risk, decision velocity, shadow org chart. This is the premium tier, and the market gap is open.

## Who This Serves

**Primary buyer:** CTO, CEO, or Head of Operations at a 30–150 person company. This person is simultaneously the one who approves the tool, the one who installs it, and one of the people who benefits most directly from its outputs. No long procurement cycle. Decision timeline measured in days, not quarters.

**Daily users:** Everyone in the organization, in different ways. New hires use the onboarding companion to ask every question they were afraid to ask a senior colleague. Engineers use it to understand why a system was built a certain way. Managers use the pre-meeting brief and bus factor alerts. Executives use the intelligence dashboard.

## Scope: Version 1

**In:**
- Passive ingestion from Slack, calendar, and meeting transcripts
- Import feature for existing organizational history
- Conversational query interface with confidence scoring and source attribution
- Knowledge audit trail (decisions versioned, not overwritten)
- Pre-meeting brief delivered before every calendar event
- Bus factor dashboard for team leads
- Onboarding companion for new hires

## Vision

In three years, OrgBrain is the organizational nervous system — not a tool anyone opens intentionally, but infrastructure as assumed as email or Slack. The company that loses a key person and does not lose their knowledge. The executive team that can see, for the first time, whether the company is actually doing what it says it is doing.

A longer-term question worth holding: what if the knowledge employees generate followed them across their careers? Organizations contributing to a person's professional memory, but not owning it. That is a harder problem and a different product — but it is where the logic of OrgBrain eventually leads.

OrgBrain starts by making organizational memory survivable. It ends by making it portable.

---

*Based on brainstorming session 2026-05-22. 33 ideas across 5 product layers.*

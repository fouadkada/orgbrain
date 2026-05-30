package filter

// Must run before any LLM call. Rejects DMs (channel_type=im|mpim) with audit log event.
// Three-layer DM exclusion: Slack app scope restriction + this filter + runtime scope assertion.

// TODO: Implement IngestionFilter (Story 1.3).

package slack

// TODO: Reconciliation job — polls conversations.history API on configurable interval (default 30m, 1h lookback).
// Idempotency key: (workspace_id, channel_id, message_ts, thread_ts).

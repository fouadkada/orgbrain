package store

// TODO: Sliding window rate limit counter in Postgres — 30 queries/hour on query endpoint, 100 req/min on admin.
// Uses SKIP LOCKED to avoid contention. No Redis.

package store

// ALL tenant storage reads and writes go through this adapter. No raw db.Query() in handlers or workers.
// KnowledgeStoreAdapter sets SET LOCAL search_path = org_{id} at the start of every transaction.
// Enforces sensitivity_tier at query layer before SQL executes; raises SensitivityTierViolation for unauthorized access.

// TODO: Implement KnowledgeStoreAdapter interface and pgx-backed implementation (Story 1.3).

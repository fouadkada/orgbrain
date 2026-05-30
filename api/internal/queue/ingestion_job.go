package queue

// TODO: River ingestion job definition — filter → embed → extract → write pipeline with per-stage checkpointing.
// On failure: re-enqueue at failed stage only, max 3 retries exponential backoff, then dead-letter.

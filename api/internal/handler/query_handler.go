package handler

// TODO: Query handler — receives natural language query, calls rag service, streams SSE response.
// SSE event order: meta (confidence + sources) → data (tokens) → error (on timeout/fallback).

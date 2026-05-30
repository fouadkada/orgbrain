package slack

// TODO: Runtime scope assertion — ingestion worker startup fails hard if DM-related scope present.
// Scopes checked: no im:read, no mpim:read. Assert on worker startup, not webhook receive.

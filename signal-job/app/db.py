import psycopg2

# SET LOCAL search_path = org_{id} per transaction — never session-level.
# Uses psycopg2 (synchronous) — do NOT use asyncpg here; signal-job is a blocking cron script.

# TODO: psycopg2 connection and tenant transaction context (Story 1.3).

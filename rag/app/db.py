import asyncpg

# SET LOCAL search_path = org_{id} per transaction — never session-level.
# PgBouncer transaction-mode requirement: session-level search_path causes tenant data leakage.

# TODO: asyncpg connection pool and tenant transaction context (Story 1.3).

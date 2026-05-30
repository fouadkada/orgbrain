package store

// TODO: Tenant provisioning — CREATE SCHEMA org_{id}, apply tenant migrations, flip org_status to 'active'.
// SET LOCAL search_path = org_{id} inside every transaction, never outside (PgBouncer transaction mode).

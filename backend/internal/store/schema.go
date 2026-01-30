package store

// createGatewaysTableSQL is the DDL for the gateways table.
// Compatible with both PostgreSQL and SQLite.
const createGatewaysTableSQL = `
CREATE TABLE IF NOT EXISTS gateways (
    id               TEXT PRIMARY KEY,
    name             TEXT NOT NULL,
    description      TEXT NOT NULL DEFAULT '',
    endpoint         TEXT NOT NULL,
    transport_type   TEXT NOT NULL DEFAULT 'https',
    transport_params TEXT NOT NULL DEFAULT '{}',
    auth_type        TEXT NOT NULL DEFAULT '',
    auth_params      TEXT NOT NULL DEFAULT '{}',
    auth_secret_ref  TEXT NOT NULL DEFAULT '',
    status           TEXT NOT NULL DEFAULT 'unknown',
    labels           TEXT NOT NULL DEFAULT '{}',
    enrolled_at      TIMESTAMP NOT NULL,
    last_seen_at     TIMESTAMP,
    ttl_seconds      INTEGER
)`

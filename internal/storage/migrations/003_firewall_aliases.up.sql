CREATE TABLE firewall_ip_whitelist_entries (
    id SERIAL PRIMARY KEY,

    owner_iss TEXT NOT NULL,
    owner_sub TEXT NOT NULL,

    alias_name TEXT NOT NULL,
    alias_uuid UUID NOT NULL,

    ip_address INET NOT NULL,
    ip_version INTEGER GENERATED ALWAYS AS (family(ip_address)) STORED,

    description TEXT,

    status TEXT NOT NULL DEFAULT 'requested',

    requested_at TIMESTAMP NOT NULL DEFAULT NOW(),
    added_at TIMESTAMP,
    removed_at TIMESTAMP,
    expires_at TIMESTAMP,

    removed_by_iss TEXT,
    removed_by_sub TEXT,
    removal_reason TEXT,

    CONSTRAINT whitelist_owner_not_empty CHECK (owner_iss != '' AND owner_sub != ''),
    CONSTRAINT whitelist_remover_not_empty CHECK (
        (removed_by_iss IS NULL AND removed_by_sub IS NULL) OR
        (removed_by_iss != '' AND removed_by_sub != '')
    ),
    CONSTRAINT valid_status CHECK (status IN ('requested', 'added', 'removed', 'removed_by_admin', 'blacklisted_by_admin')),

    UNIQUE(alias_name, ip_address),

    CONSTRAINT valid_expiration CHECK (expires_at IS NULL OR expires_at > requested_at),
    CONSTRAINT valid_added_at CHECK (added_at IS NULL OR added_at >= requested_at),
    CONSTRAINT valid_removed_at CHECK (removed_at IS NULL OR removed_at >= requested_at)
);

CREATE INDEX idx_whitelist_owner ON firewall_ip_whitelist_entries(owner_iss, owner_sub);
CREATE INDEX idx_whitelist_alias_name ON firewall_ip_whitelist_entries(alias_name);
CREATE INDEX idx_whitelist_alias_uuid ON firewall_ip_whitelist_entries(alias_uuid);
CREATE INDEX idx_whitelist_status ON firewall_ip_whitelist_entries(status);
CREATE INDEX idx_whitelist_active_ips ON firewall_ip_whitelist_entries(alias_name, status) WHERE status = 'added';
CREATE INDEX idx_whitelist_pending_ips ON firewall_ip_whitelist_entries(alias_name, status) WHERE status = 'requested';
CREATE INDEX idx_whitelist_removed_ips ON firewall_ip_whitelist_entries(alias_name, status) WHERE status = 'removed' OR status = 'removed_by_admin';
CREATE INDEX idx_whitelist_expires ON firewall_ip_whitelist_entries(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_whitelist_ip_address ON firewall_ip_whitelist_entries(ip_address);
CREATE INDEX idx_whitelist_banned_ips ON firewall_ip_whitelist_entries(alias_name, ip_address, status) WHERE status = 'blacklisted_by_admin';

CREATE TABLE firewall_whitelist_events (
    id SERIAL PRIMARY KEY,
    whitelist_id INTEGER NOT NULL,

    actor_iss TEXT NOT NULL,
    actor_sub TEXT NOT NULL,

    event_type TEXT NOT NULL,
    notes TEXT,

    client_ip INET,
    user_agent TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    FOREIGN KEY (whitelist_id) REFERENCES firewall_ip_whitelist_entries(id) ON DELETE CASCADE,

    CONSTRAINT event_actor_not_empty CHECK (actor_iss != '' AND actor_sub != ''),
    CONSTRAINT valid_event_type CHECK (event_type IN (
    'requested', 'added', 'removed', 'removed_by_admin', 'blacklisted_by_admin', 'expired', 'sync_failed'
    ))
);

CREATE INDEX idx_whitelist_events_whitelist ON firewall_whitelist_events(whitelist_id);
CREATE INDEX idx_whitelist_events_actor ON firewall_whitelist_events(actor_iss, actor_sub);
CREATE INDEX idx_whitelist_events_type ON firewall_whitelist_events(event_type);
CREATE INDEX idx_whitelist_events_created ON firewall_whitelist_events(created_at);
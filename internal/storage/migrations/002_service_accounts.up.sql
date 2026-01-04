CREATE TABLE service_accounts(
    iss TEXT NOT NULL,
    sub TEXT NOT NULL,
    name TEXT NOT NULL,

    lookup_id TEXT NOT NULL UNIQUE,
    token_hash TEXT NOT NULL,
    token_expires_at TEXT,

    is_disabled BOOLEAN NOT NULL DEFAULT FALSE,

    created_by_iss TEXT NOT NULL,
    created_by_sub TEXT NOT NULL,
    created_at TIMESTAMP,

    PRIMARY KEY (iss, sub),
    CONSTRAINT lookup_id_length CHECK (char_length(lookup_id) >= 16)
);

CREATE UNIQUE INDEX idx_service_accounts_token_hash ON service_accounts(token_hash);
CREATE UNIQUE INDEX idx_service_accounts_owner ON service_accounts(created_by_iss, created_by_sub);
CREATE UNIQUE INDEX idx_service_accounts_lookup_id ON service_accounts(lookup_id);

CREATE TABLE service_account_scopes(
    owner_iss TEXT NOT NULL,
    owner_sub TEXT NOT NULL,
    scope_name TEXT NOT NULL,

    PRIMARY KEY (owner_iss, owner_sub, scope_name),
    FOREIGN KEY (owner_iss, owner_sub) REFERENCES service_accounts(iss, sub) ON DELETE CASCADE;
)

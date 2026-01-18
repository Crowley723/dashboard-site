CREATE TABLE service_accounts(
    id SERIAL PRIMARY KEY,
    iss TEXT NOT NULL,
    sub TEXT NOT NULL,
    name TEXT NOT NULL,

    lookup_id TEXT NOT NULL UNIQUE,
    token_hash TEXT NOT NULL,
    token_expires_at TIMESTAMP,

    is_disabled BOOLEAN NOT NULL DEFAULT FALSE,
    deleted_at TIMESTAMP,

    created_by_iss TEXT NOT NULL,
    created_by_sub TEXT NOT NULL,
    created_at TIMESTAMP,

    CONSTRAINT lookup_id_length CHECK (char_length(lookup_id) >= 16)
);

CREATE UNIQUE INDEX idx_service_accounts_iss_sub ON service_accounts(iss, sub);
CREATE INDEX idx_service_accounts_token_hash ON service_accounts(token_hash);
CREATE INDEX idx_service_accounts_owner ON service_accounts(created_by_iss, created_by_sub);
CREATE INDEX idx_service_accounts_lookup_id ON service_accounts(lookup_id);
CREATE INDEX idx_service_accounts_deleted_at ON service_accounts(deleted_at);

CREATE TABLE service_account_scopes(
    owner_iss TEXT NOT NULL,
    owner_sub TEXT NOT NULL,
    scope_name TEXT NOT NULL,

    PRIMARY KEY (owner_iss, owner_sub, scope_name),
    FOREIGN KEY (owner_iss, owner_sub) REFERENCES service_accounts(iss, sub) ON DELETE CASCADE
);

-- Remove foreign key constraints that only allow users, to support both users and service accounts
ALTER TABLE certificate_requests
DROP CONSTRAINT certificate_requests_owner_iss_owner_sub_fkey;

ALTER TABLE certificate_requests
ADD CONSTRAINT certificate_requests_owner_not_empty
CHECK (owner_iss != '' AND owner_sub != '');

ALTER TABLE certificate_downloads
DROP CONSTRAINT certificate_downloads_downloader_iss_downloader_sub_fkey;

ALTER TABLE certificate_downloads
ADD CONSTRAINT certificate_downloads_downloader_not_empty
CHECK (downloader_iss != '' AND downloader_sub != '');

ALTER TABLE certificate_events
DROP CONSTRAINT certificate_events_reviewer_iss_reviewer_sub_fkey;

ALTER TABLE certificate_events
DROP CONSTRAINT certificate_events_requester_iss_requester_sub_fkey;

ALTER TABLE certificate_events
ADD CONSTRAINT certificate_events_reviewer_not_empty
CHECK (reviewer_iss != '' AND reviewer_sub != '');

ALTER TABLE certificate_events
ADD CONSTRAINT certificate_events_requester_not_empty
CHECK (requester_iss != '' AND requester_sub != '');

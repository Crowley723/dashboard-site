-- Restore foreign key constraints (will fail if service account references exist)
ALTER TABLE certificate_events
DROP CONSTRAINT IF EXISTS certificate_events_requester_not_empty;

ALTER TABLE certificate_events
DROP CONSTRAINT IF EXISTS certificate_events_reviewer_not_empty;

ALTER TABLE certificate_events
ADD CONSTRAINT certificate_events_requester_iss_requester_sub_fkey
FOREIGN KEY (requester_iss, requester_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT;

ALTER TABLE certificate_events
ADD CONSTRAINT certificate_events_reviewer_iss_reviewer_sub_fkey
FOREIGN KEY (reviewer_iss, reviewer_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT;

ALTER TABLE certificate_downloads
DROP CONSTRAINT IF EXISTS certificate_downloads_downloader_not_empty;

ALTER TABLE certificate_downloads
ADD CONSTRAINT certificate_downloads_downloader_iss_downloader_sub_fkey
FOREIGN KEY (downloader_iss, downloader_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT;

ALTER TABLE certificate_requests
DROP CONSTRAINT IF EXISTS certificate_requests_owner_not_empty;

ALTER TABLE certificate_requests
ADD CONSTRAINT certificate_requests_owner_iss_owner_sub_fkey
FOREIGN KEY (owner_iss, owner_sub) REFERENCES users(iss, sub) ON DELETE RESTRICT;

-- Drop service account tables
DROP INDEX IF EXISTS idx_service_accounts_deleted_at;
DROP INDEX IF EXISTS idx_service_accounts_iss_sub;
DROP INDEX IF EXISTS idx_service_accounts_lookup_id;
DROP INDEX IF EXISTS idx_service_accounts_owner;
DROP INDEX IF EXISTS idx_service_accounts_token_hash;

DROP TABLE IF EXISTS service_account_scopes;
DROP TABLE IF EXISTS service_accounts;
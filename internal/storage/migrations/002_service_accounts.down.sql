DROP INDEX IF EXISTS idx_service_accounts_iss_sub;
DROP INDEX IF EXISTS idx_service_accounts_lookup_id;
DROP INDEX IF EXISTS idx_service_accounts_owner;
DROP INDEX IF EXISTS idx_service_accounts_token_hash;

DROP TABLE IF EXISTS service_account_scopes;
DROP TABLE IF EXISTS service_accounts;
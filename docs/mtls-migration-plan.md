# mTLS Service Migration Plan

This document describes the plan to extract the mTLS certificate management system,
OIDC authentication, service accounts, and scope-based authorization from
`dashboard-site` into a new standalone repository.

## Overview

The new repository (referred to below as **`conduit`**) will be an independent service
responsible for:

- OIDC-based user authentication
- Service account management (token-based API authentication)
- Scope-based authorization
- mTLS certificate lifecycle management (request → approve → issue → download)
- Certificate provider backends: database (self-signed CA) and Kubernetes (cert-manager)

`dashboard-site` will retain:

- Prometheus/Mimir metrics dashboard
- Firewall IP whitelist management (OPNsense integration)
- Its own authentication (OIDC session + service account tokens), since the firewall
  feature requires access control

> **Decision point**: Because the firewall feature also requires auth, service accounts,
> and scopes, the auth/service-account infrastructure will be **duplicated** rather than
> removed from `dashboard-site`. An alternative is to have `dashboard-site` delegate auth
> to `conduit` via a shared OIDC provider (both apps share the same IdP), which is the
> recommended long-term approach. See [Auth Coupling](#auth-coupling) below.

---

## Files to Move to `conduit`

### Certificate Providers (core reason for the new service)

| Source Path | Destination |
|---|---|
| `internal/services/certificate/certificate_provider.go` | `internal/services/certificate/certificate_provider.go` |
| `internal/services/certificate/database_certificate_provider.go` | `internal/services/certificate/database_certificate_provider.go` |
| `internal/services/certificate/kubernetes_certificate_provider.go` | `internal/services/certificate/kubernetes_certificate_provider.go` |
| `internal/services/certificate/utils.go` (crypto helpers) | `internal/services/certificate/utils.go` |

### Certificate Handlers

| Source Path | Destination |
|---|---|
| `internal/handlers/handler_certificate_request.go` | `internal/handlers/handler_certificate_request.go` |
| `internal/handlers/handler_certificate_download.go` | `internal/handlers/handler_certificate_download.go` |

### Certificate Background Jobs

| Source Path | Destination |
|---|---|
| `internal/jobs/certificate_creation_job.go` | `internal/jobs/certificate_creation_job.go` |
| `internal/jobs/certificate_status_job.go` | `internal/jobs/certificate_status_job.go` |

### Authentication

| Source Path | Destination |
|---|---|
| `internal/authentication/oidc.go` | `internal/authentication/oidc.go` |
| `internal/authentication/session.go` | `internal/authentication/session.go` |
| `internal/handlers/handler_oidc_login.go` | `internal/handlers/handler_oidc_login.go` |
| `internal/handlers/handler_oidc_callback.go` | `internal/handlers/handler_oidc_callback.go` |
| `internal/handlers/handler_logout.go` | `internal/handlers/handler_logout.go` |
| `internal/handlers/handler_auth_status.go` | `internal/handlers/handler_auth_status.go` |

### Service Accounts

| Source Path | Destination |
|---|---|
| `internal/handlers/handler_service_accounts.go` | `internal/handlers/handler_service_accounts.go` |
| `internal/models/service_account.go` | `internal/models/service_account.go` |
| `internal/storage/service_account_queries.go` | `internal/storage/service_account_queries.go` |

### Authorization

| Source Path | Destination |
|---|---|
| `internal/authorization/const.go` | `internal/authorization/const.go` |

Only the `mtls:*` scopes are strictly needed in `conduit`. The `firewall:*` scopes stay
in `dashboard-site`. If a single `authorization` package is shared via a Go module, the
whole file moves to `conduit` and `dashboard-site` defines its own firewall scopes locally.

### Models

| Source Path | Destination | Notes |
|---|---|---|
| `internal/models/user.go` | `internal/models/user.go` | Keep copy in both repos |
| `internal/models/certificate.go` | `internal/models/certificate.go` | Move entirely |

### Storage / Database

| Source Path | Destination |
|---|---|
| `internal/storage/certificate_queries.go` | `internal/storage/certificate_queries.go` |
| `internal/storage/migrations/001_initial_schema.up.sql` | `internal/storage/migrations/001_initial_schema.up.sql` |
| `internal/storage/migrations/002_service_accounts.up.sql` | `internal/storage/migrations/002_service_accounts.up.sql` |
| `internal/storage/migrations/004_certificate_provider_abstraction.up.sql` | `internal/storage/migrations/004_certificate_provider_abstraction.up.sql` |

The **storage interface** (`internal/storage/storage.go`) and **database connection**
(`internal/storage/database.go`) must be copied and then trimmed to only the methods
`conduit` requires.

### Shared Infrastructure to Copy (not remove from `dashboard-site`)

| Source Path | Notes |
|---|---|
| `internal/middlewares/require_auth.go` | Copy; both services need auth middleware |
| `internal/middlewares/principal.go` | Copy; both services use the Principal interface |
| `internal/middlewares/context.go` | Copy; trim to conduit's AppContext fields |
| `internal/middlewares/oidc_provider.go` | Copy |
| `internal/middlewares/session_provider.go` | Copy |
| `internal/config/config.go` | Copy; trim to conduit's config schema |
| `internal/config/schema.go` | Copy; trim to conduit's config fields |
| `internal/utils/` | Copy all crypto / token helpers |
| `internal/jobs/job_manager.go` | Copy |
| `internal/distributed/` | Copy (Redis leader election) |
| `internal/metrics/` | Copy |

---

## New Repository Structure

```
conduit/
├── main.go
├── go.mod                        # module: github.com/crowley723/conduit
├── config.yaml.template
├── Dockerfile
├── helm/
├── internal/
│   ├── authentication/
│   │   ├── oidc.go
│   │   └── session.go
│   ├── authorization/
│   │   └── const.go              # mtls:* scopes only
│   ├── config/
│   │   ├── config.go
│   │   └── schema.go             # trimmed: storage, oidc, mtls, auth, session, redis
│   ├── distributed/              # Redis leader election (unchanged)
│   ├── handlers/
│   │   ├── handler_auth_status.go
│   │   ├── handler_oidc_login.go
│   │   ├── handler_oidc_callback.go
│   │   ├── handler_logout.go
│   │   ├── handler_certificate_request.go
│   │   ├── handler_certificate_download.go
│   │   ├── handler_service_accounts.go
│   │   └── handler_health.go
│   ├── jobs/
│   │   ├── job_manager.go
│   │   ├── certificate_creation_job.go
│   │   └── certificate_status_job.go
│   ├── middlewares/
│   │   ├── context.go
│   │   ├── principal.go
│   │   ├── require_auth.go
│   │   ├── oidc_provider.go
│   │   └── session_provider.go
│   ├── metrics/
│   ├── models/
│   │   ├── user.go
│   │   ├── certificate.go
│   │   └── service_account.go
│   ├── server/
│   │   ├── server.go
│   │   └── handlers.go
│   ├── services/
│   │   └── certificate/
│   │       ├── certificate_provider.go
│   │       ├── database_certificate_provider.go
│   │       ├── kubernetes_certificate_provider.go
│   │       └── utils.go
│   ├── storage/
│   │   ├── storage.go            # trimmed interface
│   │   ├── database.go
│   │   ├── migrations/
│   │   │   ├── 001_initial_schema.up.sql
│   │   │   ├── 002_service_accounts.up.sql
│   │   │   └── 003_certificate_provider_abstraction.up.sql
│   │   ├── user_queries.go
│   │   ├── certificate_queries.go
│   │   └── service_account_queries.go
│   ├── utils/
│   └── version/
└── web/                          # New React frontend (certificate UI only)
    └── src/
        └── routes/
            ├── index.tsx         # Certificate request list
            ├── certificates/
            └── settings/
                └── service-accounts/
```

---

## Database Strategy

`conduit` needs its own PostgreSQL database (or schema). **Do not share a database
between the two services.** Both services will have independent user tables seeded from
the same OIDC provider.

### `conduit` database tables

```
users
user_groups
service_accounts
service_account_scopes
certificate_requests
certificate_downloads
certificate_events
certificate_authority        (database cert provider)
issued_certificates          (database cert provider)
encryption                   (encryption key validation)
```

### `dashboard-site` database tables (unchanged)

```
users
user_groups
service_accounts             (keep a copy — needed for firewall API access)
service_account_scopes
firewall_ip_whitelist_entries
firewall_whitelist_events
```

### Migration renumbering in `conduit`

The migrations need to be renumbered since migration 003 in `dashboard-site` is the
firewall (which does not move):

| conduit migration | Source migration |
|---|---|
| `001_initial_schema` | `001_initial_schema` (users, cert requests, cert events) |
| `002_service_accounts` | `002_service_accounts` |
| `003_certificate_provider_abstraction` | `004_certificate_provider_abstraction` |

---

## Config Schema for `conduit`

The new service needs a trimmed configuration. Remove all firewall and data/Prometheus
config. Keep:

```yaml
server:
  port: 8080
  external_url: "https://conduit.example.com"

log_level: info

storage:
  enabled: true
  host: "postgres"
  port: 5432
  database: "conduit"
  user: "conduit"
  password: "..."
  encryption_key: "base64-32-bytes"

session:
  duration: 24h
  redis:
    enabled: false
    address: "redis:6379"

oidc:
  issuer_url: "https://sso.example.com"
  client_id: "conduit"
  client_secret: "..."
  redirect_url: "https://conduit.example.com/api/auth/callback"

authorization:
  group_scopes:
    "conduit:mtls:admin":
      - mtls:request
      - mtls:read_all
      - mtls:approve
      - mtls:auto_approve
      - mtls:self_approve_certs
      - mtls:renew
      - mtls:revoke
      - mtls:download_all
    "conduit:mtls:user":
      - mtls:request
      - mtls:read
      - mtls:renew
      - mtls:download

features:
  mtls_management:
    enabled: true
    download_token_hmac_key: "base64-key"
    auto_approve_admin_requests: false
    allow_admins_to_approve_own_requests: true
    min_certificate_validity_days: 30
    max_certificate_validity_days: 365
    certificate_subject:
      organization: "Homelab"
      country: "US"
    background_job_config:
      approved_certificate_polling_interval: 30s
      issued_certificate_polling_interval: 30s
    # Choose one provider:
    database:
      enabled: true
      key_algorithm: "ECDSA-P256"
    # OR:
    kubernetes:
      enabled: false
      in_cluster: true
      namespace: "conduit"
      issuer:
        name: "letsencrypt-prod"
        kind: "ClusterIssuer"

distributed:
  enabled: false
  redis:
    address: "redis:6379"
```

---

## Auth Coupling

Both services share the same OIDC identity provider (IdP). Each service independently:

1. Redirects unauthenticated users to the IdP for login
2. Receives the ID token on callback and reads `sub`, `iss`, `groups` claims
3. Maps groups to its own scopes via its own `authorization.group_scopes` config

This means the user logs in separately to each service (two cookies, two sessions).
There is no single sign-out across services unless the IdP handles it via
front-channel logout.

**No cross-service API calls are required for auth.** Each service is self-contained.

---

## HTTP API Surface of `conduit`

```
POST   /api/auth/login
GET    /api/auth/callback
POST   /api/auth/logout
GET    /api/auth/status

GET    /api/certificates                     # list own requests
POST   /api/certificates                     # submit new request
GET    /api/certificates/:id                 # get request details
POST   /api/certificates/:id/review          # approve or reject (admin)
GET    /api/certificates/:id/download        # download certificate bundle
DELETE /api/certificates/:id                 # revoke

GET    /api/service-accounts                 # list own service accounts
POST   /api/service-accounts                 # create service account
DELETE /api/service-accounts/:id             # delete service account
POST   /api/service-accounts/:id/pause       # pause
POST   /api/service-accounts/:id/unpause     # unpause

GET    /api/v1/health
```

---

## Files That Stay in `dashboard-site`

The following files are **not** moved. They remain in `dashboard-site` as-is:

- `internal/data/` — Prometheus client and caching
- `internal/services/firewall/` — OPNsense router client
- `internal/handlers/handler_firewall_aliases.go`
- `internal/handlers/handler_data.go`
- `internal/jobs/data_fetch_job.go`
- `internal/jobs/firewall_sync_job.go`
- `internal/jobs/firewall_expiration_job.go`
- `internal/storage/migrations/003_firewall_aliases.up.sql`
- `internal/storage/firewall_ip_whitelist_queries.go`
- `internal/models/ip_whitelist.go`
- `web/src/` (all frontend; a new minimal frontend will be built for conduit)

The authentication stack (`internal/authentication/`, `internal/middlewares/`) is
**kept** in `dashboard-site` (not removed), because the firewall feature also requires
auth. A copy of the relevant files goes to `conduit`.

The certificate-related code (`internal/services/certificate/`, certificate handlers,
certificate jobs, `internal/storage/certificate_queries.go`) is **removed** from
`dashboard-site` after `conduit` is live and validated.

---

## Changes Required in `dashboard-site` After Migration

1. **Remove mTLS config block** from `config.yaml` and `config.docker.yaml.template`
2. **Remove certificate routes** from `internal/server/handlers.go`
3. **Remove** `CertificateManager` field from `AppContext` in `internal/middlewares/context.go`
4. **Remove** certificate provider initialization from `internal/server/server.go`
5. **Remove** `CertificateCreationJob` and `CertificateIssuedStatusJob` from job registration
6. **Remove** `mtls:*` scopes from `internal/authorization/const.go`
7. **Drop** mTLS-only database tables (or leave them; they will be ignored without code references):
   - `certificate_requests`, `certificate_events`, `certificate_downloads`
   - `certificate_authority`, `issued_certificates`, `encryption`
8. **Remove** now-unused Go dependencies:
   - `github.com/cert-manager/cert-manager`
   - `k8s.io/client-go`
   - `software.sslmate.com/src/go-pkcs12`
9. **Remove** frontend certificate pages from `web/src/routes/settings/certificates/`

---

## External Dependencies for `conduit`

All of these are already in the `dashboard-site` `go.mod` and carry over to `conduit`:

```
github.com/alexedwards/scs/v2           # sessions
github.com/alexedwards/scs/goredisstore # Redis sessions
github.com/coreos/go-oidc/v3            # OIDC
github.com/go-chi/chi/v5                # HTTP router
github.com/go-chi/cors                  # CORS
github.com/jackc/pgx/v5                 # PostgreSQL
github.com/redis/go-redis/v9            # Redis
github.com/cert-manager/cert-manager    # K8s cert-manager CRDs
k8s.io/client-go                        # Kubernetes client
golang.org/x/oauth2                     # OAuth2
github.com/go-crypt/crypt               # Argon2
software.sslmate.com/src/go-pkcs12      # PKCS12
github.com/google/uuid
gopkg.in/yaml.v3
github.com/prometheus/client_golang     # metrics endpoint
```

---

## Helm / Deployment Changes

`conduit` will need its own Helm chart. The existing `helm/` directory in `dashboard-site`
can be used as a template. Key additions:

- `Secret` for `encryption_key`, OIDC credentials, HMAC key
- `ServiceAccount` + RBAC rules if using the Kubernetes cert provider (needs access to
  `Certificate` CRDs and `Secret` resources in the target namespace)
- `CronJob` or init container for database migrations
- A separate PostgreSQL instance (or a new database within an existing cluster)
- If using the Kubernetes cert provider, the `conduit` pod needs a `ClusterRole` with:
  ```yaml
  rules:
    - apiGroups: ["cert-manager.io"]
      resources: ["certificates"]
      verbs: ["create", "get", "delete"]
    - apiGroups: [""]
      resources: ["secrets"]
      verbs: ["get"]
  ```

---

## Suggested Migration Order

1. **Create the `conduit` repository** with the file structure above. Copy files;
   do not delete from `dashboard-site` yet.
2. **Wire up `conduit`**: write `main.go`, `server/server.go`, `server/handlers.go`
   using the copied components. Trim the storage interface to only the methods conduit uses.
3. **Write database migrations** (renumbered 001–003) and verify the schema boots cleanly.
4. **Build and test `conduit`** locally: OIDC login, certificate request → approval →
   issuance → download, service account CRUD.
5. **Deploy `conduit`** to staging alongside the existing `dashboard-site`. Both share
   the same IdP; the OIDC client is a new registration.
6. **Validate** the full certificate lifecycle in staging.
7. **Update `dashboard-site`**: remove cert routes, cert jobs, cert storage queries,
   cert frontend pages, and the K8s/cert-manager Go dependencies.
8. **Deploy updated `dashboard-site`** to staging; verify firewall and metrics still work.
9. **Promote both** to production. Run a DB migration script to move any existing
   `certificate_*` rows from the old database to the new `conduit` database if needed.
10. **Drop the mTLS tables** from the `dashboard-site` database after a burn-in period.

---

## Key Risks and Mitigations

| Risk | Mitigation |
|---|---|
| Existing certificate data is lost | Export and import cert rows before cutover; keep old tables read-only during transition |
| Users must log in twice (once per service) | Acceptable for a homelab; can be addressed later with OIDC front-channel logout or a shared session store |
| Service accounts used in both services diverge | Each service has independent service account tables; tokens are not cross-service |
| Kubernetes RBAC for cert-manager is tricky | Use a dedicated `ServiceAccount` in the `conduit` namespace with a scoped `ClusterRole` |
| `storage.Provider` interface is large (~100 methods) | Trim it aggressively in `conduit` — only include methods that certificate and service account queries actually call |

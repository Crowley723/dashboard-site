# Conduit Project Summary

This document summarizes all design decisions and planned work for the `conduit` project,
extracted from `dashboard-site`.

---

## What Is Conduit

`conduit` is a standalone SaaS-style web application spun out of `dashboard-site`. It
handles identity, authentication, and mTLS certificate lifecycle management for a homelab
environment. It is fully self-contained — it does not share a database, codebase, or
process with `dashboard-site`.

**`dashboard-site` retains:** Prometheus metrics dashboard, OPNsense firewall IP
whitelist management, and its own copy of the auth stack (needed for firewall access
control).

---

## Core Features

### 1. mTLS Certificate Management

Full certificate lifecycle:

```
request (user) → review (admin) → approved → [background job] → issued → downloaded → completed
```

**Certificate provider backends** (choose one per deployment):

| Backend | How it works |
|---|---|
| **Database** | App acts as its own CA. Generates ECDSA/RSA keypairs, signs certificates, stores encrypted PEMs in PostgreSQL. |
| **Kubernetes** | Creates `cert-manager` `Certificate` CRDs in a configured namespace. Polls until the Secret is populated, then reads the TLS data. |

**Certificate request fields:** common name, DNS SANs, organizational units, validity days.

**Approval workflow:**
- Requests land in `awaiting_review` status
- Principals with `mtls:approve` scope can approve/reject
- Principals with `mtls:auto_approve` skip the review queue
- `allow_admins_to_approve_own_requests` is a config toggle
- Approved requests are picked up by a background job every N seconds

**Download:** HMAC-signed download token; first download transitions to `completed`.

---

### 2. Authentication — Three Providers

#### OIDC (multi-provider)

Multiple OIDC providers can be configured simultaneously. The login portal renders a
button for each. Routes are provider-scoped:

```
GET /api/auth/login/:slug        initiate OIDC flow
GET /api/auth/callback/:slug     receive token
```

Config:
```yaml
auth:
  providers:
    - name: "Authentik"
      slug: "authentik"
      issuer_url: "https://sso.example.com/application/o/conduit/"
      client_id: "..."
      client_secret: "..."
    - name: "GitHub"
      slug: "github"
      issuer_url: "https://..."
      client_id: "..."
      client_secret: "..."
```

`redirect_url` is derived automatically from `server.external_url + /api/auth/callback/:slug`.
Each OIDC provider must have its callback URL updated in the IdP when migrating from the
single-provider setup.

#### Local (database) user provider

Users stored in PostgreSQL with Argon2id-hashed passwords. Auth flow is a direct
`POST /api/auth/login/local` with email + password — no OAuth redirect.

Local users are inserted into the shared `users` table with `iss = "conduit://local"` and
`sub = <uuid>`, so all existing foreign keys (`certificate_requests`, etc.) work unchanged.

Groups are stored in `local_user_groups` and map to scopes via the same
`authorization.group_scopes` config as OIDC users.

Config:
```yaml
auth:
  local:
    enabled: true
    allow_registration: false    # accounts created by admin only (v1)
```

Password reset in v1: CLI command (`conduit users reset-password <email>`), no email flow.

#### Service accounts

Token-based API authentication for programmatic access. Format:
`conduit_sa.<22-char-base64-lookupId>.<43-char-base64-secret>`

Tokens are scoped — a service account only holds the scopes explicitly granted at creation
time by the creating user (who must themselves hold those scopes).

---

### 3. Scope-Based Authorization

All authorization is expressed as scopes. Principals (users and service accounts) are
checked with `principal.HasScope(scope)` at the handler level.

**mTLS scopes:**

| Scope | Meaning |
|---|---|
| `mtls:request` | Submit a certificate request |
| `mtls:read` | Read own requests |
| `mtls:read_all` | Read all requests (admin) |
| `mtls:approve` | Approve or reject requests |
| `mtls:auto_approve` | Auto-approve own requests |
| `mtls:self_approve_certs` | Approve own requests (even without auto-approve) |
| `mtls:renew` | Renew a certificate |
| `mtls:revoke` | Revoke a certificate |
| `mtls:download` | Download own certificates |
| `mtls:download_all` | Download any certificate (admin) |

Groups → scopes mapping is configured in YAML:
```yaml
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
```

---

## Frontend: SaaS-Style Portal

The frontend is being rebuilt from the `dashboard-site` React/TypeScript base. Key
changes from the original:

- **Root route `/` is a login portal**, not a dashboard.
- **After login**, redirect to `/certs` (certificate list).
- **Provider discovery**: frontend calls `GET /api/auth/providers` on load to render the
  correct set of login buttons/form dynamically.
- No metrics dashboard, no firewall UI — those stay in `dashboard-site`.

**Planned route structure:**
```
/                        Login portal (public)
/certs                   Certificate request list (protected)
/certs/:id               Request detail + status (protected)
/service-accounts        Service account management (protected)
/settings/profile        User profile (protected)
```

**Components to keep from `dashboard-site`:**
- All `web/src/components/ui/` (Radix/shadcn primitives)
- `RequestCertificateDialog`, `RequestCertificateForm`, `DownloadCertificateDialog`
- `CreateServiceAccountDialog`
- `LoginDialog`, `LoginForm` (to be reworked into full-page portal)
- `UserDisplay`, `UserDropdown`, `header`, `app-sidebar` (trimmed)

**Components removed (firewall/metrics):**
- All `AddIPWhitelist*`, `Traefik*`, `Cluster*`, `NodeStatus*`, `PodUptime*`, `LineChartCard`

---

## Database Schema

`conduit` uses its own PostgreSQL database (not shared with `dashboard-site`).

### Tables

| Table | Purpose |
|---|---|
| `users` | Identity records for OIDC and local users |
| `user_groups` | Group membership (for OIDC users) |
| `local_users` | Local auth credentials (email, Argon2id hash) |
| `local_user_groups` | Group membership for local users |
| `service_accounts` | Service account metadata + token lookup |
| `service_account_scopes` | Scopes granted to each service account |
| `certificate_requests` | Full request lifecycle record |
| `certificate_downloads` | Download audit log |
| `certificate_events` | Approval/rejection audit trail |
| `certificate_authority` | Self-signed CA (database provider only) |
| `issued_certificates` | Issued cert PEMs, encrypted at rest |
| `encryption` | Encryption key validation record |

### Migrations (conduit numbering)

| Migration | Content |
|---|---|
| `001_initial_schema` | users, user_groups, certificate_requests, certificate_events, certificate_downloads |
| `002_service_accounts` | service_accounts, service_account_scopes |
| `003_certificate_provider_abstraction` | certificate_authority, issued_certificates, encryption |
| `004_local_users` *(new)* | local_users, local_user_groups |

---

## HTTP API

```
# Auth
GET  /api/auth/providers
GET  /api/auth/login/:slug          OIDC initiation
GET  /api/auth/callback/:slug       OIDC callback
POST /api/auth/login/local          Email + password
POST /api/auth/logout
GET  /api/auth/status

# Certificates
GET    /api/certificates
POST   /api/certificates
GET    /api/certificates/:id
POST   /api/certificates/:id/review
GET    /api/certificates/:id/download
DELETE /api/certificates/:id

# Service Accounts
GET    /api/service-accounts
POST   /api/service-accounts
DELETE /api/service-accounts/:id
POST   /api/service-accounts/:id/pause
POST   /api/service-accounts/:id/unpause

# System
GET    /api/v1/health
```

---

## Config Reference (conduit)

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
  encryption_key: "base64-32-bytes"   # AES-256 for cert PEMs at rest

session:
  duration: 24h
  redis:
    enabled: false
    address: "redis:6379"

auth:
  providers:
    - name: "Authentik"
      slug: "authentik"
      issuer_url: "https://sso.example.com/application/o/conduit/"
      client_id: "..."
      client_secret: "..."
  local:
    enabled: false
    allow_registration: false

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
    database:
      enabled: true
      key_algorithm: "ECDSA-P256"
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

## Deployment

- Own Helm chart (based on `dashboard-site` chart as template)
- Own PostgreSQL database
- If using Kubernetes cert provider, pod needs a `ClusterRole`:
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

## Open Questions

| Question | Status |
|---|---|
| Host under a GitHub organization or personal account? | Undecided |
| Password reset mechanism for local users (CLI vs email) | CLI preferred for v1 |
| Allow self-registration for local users? | Off by default, admin-created only |
| `/api/auth/providers` response shape — include icon URL? | TBD |
| Post-login redirect target (`/certs` vs configurable) | TBD |

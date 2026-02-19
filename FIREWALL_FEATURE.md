# Firewall IP Whitelisting Feature

## Overview

This feature allows users to manage IP addresses in firewall aliases (whitelists) through a web interface. It includes both backend API endpoints and a React frontend.

## Backend Components

### API Endpoints

**User Endpoints:**
- `GET /api/firewall/aliases` - Get available firewall aliases for the user
- `GET /api/firewall/entries` - Get user's whitelist entries
- `POST /api/firewall/entries` - Add an IP to whitelist
- `DELETE /api/firewall/entries/{id}` - Remove an IP from whitelist

**Admin Endpoints:**
- `DELETE /api/firewall/entries/{id}/blacklist` - Blacklist an IP (all duplicates)
  - Requires `firewall:blacklist` scope
  - Accepts optional `{"reason": "..."}` in request body

### Background Jobs

**FirewallSyncJob** (`internal/jobs/firewall_sync_job.go`):
- Runs every 5 minutes (configurable)
- Syncs database state to OPNsense firewall
- Smart IP removal: Only removes IP when ALL entries are inactive
- Marks pending IPs as "added" after successful sync
- Creates audit events on failures

**FirewallExpirationJob** (`internal/jobs/firewall_expiration_job.go`):
- Runs every 1 hour (configurable)
- Marks expired entries as "removed"
- Creates "expired" audit events

### Database Schema

**Tables:**
- `firewall_ip_whitelist_entries` - IP whitelist entries
- `firewall_whitelist_events` - Audit trail of all actions

**Entry Statuses:**
- `requested` - Pending addition to firewall
- `added` - Active in firewall
- `removed` - User removed
- `removed_by_admin` - Admin removed
- `blacklisted_by_admin` - Blacklisted (cannot be re-added)

### Configuration

```yaml
features:
  firewall_management:
    enabled: true
    router_endpoint: "https://opnsense.example.com"
    router_api_key: "your-api-key"
    router_api_secret: "your-api-secret"
    background_job_config:
      sync_interval: 5m        # Sync to firewall every 5 minutes
      expiration_interval: 1h   # Check for expired IPs every hour
    aliases:
      - name: "vpn_users"
        uuid: "12345678-1234-1234-1234-123456789abc"
        auth_group: "vpn-users-group"
        max_ips_per_user: 5
        max_total_ips: 100
        default_ttl: 720h  # 30 days
        description: "VPN user whitelist"
```

## Frontend Components

### Pages

**Firewall Whitelist** (`/settings/firewall`):
- View all user's whitelist entries
- Search and filter entries
- View entry details (IP, status, expiration, events)
- Remove active entries
- Expandable accordion UI similar to certificates page

**Add IP** (`/settings/firewall/add`):
- Form to add new IP addresses
- Select from available aliases
- Optional description
- Optional custom TTL (time-to-live)
- IPv4 and IPv6 support

### Components

**AddIPWhitelistForm** (`web/src/components/AddIPWhitelistForm.tsx`):
- Reusable form component
- IP validation (IPv4/IPv6)
- Alias selection dropdown
- Optional TTL configuration
- Error and success message handling

**API Hooks** (`web/src/api/Firewall.tsx`):
- `useAvailableAliases()` - Fetch available aliases
- `useUserEntries()` - Fetch user's entries (auto-refresh every 30s)
- `useAddIPWhitelistEntry()` - Add IP mutation
- `useRemoveIPWhitelistEntry()` - Remove IP mutation
- `useBlacklistIPEntry()` - Blacklist IP mutation (admin only)

**Types** (`web/src/types/Firewall.ts`):
- TypeScript interfaces for all firewall-related data

### UI Features

- **Status Badges**: Color-coded status indicators (Pending, Active, Removed, Blacklisted)
- **Expiration Warnings**: Visual alerts for IPs expiring within 7 days
- **Event History**: Complete audit trail for each entry
- **Confirmation Dialogs**: Safe removal with confirmation
- **Real-time Updates**: Auto-refresh entries every 30 seconds
- **Search & Filter**: Quick search across IP, alias, description

### Navigation

The firewall section appears in the Settings sidebar when firewall management is enabled:

```
Settings
├── Certificates
│   ├── Certificates
│   ├── Requests
│   └── Admin Requests
├── Firewall          ← NEW
│   ├── Whitelist     ← View entries
│   └── Add IP        ← Add new entry
├── Service Accounts
└── General
```

## Key Features

### Security

- IP validation on both frontend and backend
- Authorization checks per alias (group-based)
- Per-user and global IP limits
- Audit trail with client IP and user agent
- Admin-only blacklist capability

### Smart IP Management

- **Multiple Users, Same IP**: Multiple users can whitelist the same IP
- **Smart Removal**: IP only removed from firewall when ALL entries are inactive
- **Auto-Expiration**: Automatic cleanup of expired entries
- **Duplicate Prevention**: Prevents duplicate active entries per user/alias

### Audit & Observability

- Complete event history per entry
- Captures: requester, action type, timestamp, client IP, user agent
- Sync failure tracking
- Structured logging

## Testing

The feature has been fully implemented with:
- ✅ All backend endpoints
- ✅ Background sync and expiration jobs
- ✅ Database migrations
- ✅ Complete frontend UI
- ✅ TypeScript types and API hooks
- ✅ Form validation
- ✅ Error handling
- ✅ Compilation verified (both backend and frontend)

## Next Steps

1. **Test with real OPNsense instance**
   - Configure router credentials
   - Verify API communication
   - Test IP sync operations

2. **End-to-end testing**
   - Add IP through UI
   - Verify background sync
   - Test expiration
   - Test removal
   - Test blacklist (admin)

3. **Admin UI** (Optional future enhancement)
   - View all entries across all users
   - Blacklist UI with reason input
   - Statistics dashboard

4. **Additional features** (Optional)
   - Email notifications for expiring IPs
   - Bulk operations
   - IP range support (CIDR)
   - Geolocation display

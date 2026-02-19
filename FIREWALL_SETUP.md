# Firewall IP Whitelisting - Setup Guide

## To View the Firewall Page

### 1. Enable Firewall Management in Configuration

Add or update your `config.yaml`:

```yaml
features:
  firewall_management:
    enabled: true  # Set this to true
    router_endpoint: "https://your-opnsense-router.local"
    router_api_key: "your-api-key-here"
    router_api_secret: "your-api-secret-here"
    background_job_config:
      sync_interval: 5m
      expiration_interval: 1h
    aliases:
      - name: "vpn_users"
        uuid: "your-alias-uuid-from-opnsense"
        description: "VPN user whitelist"
        max_ips_per_user: 5
        max_total_ips: 100
        default_ttl: 720h  # 30 days
        auth_group: "vpn-users"  # Users in this group can use this alias
```

### 2. Configure OPNsense API Access

In your OPNsense firewall:

1. Go to **System > Access > Users**
2. Create an API user or use existing one
3. Generate API credentials (Key + Secret)
4. Note the Alias UUID from **Firewall > Aliases**

### 3. Set Up Authorization Groups

In your OIDC provider, ensure users have the appropriate groups:

**For regular users:**
- Add users to a group that matches the `auth_group` in your alias config (e.g., `vpn-users`)

**For admins (optional):**
- Add admin users to `conduit:firewall:admin` group for blacklist capabilities

### 4. Run Database Migrations

The migrations should run automatically on startup, but you can verify:

```bash
# Check if migration 003 has run
psql -d your_database -c "SELECT * FROM schema_migrations WHERE version = 3;"
```

### 5. Start the Application

```bash
# Development
make dev

# Or backend only
make dev-backend

# Production
./homelab-dashboard
```

### 6. Access the Firewall Page

1. Log in to the application
2. Go to **Settings** (hamburger menu)
3. You should see a **Firewall** section in the sidebar with:
   - **Whitelist** - View your IP entries
   - **Add IP** - Add a new IP to the whitelist

### Troubleshooting

**"Firewall section not visible in sidebar"**
- Check that `firewall_management.enabled: true` in config
- Verify user is authenticated
- Check browser console for errors
- Verify the auth status endpoint returns firewall config:
  ```bash
  curl -b cookies.txt http://localhost:8080/api/auth/status
  # Should include: {"config": {"firewall": {"enabled": true}}}
  ```

**"No firewall aliases available"**
- Verify your user's groups match the `auth_group` in the alias config
- Check that the alias is properly configured in `config.yaml`
- Check logs for authorization errors

**"IPs not syncing to OPNsense"**
- Verify router credentials are correct
- Check router endpoint is accessible
- Check background job logs for sync errors
- Verify the alias UUID matches OPNsense

### Testing Without OPNsense

You can test the UI without a real OPNsense instance:

1. Set `enabled: true` in config
2. Use dummy values for router credentials
3. The UI will work, but background sync will fail (logged as errors)
4. You can add IPs, view them, remove them - all database operations work
5. Status will remain "Pending" since sync job can't reach the router

### Key Configuration Fields

| Field | Description | Example |
|-------|-------------|---------|
| `enabled` | Enable/disable feature | `true` |
| `router_endpoint` | OPNsense API URL | `https://192.168.1.1` |
| `router_api_key` | API key from OPNsense | `abc123...` |
| `router_api_secret` | API secret from OPNsense | `xyz789...` |
| `aliases[].uuid` | Alias UUID from OPNsense | `12345678-1234-...` |
| `aliases[].auth_group` | OIDC group name | `vpn-users` |
| `aliases[].max_ips_per_user` | Limit per user | `5` |
| `aliases[].max_total_ips` | Global limit | `100` |
| `aliases[].default_ttl` | Default expiration | `720h` (30 days) |

### Next Steps

1. **Add IPs** - Use the "Add IP" page to whitelist an IP
2. **Monitor** - Watch the background sync job logs
3. **Verify** - Check OPNsense to confirm IPs are added to the alias
4. **Test Expiration** - Set short TTL to test auto-expiration
5. **Admin Features** - Test blacklist endpoint if you have admin access

### Logs to Monitor

```bash
# Backend logs
tail -f logs/app.log

# Look for:
# - "firewall management jobs registered" - Jobs started
# - "syncing firewall alias" - Sync job running
# - "firewall alias synced successfully" - Sync succeeded
# - "expired IP whitelist entries" - Expiration job running
```

### API Endpoints (for testing)

```bash
# Get available aliases
curl -b cookies.txt http://localhost:8080/api/firewall/aliases

# Get your entries
curl -b cookies.txt http://localhost:8080/api/firewall/entries

# Add an IP
curl -b cookies.txt -X POST http://localhost:8080/api/firewall/entries \
  -H "Content-Type: application/json" \
  -d '{"alias_name":"vpn_users","ip_address":"192.168.1.100","description":"Home network"}'

# Remove an IP
curl -b cookies.txt -X DELETE http://localhost:8080/api/firewall/entries/1
```

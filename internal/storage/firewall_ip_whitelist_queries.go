package storage

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
)

// AddIPToWhitelist adds a firewall ip whitelist entry.
func (p *DatabaseProvider) AddIPToWhitelist(ctx context.Context, ownerIss, ownerSub, aliasName, aliasUUID, ipAddress, description string, expiresAt *time.Time) (*models.FirewallIPWhitelistEntry, error) {
	query := `
		INSERT INTO firewall_ip_whitelist_entries (owner_iss, owner_sub, alias_name, alias_uuid, ip_address, description, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		returning id
	`

	var recordId int

	err := p.pool.QueryRow(ctx, query,
		ownerIss, ownerSub, aliasName, aliasUUID, ipAddress, description, expiresAt).Scan(&recordId)

	if err != nil {
		return nil, fmt.Errorf("failed to add IP to whitelist: %w", err)
	}

	return p.GetWhitelistEntryByID(ctx, recordId)
}

// GetWhitelistEntryByID returns a firewall whitelist entry, including events for a specific id.
func (p *DatabaseProvider) GetWhitelistEntryByID(ctx context.Context, id int) (*models.FirewallIPWhitelistEntry, error) {
	query := `
        SELECT id, owner_iss, owner_sub, alias_name, alias_uuid, ip_address::text, ip_version, description, status, 
               requested_at, added_at, removed_at, expires_at, removed_by_iss, removed_by_sub, removal_reason
        FROM firewall_ip_whitelist_entries
        WHERE id = $1
    `

	var whitelistEntry models.FirewallIPWhitelistEntry
	err := p.pool.QueryRow(ctx, query, id).Scan(
		&whitelistEntry.ID,
		&whitelistEntry.OwnerIss,
		&whitelistEntry.OwnerSub,
		&whitelistEntry.AliasName,
		&whitelistEntry.AliasUUID,
		&whitelistEntry.IPAddress,
		&whitelistEntry.IPVersion,
		&whitelistEntry.Description,
		&whitelistEntry.Status,
		&whitelistEntry.RequestedAt,
		&whitelistEntry.AddedAt,
		&whitelistEntry.RemovedAt,
		&whitelistEntry.ExpiresAt,
		&whitelistEntry.RemovedByIss,
		&whitelistEntry.RemovedBySub,
		&whitelistEntry.RemovalReason,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("whitelist entry not found")
		}
		return nil, fmt.Errorf("failed to get whitelist entry with id '%d': %w", id, err)
	}

	events, err := p.GetWhitelistEventsByEntry(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get events for whitelist entry '%d': %w", id, err)
	}

	whitelistEntry.Events = make([]models.FirewallIPWhitelistEvent, len(events))
	for i, event := range events {
		whitelistEntry.Events[i] = *event
	}

	return &whitelistEntry, nil
}

func (p *DatabaseProvider) GetUserWhitelistEntries(ctx context.Context, ownerIss, ownerSub string) ([]*models.FirewallIPWhitelistEntry, error) {
	query := `
        SELECT fiwe.id, fiwe.owner_iss, fiwe.owner_sub, fiwe.alias_name, fiwe.alias_uuid, fiwe.ip_address::text, fiwe.ip_version,
               fiwe.description, fiwe.status, fiwe.requested_at, fiwe.added_at, fiwe.removed_at, fiwe.expires_at, 
               fiwe.removed_by_iss, fiwe.removed_by_sub, fiwe.removal_reason,
               owner.username as owner_username,
               owner.display_name as owner_display_name
        FROM firewall_ip_whitelist_entries fiwe
        JOIN users owner ON fiwe.owner_iss = owner.iss AND fiwe.owner_sub = owner.sub
        WHERE fiwe.owner_iss = $1 AND fiwe.owner_sub = $2
        ORDER BY fiwe.requested_at DESC
    `

	eventsQuery := `
        SELECT fwe.id, fwe.whitelist_id, fwe.actor_iss, fwe.actor_sub, fwe.event_type, fwe.notes, fwe.client_ip::text,
               fwe.user_agent, fwe.created_at,
               COALESCE(actor.username, sa_creator.username) as actor_username,
               COALESCE(actor.display_name, sa_creator.display_name) as actor_display_name
        FROM firewall_whitelist_events fwe
        LEFT JOIN users actor ON fwe.actor_iss = actor.iss AND fwe.actor_sub = actor.sub
        LEFT JOIN service_accounts sa ON fwe.actor_iss = sa.iss AND fwe.actor_sub = sa.sub
        LEFT JOIN users sa_creator ON sa.created_by_iss = sa_creator.iss AND sa.created_by_sub = sa_creator.sub
        WHERE fwe.whitelist_id = $1
        ORDER BY fwe.created_at DESC
    `

	rows, err := p.pool.Query(ctx, query, ownerIss, ownerSub)
	if err != nil {
		return nil, fmt.Errorf("failed to get user whitelist entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.FirewallIPWhitelistEntry
	for rows.Next() {
		var entry models.FirewallIPWhitelistEntry

		err := rows.Scan(
			&entry.ID,
			&entry.OwnerIss,
			&entry.OwnerSub,
			&entry.AliasName,
			&entry.AliasUUID,
			&entry.IPAddress,
			&entry.IPVersion,
			&entry.Description,
			&entry.Status,
			&entry.RequestedAt,
			&entry.AddedAt,
			&entry.RemovedAt,
			&entry.ExpiresAt,
			&entry.RemovedByIss,
			&entry.RemovedBySub,
			&entry.RemovalReason,
			&entry.OwnerUsername,
			&entry.OwnerDisplayName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan whitelist entry: %w", err)
		}

		eventRows, err := p.pool.Query(ctx, eventsQuery, entry.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get events for whitelist entry '%d': %w", entry.ID, err)
		}

		var events []models.FirewallIPWhitelistEvent
		for eventRows.Next() {
			var event models.FirewallIPWhitelistEvent

			err := eventRows.Scan(
				&event.ID,
				&event.WhitelistID,
				&event.ActorISS,
				&event.ActorSub,
				&event.EventType,
				&event.Notes,
				&event.ClientIP,
				&event.UserAgent,
				&event.CreatedAt,
				&event.ActorUsername,
				&event.ActorDisplayName,
			)
			if err != nil {
				eventRows.Close()
				return nil, fmt.Errorf("failed to scan firewall event: %w", err)
			}
			events = append(events, event)
		}
		eventRows.Close()

		if err := eventRows.Err(); err != nil {
			return nil, fmt.Errorf("failed to iterate events: %w", err)
		}

		entry.Events = events
		if entry.Events == nil {
			entry.Events = []models.FirewallIPWhitelistEvent{}
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate whitelist entries: %w", err)
	}

	return entries, nil
}

func (p *DatabaseProvider) GetAllWhitelistEntries(ctx context.Context) ([]*models.FirewallIPWhitelistEntry, error) {
	query := `
        SELECT fiwe.id, fiwe.owner_iss, fiwe.owner_sub, fiwe.alias_name, fiwe.alias_uuid, 
               fiwe.ip_address::text, fiwe.ip_version, fiwe.description, fiwe.status, 
               fiwe.requested_at, fiwe.added_at, fiwe.removed_at, fiwe.expires_at, 
               fiwe.removed_by_iss, fiwe.removed_by_sub, fiwe.removal_reason,
               owner.username as owner_username,
               owner.display_name as owner_display_name
        FROM firewall_ip_whitelist_entries fiwe
        JOIN users owner ON fiwe.owner_iss = owner.iss AND fiwe.owner_sub = owner.sub
        ORDER BY fiwe.requested_at DESC
    `

	eventsQuery := `
        SELECT fwe.id, fwe.whitelist_id, fwe.actor_iss, fwe.actor_sub, fwe.event_type, fwe.notes, fwe.client_ip::text,
               fwe.user_agent, fwe.created_at,
               COALESCE(actor.username, sa_creator.username) as actor_username,
               COALESCE(actor.display_name, sa_creator.display_name) as actor_display_name
        FROM firewall_whitelist_events fwe
        LEFT JOIN users actor ON fwe.actor_iss = actor.iss AND fwe.actor_sub = actor.sub
        LEFT JOIN service_accounts sa ON fwe.actor_iss = sa.iss AND fwe.actor_sub = sa.sub
        LEFT JOIN users sa_creator ON sa.created_by_iss = sa_creator.iss AND sa.created_by_sub = sa_creator.sub
        ORDER BY fwe.created_at DESC
    `

	rows, err := p.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get user whitelist entries: %w", err)
	}
	defer rows.Close()

	var entries []*models.FirewallIPWhitelistEntry
	for rows.Next() {
		var entry models.FirewallIPWhitelistEntry

		err := rows.Scan(
			&entry.ID,
			&entry.OwnerIss,
			&entry.OwnerSub,
			&entry.AliasName,
			&entry.AliasUUID,
			&entry.IPAddress,
			&entry.IPVersion,
			&entry.Description,
			&entry.Status,
			&entry.RequestedAt,
			&entry.AddedAt,
			&entry.RemovedAt,
			&entry.ExpiresAt,
			&entry.RemovedByIss,
			&entry.RemovedBySub,
			&entry.RemovalReason,
			&entry.OwnerUsername,
			&entry.OwnerDisplayName,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan whitelist entry: %w", err)
		}

		eventRows, err := p.pool.Query(ctx, eventsQuery, entry.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get events for whitelist entry '%d': %w", entry.ID, err)
		}

		var events []models.FirewallIPWhitelistEvent
		for eventRows.Next() {
			var event models.FirewallIPWhitelistEvent

			err := eventRows.Scan(
				&event.ID,
				&event.WhitelistID,
				&event.ActorISS,
				&event.ActorSub,
				&event.EventType,
				&event.Notes,
				&event.ClientIP,
				&event.UserAgent,
				&event.CreatedAt,
				&event.ActorUsername,
				&event.ActorDisplayName,
			)
			if err != nil {
				eventRows.Close()
				return nil, fmt.Errorf("failed to scan firewall event: %w", err)
			}
			events = append(events, event)
		}
		eventRows.Close()

		if err := eventRows.Err(); err != nil {
			return nil, fmt.Errorf("failed to iterate events: %w", err)
		}

		entry.Events = events
		if entry.Events == nil {
			entry.Events = []models.FirewallIPWhitelistEvent{}
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate whitelist entries: %w", err)
	}

	return entries, nil
}

// RemoveIPFromWhitelist removes an IP from the whitelist
func (p *DatabaseProvider) RemoveIPFromWhitelist(ctx context.Context, id int, ownerIss, ownerSub string) error {
	query := `
        UPDATE firewall_ip_whitelist_entries
        SET status = 'removed', removed_at = NOW()
        WHERE id = $1
    `

	result, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to remove IP from whitelist: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("whitelist entry not found")
	}

	err = p.CreateWhitelistEvent(ctx, id, ownerIss, ownerSub, "removed", "", nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create removal event: %w", err)
	}

	return nil
}

// BlacklistIP blacklists an IP address (prevents re-adding)
func (p *DatabaseProvider) BlacklistIP(ctx context.Context, id int, adminIss, adminSub, reason string) error {
	query := `
        UPDATE firewall_ip_whitelist_entries
        SET status = 'blacklisted_by_admin',
            removed_at = NOW(),
            removed_by_iss = $2,
            removed_by_sub = $3,
            removal_reason = $4
        WHERE id = $1
    `

	result, err := p.pool.Exec(ctx, query, id, adminIss, adminSub, reason)
	if err != nil {
		return fmt.Errorf("failed to blacklist IP: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("whitelist entry not found")
	}

	err = p.CreateWhitelistEvent(ctx, id, adminIss, adminSub, "blacklisted_by_admin", reason, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create blacklist event: %w", err)
	}

	return nil
}

// IsIPBlacklisted checks if an IP is blacklisted for an alias
func (p *DatabaseProvider) IsIPBlacklisted(ctx context.Context, aliasUUID, ipAddress string) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM firewall_ip_whitelist_entries
            WHERE alias_uuid = $1 
              AND ip_address = $2 
              AND status = 'blacklisted_by_admin'
        )
    `

	var exists bool
	err := p.pool.QueryRow(ctx, query, aliasUUID, ipAddress).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if IP is blacklisted: %w", err)
	}

	return exists, nil
}

// GetPendingIPs gets all IPs that need to be added to the firewall for a specific alias
func (p *DatabaseProvider) GetPendingIPs(ctx context.Context, aliasUUID string) ([]*models.FirewallIPWhitelistEntry, error) {
	query := `
        SELECT id, owner_iss, owner_sub, alias_name, alias_uuid, ip_address::text, ip_version, description, status, 
               requested_at, added_at, removed_at, expires_at, removed_by_iss, removed_by_sub, removal_reason
        FROM firewall_ip_whitelist_entries
        WHERE alias_uuid = $1 
          AND status = 'requested'
          AND (expires_at IS NULL OR expires_at > NOW())
        ORDER BY requested_at ASC
    `

	rows, err := p.pool.Query(ctx, query, aliasUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending IPs: %w", err)
	}
	defer rows.Close()

	var entries []*models.FirewallIPWhitelistEntry
	for rows.Next() {
		var entry models.FirewallIPWhitelistEntry
		err := rows.Scan(
			&entry.ID,
			&entry.OwnerIss,
			&entry.OwnerSub,
			&entry.AliasName,
			&entry.AliasUUID,
			&entry.IPAddress,
			&entry.IPVersion,
			&entry.Description,
			&entry.Status,
			&entry.RequestedAt,
			&entry.AddedAt,
			&entry.RemovedAt,
			&entry.ExpiresAt,
			&entry.RemovedByIss,
			&entry.RemovedBySub,
			&entry.RemovalReason,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending IP: %w", err)
		}
		entry.Events = []models.FirewallIPWhitelistEvent{}
		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate pending IPs: %w", err)
	}

	return entries, nil
}

// MarkIPsAsAdded marks IPs as successfully added to the firewall
func (p *DatabaseProvider) MarkIPsAsAdded(ctx context.Context, ids []int, systemUserIss, systemUserSub string) error {
	query := `
        UPDATE firewall_ip_whitelist_entries
        SET status = 'added', added_at = NOW()
        WHERE id = ANY($1) AND status = 'requested'
    `

	_, err := p.pool.Exec(ctx, query, ids)
	if err != nil {
		return fmt.Errorf("failed to mark IPs as added: %w", err)
	}

	for _, id := range ids {
		err = p.CreateWhitelistEvent(ctx, id, systemUserIss, systemUserSub, "added", "", nil, nil)
		if err != nil {
			return fmt.Errorf("failed to create added event for IP %d: %w", id, err)
		}
	}

	return nil
}

// ExpireOldIPs marks expired IPs as removed and returns the count of expired IPs
func (p *DatabaseProvider) ExpireOldIPs(ctx context.Context, systemUserIss, systemUserSub string) (int, error) {
	selectQuery := `
        SELECT id
        FROM firewall_ip_whitelist_entries
        WHERE status IN ('requested', 'added')
          AND expires_at IS NOT NULL
          AND expires_at <= NOW()
    `

	rows, err := p.pool.Query(ctx, selectQuery)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired IPs: %w", err)
	}
	defer rows.Close()

	var expiredIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("failed to scan expired IP ID: %w", err)
		}
		expiredIDs = append(expiredIDs, id)
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("failed to iterate expired IPs: %w", err)
	}

	if len(expiredIDs) == 0 {
		return 0, nil
	}

	updateQuery := `
        UPDATE firewall_ip_whitelist_entries
        SET status = 'removed', removed_at = NOW()
        WHERE id = ANY($1)
    `

	_, err = p.pool.Exec(ctx, updateQuery, expiredIDs)
	if err != nil {
		return 0, fmt.Errorf("failed to mark IPs as expired: %w", err)
	}

	for _, id := range expiredIDs {
		err = p.CreateWhitelistEvent(ctx, id, systemUserIss, systemUserSub, "expired", "", nil, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to create expired event for IP %d: %w", id, err)
		}
	}

	return len(expiredIDs), nil
}

// CountUserActiveIPs counts how many active IPs a user has for a specific alias
func (p *DatabaseProvider) CountUserActiveIPs(ctx context.Context, ownerIss, ownerSub, aliasUUID string) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM firewall_ip_whitelist_entries
        WHERE owner_iss = $1 AND owner_sub = $2 AND alias_uuid = $3 
          AND status IN ('requested', 'added')
    `

	var count int
	err := p.pool.QueryRow(ctx, query, ownerIss, ownerSub, aliasUUID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count user active IPs: %w", err)
	}

	return count, nil
}

// CountTotalActiveIPs counts total active IPs for a specific alias
func (p *DatabaseProvider) CountTotalActiveIPs(ctx context.Context, aliasUUID string) (int, error) {
	query := `
        SELECT COUNT(*) 
        FROM firewall_ip_whitelist_entries
        WHERE alias_uuid = $1 
          AND status IN ('requested', 'added')
    `

	var count int
	err := p.pool.QueryRow(ctx, query, aliasUUID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count total active IPs: %w", err)
	}

	return count, nil
}

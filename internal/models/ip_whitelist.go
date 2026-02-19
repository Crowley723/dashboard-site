package models

import (
	"time"
)

type FirewallIPWhitelistEntry struct {
	ID               int    `json:"id"`
	OwnerIss         string `json:"owner_iss"`
	OwnerSub         string `json:"owner_sub"`
	OwnerUsername    string `json:"owner_username"`
	OwnerDisplayName string `json:"owner_display_name"`

	AliasName string `json:"alias_name"`
	AliasUUID string `json:"alias_uuid"`

	IPAddress string `json:"ip_address"`
	IPVersion int    `json:"ip_version"`

	Description string `json:"description"`

	Status FirewallIPWhitelistStatus `json:"status"`

	RequestedAt time.Time  `json:"requested_at"`
	AddedAt     *time.Time `json:"added_at,omitempty"`
	RemovedAt   *time.Time `json:"removed_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`

	RemovedByIss  *string `json:"removed_by_iss,omitempty"`
	RemovedBySub  *string `json:"removed_by_sub,omitempty"`
	RemovalReason *string `json:"removal_reason,omitempty"`
	
	Events []FirewallIPWhitelistEvent `json:"events"`
}

type FirewallIPWhitelistEvent struct {
	ID          int `json:"id"`
	WhitelistID int `json:"whitelist_id"`

	ActorISS         string `json:"actor_iss"`
	ActorSub         string `json:"actor_sub"`
	ActorUsername    string `json:"actor_username"`
	ActorDisplayName string `json:"actor_display_name"`

	EventType string  `json:"event_type"`
	Notes     *string `json:"notes,omitempty"`

	ClientIP  *string `json:"client_ip,omitempty"`
	UserAgent *string `json:"user_agent,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

type FirewallIPWhitelistStatus string

const (
	StatusAdded              FirewallIPWhitelistStatus = "added"
	StatusRequested          FirewallIPWhitelistStatus = "requested"
	StatusRemoved            FirewallIPWhitelistStatus = "removed"
	StatusRemovedByAdmin     FirewallIPWhitelistStatus = "removed_by_admin"
	StatusBlacklistedByAdmin FirewallIPWhitelistStatus = "blacklisted_by_admin"
)

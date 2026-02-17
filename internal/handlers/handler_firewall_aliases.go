package handlers

import (
	"encoding/json"
	"fmt"
	"homelab-dashboard/internal/authorization"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type AvailableAliasResponse struct {
	Name          string `json:"name"`
	UUID          string `json:"uuid"`
	Description   string `json:"description"`
	MaxIPsPerUser int    `json:"max_ips_per_user"`
	MaxTotalIPs   int    `json:"max_total_ips"`
	DefaultTTL    *int64 `json:"default_ttl_hours,omitempty"` // hours, null = no expiration
}

func GETAvailableAliases(ctx *middlewares.AppContext) {
	principal := ctx.GetPrincipal()
	if principal == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if !principal.HasScope(ctx.Config, authorization.ScopeFirewallReadOwn) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	principalGroups := principal.GetGroups()
	if principalGroups == nil {
		principalGroups = []string{}
	}

	principalGroupMap := make(map[string]bool)
	for _, group := range principalGroups {
		principalGroupMap[group] = true
	}

	var availableAliases []AvailableAliasResponse

	for _, alias := range ctx.Config.Features.FirewallManagement.Aliases {
		if !principalGroupMap[alias.AuthGroup] {
			continue
		}

		groupScopes, exists := ctx.Config.Authorization.GroupScopes[alias.AuthGroup]
		if !exists {
			continue
		}

		hasReadScope := false
		for _, scope := range groupScopes {
			if scope == authorization.ScopeFirewallReadOwn {
				hasReadScope = true
				break
			}
		}

		if !hasReadScope {
			continue
		}

		var ttlHours *int64
		if alias.DefaultTTL != nil {
			hours := int64(alias.DefaultTTL.Hours())
			ttlHours = &hours
		}

		availableAliases = append(availableAliases, AvailableAliasResponse{
			Name:          alias.Name,
			UUID:          alias.UUID,
			Description:   alias.Description,
			MaxIPsPerUser: alias.MaxIPsPerUser,
			MaxTotalIPs:   alias.MaxTotalIPs,
			DefaultTTL:    ttlHours,
		})
	}

	if availableAliases == nil {
		availableAliases = []AvailableAliasResponse{}
	}

	ctx.WriteJSON(http.StatusOK, availableAliases)
}

func GETUserEntries(ctx *middlewares.AppContext) {
	principal := ctx.GetPrincipal()
	if principal == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if !principal.HasScope(ctx.Config, authorization.ScopeFirewallReadOwn) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	var entries []*models.FirewallIPWhitelistEntry
	var err error

	allUsers := ctx.Request.URL.Query().Get("all_users")
	ownerSub := ctx.Request.URL.Query().Get("owner_sub")
	ownerIss := ctx.Request.URL.Query().Get("owner_iss")

	isAdminQuery := allUsers == "1" || ownerSub != "" || ownerIss != ""

	if isAdminQuery {
		if !principal.HasScope(ctx.Config, authorization.ScopeFirewallReadAll) {
			ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
			return
		}

		if allUsers == "1" {
			entries, err = ctx.Storage.GetAllWhitelistEntries(ctx)
			if err != nil {
				ctx.Logger.Error("failed to get all whitelist entries", "error", err)
				ctx.SetJSONError(http.StatusInternalServerError, "Failed to get whitelist entries")
				return
			}
		} else {
			if ownerSub == "" || ownerIss == "" {
				ctx.SetJSONError(http.StatusBadRequest, "owner_sub and owner_iss must both be provided")
				return
			}

			entries, err = ctx.Storage.GetUserWhitelistEntries(ctx, ownerIss, ownerSub)
			if err != nil {
				ctx.Logger.Error("failed to get whitelist entries for user",
					"error", err,
					"owner_sub", ownerSub,
					"owner_iss", ownerIss,
				)
				ctx.SetJSONError(http.StatusInternalServerError, "Failed to get whitelist entries")
				return
			}
		}
	} else {
		entries, err = ctx.Storage.GetUserWhitelistEntries(ctx, principal.GetIss(), principal.GetSub())
		if err != nil {
			ctx.Logger.Error("failed to get user whitelist entries",
				"error", err,
				"user", principal.GetUsername(),
			)
			ctx.SetJSONError(http.StatusInternalServerError, "Failed to get whitelist entries")
			return
		}
	}

	if entries == nil {
		ctx.WriteJSON(http.StatusOK, []interface{}{})
		return
	}

	ctx.WriteJSON(http.StatusOK, entries)
}

func POSTAddIPEntry(ctx *middlewares.AppContext) {
	principal := ctx.GetPrincipal()
	if principal == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	// Check if user has the firewall:request:own scope
	if !principal.HasScope(ctx.Config, authorization.ScopeFirewallRequestOwn) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	var req struct {
		AliasName   string `json:"alias_name"`
		IPAddress   string `json:"ip_address"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		ctx.Logger.Error("failed to decode request body", "error", err)
		ctx.SetJSONError(http.StatusBadRequest, "Invalid request body")
		return
	}

	req.AliasName = strings.TrimSpace(req.AliasName)
	req.IPAddress = strings.TrimSpace(req.IPAddress)

	if req.AliasName == "" {
		ctx.SetJSONError(http.StatusBadRequest, "alias_name is required")
		return
	}

	if req.IPAddress == "" {
		ctx.SetJSONError(http.StatusBadRequest, "ip_address is required")
		return
	}

	// CRITICAL SECURITY FIX: Validate IP address format
	parsedIP := net.ParseIP(strings.TrimSpace(req.IPAddress))
	if parsedIP == nil {
		ctx.SetJSONError(http.StatusBadRequest, "Invalid IP address format")
		return
	}
	// Use canonical form from parsing
	req.IPAddress = parsedIP.String()

	user, ok := principal.(*models.User)
	if !ok {
		ctx.SetJSONError(http.StatusForbidden, "Firewall management is only available for user accounts")
		return
	}

	userGroups := user.Groups
	if userGroups == nil {
		userGroups = []string{}
	}

	var matchedAlias *config.FirewallAliasConfig
	userGroupSet := make(map[string]bool)
	for _, group := range userGroups {
		userGroupSet[group] = true
	}

	for _, alias := range ctx.Config.Features.FirewallManagement.Aliases {
		if alias.Name == req.AliasName && userGroupSet[alias.AuthGroup] {
			matchedAlias = &alias
			break
		}
	}

	if matchedAlias == nil {
		ctx.SetJSONError(http.StatusForbidden, "You do not have access to this alias")
		return
	}

	isBlacklisted, err := ctx.Storage.IsIPBlacklisted(ctx, matchedAlias.UUID, req.IPAddress)
	if err != nil {
		ctx.Logger.Error("failed to check if IP is blacklisted",
			"error", err,
			"ip", req.IPAddress,
			"alias", req.AliasName,
		)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to verify IP status")
		return
	}

	if isBlacklisted {
		ctx.SetJSONError(http.StatusForbidden, "This IP address has been blacklisted and cannot be added")
		return
	}

	userCount, err := ctx.Storage.CountUserActiveIPs(ctx, principal.GetIss(), principal.GetSub(), matchedAlias.UUID)
	if err != nil {
		ctx.Logger.Error("failed to count user active IPs",
			"error", err,
			"user", principal.GetUsername(),
			"alias", req.AliasName,
		)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to verify IP limits")
		return
	}

	if userCount >= matchedAlias.MaxIPsPerUser {
		ctx.SetJSONError(http.StatusBadRequest,
			fmt.Sprintf("You have reached the maximum IP limit (%d) for this alias", matchedAlias.MaxIPsPerUser))
		return
	}

	// Check total alias limit
	totalCount, err := ctx.Storage.CountTotalActiveIPs(ctx, matchedAlias.UUID)
	if err != nil {
		ctx.Logger.Error("failed to count total active IPs",
			"error", err,
			"alias", req.AliasName,
		)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to verify alias limits")
		return
	}

	if totalCount >= matchedAlias.MaxTotalIPs {
		ctx.SetJSONError(http.StatusBadRequest,
			fmt.Sprintf("This alias has reached its maximum total IP limit (%d)", matchedAlias.MaxTotalIPs))
		return
	}

	// Calculate expiration based on alias config
	var expiresAt *time.Time
	if matchedAlias.DefaultTTL != nil {
		expiry := time.Now().Add(*matchedAlias.DefaultTTL)
		expiresAt = &expiry
	}

	// Extract client IP from request
	clientIP := ""
	host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
	if err == nil {
		clientIP = host
	}

	// Extract user agent
	userAgentStr := ctx.Request.UserAgent()

	// Convert to pointers for storage function
	var clientIPPtr, userAgentPtr *string
	if clientIP != "" {
		clientIPPtr = &clientIP
	}
	if userAgentStr != "" {
		userAgentPtr = &userAgentStr
	}

	// Add IP to whitelist
	entry, err := ctx.Storage.AddIPToWhitelist(
		ctx,
		principal.GetIss(),
		principal.GetSub(),
		matchedAlias.Name,
		matchedAlias.UUID,
		req.IPAddress,
		req.Description,
		expiresAt,
		clientIPPtr,
		userAgentPtr,
	)
	if err != nil {
		// Check if it's a duplicate IP error
		if strings.Contains(err.Error(), "you already have this IP address whitelisted") {
			ctx.SetJSONError(http.StatusConflict, err.Error())
			return
		}

		ctx.Logger.Error("failed to add IP to whitelist",
			"error", err,
			"user", principal.GetUsername(),
			"ip", req.IPAddress,
			"alias", req.AliasName,
		)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to add IP to whitelist")
		return
	}

	ctx.Logger.Info("IP added to whitelist",
		"user", principal.GetUsername(),
		"ip", req.IPAddress,
		"alias", req.AliasName,
		"entry_id", entry.ID,
	)

	ctx.WriteJSON(http.StatusCreated, entry)
}

func DELETERemoveIPEntry(ctx *middlewares.AppContext) {
	principal := ctx.GetPrincipal()
	if principal == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	if !principal.HasScope(ctx.Config, authorization.ScopeFirewallRevokeOwn) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	// Get entry ID from URL
	idParam := chi.URLParam(ctx.Request, "id")
	if idParam == "" {
		ctx.SetJSONError(http.StatusBadRequest, "Entry ID is required")
		return
	}

	entryID, err := strconv.Atoi(strings.TrimSpace(idParam))
	if err != nil {
		ctx.SetJSONError(http.StatusBadRequest, "Invalid entry ID")
		return
	}

	entry, err := ctx.Storage.GetWhitelistEntryByID(ctx, entryID)
	if err != nil {
		ctx.Logger.Error("failed to get whitelist entry",
			"error", err,
			"entry_id", entryID,
		)
		ctx.SetJSONError(http.StatusNotFound, "Whitelist entry not found")
		return
	}

	if !principal.MatchesOwner(entry.OwnerIss, entry.OwnerSub) {
		ctx.Logger.Warn("user attempted to remove IP they don't own",
			"user", principal.GetUsername(),
			"entry_id", entryID,
			"owner", entry.OwnerUsername,
		)
		ctx.SetJSONError(http.StatusForbidden, "You can only remove your own IP addresses")
		return
	}

	if entry.Status == models.StatusRemoved ||
		entry.Status == models.StatusRemovedByAdmin ||
		entry.Status == models.StatusBlacklistedByAdmin {
		ctx.SetJSONError(http.StatusBadRequest,
			fmt.Sprintf("IP address is already removed (status: %s)", entry.Status))
		return
	}

	// Extract client IP from request
	clientIP := ""
	host, _, err := net.SplitHostPort(ctx.Request.RemoteAddr)
	if err == nil {
		clientIP = host
	}

	// Extract user agent
	userAgentStr := ctx.Request.UserAgent()

	// Convert to pointers
	var clientIPPtr, userAgentPtr *string
	if clientIP != "" {
		clientIPPtr = &clientIP
	}
	if userAgentStr != "" {
		userAgentPtr = &userAgentStr
	}

	err = ctx.Storage.RemoveIPFromWhitelist(ctx, entryID, principal.GetIss(), principal.GetSub(), clientIPPtr, userAgentPtr)
	if err != nil {
		ctx.Logger.Error("failed to remove IP from whitelist",
			"error", err,
			"user", principal.GetUsername(),
			"entry_id", entryID,
		)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to remove IP from whitelist")
		return
	}

	ctx.Logger.Info("IP removed from whitelist",
		"user", principal.GetUsername(),
		"entry_id", entryID,
		"ip", entry.IPAddress,
		"alias", entry.AliasName,
	)

	ctx.Response.WriteHeader(http.StatusNoContent)
}

// DELETEBlacklistIPEntry blacklists an IP address (admin-only).
// This blacklists ALL entries with the same IP address, preventing it from being re-added.
func DELETEBlacklistIPEntry(ctx *middlewares.AppContext) {
	// 1. Authenticate
	principal := ctx.GetPrincipal()
	if principal == nil {
		ctx.SetJSONError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
		return
	}

	// 2. Authorize (admin-only scope)
	if !principal.HasScope(ctx.Config, authorization.ScopeFirewallBlacklist) {
		ctx.SetJSONError(http.StatusForbidden, http.StatusText(http.StatusForbidden))
		return
	}

	// 3. Extract and validate path parameter
	idParam := chi.URLParam(ctx.Request, "id")
	if idParam == "" {
		ctx.SetJSONError(http.StatusBadRequest, "Entry ID is required")
		return
	}

	entryID, err := strconv.Atoi(strings.TrimSpace(idParam))
	if err != nil {
		ctx.SetJSONError(http.StatusBadRequest, "Invalid entry ID")
		return
	}

	// 4. Decode optional request body (reason)
	var req struct {
		Reason string `json:"reason"`
	}
	// Ignore decode errors - reason is optional
	_ = json.NewDecoder(ctx.Request.Body).Decode(&req)
	req.Reason = strings.TrimSpace(req.Reason)

	// 5. Get the entry to obtain the IP address
	entry, err := ctx.Storage.GetWhitelistEntryByID(ctx, entryID)
	if err != nil {
		ctx.Logger.Error("failed to get whitelist entry",
			"error", err,
			"entry_id", entryID,
		)
		ctx.SetJSONError(http.StatusNotFound, "Whitelist entry not found")
		return
	}

	// 6. Validate state - can't blacklist if already blacklisted
	if entry.Status == models.StatusBlacklistedByAdmin {
		ctx.SetJSONError(http.StatusBadRequest, "IP address is already blacklisted")
		return
	}

	// 7. Blacklist ALL entries with this IP address
	count, err := ctx.Storage.BlacklistIPAddress(
		ctx,
		entry.AliasUUID,
		entry.IPAddress,
		principal.GetIss(),
		principal.GetSub(),
		req.Reason,
	)
	if err != nil {
		ctx.Logger.Error("failed to blacklist IP address",
			"error", err,
			"admin", principal.GetUsername(),
			"ip", entry.IPAddress,
			"alias_uuid", entry.AliasUUID,
		)
		ctx.SetJSONError(http.StatusInternalServerError, "Failed to blacklist IP address")
		return
	}

	// 8. Audit logging
	ctx.Logger.Info("IP address blacklisted",
		"admin", principal.GetUsername(),
		"ip", entry.IPAddress,
		"alias_uuid", entry.AliasUUID,
		"alias_name", entry.AliasName,
		"entries_affected", count,
		"reason", req.Reason,
	)

	// 9. Return success
	ctx.Response.WriteHeader(http.StatusNoContent)
}

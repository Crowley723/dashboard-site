package jobs

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/models"
	"homelab-dashboard/internal/services/firewall"
	"log/slog"
	"strings"
	"time"
)

type FirewallSyncJob struct {
	appCtx       *middlewares.AppContext
	routerClient *firewall.RouterClient
	interval     time.Duration
	logger       *slog.Logger
}

func NewFirewallSyncJob(appCtx *middlewares.AppContext, routerClient *firewall.RouterClient, interval time.Duration, logger *slog.Logger) *FirewallSyncJob {
	return &FirewallSyncJob{
		appCtx:       appCtx,
		routerClient: routerClient,
		interval:     interval,
		logger:       logger,
	}
}

func (j *FirewallSyncJob) Name() string {
	return "firewall_ip_sync"
}

func (j *FirewallSyncJob) RequiresLeadership() bool {
	return true // Only leader should sync to firewall
}

func (j *FirewallSyncJob) Interval() time.Duration {
	return j.interval
}

func (j *FirewallSyncJob) Run(ctx context.Context) error {
	if j.interval <= 0 {
		return fmt.Errorf("firewall sync job interval must be positive")
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	if err := j.syncAllAliases(ctx); err != nil && !errors.Is(err, context.Canceled) {
		j.logger.Error("initial firewall sync failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := j.syncAllAliases(ctx); err != nil && !errors.Is(err, context.Canceled) {
				j.logger.Error("firewall sync failed", "error", err)
			}
		}
	}
}

func (j *FirewallSyncJob) syncAllAliases(ctx context.Context) error {
	systemUserIss, systemUserSub, err := j.appCtx.Storage.GetSystemUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system user: %w", err)
	}

	for _, aliasConfig := range j.appCtx.Config.Features.FirewallManagement.Aliases {
		if err := j.syncAlias(ctx, &aliasConfig, systemUserIss, systemUserSub); err != nil {
			j.logger.Error("failed to sync alias",
				"alias_name", aliasConfig.Name,
				"alias_uuid", aliasConfig.UUID,
				"error", err,
			)
		}
	}

	return nil
}

func (j *FirewallSyncJob) syncAlias(ctx context.Context, aliasConfig *config.FirewallAliasConfig, systemUserIss, systemUserSub string) error {
	currentFirewallIPs, err := j.routerClient.GetAliasIPs(ctx, aliasConfig.UUID)
	if err != nil {
		return fmt.Errorf("failed to get current firewall IPs: %w", err)
	}

	allEntries, err := j.appCtx.Storage.GetAllWhitelistEntries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get all whitelist entries: %w", err)
	}

	var aliasEntries []*models.FirewallIPWhitelistEntry
	for _, entry := range allEntries {
		if entry.AliasUUID == aliasConfig.UUID {
			aliasEntries = append(aliasEntries, entry)
		}
	}

	ipStatusMap := make(map[string]string) // IP -> desired status ("add" or "remove")
	ipToEntries := make(map[string][]*models.FirewallIPWhitelistEntry)

	// Helper function to strip CIDR notation
	stripCIDR := func(ip string) string {
		if idx := strings.Index(ip, "/"); idx != -1 {
			return ip[:idx]
		}
		return ip
	}

	for _, entry := range aliasEntries {
		// Strip CIDR notation for comparison
		cleanIP := stripCIDR(entry.IPAddress)
		ipToEntries[cleanIP] = append(ipToEntries[cleanIP], entry)
	}

	for ip, entries := range ipToEntries {
		hasActiveEntry := false
		pendingIDs := []int{}

		for _, entry := range entries {
			if entry.Status == models.StatusRequested {
				pendingIDs = append(pendingIDs, entry.ID)
				hasActiveEntry = true
			} else if entry.Status == models.StatusAdded {
				hasActiveEntry = true
			}
		}

		if hasActiveEntry {
			ipStatusMap[ip] = "add"
			if len(pendingIDs) > 0 {
				err := j.appCtx.Storage.MarkIPsAsAdded(ctx, pendingIDs, systemUserIss, systemUserSub)
				if err != nil {
					j.logger.Error("failed to mark IPs as added",
						"ip", ip,
						"alias", aliasConfig.Name,
						"error", err,
					)
				}
			}
		} else {
			ipStatusMap[ip] = "remove"
		}
	}

	var ipsToAdd []string
	var ipsToRemove []string

	currentIPSet := make(map[string]bool)
	for _, ip := range currentFirewallIPs {
		currentIPSet[ip] = true
	}

	for ip, status := range ipStatusMap {
		if status == "add" && !currentIPSet[ip] {
			ipsToAdd = append(ipsToAdd, ip)
		}
	}

	for ip := range currentIPSet {
		status, exists := ipStatusMap[ip]
		if !exists || status == "remove" {
			ipsToRemove = append(ipsToRemove, ip)
		}
	}

	if len(ipsToAdd) > 0 || len(ipsToRemove) > 0 {
		j.logger.Info("syncing firewall alias",
			"alias", aliasConfig.Name,
			"ips_to_add", len(ipsToAdd),
			"ips_to_remove", len(ipsToRemove),
		)

		err := j.routerClient.UpdateAlias(ctx, aliasConfig.UUID, ipsToAdd, ipsToRemove)
		if err != nil {
			for ip := range ipStatusMap {
				for _, entry := range ipToEntries[ip] {
					_ = j.appCtx.Storage.CreateWhitelistEvent(
						ctx,
						entry.ID,
						systemUserIss,
						systemUserSub,
						"sync_failed",
						err.Error(),
						nil,
						nil,
					)
				}
			}
			return fmt.Errorf("failed to update firewall alias: %w", err)
		}

		j.logger.Info("firewall alias synced successfully",
			"alias", aliasConfig.Name,
			"added", ipsToAdd,
			"removed", ipsToRemove,
		)
	}

	return nil
}

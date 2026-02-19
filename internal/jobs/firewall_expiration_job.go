package jobs

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/middlewares"
	"log/slog"
	"time"
)

type FirewallExpirationJob struct {
	appCtx   *middlewares.AppContext
	interval time.Duration
	logger   *slog.Logger
}

func NewFirewallExpirationJob(appCtx *middlewares.AppContext, interval time.Duration, logger *slog.Logger) *FirewallExpirationJob {
	return &FirewallExpirationJob{
		appCtx:   appCtx,
		interval: interval,
		logger:   logger,
	}
}

func (j *FirewallExpirationJob) Name() string {
	return "firewall_ip_expiration"
}

func (j *FirewallExpirationJob) RequiresLeadership() bool {
	return true // Only leader should mark entries as expired
}

func (j *FirewallExpirationJob) Interval() time.Duration {
	return j.interval
}

func (j *FirewallExpirationJob) Run(ctx context.Context) error {
	if j.interval <= 0 {
		return fmt.Errorf("firewall expiration job interval must be positive")
	}

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	if err := j.expireOldIPs(ctx); err != nil && !errors.Is(err, context.Canceled) {
		j.logger.Error("initial expiration check failed", "error", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := j.expireOldIPs(ctx); err != nil && !errors.Is(err, context.Canceled) {
				j.logger.Error("expiration check failed", "error", err)
			}
		}
	}
}

func (j *FirewallExpirationJob) expireOldIPs(ctx context.Context) error {
	systemUserIss, systemUserSub, err := j.appCtx.Storage.GetSystemUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get system user: %w", err)
	}

	count, err := j.appCtx.Storage.ExpireOldIPs(ctx, systemUserIss, systemUserSub)
	if err != nil {
		return fmt.Errorf("failed to expire IPs: %w", err)
	}

	if count > 0 {
		j.logger.Info("expired IP whitelist entries", "count", count)
	}

	return nil
}

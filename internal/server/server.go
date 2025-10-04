package server

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/auth"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/data"
	"homelab-dashboard/internal/middlewares"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(cfg *config.Config) error {
	logger := setupLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionManager, err := auth.NewSessionManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	oidcProvider, err := auth.NewRealOIDCProvider(ctx, cfg.OIDC)

	dataService, cache, err := setupDataService(cfg, logger)
	if err != nil {
		return err
	}

	timeInterval := calculateFetchInterval(cfg, 10*time.Minute)
	go func() {
		if err := runBackgroundDataFetching(ctx, dataService, logger, timeInterval); err != nil {
			logger.Error("background data fetching stopped", "error", err)
		}
	}()

	appCtx := middlewares.NewAppContext(ctx, cfg, logger, cache, sessionManager, oidcProvider)

	router := setupRouter(appCtx)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	go func() {
		logger.Info("Server starting", "port", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server failed to start", "error", err)
			cancel()
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		logger.Info("Shutdown signal received")
	case <-ctx.Done():
		logger.Info("Context canceled")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	logger.Info("Shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	logger.Info("Server exited")
	return nil
}

func setupDataService(cfg *config.Config, logger *slog.Logger) (*data.Service, data.CacheProvider, error) {
	mimirClient, err := data.NewMimirClient(
		cfg.Data.PrometheusURL,
		cfg.Data.BasicAuth.Username,
		cfg.Data.BasicAuth.Password,
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to create new mimir client: %w", err)
	}

	cache := data.NewCacheProvider(cfg, logger)
	return data.NewService(mimirClient, cache, logger, cfg.Data.Queries), cache, nil
}

// calculateFetchInterval determines how often the background data fetching should happen, based completely on the shortest configured ttl, falling back to the default if there are no ttl configured.
func calculateFetchInterval(cfg *config.Config, defaultInterval time.Duration) time.Duration {
	var minTTL time.Duration

	for _, q := range cfg.Data.Queries {
		if q.TTL <= 0 {
			continue
		}

		if q.TTL == 0 || q.TTL < minTTL {
			minTTL = q.TTL
		}
	}

	if minTTL == 0 {
		minTTL = defaultInterval
	}

	if minTTL < time.Second {
		minTTL = time.Second
	}

	return minTTL
}

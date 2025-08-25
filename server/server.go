package server

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/auth"
	"homelab-dashboard/config"
	"homelab-dashboard/middlewares"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Start(cfg *config.Config) error {
	logger := setupLogger(cfg)

	sessionManager, err := auth.NewSessionManager(cfg.Sessions)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	oidcProvider, oauth2Config, err := auth.NewOIDCProvider(ctx, cfg.OIDC)

	appCtx := middlewares.NewAppContext(ctx, cfg, logger, sessionManager, oidcProvider, oauth2Config)

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

func setupLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	var level slog.Level
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(os.Stderr, opts)
	}

	return slog.New(handler)
}

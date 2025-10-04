package server

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/auth"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/data"
	"homelab-dashboard/internal/leader"
	"homelab-dashboard/internal/middlewares"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	cfg         *config.Config
	logger      *slog.Logger
	appCtx      *middlewares.AppContext
	httpServer  *http.Server
	debugServer *http.Server
	dataService *data.Service
	cache       *data.CacheProvider
	election    *leader.Election
	ctx         *context.Context
	cancel      context.CancelFunc
}

func New(cfg *config.Config) (*Server, error) {
	logger := setupLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	sessionManager, err := auth.NewSessionManager(cfg)
	if err != nil {
		cancel()
		return nil, err
	}

	oidcProvider, err := auth.NewRealOIDCProvider(ctx, cfg.OIDC)

	dataService, cache, err := setupDataService(cfg, logger)
	if err != nil {
		cancel()
		return nil, err
	}

	var election *leader.Election
	if cfg.Distributed.Enabled {
		redisClient := redis.NewClient(&redis.Options{
			Addr: cfg.Redis.Address,
			DB:   cfg.Redis.LeaderIndex,
		})

		hostname := os.Getenv("HOSTNAME")
		if hostname == "" {
			hostname = uuid.New().String()
		}

		election = &leader.Election{
			Redis:      redisClient,
			InstanceID: hostname,
			TTL:        cfg.Distributed.TTL,
		}
	}

	appCtx := middlewares.NewAppContext(ctx, cfg, logger, cache, sessionManager, oidcProvider)

	router := setupRouter(appCtx)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	var debugServer *http.Server
	if cfg.Server.Debug.Enabled {
		debugRouter := setupDebugRouter()
		debugServer = &http.Server{
			Addr:    fmt.Sprintf("%s:%d", cfg.Server.Debug.Host, cfg.Server.Debug.Port),
			Handler: debugRouter,
		}

	}

	return &Server{
		cfg:         cfg,
		logger:      logger,
		appCtx:      appCtx,
		httpServer:  server,
		debugServer: debugServer,
		dataService: dataService,
		election:    election,
		ctx:         &ctx,
		cancel:      cancel,
	}, nil
}

func (s *Server) Start() error {

	timeInterval := calculateFetchInterval(s.cfg, 10*time.Minute)
	go func() {
		if err := s.runBackgroundDataFetching(timeInterval); err != nil {
			s.logger.Error("background data fetching stopped", "error", err)
		}
	}()

	router := setupRouter(s.appCtx)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Server.Port),
		Handler: router,
	}

	go func() {
		s.logger.Info("Server starting", "port", s.cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Server failed to start", "error", err)
			s.cancel()
		}
	}()

	if s.cfg.Server.Debug.Enabled {

		go func() {
			s.logger.Info("Metrics server starting", "address", fmt.Sprintf("%s:%d", s.cfg.Server.Debug.Host, s.cfg.Server.Debug.Port))
			if err := s.debugServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				s.logger.Error("Metrics server failed to start", "error", err)
				s.cancel()
			}
		}()
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		s.logger.Info("Shutdown signal received")
	case <-s.appCtx.Done():
		s.logger.Info("Context canceled")
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	s.logger.Info("Shutting down server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		s.logger.Error("Server forced to shutdown", "error", err)
		return err
	}

	if s.debugServer != nil && s.cfg.Server.Debug.Enabled {
		if err := s.debugServer.Shutdown(shutdownCtx); err != nil {
			s.logger.Error("Debug server forced to shutdown", "error", err)
		}
	}

	s.logger.Info("Server exited")
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

	cache, err := data.NewCacheProvider(cfg, logger)
	if err != nil {
		logger.Error("error setting up cache provider", "error", err)
	}
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

func (s *Server) monitorLeadership(ctx context.Context) {
	var cancel context.CancelFunc
	var fetcherCtx context.Context

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if cancel != nil {
				cancel()
			}
			return

		case <-ticker.C:
			isLeader := s.election.IsLeader()

			if isLeader && cancel == nil {
				fetcherCtx, cancel = context.WithCancel(ctx)
				go s.runBackgroundDataFetcher(fetcherCtx)

			} else if !isLeader && cancel != nil {
				cancel()
				cancel = nil
			}
		}
	}
}

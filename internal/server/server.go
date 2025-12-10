package server

import (
	"context"
	"errors"
	"fmt"
	"homelab-dashboard/internal/auth"
	"homelab-dashboard/internal/config"
	"homelab-dashboard/internal/data"
	"homelab-dashboard/internal/distributed"
	"homelab-dashboard/internal/metrics"
	"homelab-dashboard/internal/middlewares"
	"homelab-dashboard/internal/storage"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/extra/redisprometheus/v9"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	cfg         *config.Config
	logger      *slog.Logger
	appCtx      *middlewares.AppContext
	httpServer  *http.Server
	debugServer *http.Server
	dataService *data.Service
	cache       data.CacheProvider
	election    *distributed.Election
	ctx         *context.Context
	cancel      context.CancelFunc
}

func New(cfg *config.Config) (*Server, error) {
	logger := setupLogger(cfg)

	ctx, cancel := context.WithCancel(context.Background())

	sessionManager, err := auth.NewSessionManager(logger, cfg)
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

	var election *distributed.Election
	if cfg.Distributed != nil && cfg.Distributed.Enabled {
		var client *redis.Client

		if cfg.Redis.Sentinel != nil {
			logger.Info("connecting to redis via sentinel",
				"master", cfg.Redis.Sentinel.MasterName,
				"sentinels", cfg.Redis.Sentinel.SentinelAddresses)

			client = redis.NewFailoverClient(&redis.FailoverOptions{
				MasterName:       cfg.Redis.Sentinel.MasterName,
				SentinelAddrs:    cfg.Redis.Sentinel.SentinelAddresses,
				SentinelPassword: cfg.Redis.Sentinel.SentinelPassword,
				Password:         cfg.Redis.Password,
				DB:               cfg.Redis.LeaderIndex,
				MinIdleConns:     2,
			})
		} else {
			client = redis.NewClient(&redis.Options{
				Addr:         cfg.Redis.Address,
				Password:     cfg.Redis.Password,
				DB:           cfg.Redis.LeaderIndex,
				MinIdleConns: 2,
			})
		}

		if cfg.Server.Debug != nil && cfg.Server.Debug.Enabled {
			collector := redisprometheus.NewCollector(metrics.Namespace, "election", client)
			if err := prometheus.Register(collector); err != nil {
				logger.Debug("failed to register redis election collector: already registered", "error", err)
			}
		}

		hostname := os.Getenv("HOSTNAME")
		if hostname == "" {
			hostname = uuid.New().String()
		}

		election = &distributed.Election{
			Redis:      client,
			InstanceID: hostname,
			TTL:        cfg.Distributed.TTL,
		}
	}

	var database *storage.DatabaseProvider
	if cfg.Storage.Enabled == true {
		dbProvider, err := storage.NewDatabaseProvider(ctx, cfg)
		if err != nil {
			logger.Error("failed to initialize database provider", "error", err)
			cancel()
			return nil, err
		}

		logger.Info("Running database migrations")
		if err := dbProvider.RunMigrations(ctx); err != nil {
			logger.Error("failed to run database migrations", "error", err)
			cancel()
			return nil, err
		}
		logger.Info("Database migrations completed successfully")

		database = dbProvider
	}

	appCtx := middlewares.NewAppContext(ctx, cfg, logger, cache, sessionManager, oidcProvider, database)

	router := setupRouter(appCtx)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	var debugServer *http.Server
	if cfg.Server.Debug != nil && cfg.Server.Debug.Enabled {
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

	if s.election != nil {
		go s.election.Start(*s.appCtx)

		go s.monitorLeadership(*s.appCtx)
	} else {
		timeInterval := calculateFetchInterval(s.cfg, s.cfg.Data.FallbackFetchInterval)
		go func() {
			if err := s.runBackgroundDataFetching(s.appCtx, timeInterval); err != nil {
				s.logger.Error("background data fetching stopped", "error", err)
			}
		}()
	}

	router := setupRouter(s.appCtx)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.cfg.Server.Port),
		Handler: router,
	}

	go func() {
		if s.cfg.Distributed != nil && s.cfg.Distributed.Enabled {
			s.logger.Info("Server starting", "port", s.cfg.Server.Port, "instance", s.election.InstanceID)
		} else {
			s.logger.Info("Server starting", "port", s.cfg.Server.Port)
		}
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("Server failed to start", "error", err)
			s.cancel()
		}
	}()

	if s.cfg.Server.Debug != nil && s.cfg.Server.Debug.Enabled {
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
	found := false

	for _, q := range cfg.Data.Queries {
		if q.Disabled || q.TTL <= 0 {
			continue
		}

		if !found || q.TTL < minTTL {
			minTTL = q.TTL
			found = true
		}
	}

	if !found {
		minTTL = defaultInterval
	}

	if minTTL < time.Second*30 {
		minTTL = time.Second * 30
	}
	return minTTL
}

func (s *Server) monitorLeadership(ctx context.Context) {
	var fetchCtx context.Context
	var fetchCancel context.CancelFunc

	ticker := time.NewTicker(s.cfg.Distributed.TTL / 3)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if fetchCancel != nil {
				s.logger.Info("stopping data fetching due to shutdown")
				fetchCancel()
			}
			return

		case <-ticker.C:
			isLeader := s.election.IsLeader()

			if isLeader && fetchCancel == nil {
				s.logger.Info("starting data fetching as leader")

				fetchCtx, fetchCancel = context.WithCancel(ctx)
				interval := calculateFetchInterval(s.cfg, s.cfg.Data.FallbackFetchInterval)

				go func(fctx context.Context) {
					if err := s.runBackgroundDataFetching(fctx, interval); err != nil {
						s.logger.Error("data fetching stopped", "error", err)
					}
				}(fetchCtx)

			} else if !isLeader && fetchCancel != nil {
				s.logger.Info("stopping data fetching, lost leadership")
				fetchCancel()
				fetchCancel = nil
			}
		}
	}
}

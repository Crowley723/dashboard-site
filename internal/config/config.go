package config

import (
	"fmt"
	"homelab-dashboard/internal/authorization"
	"log/slog"
	"net"
	"os"
	"slices"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config file path is required (use -config or -c)")
	}

	// Read and parse YAML
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	applyEnvironmentOverrides(&config)

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

var (
	EnvOIDCClientID             = "DASHBOARD_OIDC_CLIENT_ID"
	EnvOIDCClientSecret         = "DASHBOARD_OIDC_CLIENT_SECRET"
	EnvOIDCIssuerURL            = "DASHBOARD_OIDC_ISSUER_URL"
	EnvOIDCRedirectURL          = "DASHBOARD_OIDC_REDIRECT_URL"
	EnvDataPrometheusURL        = "DASHBOARD_DATA_PROMETHEUS_URL"
	EnvDataBasicAuthUsername    = "DASHBOARD_DATA_BASIC_AUTH_USERNAME"
	EnvDataBasicAuthPassword    = "DASHBOARD_DATA_BASIC_AUTH_PASSWORD"
	EnvRedisPassword            = "DASHBOARD_REDIS_PASSWORD"
	EnvRedisUsername            = "DASHBOARD_REDIS_USERNAME"
	EnvRedisSentinelUsername    = "DASHBOARD_REDIS_SENTINEL_USERNAME"
	EnvRedisSentinelPassword    = "DASHBOARD_REDIS_SENTINEL_PASSWORD"
	EnvMTLSDownloadTokenHMACKey = "DASHBOARD_MTLS_DOWNLOAD_TOKEN_HMAC_KEY"
	EnvStorageHost              = "DASHBOARD_STORAGE_HOST"
	EnvStoragePort              = "DASHBOARD_STORAGE_PORT"
	EnvStorageUsername          = "DASHBOARD_STORAGE_USERNAME"
	EnvStoragePassword          = "DASHBOARD_STORAGE_PASSWORD"
	EnvStorageDatabase          = "DASHBOARD_STORAGE_DATABASE"
	EnvFirewallRouterEndpoint   = "DASHBOARD_FIREWALL_ROUTER_ENDPOINT"
	EnvFirewallRouterAPIKey     = "DASHBOARD_FIREWALL_ROUTER_API_KEY"
	EnvFirewallRouterAPISecret  = "DASHBOARD_FIREWALL_ROUTER_API_SECRET"
)

func applyEnvironmentOverrides(config *Config) {
	if clientID := os.Getenv(EnvOIDCClientID); clientID != "" {
		config.OIDC.ClientID = clientID
	}

	if clientSecret := os.Getenv(EnvOIDCClientSecret); clientSecret != "" {
		config.OIDC.ClientSecret = clientSecret
	}

	if issuerURL := os.Getenv(EnvOIDCIssuerURL); issuerURL != "" {
		config.OIDC.IssuerURL = issuerURL
	}

	if redirectURL := os.Getenv(EnvOIDCRedirectURL); redirectURL != "" {
		config.OIDC.RedirectURI = redirectURL
	}

	if prometheusURL := os.Getenv(EnvDataPrometheusURL); prometheusURL != "" {
		config.Data.PrometheusURL = prometheusURL
	}

	if username := os.Getenv(EnvDataBasicAuthUsername); username != "" {
		if config.Data.BasicAuth == nil {
			config.Data.BasicAuth = &BasicAuth{}
		}
		config.Data.BasicAuth.Username = username
	}

	if password := os.Getenv(EnvDataBasicAuthPassword); password != "" {
		if config.Data.BasicAuth == nil {
			config.Data.BasicAuth = &BasicAuth{}
		}
		config.Data.BasicAuth.Password = password
	}

	if redisPassword := os.Getenv(EnvRedisPassword); redisPassword != "" {
		if config.Redis == nil {
			config.Redis = &RedisConfig{}
		}
		config.Redis.Password = redisPassword
	}

	if redisUsername := os.Getenv(EnvRedisUsername); redisUsername != "" {
		if config.Redis == nil {
			config.Redis = &RedisConfig{}
		}
		config.Redis.Username = redisUsername
	}

	if sentinelUsername := os.Getenv(EnvRedisSentinelUsername); sentinelUsername != "" {
		if config.Redis == nil {
			config.Redis = &RedisConfig{}
		}
		if config.Redis.Sentinel == nil {
			config.Redis.Sentinel = &RedisSentinelConfig{}
		}
		config.Redis.Sentinel.SentinelUsername = sentinelUsername
	}

	if sentinelPassword := os.Getenv(EnvRedisSentinelPassword); sentinelPassword != "" {
		if config.Redis == nil {
			config.Redis = &RedisConfig{}
		}
		if config.Redis.Sentinel == nil {
			config.Redis.Sentinel = &RedisSentinelConfig{}
		}
		config.Redis.Sentinel.SentinelPassword = sentinelPassword
	}

	if hmacKey := os.Getenv(EnvMTLSDownloadTokenHMACKey); hmacKey != "" {
		if config.Features == nil {
			config.Features = &FeaturesConfig{}
		}
		config.Features.MTLSManagement.DownloadTokenHMACKey = hmacKey
	}

	if host := os.Getenv(EnvStorageHost); host != "" {
		if config.Storage == nil {
			config.Storage = &StorageConfig{}
		}
		config.Storage.Host = host
	}

	if portStr := os.Getenv(EnvStoragePort); portStr != "" {
		if config.Storage == nil {
			config.Storage = &StorageConfig{}
		}
		if port, err := strconv.Atoi(portStr); err == nil {
			config.Storage.Port = port
		}
	}

	if username := os.Getenv(EnvStorageUsername); username != "" {
		if config.Storage == nil {
			config.Storage = &StorageConfig{}
		}
		config.Storage.Username = username
	}

	if password := os.Getenv(EnvStoragePassword); password != "" {
		if config.Storage == nil {
			config.Storage = &StorageConfig{}
		}
		config.Storage.Password = password
	}

	if database := os.Getenv(EnvStorageDatabase); database != "" {
		if config.Storage == nil {
			config.Storage = &StorageConfig{}
		}
		config.Storage.Database = database
	}

	if endpoint := os.Getenv(EnvFirewallRouterEndpoint); endpoint != "" {
		if config.Features == nil {
			config.Features = &FeaturesConfig{}
		}
		config.Features.FirewallManagement.RouterEndpoint = endpoint
	}

	if apiKey := os.Getenv(EnvFirewallRouterAPIKey); apiKey != "" {
		if config.Features == nil {
			config.Features = &FeaturesConfig{}
		}
		config.Features.FirewallManagement.RouterAPIKey = apiKey
	}

	if apiSecret := os.Getenv(EnvFirewallRouterAPISecret); apiSecret != "" {
		if config.Features == nil {
			config.Features = &FeaturesConfig{}
		}
		config.Features.FirewallManagement.RouterAPISecret = apiSecret
	}
}

func validateConfig(config *Config) error {

	err := config.validateServerConfig()
	if err != nil {
		return err
	}

	err = config.validateOIDCConfig()
	if err != nil {
		return err
	}

	err = config.validateLogConfig()
	if err != nil {
		return err
	}

	err = config.validateCORSConfig()
	if err != nil {
		return err
	}

	err = config.validateSessionConfig()
	if err != nil {
		return err
	}

	err = config.validateDataConfig()
	if err != nil {
		return err
	}

	err = config.validateCacheConfig()
	if err != nil {
		return err
	}

	if config.Cache.Type == "redis" || config.Sessions.Store == "redis" {
		err = config.validateRedisConfig()
		if err != nil {
			return err
		}
	}

	err = config.validateDistributedConfig()
	if err != nil {
		return err
	}

	err = config.validateStorageConfig()
	if err != nil {
		return err
	}

	err = config.validateAuthorizationConfig()
	if err != nil {
		return err
	}

	err = config.ValidateFeaturesConfig()
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) validateOIDCConfig() error {
	if c.OIDC.ClientID == "" {
		return fmt.Errorf("oidc client id is required")
	}

	if c.OIDC.ClientSecret == "" {
		return fmt.Errorf("OIDC clientSecret is required")
	}

	if err := validateURL(c.OIDC.IssuerURL, "issuer_url"); err != nil {
		return err
	}

	if err := validateURL(c.OIDC.RedirectURI, "redirect_url"); err != nil {
		return err
	}

	if len(c.OIDC.Scopes) == 0 {
		c.OIDC.Scopes = DefaultOIDCConfig.Scopes
	}

	return nil
}

func (c *Config) validateServerConfig() error {
	if c.Server.Port == 0 {
		c.Server.Port = DefaultServerConfig.Port
	}

	if c.Server.ExternalURL == "" {
		return fmt.Errorf("server.external_url is required")
	}

	if c.Server.Debug != nil && c.Server.Debug.Enabled {
		if c.Server.Debug.Host == "" {
			c.Server.Debug.Host = DefaultDebugConfig.Host
		}
		if c.Server.Debug.Port <= 0 || c.Server.Debug.Port >= 65535 {
			c.Server.Debug.Port = DefaultDebugConfig.Port
		}
	}

	return nil
}

func (c *Config) validateLogConfig() error {
	if c.Log.Format == "" {
		c.Log.Format = DefaultLogConfig.Format
	} else {
		switch c.Log.Format {
		case "text":
			c.Log.Format = "text"
		case "json":
			c.Log.Format = "json"
		default:
			return fmt.Errorf("invalid log format: %s, options are text or json", c.Log.Format)
		}
	}

	if c.Log.Level == "" {
		c.Log.Level = DefaultLogConfig.Level
	} else {
		switch c.Log.Level {
		case "debug":
			c.Log.Level = string(rune(slog.LevelDebug))
		case "info":
			c.Log.Level = string(rune(slog.LevelInfo))
		case "warn":
			c.Log.Level = string(rune(slog.LevelWarn))
		case "error":
			c.Log.Level = string(rune(slog.LevelError))
		default:
			return fmt.Errorf("invalid log level: %s, options are debug, info, warn, error", c.Log.Level)
		}
	}

	return nil
}

func (c *Config) validateCORSConfig() error {
	if len(c.CORS.AllowedOrigins) == 0 {
		c.CORS.AllowedOrigins = DefaultCORSConfig.AllowedOrigins
	}
	if len(c.CORS.AllowedMethods) == 0 {
		c.CORS.AllowedMethods = DefaultCORSConfig.AllowedMethods
	}
	if len(c.CORS.AllowedHeaders) == 0 {
		c.CORS.AllowedHeaders = DefaultCORSConfig.AllowedHeaders
	}
	if c.CORS.MaxAgeSeconds == 0 {
		c.CORS.MaxAgeSeconds = DefaultCORSConfig.MaxAgeSeconds
	}

	return nil
}

func (c *Config) validateSessionConfig() error {
	if c == nil {
		return fmt.Errorf("session config is required")
	}

	if c.Sessions.Store == "" {
		c.Sessions.Store = DefaultSessionConfig.Store
	} else {
		switch c.Sessions.Store {
		case "memory":
			c.Sessions.Store = "memory"
		case "redis":
			c.Sessions.Store = "redis"
		default:
			return fmt.Errorf("invalid session store: %s, options are 'memory' or 'redis'", c.Sessions.Store)
		}
	}

	if c.Sessions.DurationSource == "" {
		c.Sessions.DurationSource = DefaultSessionConfig.DurationSource
	} else {
		switch c.Sessions.DurationSource {
		case "fixed":
			c.Sessions.DurationSource = "fixed"
		case "oidc_tokens":
			c.Sessions.DurationSource = "oidc_tokens"
		default:
			return fmt.Errorf("invalid session duration source: %s, options are 'fixed' or 'oidc_tokens'", c.Sessions.DurationSource)
		}
	}

	if c.Sessions.Name == "" {
		c.Sessions.Name = DefaultSessionConfig.Name
	}

	if c.Sessions.FixedTimeout == 0 {
		c.Sessions.FixedTimeout = DefaultSessionConfig.FixedTimeout
	}

	return nil
}

func (c *Config) validateDataConfig() (err error) {
	if c.Data.PrometheusURL == "" {
		return fmt.Errorf("data.prometheus_url is required")
	}

	if c.Data.BasicAuth != nil {
		if c.Data.BasicAuth.Username == "" {
			return fmt.Errorf("data.basic_auth.username is required")
		}
		if c.Data.BasicAuth.Password == "" {
			return fmt.Errorf("data.basic_auth.password is required")
		}
	}

	if len(c.Data.Queries) > 0 {
		if err = c.validateDataQueriesConfig(); err != nil {
			return err
		}
	}

	if c.Data.FallbackFetchInterval.Seconds() < 0 {
		c.Data.FallbackFetchInterval = defaultDataConfig.FallbackFetchInterval
	} else if c.Data.FallbackFetchInterval.Seconds() < 30 && c.Data.FallbackFetchInterval.Seconds() > 0 {
		return fmt.Errorf("data.fallback_fetch_interval cannot be less than 30 seconds")
	}

	return nil
}

func (c *Config) validateDataQueriesConfig() (err error) {

	queries := c.Data.Queries

	for i, query := range queries {
		if query.Disabled {
			continue
		}

		if query.Name == "" {
			return fmt.Errorf("data.queries[%d].name is required", i)
		}

		if query.Query == "" {
			return fmt.Errorf("data.queries[%d].query is required", i)
		}

		if query.TTL.Seconds() == 0 {
			queries[i].TTL = 30 * time.Second
		} else if query.TTL.Seconds() < 30 {
			return fmt.Errorf("data.queries[%d].ttl cannot be less than 30s", i)
		}

		if query.Type == "range" {
			if query.Range == "" {
				return fmt.Errorf("data.queries[%d].range is required for range queries", i)
			}

			if query.Step == "" {
				return fmt.Errorf("data.queries[%d].step is required for range queries", i)
			}
		} else if query.Type != "" {
			return fmt.Errorf("invalid query type: %s", query.Type)
		}
	}

	return nil
}

func (c *Config) validateCacheConfig() error {
	if c.Cache.Type == "" {
		c.Cache.Type = "memory"
	}

	switch c.Cache.Type {
	case "memory":
		break
	case "redis":
		if c.Redis == nil {
			return fmt.Errorf("redis configuration must be enabled to use redis for data cache")
		}
	default:
		return fmt.Errorf("invalid cache type: %s, must be 'memory' or 'redis'", c.Cache.Type)
	}

	return nil
}

func (c *Config) validateRedisConfig() error {
	if c.Redis == nil {
		return fmt.Errorf("redis config is nil")
	}

	if c.Redis.Address == "" {
		return fmt.Errorf("redis address is required")
	}

	if _, _, err := net.SplitHostPort(c.Redis.Address); err != nil {
		return fmt.Errorf("invalid redis address format (expected host:port): %w", err)
	}

	// Apply default indices if not set
	if c.Redis.SessionIndex == 0 && c.Redis.CacheIndex == 0 && c.Redis.LeaderIndex == 0 {
		c.Redis.SessionIndex = DefaultRedisConfig.SessionIndex
		c.Redis.CacheIndex = DefaultRedisConfig.CacheIndex
		c.Redis.LeaderIndex = DefaultRedisConfig.LeaderIndex
	}

	if c.Redis.SessionIndex < 0 {
		return fmt.Errorf("redis session_index must be non-negative, got %d", c.Redis.SessionIndex)
	}

	if c.Redis.CacheIndex < 0 {
		return fmt.Errorf("redis cache_index must be non-negative, got %d", c.Redis.CacheIndex)
	}

	if c.Redis.LeaderIndex < 0 {
		return fmt.Errorf("redis leader_index must be non-negative, got %d", c.Redis.CacheIndex)
	}

	if c.Redis.SessionIndex == c.Redis.CacheIndex {
		return fmt.Errorf("redis session_index and cache_index should be different to avoid data collision (both are %d)", c.Redis.SessionIndex)
	}

	if c.Redis.LeaderIndex == c.Redis.CacheIndex {
		return fmt.Errorf("redis leader_index and cache_index should be different to avoid data collision (both are %d)", c.Redis.LeaderIndex)
	}

	if c.Redis.LeaderIndex == c.Redis.SessionIndex {
		return fmt.Errorf("redis leader_index and session_index should be different to avoid data collision (both are %d)", c.Redis.LeaderIndex)
	}

	const maxRedisDB = 15
	if c.Redis.SessionIndex > maxRedisDB {
		return fmt.Errorf("redis session_index %d exceeds typical maximum of %d", c.Redis.SessionIndex, maxRedisDB)
	}

	if c.Redis.CacheIndex > maxRedisDB {
		return fmt.Errorf("redis cache_index %d exceeds typical maximum of %d", c.Redis.CacheIndex, maxRedisDB)
	}

	if c.Redis.LeaderIndex > maxRedisDB {
		return fmt.Errorf("redis leader_index %d exceeds typical maximum of %d", c.Redis.LeaderIndex, maxRedisDB)
	}

	if c.Redis.Sentinel != nil {
		if c.Redis.Sentinel.MasterName == "" {
			return fmt.Errorf("sentinel master_name is required")
		}
		if len(c.Redis.Sentinel.SentinelAddresses) == 0 {
			return fmt.Errorf("at least one sentinel address is required")
		}
	}
	return nil
}

func (c *Config) validateDistributedConfig() error {
	if c.Distributed == nil {
		return nil
	}

	// Apply default enabled state if not explicitly set
	if !c.Distributed.Enabled {
		return nil
	}

	if c.Distributed.TTL.Seconds() <= 0 {
		c.Distributed.TTL = DefaultDistributedConfig.TTL
	} else if c.Distributed.TTL > time.Minute {
		return fmt.Errorf("distributed ttl cannot be more than 1 minute")
	}

	return nil
}

func (c *Config) validateStorageConfig() error {
	if c.Storage == nil || !c.Storage.Enabled {
		return nil
	}

	if c.Storage.Host == "" {
		return fmt.Errorf("storage.host is required when storage is enabled")
	}

	if c.Storage.Port <= 0 || c.Storage.Port > 65535 {
		return fmt.Errorf("storage.port must be between 1 and 65535, got %d", c.Storage.Port)
	}

	if c.Storage.Database == "" {
		return fmt.Errorf("storage.database is required when storage is enabled")
	}

	return nil
}

func (c *Config) ValidateFeaturesConfig() error {
	// Initialize Features with defaults if not set
	if c.Features == nil {
		defaultConfig := DefaultFeaturesConfig
		c.Features = &defaultConfig
	}

	if c.Features.MTLSManagement.Enabled {
		if err := c.ValidateMTLSManagementConfig(); err != nil {
			return err
		}
	}

	if c.Features.FirewallManagement.Enabled {
		if err := c.ValidateFirewallManagementConfig(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) ValidateMTLSManagementConfig() error {
	if c.Features == nil || !c.Features.MTLSManagement.Enabled {
		return nil
	}

	// mTLS management requires storage to be enabled
	if c.Storage == nil || !c.Storage.Enabled {
		return fmt.Errorf("storage must be enabled when mtls_management is enabled")
	}

	if c.Features.MTLSManagement.DownloadTokenHMACKey == "" {
		return fmt.Errorf("features.mtls_management.download_token_hmac_key is required when mtls_management is enabled")
	}

	if len(c.Features.MTLSManagement.DownloadTokenHMACKey) <= 32 {
		return fmt.Errorf("features.mtls_management.download_token_hmac_key must be at least 32 characters")
	}

	// Apply default validity days if not set
	if c.Features.MTLSManagement.MinCertificateValidityDays == 0 {
		c.Features.MTLSManagement.MinCertificateValidityDays = DefaultMTLSIssuerConfig.MinCertificateValidityDays
	}

	if c.Features.MTLSManagement.MaxCertificateValidityDays == 0 {
		c.Features.MTLSManagement.MaxCertificateValidityDays = DefaultMTLSIssuerConfig.MaxCertificateValidityDays
	}

	if c.Features.MTLSManagement.MaxCertificateValidityDays < c.Features.MTLSManagement.MinCertificateValidityDays {
		return fmt.Errorf("features.mtls_management.max_certificate_validity_days cannot be less than min_certificate_validity_days")
	}

	// Apply default Kubernetes configuration if not set
	if c.Features.MTLSManagement.Kubernetes == nil {
		c.Features.MTLSManagement.Kubernetes = DefaultKubernetesConfig
	}

	if c.Features.MTLSManagement.Kubernetes.Namespace == "" {
		c.Features.MTLSManagement.Kubernetes.Namespace = DefaultKubernetesConfig.Namespace
	}

	if c.Features.MTLSManagement.CertificateSubject == nil {
		c.Features.MTLSManagement.CertificateSubject = DefaultCertificateSubject
	}

	if c.Features.MTLSManagement.CertificateSubject.Organization == "" {
		c.Features.MTLSManagement.CertificateSubject.Organization = DefaultCertificateSubject.Organization
	}

	// Apply default background job config if not set
	if c.Features.MTLSManagement.BackgroundJobConfig == nil {
		c.Features.MTLSManagement.BackgroundJobConfig = DefaultMTLSBackgroundJobConfig
	}

	if c.Features.MTLSManagement.BackgroundJobConfig.ApprovedCertificatePollingInterval == 0 {
		c.Features.MTLSManagement.BackgroundJobConfig.ApprovedCertificatePollingInterval = DefaultMTLSBackgroundJobConfig.ApprovedCertificatePollingInterval
	}

	if c.Features.MTLSManagement.BackgroundJobConfig.IssuedCertificatePollingInterval == 0 {
		c.Features.MTLSManagement.BackgroundJobConfig.IssuedCertificatePollingInterval = DefaultMTLSBackgroundJobConfig.IssuedCertificatePollingInterval
	}

	// Validate certificate issuer configuration
	if c.Features.MTLSManagement.CertificateIssuer == nil {
		return fmt.Errorf("features.mtls_management.certificate_issuer is required when mtls_management is enabled")
	}

	if c.Features.MTLSManagement.CertificateIssuer.Name == "" {
		return fmt.Errorf("features.mtls_management.certificate_issuer.name is required when mtls_management is enabled")
	}

	if c.Features.MTLSManagement.CertificateIssuer.Kind == "" {
		return fmt.Errorf("features.mtls_management.certificate_issuer.kind is required when mtls_management is enabled")
	}

	// Validate issuer kind is either Issuer or ClusterIssuer
	kind := c.Features.MTLSManagement.CertificateIssuer.Kind
	if kind != "Issuer" && kind != "ClusterIssuer" {
		return fmt.Errorf("features.mtls_management.certificate_issuer.kind must be either 'Issuer' or 'ClusterIssuer', got '%s'", kind)
	}

	return nil
}

func (c *Config) ValidateFirewallManagementConfig() error {
	if c.Features == nil || !c.Features.FirewallManagement.Enabled {
		return nil
	}

	if c.Storage == nil || !c.Storage.Enabled {
		return fmt.Errorf("storage must be enabled when firewall_management is enabled")
	}

	if c.Features.FirewallManagement.RouterEndpoint == "" {
		return fmt.Errorf("features.firewall_management.router_endpoint is required when firewall_management is enabled")
	}

	if c.Features.FirewallManagement.RouterEndpoint[len(c.Features.FirewallManagement.RouterEndpoint)-1] == '/' {
		c.Features.FirewallManagement.RouterEndpoint = c.Features.FirewallManagement.RouterEndpoint[:len(c.Features.FirewallManagement.RouterEndpoint)-1]
	}

	if err := validateURL(c.Features.FirewallManagement.RouterEndpoint, "features.firewall_management.router_endpoint"); err != nil {
		return err
	}

	if c.Features.FirewallManagement.RouterAPIKey == "" {
		return fmt.Errorf("features.firewall_management.router_api_key is required when firewall_management is enabled")
	}

	if c.Features.FirewallManagement.RouterAPISecret == "" {
		return fmt.Errorf("features.firewall_management.router_api_secret is required when firewall_management is enabled")
	}

	if c.Features.FirewallManagement.BackgroundJobConfig == nil {
		c.Features.FirewallManagement.BackgroundJobConfig = DefaultFirewallBackgroundJobConfig
	}

	if c.Features.FirewallManagement.BackgroundJobConfig.SyncInterval == 0 {
		c.Features.FirewallManagement.BackgroundJobConfig.SyncInterval = DefaultFirewallBackgroundJobConfig.SyncInterval
	}

	if c.Features.FirewallManagement.BackgroundJobConfig.SyncInterval < 30*time.Second {
		return fmt.Errorf("features.firewall_management.background_job_config.sync_interval cannot be less than 30 seconds")
	}

	if c.Features.FirewallManagement.BackgroundJobConfig.ExpirationInterval == 0 {
		c.Features.FirewallManagement.BackgroundJobConfig.ExpirationInterval = DefaultFirewallBackgroundJobConfig.ExpirationInterval
	}

	if c.Features.FirewallManagement.BackgroundJobConfig.ExpirationInterval < 1*time.Minute {
		return fmt.Errorf("features.firewall_management.background_job_config.expiration_interval cannot be less than 1 minute")
	}

	if len(c.Features.FirewallManagement.Aliases) == 0 {
		return fmt.Errorf("features.firewall_management.aliases must have at least one alias configured when firewall_management is enabled")
	}

	aliasNames := make(map[string]bool)
	for i, alias := range c.Features.FirewallManagement.Aliases {
		if alias.Name == "" {
			return fmt.Errorf("features.firewall_management.aliases[%d].name is required", i)
		}

		if alias.UUID == "" {
			return fmt.Errorf("features.firewall_management.aliases[%d].uuid is required", i)
		}

		if len(alias.UUID) != 36 {
			return fmt.Errorf("features.firewall_management.aliases[%d].uuid must be a valid UUID", i)
		}

		if alias.AuthGroup == "" {
			return fmt.Errorf("features.firewall_management.aliases[%d].auth_group is required", i)
		}

		if _, exists := c.Authorization.GroupScopes[alias.AuthGroup]; !exists {
			return fmt.Errorf("features.firewall_management.aliases[%d].auth_group '%s' does not exist in authorization.group_scopes", i, alias.AuthGroup)
		}

		if alias.MaxIPsPerUser <= 0 {
			return fmt.Errorf("features.firewall_management.aliases[%d].max_ips_per_user must be greater than 0", i)
		}

		if alias.MaxTotalIPs <= 0 {
			return fmt.Errorf("features.firewall_management.aliases[%d].max_total_ips must be greater than 0", i)
		}

		if alias.MaxIPsPerUser > alias.MaxTotalIPs {
			return fmt.Errorf("features.firewall_management.aliases[%d].max_ips_per_user (%d) cannot be greater than max_total_ips (%d)",
				i, alias.MaxIPsPerUser, alias.MaxTotalIPs)
		}

		if alias.DefaultTTL != nil && *alias.DefaultTTL < 1*time.Hour {
			return fmt.Errorf("features.firewall_management.aliases[%d].default_ttl cannot be less than 1 hour if set", i)
		}

		key := fmt.Sprintf("%s:%s", alias.Name, alias.UUID)
		if aliasNames[key] {
			// This is actually intentional - same UUID with different groups/limits
			// So we just note it but don't error
		}
		aliasNames[key] = true
	}

	return nil
}

func (c *Config) validateAuthorizationConfig() error {
	// Apply default authorization config if not set
	if c.Authorization.GroupScopes == nil || len(c.Authorization.GroupScopes) == 0 {
		c.Authorization = DefaultAuthorizationConfig
	}

	// Validate that each scope is a known/valid scope
	validScopes := authorization.GetAllValidScopes()
	for group, scopes := range c.Authorization.GroupScopes {
		if len(scopes) == 0 {
			return fmt.Errorf("authorization group '%s' has no scopes defined", group)
		}

		for _, scope := range scopes {
			if !slices.Contains(validScopes, scope) {
				return fmt.Errorf("authorization group '%s' contains invalid scope '%s'", group, scope)
			}
		}
	}

	return nil
}

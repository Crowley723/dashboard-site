package config

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadConfig() (*Config, error) {
	configPath := flag.String("config", "", "path to config file")
	flag.StringVar(configPath, "c", "", "path to config file (shorthand)")
	flag.Parse()

	if *configPath == "" {
		return nil, fmt.Errorf("config file path is required (use -config or -c)")
	}

	// Read and parse YAML
	data, err := os.ReadFile(*configPath)
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
	EnvOIDCClientID          = "DASHBOARD_OIDC_CLIENT_ID"
	EnvOIDCClientSecret      = "DASHBOARD_OIDC_CLIENT_SECRET"
	EnvOIDCIssuerURL         = "DASHBOARD_OIDC_ISSUER_URL"
	EnvOIDCRedirectURL       = "DASHBOARD_OIDC_REDIRECT_URL"
	EnvDataPrometheusURL     = "DASHBOARD_DATA_PROMETHEUS_URL"
	EnvDataBasicAuthUsername = "DASHBOARD_DATA_BASIC_AUTH_USERNAME"
	EnvDataBasicAuthPassword = "DASHBOARD_DATA_BASIC_AUTH_PASSWORD"
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

	err = config.validateRedisConfig()
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

	if c.Server.Debug != nil || c.Server.Debug.Enabled {
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
		c.Sessions.Store = "memory"
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

	if c.Redis.SessionIndex < 0 {
		return fmt.Errorf("redis session_index must be non-negative, got %d", c.Redis.SessionIndex)
	}

	if c.Redis.CacheIndex < 0 {
		return fmt.Errorf("redis cache_index must be non-negative, got %d", c.Redis.CacheIndex)
	}

	if c.Redis.LeaderIndex < 0 {
		return fmt.Errorf("redis cache_index must be non-negative, got %d", c.Redis.CacheIndex)
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
	return nil
}

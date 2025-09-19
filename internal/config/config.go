package config

import (
	"flag"
	"fmt"
	"log/slog"
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

	err := validateServerConfig(&config.Server)
	if err != nil {
		return err
	}

	err = validateOIDCConfig(&config.OIDC)
	if err != nil {
		return err
	}

	err = validateLogConfig(&config.Log)
	if err != nil {
		return err
	}

	err = validateCORSConfig(&config.CORS)
	if err != nil {
		return err
	}

	err = validateSessionConfig(&config.Sessions)
	if err != nil {
		return err
	}

	err = validateDataConfig(&config.Data)
	if err != nil {
		return err
	}

	err = validateCacheConfig(&config.Cache)

	return nil
}

func validateOIDCConfig(oidcConfig *OIDCConfig) error {
	if oidcConfig == nil {
		return fmt.Errorf("oidc config is required")
	}

	if oidcConfig.ClientID == "" {
		return fmt.Errorf("oidc client id is required")
	}

	if oidcConfig.ClientSecret == "" {
		return fmt.Errorf("OIDC clientSecret is required")
	}

	if err := validateURL(oidcConfig.IssuerURL, "issuer_url"); err != nil {
		return err
	}

	if err := validateURL(oidcConfig.RedirectURI, "redirect_url"); err != nil {
		return err
	}

	if len(oidcConfig.Scopes) == 0 {
		oidcConfig.Scopes = DefaultOIDCConfig.Scopes
	}

	return nil
}

func validateServerConfig(config *ServerConfig) error {
	if config == nil {
		return fmt.Errorf("server config is required")
	}

	if config.Port == 0 {
		config.Port = DefaultServerConfig.Port
	}

	return nil
}

func validateLogConfig(config *LogConfig) error {
	if config.Format == "" {
		config.Format = DefaultLogConfig.Format
	} else {
		switch config.Format {
		case "text":
			config.Format = "text"
		case "json":
			config.Format = "json"
		default:
			return fmt.Errorf("invalid log format: %s, options are text or json", config.Format)
		}
	}

	if config.Level == "" {
		config.Level = DefaultLogConfig.Level
	} else {
		switch config.Level {
		case "debug":
			config.Level = string(rune(slog.LevelDebug))
		case "info":
			config.Level = string(rune(slog.LevelInfo))
		case "warn":
			config.Level = string(rune(slog.LevelWarn))
		case "error":
			config.Level = string(rune(slog.LevelError))
		default:
			return fmt.Errorf("invalid log level: %s, options are debug, info, warn, error", config.Level)
		}
	}

	return nil
}

func validateCORSConfig(config *CORSConfig) error {
	if len(config.AllowedOrigins) == 0 {
		config.AllowedOrigins = DefaultCORSConfig.AllowedOrigins
	}
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = DefaultCORSConfig.AllowedMethods
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = DefaultCORSConfig.AllowedHeaders
	}
	if config.MaxAgeSeconds == 0 {
		config.MaxAgeSeconds = DefaultCORSConfig.MaxAgeSeconds
	}

	return nil
}

func validateSessionConfig(config *SessionConfig) error {
	if config == nil {
		return fmt.Errorf("session config is required")
	}

	if config.Store == "" {
		config.Store = "memory"
	} else {
		switch config.Store {
		case "memory":
			config.Store = "memory"
		case "redis":
			config.Store = "redis"
		default:
			return fmt.Errorf("invalid session store: %s, options are 'memory' or 'redis'", config.Store)
		}
	}

	if config.DurationSource == "" {
		config.DurationSource = DefaultSessionConfig.DurationSource
	} else {
		switch config.DurationSource {
		case "fixed":
			config.DurationSource = "fixed"
		case "oidc_tokens":
			config.DurationSource = "oidc_tokens"
		default:
			return fmt.Errorf("invalid session duration source: %s, options are 'fixed' or 'oidc_tokens'", config.DurationSource)
		}
	}

	if config.Name == "" {
		config.Name = DefaultSessionConfig.Name
	}

	if config.FixedTimeout == 0 {
		config.FixedTimeout = DefaultSessionConfig.FixedTimeout
	}

	return nil
}

func validateDataConfig(config *DataConfig) (err error) {
	if config == nil {
		return fmt.Errorf("data config is required")
	}

	if config.PrometheusURL == "" {
		return fmt.Errorf("data.prometheus_url is required")
	}

	if config.BasicAuth != nil {
		if config.BasicAuth.Username == "" {
			return fmt.Errorf("data.basic_auth.username is required")
		}
		if config.BasicAuth.Password == "" {
			return fmt.Errorf("data.basic_auth.password is required")
		}
	}

	if config.TimeInterval == 0 {
		config.TimeInterval, err = time.ParseDuration("1h")
		if err != nil {
			return fmt.Errorf("unable to parse default duration: %v", err)
		}
	}

	if len(config.Queries) > 0 {
		if err = validateDataQueriesConfig(config); err != nil {
			return err
		}
	}

	return nil
}

func validateDataQueriesConfig(config *DataConfig) (err error) {

	queries := config.Queries

	for i, query := range queries {

		if query.Name == "" {
			return fmt.Errorf("data.queries[%d].name is required", i)
		}

		if query.Query == "" {
			return fmt.Errorf("data.queries[%d].query is required", i)
		}

		if query.TTL.Seconds() == 0 {
			query.TTL = config.TimeInterval
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

func validateCacheConfig(c *CacheConfig) error {
	switch c.Type {
	case "memory":
		break
	case "redis":
		if c.Redis.Address == "" {
			return fmt.Errorf("redis.address is required")
		}

		if c.Redis.Password == "" {
			return fmt.Errorf("redis.password is required")
		}
	default:
		return fmt.Errorf("invalid cache type: %s, must be 'memory' or 'redis'", c.Type)
	}

	return nil
}

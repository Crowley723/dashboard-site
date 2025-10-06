package config

import (
	"time"
)

type Config struct {
	Server      ServerConfig       `yaml:"server"`
	OIDC        OIDCConfig         `yaml:"oidc"`
	Log         LogConfig          `yaml:"log"`
	CORS        CORSConfig         `yaml:"cors"`
	Sessions    SessionConfig      `yaml:"sessions"`
	Data        DataConfig         `yaml:"data"`
	Cache       CacheConfig        `yaml:"cache"`
	Redis       *RedisConfig       `yaml:"redis"`
	Distributed *DistributedConfig `yaml:"distributed"`
}

type ServerConfig struct {
	Port  int                `yaml:"port"`
	Debug *ServerDebugConfig `yaml:"debug"`
}

var DefaultServerConfig = ServerConfig{
	Port: 8080,
}

type ServerDebugConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
}

var DefaultDebugConfig = ServerDebugConfig{
	Enabled: false,
	Host:    "localhost",
	Port:    5123,
}

type OIDCConfig struct {
	ClientID     string   `yaml:"client_id"`
	ClientSecret string   `yaml:"client_secret"`
	IssuerURL    string   `yaml:"issuer_url"`
	RedirectURI  string   `yaml:"redirect_url"`
	Scopes       []string `yaml:"scopes"`
}

var DefaultOIDCConfig = OIDCConfig{
	Scopes: []string{"openid", "profile", "email", "groups"},
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

var DefaultLogConfig = LogConfig{
	Level:  "info",
	Format: "text",
}

type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAgeSeconds    int      `yaml:"max_age_seconds"`
}

var DefaultCORSConfig = CORSConfig{
	AllowedOrigins: []string{"http://localhost:5173"},
	AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowedHeaders: []string{"*"},
	MaxAgeSeconds:  300,
}

type SessionConfig struct {
	Store          string        `yaml:"store"`
	DurationSource string        `yaml:"duration_source"`
	FixedTimeout   time.Duration `yaml:"fixed_timeout"`
	Name           string        `yaml:"name"`
	Secure         bool          `yaml:"secure"`
}

var DefaultSessionConfig = SessionConfig{
	Store:          "memory",
	DurationSource: "fixed",
	FixedTimeout:   24 * time.Hour,
	Name:           "session_id",
	Secure:         true,
}

type DataConfig struct {
	PrometheusURL         string            `yaml:"prometheus_url"`
	BasicAuth             *BasicAuth        `yaml:"basic_auth"`
	Queries               []PrometheusQuery `yaml:"queries"`
	FallbackFetchInterval time.Duration     `yaml:"fallback_fetch_interval"`
}

var defaultDataConfig = DataConfig{
	FallbackFetchInterval: 10 * time.Minute,
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type PrometheusQuery struct {
	Name          string        `yaml:"name"`
	Disabled      bool          `yaml:"disabled"`
	Query         string        `yaml:"query"`
	Type          string        `yaml:"type"`
	TTL           time.Duration `yaml:"ttl"`
	Range         string        `yaml:"range"`
	Step          string        `yaml:"step"`
	RequireAuth   bool          `yaml:"require_auth"`
	RequiredGroup string        `yaml:"required_group"`
}

type CacheConfig struct {
	Type string `yaml:"type"` //  "memory" or "redis"
}

type RedisConfig struct {
	Address      string               `yaml:"address"`
	Username     string               `yaml:"username"`
	Password     string               `yaml:"password"`
	Sentinel     *RedisSentinelConfig `yaml:"sentinel"`
	SessionIndex int                  `yaml:"session_index"`
	CacheIndex   int                  `yaml:"cache_index"`
	LeaderIndex  int                  `yaml:"leader_index"`
}

var DefaultRedisConfig = RedisConfig{
	SessionIndex: 0,
	CacheIndex:   1,
	LeaderIndex:  2,
}

type RedisSentinelConfig struct {
	MasterName        string   `yaml:"master_name"`
	SentinelAddresses []string `yaml:"addresses"`
	SentinelPassword  string   `yaml:"password"`
	SentinelUsername  string   `yaml:"username"`
}

type DistributedConfig struct {
	Enabled bool          `yaml:"enabled"`
	TTL     time.Duration `yaml:"ttl"`
}

var DefaultDistributedConfig = DistributedConfig{
	Enabled: false,
	TTL:     30 * time.Second,
}

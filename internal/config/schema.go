package config

import (
	"time"
)

type Config struct {
	Server   ServerConfig  `yaml:"server"`
	OIDC     OIDCConfig    `yaml:"oidc"`
	Log      LogConfig     `yaml:"log"`
	CORS     CORSConfig    `yaml:"cors"`
	Sessions SessionConfig `yaml:"sessions"`
	Data     DataConfig    `yaml:"data"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

var DefaultServerConfig = ServerConfig{
	Port: 8080,
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
	PrometheusURL string            `yaml:"prometheus_url"`
	TimeInterval  time.Duration     `yaml:"time_interval"`
	BasicAuth     *BasicAuth        `yaml:"basic_auth"`
	Queries       []PrometheusQuery `yaml:"queries"`
}

type BasicAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type PrometheusQuery struct {
	Name          string        `yaml:"name"`
	Query         string        `yaml:"query"`
	Type          string        `yaml:"type"`
	TTL           time.Duration `yaml:"ttl"`
	Range         string        `yaml:"range"`
	Step          string        `yaml:"step"`
	RequireAuth   bool          `yaml:"require_auth"`
	RequiredGroup string        `yaml:"required_group"`
}

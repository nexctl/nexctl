package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config stores all server runtime configuration.
type Config struct {
	App   AppConfig   `yaml:"app"`
	MySQL MySQLConfig `yaml:"mysql"`
	Redis RedisConfig `yaml:"redis"`
	Auth  AuthConfig  `yaml:"auth"`
	Node  NodeConfig  `yaml:"node"`
}

// AppConfig stores generic application settings.
type AppConfig struct {
	Name                    string   `yaml:"name"`
	Env                     string   `yaml:"env"`
	ListenAddr              string   `yaml:"listen_addr"`
	ExternalURL             string   `yaml:"external_url"`
	ShutdownTimeoutSeconds  int      `yaml:"shutdown_timeout_seconds"`
	WebSocketAllowedOrigins []string `yaml:"websocket_allowed_origins"`
}

// MySQLConfig stores MySQL connection settings.
type MySQLConfig struct {
	DSN                    string `yaml:"dsn"`
	MaxOpenConns           int    `yaml:"max_open_conns"`
	MaxIdleConns           int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeMinutes int    `yaml:"conn_max_lifetime_minutes"`
}

// RedisConfig stores Redis connection settings.
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// AuthConfig stores JWT settings.
type AuthConfig struct {
	JWTSecret      string `yaml:"jwt_secret"`
	JWTExpireHours int    `yaml:"jwt_expire_hours"`
}

// NodeConfig stores node session and runtime settings.
type NodeConfig struct {
	HeartbeatTimeoutSeconds int `yaml:"heartbeat_timeout_seconds"`
	UnstableTimeoutSeconds  int `yaml:"unstable_timeout_seconds"`
	RuntimePointsTTLSeconds int `yaml:"runtime_points_ttl_seconds"`
	RuntimePointsMaxCount   int `yaml:"runtime_points_max_count"`
}

// Load loads config from yaml file and overlays supported env vars.
func Load(path string) (Config, error) {
	var cfg Config
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, fmt.Errorf("unmarshal config: %w", err)
	}
	if v := os.Getenv("OPSPILOT_SERVER_LISTEN_ADDR"); v != "" {
		cfg.App.ListenAddr = v
	}
	if v := os.Getenv("OPSPILOT_SERVER_EXTERNAL_URL"); v != "" {
		cfg.App.ExternalURL = v
	}
	if v := os.Getenv("OPSPILOT_MYSQL_DSN"); v != "" {
		cfg.MySQL.DSN = v
	}
	if v := os.Getenv("OPSPILOT_REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("OPSPILOT_REDIS_PASSWORD"); v != "" {
		cfg.Redis.Password = v
	}
	if v := os.Getenv("OPSPILOT_REDIS_DB"); v != "" {
		db, err := strconv.Atoi(v)
		if err != nil {
			return cfg, fmt.Errorf("parse OPSPILOT_REDIS_DB: %w", err)
		}
		cfg.Redis.DB = db
	}
	if v := os.Getenv("OPSPILOT_JWT_SECRET"); v != "" {
		cfg.Auth.JWTSecret = v
	}
	if v := os.Getenv("OPSPILOT_WEBSOCKET_ALLOWED_ORIGINS"); v != "" {
		cfg.App.WebSocketAllowedOrigins = splitCommaTrim(v)
	}
	return cfg, nil
}

func splitCommaTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// ShutdownTimeout returns the configured graceful shutdown timeout.
func (c Config) ShutdownTimeout() time.Duration {
	if c.App.ShutdownTimeoutSeconds <= 0 {
		return 10 * time.Second
	}
	return time.Duration(c.App.ShutdownTimeoutSeconds) * time.Second
}

// Package config provides application-wide configuration loaded from environment variables.
// All configuration is immutable after initialization — no global mutable state.
package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all application configuration parsed from environment variables.
// Each field maps to a specific .env key documented in .env.example.
type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	WebSocket WebSocketConfig
	Fetcher   FetcherConfig
}

// AppConfig holds HTTP server settings.
type AppConfig struct {
	Env  string
	Port string
}

// DatabaseConfig holds MySQL connection parameters and pool settings.
type DatabaseConfig struct {
	Host                  string
	Port                  string
	User                  string
	Password              string
	Name                  string
	Charset               string
	ParseTime             bool
	Loc                   string
	MaxOpenConns          int
	MaxIdleConns          int
	ConnMaxLifetimeMinutes int
}

// WebSocketConfig holds Gorilla WebSocket buffer and channel settings.
type WebSocketConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	ClientSendBuffer int
}

// FetcherConfig controls the Yahoo Finance real data fetcher.
type FetcherConfig struct {
	QuoteIntervalSec int
	OHLCVIntervalSec int
	Enabled          bool
}

// Load reads .env (if present) then parses all environment variables.
// Returns an error if any required variable is missing or malformed.
func Load() (*Config, error) {
	// godotenv.Load silently skips if .env does not exist (production case).
	_ = godotenv.Load()

	cfg := &Config{}

	if err := cfg.loadApp(); err != nil {
		return nil, fmt.Errorf("config.Load app: %w", err)
	}
	if err := cfg.loadDatabase(); err != nil {
		return nil, fmt.Errorf("config.Load database: %w", err)
	}
	if err := cfg.loadWebSocket(); err != nil {
		return nil, fmt.Errorf("config.Load websocket: %w", err)
	}
	if err := cfg.loadFetcher(); err != nil {
		return nil, fmt.Errorf("config.Load fetcher: %w", err)
	}

	return cfg, nil
}

func (c *Config) loadApp() error {
	c.App = AppConfig{
		Env:  getEnvOrDefault("APP_ENV", "development"),
		Port: getEnvOrDefault("APP_PORT", "8080"),
	}
	return nil
}

func (c *Config) loadDatabase() error {
	parseTime, err := strconv.ParseBool(getEnvOrDefault("DB_PARSE_TIME", "true"))
	if err != nil {
		return fmt.Errorf("DB_PARSE_TIME is not a valid bool: %w", err)
	}

	maxOpen, err := strconv.Atoi(getEnvOrDefault("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return fmt.Errorf("DB_MAX_OPEN_CONNS is not a valid int: %w", err)
	}

	maxIdle, err := strconv.Atoi(getEnvOrDefault("DB_MAX_IDLE_CONNS", "5"))
	if err != nil {
		return fmt.Errorf("DB_MAX_IDLE_CONNS is not a valid int: %w", err)
	}

	connLifetime, err := strconv.Atoi(getEnvOrDefault("DB_CONN_MAX_LIFETIME_MINUTES", "5"))
	if err != nil {
		return fmt.Errorf("DB_CONN_MAX_LIFETIME_MINUTES is not a valid int: %w", err)
	}

	c.Database = DatabaseConfig{
		Host:                  getEnvOrDefault("DB_HOST", "127.0.0.1"),
		Port:                  getEnvOrDefault("DB_PORT", "8889"),
		User:                  requireEnv("DB_USER"),
		Password:              getEnvOrDefault("DB_PASSWORD", ""),
		Name:                  requireEnv("DB_NAME"),
		Charset:               getEnvOrDefault("DB_CHARSET", "utf8mb4"),
		ParseTime:             parseTime,
		Loc:                   getEnvOrDefault("DB_LOC", "Local"),
		MaxOpenConns:          maxOpen,
		MaxIdleConns:          maxIdle,
		ConnMaxLifetimeMinutes: connLifetime,
	}
	return nil
}

func (c *Config) loadWebSocket() error {
	readBuf, err := strconv.Atoi(getEnvOrDefault("WS_READ_BUFFER_SIZE", "1024"))
	if err != nil {
		return fmt.Errorf("WS_READ_BUFFER_SIZE is not a valid int: %w", err)
	}

	writeBuf, err := strconv.Atoi(getEnvOrDefault("WS_WRITE_BUFFER_SIZE", "1024"))
	if err != nil {
		return fmt.Errorf("WS_WRITE_BUFFER_SIZE is not a valid int: %w", err)
	}

	clientBuf, err := strconv.Atoi(getEnvOrDefault("WS_CLIENT_SEND_BUFFER", "256"))
	if err != nil {
		return fmt.Errorf("WS_CLIENT_SEND_BUFFER is not a valid int: %w", err)
	}

	c.WebSocket = WebSocketConfig{
		ReadBufferSize:   readBuf,
		WriteBufferSize:  writeBuf,
		ClientSendBuffer: clientBuf,
	}
	return nil
}

func (c *Config) loadFetcher() error {
	fg, err := strconv.Atoi(getEnvOrDefault("FETCHER_QUOTE_INTERVAL_SEC", "5"))
	if err != nil {
		return fmt.Errorf("FETCHER_QUOTE_INTERVAL_SEC is not a valid int: %w", err)
	}

	og, err := strconv.Atoi(getEnvOrDefault("FETCHER_OHLCV_INTERVAL_SEC", "60"))
	if err != nil {
		return fmt.Errorf("FETCHER_OHLCV_INTERVAL_SEC is not a valid int: %w", err)
	}

	enabled, err := strconv.ParseBool(getEnvOrDefault("FETCHER_ENABLED", "true"))
	if err != nil {
		return fmt.Errorf("FETCHER_ENABLED is not a valid bool: %w", err)
	}

	c.Fetcher = FetcherConfig{
		QuoteIntervalSec: fg,
		OHLCVIntervalSec: og,
		Enabled:          enabled,
	}
	return nil
}

// getEnvOrDefault returns the environment variable value or a fallback default.
func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// requireEnv returns the environment variable value.
// Returns an empty string if not set — caller must validate.
func requireEnv(key string) string {
	return os.Getenv(key)
}

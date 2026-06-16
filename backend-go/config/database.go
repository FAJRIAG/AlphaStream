// Package config provides MySQL database connection setup with optimized pool settings.
package config

import (
	"database/sql"
	"fmt"
	"time"

	// Side-effect import: registers the "mysql" driver with database/sql.
	_ "github.com/go-sql-driver/mysql"
)

// NewDatabaseConnection creates and validates a MySQL *sql.DB connection pool.
// The pool is tuned for low-memory environments (Mac M1 / 8GB RAM):
//   - MaxOpenConns limits total connections to avoid exhausting MySQL's max_connections
//   - MaxIdleConns keeps a small pool warm to avoid repeated handshake overhead
//   - ConnMaxLifetime forces periodic recycling to prevent stale/zombie connections
//
// Returns an error if the DSN is invalid or the DB is unreachable.
func NewDatabaseConnection(cfg DatabaseConfig) (*sql.DB, error) {
	dsn := buildDSN(cfg)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("database.NewConnection sql.Open: %w", err)
	}

	// Apply connection pool settings before Ping to ensure they take effect.
	applyPoolSettings(db, cfg)

	// Ping validates the DSN and confirms the server is reachable.
	if err := db.Ping(); err != nil {
		// Always close the db handle on failure to prevent resource leaks.
		_ = db.Close()
		return nil, fmt.Errorf("database.NewConnection db.Ping: %w", err)
	}

	return db, nil
}

// buildDSN constructs the MySQL Data Source Name string.
// Format: user:password@tcp(host:port)/dbname?charset=...&parseTime=...&loc=...
func buildDSN(cfg DatabaseConfig) string {
	parseTime := "false"
	if cfg.ParseTime {
		parseTime = "true"
	}

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=%s&loc=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
		cfg.Charset,
		parseTime,
		cfg.Loc,
	)
}

// applyPoolSettings configures the *sql.DB connection pool.
// These values are tuned conservatively for a developer machine with 8GB RAM.
func applyPoolSettings(db *sql.DB, cfg DatabaseConfig) {
	// Total number of open connections (in-use + idle).
	db.SetMaxOpenConns(cfg.MaxOpenConns)

	// Maximum number of idle connections held ready in the pool.
	// Lower than MaxOpenConns to allow pool to shrink under low load.
	db.SetMaxIdleConns(cfg.MaxIdleConns)

	// Maximum lifetime of a connection before it is replaced.
	// Prevents use of stale connections after MySQL server restarts.
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetimeMinutes) * time.Minute)

	// Maximum time a connection may be idle before being closed.
	// Set to half of ConnMaxLifetime to aggressively reclaim idle memory.
	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxLifetimeMinutes/2+1) * time.Minute)
}

// Package entity defines the core domain structs for AlphaStream.
// These are pure data structures with no external dependencies —
// they represent Enterprise Business Rules and must not import infrastructure code.
package entity

import "time"

// ─── Stock Master Data ────────────────────────────────────────────────────────

// Stock represents a listed company/asset tracked by AlphaStream.
// This is the root aggregate for all price and indicator data.
type Stock struct {
	ID            int64      `json:"id"`
	Symbol        string     `json:"symbol"`       // e.g. "BBCA", "TLKM", "GOTO"
	Name          string     `json:"name"`         // Full company name
	Exchange      string     `json:"exchange"`     // e.g. "IDX", "NASDAQ"
	Currency      string     `json:"currency"`     // e.g. "IDR", "USD"
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Price         *float64   `json:"price"`
	Change        *float64   `json:"change"`
	ChangePercent *float64   `json:"change_percent"`
	Volume        *int64     `json:"volume"`
}

// ─── OHLCV Candlestick ────────────────────────────────────────────────────────

// OHLCV represents one candlestick data point (Open-High-Low-Close-Volume).
// Timeframe can be 1m, 5m, 15m, 1h, 1d depending on the feed source.
type OHLCV struct {
	ID        int64     `json:"id"`
	StockID   int64     `json:"stock_id"`
	Symbol    string    `json:"symbol"`
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    int64     `json:"volume"`
	Timeframe string    `json:"timeframe"` // "1m", "5m", "1h", "1d"
}

// ─── Live Ticker ──────────────────────────────────────────────────────────────

// Ticker represents a single real-time price update broadcast via WebSocket.
// It carries the minimum data needed for a live price feed display.
type Ticker struct {
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`         // Absolute change from prev close
	ChangePercent float64   `json:"change_percent"` // Percentage change
	Volume        int64     `json:"volume"`
	Timestamp     time.Time `json:"timestamp"`
}

// ─── OHLCV Query Parameters ───────────────────────────────────────────────────

// OHLCVQueryParams holds filter parameters for fetching historical OHLCV data.
type OHLCVQueryParams struct {
	Symbol    string
	Timeframe string
	Limit     int
	From      *time.Time
	To        *time.Time
}

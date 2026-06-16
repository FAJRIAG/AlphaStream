// Package repository defines the IStockRepository interface.
// This contract must be implemented by the infrastructure layer (mysql package).
// The domain layer has zero knowledge of *how* data is stored.
package repository

import (
	"context"

	"github.com/alphastream/backend-go/internal/domain/entity"
)

// IStockRepository defines all persistence operations for Stock and OHLCV data.
// Any struct implementing this interface can be injected into the usecase layer,
// enabling easy swapping of implementations (MySQL, PostgreSQL, in-memory mock).
//
//go:generate mockgen -destination=../../../../mocks/mock_stock_repository.go -package=mocks . IStockRepository
type IStockRepository interface {
	// ─── Stock Master Data ────────────────────────────────────────────────

	// GetAllStocks retrieves all active stocks from the database.
	GetAllStocks(ctx context.Context) ([]entity.Stock, error)

	// GetStockBySymbol retrieves a single stock by its ticker symbol (e.g. "BBCA").
	// Returns (nil, nil) if no stock is found — caller must check for nil.
	GetStockBySymbol(ctx context.Context, symbol string) (*entity.Stock, error)

	// CreateStock persists a new stock master record.
	// Returns the auto-generated database ID.
	CreateStock(ctx context.Context, stock entity.Stock) (int64, error)

	// UpdateStockPrice updates the current price and change metrics for a stock.
	UpdateStockPrice(ctx context.Context, symbol string, price, change, changePercent float64, volume int64) error

	// UpdateStockPricesBatch updates the current prices and change metrics for multiple stocks in a single transaction.
	UpdateStockPricesBatch(ctx context.Context, tickers []entity.Ticker) error

	// ─── OHLCV Candlestick Data ───────────────────────────────────────────

	// SaveOHLCV inserts a new candlestick record into the database.
	// Uses INSERT IGNORE to handle duplicate timestamps gracefully.
	SaveOHLCV(ctx context.Context, ohlcv entity.OHLCV) error

	// GetOHLCVHistory retrieves a slice of candlestick bars matching the query params.
	// Results are ordered by Timestamp ASC for chart rendering.
	GetOHLCVHistory(ctx context.Context, params entity.OHLCVQueryParams) ([]entity.OHLCV, error)

	// GetLatestOHLCV retrieves the most recent N candles for a symbol.
	// Used by the Quantitative Engine to compute indicators without loading all history.
	GetLatestOHLCV(ctx context.Context, symbol, timeframe string, limit int) ([]entity.OHLCV, error)
}

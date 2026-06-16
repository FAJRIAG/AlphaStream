// Package mysql provides the MySQL implementation of IStockRepository.
// All queries use prepared statements to prevent SQL injection.
// Transactions are used for multi-step write operations.
package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alphastream/backend-go/internal/domain/entity"
	domainRepo "github.com/alphastream/backend-go/internal/domain/repository"
)

// Compile-time check: stockRepository must implement IStockRepository.
var _ domainRepo.IStockRepository = (*stockRepository)(nil)

// stockRepository is the MySQL implementation of IStockRepository.
// It is unexported to enforce usage via the interface.
type stockRepository struct {
	db *sql.DB
}


// NewStockRepository creates a new stockRepository with the provided *sql.DB.
// This is the constructor pattern for Dependency Injection.
func NewStockRepository(db *sql.DB) domainRepo.IStockRepository {
	return &stockRepository{db: db}
}

// ─── Stock Master Data ────────────────────────────────────────────────────────

// GetAllStocks retrieves all active stocks ordered by symbol.
func (r *stockRepository) GetAllStocks(ctx context.Context) ([]entity.Stock, error) {
	const query = `
		SELECT id, symbol, name, exchange, currency, is_active, created_at, updated_at, price, change_val, change_percent, volume
		FROM stocks
		WHERE is_active = 1
		ORDER BY symbol ASC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("stock_repo.GetAllStocks query: %w", err)
	}
	defer rows.Close()

	return r.scanStockRows(rows)
}

// GetStockBySymbol retrieves a single stock by its ticker symbol.
// Returns (nil, nil) if no matching record is found.
func (r *stockRepository) GetStockBySymbol(ctx context.Context, symbol string) (*entity.Stock, error) {
	const query = `
		SELECT id, symbol, name, exchange, currency, is_active, created_at, updated_at, price, change_val, change_percent, volume
		FROM stocks
		WHERE symbol = ? AND is_active = 1
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, symbol)

	stock, err := r.scanStockRow(row)
	if err == sql.ErrNoRows {
		return nil, nil // Not found — caller must handle nil
	}
	if err != nil {
		return nil, fmt.Errorf("stock_repo.GetStockBySymbol scan: %w", err)
	}

	return stock, nil
}

// CreateStock inserts a new stock record and returns the auto-generated ID.
func (r *stockRepository) CreateStock(ctx context.Context, stock entity.Stock) (int64, error) {
	const query = `
		INSERT INTO stocks (symbol, name, exchange, currency, is_active)
		VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		stock.Symbol,
		stock.Name,
		stock.Exchange,
		stock.Currency,
		stock.IsActive,
	)
	if err != nil {
		return 0, fmt.Errorf("stock_repo.CreateStock exec: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("stock_repo.CreateStock last_insert_id: %w", err)
	}

	return id, nil
}

// UpdateStockPrice updates the current price and change metrics for a stock.
func (r *stockRepository) UpdateStockPrice(ctx context.Context, symbol string, price, change, changePercent float64, volume int64) error {
	const query = `
		UPDATE stocks
		SET price = ?, change_val = ?, change_percent = ?, volume = ?, updated_at = CURRENT_TIMESTAMP
		WHERE symbol = ?`

	_, err := r.db.ExecContext(ctx, query, price, change, changePercent, volume, symbol)
	if err != nil {
		return fmt.Errorf("stock_repo.UpdateStockPrice exec: %w", err)
	}
	return nil
}

// UpdateStockPricesBatch updates the current prices and change metrics for multiple stocks in a single transaction.
func (r *stockRepository) UpdateStockPricesBatch(ctx context.Context, tickers []entity.Ticker) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("stock_repo.UpdateStockPricesBatch begin tx: %w", err)
	}
	defer tx.Rollback()

	const query = `
		UPDATE stocks
		SET price = ?, change_val = ?, change_percent = ?, volume = ?, updated_at = CURRENT_TIMESTAMP
		WHERE symbol = ?`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("stock_repo.UpdateStockPricesBatch prepare: %w", err)
	}
	defer stmt.Close()

	for _, t := range tickers {
		_, err = stmt.ExecContext(ctx, t.Price, t.Change, t.ChangePercent, t.Volume, t.Symbol)
		if err != nil {
			return fmt.Errorf("stock_repo.UpdateStockPricesBatch exec %s: %w", t.Symbol, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("stock_repo.UpdateStockPricesBatch commit: %w", err)
	}
	return nil
}

// ─── OHLCV Candlestick Data ───────────────────────────────────────────────────

// SaveOHLCV inserts a candlestick record using INSERT IGNORE to handle
// duplicate (stock_id, timestamp, timeframe) gracefully without error.
func (r *stockRepository) SaveOHLCV(ctx context.Context, ohlcv entity.OHLCV) error {
	const query = `
		INSERT IGNORE INTO ohlcv (stock_id, symbol, timestamp, open, high, low, close, volume, timeframe)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		ohlcv.StockID,
		ohlcv.Symbol,
		ohlcv.Timestamp,
		ohlcv.Open,
		ohlcv.High,
		ohlcv.Low,
		ohlcv.Close,
		ohlcv.Volume,
		ohlcv.Timeframe,
	)
	if err != nil {
		return fmt.Errorf("stock_repo.SaveOHLCV exec: %w", err)
	}

	return nil
}

// GetOHLCVHistory retrieves candlestick history matching the given query params.
// It gets the latest limit records ordered descending, then returns them in chronological order.
func (r *stockRepository) GetOHLCVHistory(ctx context.Context, params entity.OHLCVQueryParams) ([]entity.OHLCV, error) {
	const query = `
		SELECT id, stock_id, symbol, timestamp, open, high, low, close, volume, timeframe
		FROM (
			SELECT id, stock_id, symbol, timestamp, open, high, low, close, volume, timeframe
			FROM ohlcv
			WHERE symbol = ? AND timeframe = ?
			ORDER BY timestamp DESC
			LIMIT ?
		) sub
		ORDER BY timestamp ASC`

	limit := params.Limit
	if limit <= 0 || limit > 1000 {
		limit = 500 // Safe default cap to prevent runaway queries
	}

	rows, err := r.db.QueryContext(ctx, query, params.Symbol, params.Timeframe, limit)
	if err != nil {
		return nil, fmt.Errorf("stock_repo.GetOHLCVHistory query: %w", err)
	}
	defer rows.Close()

	return r.scanOHLCVRows(rows)
}

// GetLatestOHLCV retrieves the most recent N candles for the Quantitative Engine.
// Results are returned in chronological order (oldest first) for indicator computation.
func (r *stockRepository) GetLatestOHLCV(ctx context.Context, symbol, timeframe string, limit int) ([]entity.OHLCV, error) {
	// Subquery selects the latest N rows ordered DESC, outer query re-orders ASC.
	const query = `
		SELECT id, stock_id, symbol, timestamp, open, high, low, close, volume, timeframe
		FROM (
			SELECT id, stock_id, symbol, timestamp, open, high, low, close, volume, timeframe
			FROM ohlcv
			WHERE symbol = ? AND timeframe = ?
			ORDER BY timestamp DESC
			LIMIT ?
		) sub
		ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, symbol, timeframe, limit)
	if err != nil {
		return nil, fmt.Errorf("stock_repo.GetLatestOHLCV query: %w", err)
	}
	defer rows.Close()

	return r.scanOHLCVRows(rows)
}

// ─── Private Scan Helpers ─────────────────────────────────────────────────────

// scanStockRows scans multiple stock rows into a slice.
func (r *stockRepository) scanStockRows(rows *sql.Rows) ([]entity.Stock, error) {
	var stocks []entity.Stock

	for rows.Next() {
		stock := entity.Stock{}
		if err := r.scanStockInto(rows.Scan, &stock); err != nil {
			return nil, fmt.Errorf("stock_repo.scanStockRows: %w", err)
		}
		stocks = append(stocks, stock)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("stock_repo.scanStockRows rows.Err: %w", err)
	}

	return stocks, nil
}

// scanStockRow scans a single *sql.Row into a Stock.
func (r *stockRepository) scanStockRow(row *sql.Row) (*entity.Stock, error) {
	stock := &entity.Stock{}
	err := row.Scan(
		&stock.ID,
		&stock.Symbol,
		&stock.Name,
		&stock.Exchange,
		&stock.Currency,
		&stock.IsActive,
		&stock.CreatedAt,
		&stock.UpdatedAt,
		&stock.Price,
		&stock.Change,
		&stock.ChangePercent,
		&stock.Volume,
	)
	return stock, err
}

// scanIntoStockFunc is a type alias for the scan function signature, enabling
// a single scanStockInto helper to work with both rows.Scan and row.Scan.
type scanIntoStockFunc func(dest ...interface{}) error

func (r *stockRepository) scanStockInto(scan scanIntoStockFunc, s *entity.Stock) error {
	return scan(
		&s.ID,
		&s.Symbol,
		&s.Name,
		&s.Exchange,
		&s.Currency,
		&s.IsActive,
		&s.CreatedAt,
		&s.UpdatedAt,
		&s.Price,
		&s.Change,
		&s.ChangePercent,
		&s.Volume,
	)
}

// scanOHLCVRows scans multiple OHLCV rows into a slice.
func (r *stockRepository) scanOHLCVRows(rows *sql.Rows) ([]entity.OHLCV, error) {
	var candles []entity.OHLCV

	for rows.Next() {
		var c entity.OHLCV
		if err := rows.Scan(
			&c.ID,
			&c.StockID,
			&c.Symbol,
			&c.Timestamp,
			&c.Open,
			&c.High,
			&c.Low,
			&c.Close,
			&c.Volume,
			&c.Timeframe,
		); err != nil {
			return nil, fmt.Errorf("stock_repo.scanOHLCVRows scan: %w", err)
		}
		candles = append(candles, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("stock_repo.scanOHLCVRows rows.Err: %w", err)
	}

	return candles, nil
}

// DeleteOHLCV deletes all historical ohlcv candles for a symbol.
func (r *stockRepository) DeleteOHLCV(ctx context.Context, symbol string) error {
	const query = `DELETE FROM ohlcv WHERE symbol = ?`
	_, err := r.db.ExecContext(ctx, query, symbol)
	if err != nil {
		return fmt.Errorf("stock_repo.DeleteOHLCV exec: %w", err)
	}
	return nil
}

// Package mysql provides the MySQL implementation of IIndicatorRepository.
package mysql

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/alphastream/backend-go/internal/domain/entity"
	domainRepo "github.com/alphastream/backend-go/internal/domain/repository"
)

// Compile-time check: indicatorRepository must implement IIndicatorRepository.
var _ domainRepo.IIndicatorRepository = (*indicatorRepository)(nil)

// indicatorRepository is the MySQL implementation of IIndicatorRepository.
type indicatorRepository struct {
	db *sql.DB
}

// NewIndicatorRepository creates a new indicatorRepository via constructor injection.
func NewIndicatorRepository(db *sql.DB) domainRepo.IIndicatorRepository {
	return &indicatorRepository{db: db}
}

// SaveIndicators persists a computed TechnicalIndicators record.
// Uses INSERT ... ON DUPLICATE KEY UPDATE to upsert by (stock_id, timestamp).
func (r *indicatorRepository) SaveIndicators(ctx context.Context, ind entity.TechnicalIndicators) error {
	const query = `
		INSERT INTO indicators
			(stock_id, symbol, timestamp, ma_20, ma_50, rsi_14, is_golden_cross, is_death_cross, atr_14)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			ma_20           = VALUES(ma_20),
			ma_50           = VALUES(ma_50),
			rsi_14          = VALUES(rsi_14),
			is_golden_cross = VALUES(is_golden_cross),
			is_death_cross  = VALUES(is_death_cross),
			atr_14          = VALUES(atr_14)`

	_, err := r.db.ExecContext(ctx, query,
		ind.StockID,
		ind.Symbol,
		ind.Timestamp,
		ind.MA20,
		ind.MA50,
		ind.RSI14,
		ind.IsGoldenCross,
		ind.IsDeathCross,
		ind.ATR14,
	)
	if err != nil {
		return fmt.Errorf("indicator_repo.SaveIndicators exec: %w", err)
	}

	return nil
}

// GetLatestIndicators retrieves the most recent indicator record for a symbol.
// Returns (nil, nil) if no record exists.
func (r *indicatorRepository) GetLatestIndicators(ctx context.Context, symbol string) (*entity.TechnicalIndicators, error) {
	const query = `
		SELECT id, stock_id, symbol, timestamp, ma_20, ma_50, rsi_14,
		       is_golden_cross, is_death_cross, atr_14, created_at
		FROM indicators
		WHERE symbol = ?
		ORDER BY timestamp DESC
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, symbol)

	ind, err := r.scanIndicatorRow(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("indicator_repo.GetLatestIndicators scan: %w", err)
	}

	return ind, nil
}

// GetIndicatorsHistory retrieves the N most recent indicator records for a symbol,
// returned in chronological order (oldest first) for chart rendering.
func (r *indicatorRepository) GetIndicatorsHistory(ctx context.Context, symbol string, limit int) ([]entity.TechnicalIndicators, error) {
	const query = `
		SELECT id, stock_id, symbol, timestamp, ma_20, ma_50, rsi_14,
		       is_golden_cross, is_death_cross, atr_14, created_at
		FROM (
			SELECT id, stock_id, symbol, timestamp, ma_20, ma_50, rsi_14,
			       is_golden_cross, is_death_cross, atr_14, created_at
			FROM indicators
			WHERE symbol = ?
			ORDER BY timestamp DESC
			LIMIT ?
		) sub
		ORDER BY timestamp ASC`

	rows, err := r.db.QueryContext(ctx, query, symbol, limit)
	if err != nil {
		return nil, fmt.Errorf("indicator_repo.GetIndicatorsHistory query: %w", err)
	}
	defer rows.Close()

	return r.scanIndicatorRows(rows)
}

// ─── Private Scan Helpers ─────────────────────────────────────────────────────

func (r *indicatorRepository) scanIndicatorRow(row *sql.Row) (*entity.TechnicalIndicators, error) {
	ind := &entity.TechnicalIndicators{}
	err := row.Scan(
		&ind.ID,
		&ind.StockID,
		&ind.Symbol,
		&ind.Timestamp,
		&ind.MA20,
		&ind.MA50,
		&ind.RSI14,
		&ind.IsGoldenCross,
		&ind.IsDeathCross,
		&ind.ATR14,
		&ind.CreatedAt,
	)
	return ind, err
}

func (r *indicatorRepository) scanIndicatorRows(rows *sql.Rows) ([]entity.TechnicalIndicators, error) {
	var results []entity.TechnicalIndicators

	for rows.Next() {
		ind := entity.TechnicalIndicators{}
		if err := rows.Scan(
			&ind.ID,
			&ind.StockID,
			&ind.Symbol,
			&ind.Timestamp,
			&ind.MA20,
			&ind.MA50,
			&ind.RSI14,
			&ind.IsGoldenCross,
			&ind.IsDeathCross,
			&ind.ATR14,
			&ind.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("indicator_repo.scanIndicatorRows scan: %w", err)
		}
		results = append(results, ind)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("indicator_repo.scanIndicatorRows rows.Err: %w", err)
	}

	return results, nil
}

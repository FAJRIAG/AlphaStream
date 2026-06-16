// Package repository defines the IIndicatorRepository interface.
// Separating indicator persistence from stock persistence follows SRP —
// each interface has exactly one reason to change.
package repository

import (
	"context"

	"github.com/alphastream/backend-go/internal/domain/entity"
)

// IIndicatorRepository defines all persistence operations for technical indicator data.
// The Quantitative Engine writes computed values through this interface.
//
//go:generate mockgen -destination=../../../../mocks/mock_indicator_repository.go -package=mocks . IIndicatorRepository
type IIndicatorRepository interface {
	// SaveIndicators persists a computed TechnicalIndicators record.
	// Called by the Quantitative Engine after each indicator computation cycle.
	SaveIndicators(ctx context.Context, indicators entity.TechnicalIndicators) error

	// GetLatestIndicators retrieves the most recently computed indicator set
	// for a given symbol. Returns (nil, nil) if no record exists yet.
	GetLatestIndicators(ctx context.Context, symbol string) (*entity.TechnicalIndicators, error)

	// GetIndicatorsHistory retrieves indicator history for a symbol within a time range.
	// Useful for rendering MA lines on the chart alongside OHLCV data.
	GetIndicatorsHistory(ctx context.Context, symbol string, limit int) ([]entity.TechnicalIndicators, error)

	// GetLatestIndicatorsAll retrieves the latest computed indicators for all stocks.
	GetLatestIndicatorsAll(ctx context.Context) ([]entity.IndicatorWithStock, error)
}

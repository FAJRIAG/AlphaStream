// Package entity defines all technical analysis indicator structs.
// These represent the output of the Quantitative Engine usecase layer.
package entity

import "time"

// ─── Technical Indicators ─────────────────────────────────────────────────────

// TechnicalIndicators holds the computed values for a single OHLCV bar.
// All fields are float64 pointers to distinguish "not yet computed" (nil)
// from "computed as zero" (0.0), especially during warm-up periods.
type TechnicalIndicators struct {
	ID        int64      `json:"id"`
	StockID   int64      `json:"stock_id"`
	Symbol    string     `json:"symbol"`
	Timestamp time.Time  `json:"timestamp"`

	// Moving Averages
	MA20 *float64 `json:"ma_20"` // Simple Moving Average over 20 periods
	MA50 *float64 `json:"ma_50"` // Simple Moving Average over 50 periods

	// Relative Strength Index
	RSI14 *float64 `json:"rsi_14"` // RSI over 14 periods (0–100 scale)

	// Cross Signals
	IsGoldenCross *bool `json:"is_golden_cross"` // MA20 crossed above MA50
	IsDeathCross  *bool `json:"is_death_cross"`  // MA20 crossed below MA50

	// Volatility
	ATR14 *float64 `json:"atr_14"` // Average True Range over 14 periods

	CreatedAt time.Time `json:"created_at"`
}

// ─── Prediction Result ────────────────────────────────────────────────────────

// TrendDirection represents the directional bias of the predicted price movement.
type TrendDirection string

const (
	TrendBullish TrendDirection = "BULLISH"
	TrendBearish TrendDirection = "BEARISH"
	TrendNeutral TrendDirection = "NEUTRAL"
)

// PredictionResult is the output of the predictive engine for a given symbol.
// It combines indicator signals into a human-readable actionable summary.
type PredictionResult struct {
	Symbol      string         `json:"symbol"`
	Timestamp   time.Time      `json:"timestamp"`
	Direction   TrendDirection `json:"direction"`

	// Probability score (0.0–1.0) derived from combined indicator signals.
	Probability float64 `json:"probability"`

	// Volatility-based target prices (using ATR multiples).
	TargetPriceUp   float64 `json:"target_price_up"`
	TargetPriceDown float64 `json:"target_price_down"`

	// Supporting signals used to derive this prediction.
	Signals PredictionSignals `json:"signals"`

	// Current price at the time of prediction.
	CurrentPrice float64 `json:"current_price"`
}

// PredictionSignals holds the individual indicator readings that fed into
// the PredictionResult. Exposed in the API response for transparency.
type PredictionSignals struct {
	MA20          float64 `json:"ma_20"`
	MA50          float64 `json:"ma_50"`
	RSI14         float64 `json:"rsi_14"`
	ATR14         float64 `json:"atr_14"`
	IsGoldenCross bool    `json:"is_golden_cross"`
	IsDeathCross  bool    `json:"is_death_cross"`
	RSIZone       string  `json:"rsi_zone"` // "OVERBOUGHT", "NEUTRAL", "OVERSOLD"
}

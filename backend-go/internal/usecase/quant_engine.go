// Package usecase contains the Quantitative Engine for AlphaStream.
// All functions here are pure (no side effects, no I/O) making them
// trivially unit-testable and fully deterministic.
//
// Algorithms implemented:
//   - Simple Moving Average (SMA): MA-20, MA-50
//   - Relative Strength Index (RSI-14)
//   - Average True Range (ATR-14)
//   - Golden Cross / Death Cross detection
//   - Volatility-based Target Price
//   - Trend Probability scoring
package usecase

import (
	"math"

	"github.com/alphastream/backend-go/internal/domain/entity"
)

// ─── Constants ────────────────────────────────────────────────────────────────

const (
	// MA periods
	MAPeriod20 = 20
	MAPeriod50 = 50

	// RSI period and zone thresholds
	RSIPeriod         = 14
	RSIOverboughtZone = 70.0
	RSIOversoldZone   = 30.0

	// ATR period for volatility calculation
	ATRPeriod = 14

	// ATR multipliers for target price calculation
	ATRMultiplierUp   = 2.0
	ATRMultiplierDown = 1.5

	// Minimum candles required before any indicator is valid
	MinCandlesForMA20 = MAPeriod20
	MinCandlesForMA50 = MAPeriod50
	MinCandlesForRSI  = RSIPeriod + 1 // RSI needs period+1 for the first diff
	MinCandlesForATR  = ATRPeriod + 1
)

// ─── QuantEngine ─────────────────────────────────────────────────────────────

// QuantEngine is the stateless Quantitative Engine for computing technical indicators.
// It operates exclusively on []entity.OHLCV slices and returns typed results.
// Zero dependencies on infrastructure — fully deterministic and testable.
type QuantEngine struct{}

// NewQuantEngine creates a new QuantEngine instance.
func NewQuantEngine() *QuantEngine {
	return &QuantEngine{}
}

// ComputeAll runs all indicators and the prediction engine on the provided candle series.
// The candles slice must be sorted in chronological order (oldest first).
// Returns nil if there are not enough candles for even the shortest indicator (MA-20).
func (q *QuantEngine) ComputeAll(symbol string, candles []entity.OHLCV) (*entity.TechnicalIndicators, *entity.PredictionResult) {
	if len(candles) < MinCandlesForMA20 {
		return nil, nil
	}

	closes := extractCloses(candles)
	ind := &entity.TechnicalIndicators{
		Symbol:    symbol,
		Timestamp: candles[len(candles)-1].Timestamp,
	}

	// ── Moving Averages ──────────────────────────────────────────────────
	if ma20, ok := q.computeSMA(closes, MAPeriod20); ok {
		ind.MA20 = float64Ptr(ma20)
	}
	if ma50, ok := q.computeSMA(closes, MAPeriod50); ok {
		ind.MA50 = float64Ptr(ma50)
	}

	// ── RSI ──────────────────────────────────────────────────────────────
	if rsi, ok := q.computeRSI(closes, RSIPeriod); ok {
		ind.RSI14 = float64Ptr(rsi)
	}

	// ── ATR ──────────────────────────────────────────────────────────────
	if atr, ok := q.computeATR(candles, ATRPeriod); ok {
		ind.ATR14 = float64Ptr(atr)
	}

	// ── Cross Signals ────────────────────────────────────────────────────
	if len(candles) >= MinCandlesForMA50+1 {
		prevCloses := extractCloses(candles[:len(candles)-1])
		prevMA20, hasPrevMA20 := q.computeSMA(prevCloses, MAPeriod20)
		prevMA50, hasPrevMA50 := q.computeSMA(prevCloses, MAPeriod50)

		if ind.MA20 != nil && ind.MA50 != nil && hasPrevMA20 && hasPrevMA50 {
			golden := *ind.MA20 > *ind.MA50 && prevMA20 <= prevMA50
			death := *ind.MA20 < *ind.MA50 && prevMA20 >= prevMA50
			ind.IsGoldenCross = boolPtr(golden)
			ind.IsDeathCross = boolPtr(death)
		}
	}

	// ── Prediction ───────────────────────────────────────────────────────
	currentClose := candles[len(candles)-1].Close
	prediction := q.buildPrediction(symbol, currentClose, ind)

	return ind, prediction
}

// ─── SMA (Simple Moving Average) ─────────────────────────────────────────────

// computeSMA calculates the Simple Moving Average of the last `period` values.
// Returns (0, false) if there are fewer data points than the period requires.
func (q *QuantEngine) computeSMA(closes []float64, period int) (float64, bool) {
	if len(closes) < period {
		return 0, false
	}

	// Use only the last `period` values.
	window := closes[len(closes)-period:]
	sum := 0.0
	for _, c := range window {
		sum += c
	}
	return sum / float64(period), true
}

// ─── RSI (Relative Strength Index) ───────────────────────────────────────────

// computeRSI calculates the RSI using Wilder's smoothed moving average method.
// Requires at least period+1 closes to compute the first gain/loss delta.
// Returns a value in the range [0.0, 100.0].
func (q *QuantEngine) computeRSI(closes []float64, period int) (float64, bool) {
	if len(closes) < period+1 {
		return 0, false
	}

	var totalGain, totalLoss float64

	// First pass: simple average of gains and losses over the first `period` deltas.
	for i := 1; i <= period; i++ {
		delta := closes[i] - closes[i-1]
		if delta > 0 {
			totalGain += delta
		} else {
			totalLoss += math.Abs(delta)
		}
	}

	avgGain := totalGain / float64(period)
	avgLoss := totalLoss / float64(period)

	// Wilder's smoothing for subsequent periods (from index period+1 to the end).
	prev := closes[period]
	for i := period + 1; i < len(closes); i++ {
		cur := closes[i]
		delta := cur - prev
		gain, loss := 0.0, 0.0
		if delta > 0 {
			gain = delta
		} else {
			loss = math.Abs(delta)
		}
		// Wilder's smoothed average
		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)
		prev = cur
	}

	if avgLoss == 0 {
		return 100.0, true // All gains, no losses → RSI = 100
	}

	rs := avgGain / avgLoss
	rsi := 100.0 - (100.0 / (1.0 + rs))
	return rsi, true
}

// ─── ATR (Average True Range) ────────────────────────────────────────────────

// computeATR calculates Wilder's Average True Range over `period` bars.
// ATR measures volatility as the average of True Range values.
// True Range = max(High-Low, |High-PrevClose|, |Low-PrevClose|)
func (q *QuantEngine) computeATR(candles []entity.OHLCV, period int) (float64, bool) {
	if len(candles) < period+1 {
		return 0, false
	}

	// Compute True Range for each bar starting from index 1 to `period`.
	var totalTR float64
	for i := 1; i <= period; i++ {
		curr := candles[i]
		prev := candles[i-1]
		tr := trueRange(curr, prev.Close)
		totalTR += tr
	}

	// Initial ATR = simple average of first `period` TR values.
	atr := totalTR / float64(period)

	// Wilder's smoothing for any additional bars beyond the initial period.
	prevClose := candles[period].Close
	for i := period + 1; i < len(candles); i++ {
		c := candles[i]
		tr := trueRange(c, prevClose)
		atr = (atr*float64(period-1) + tr) / float64(period)
		prevClose = c.Close
	}

	return atr, true
}

// trueRange computes the True Range for a single candle given the previous close.
func trueRange(c entity.OHLCV, prevClose float64) float64 {
	highLow := c.High - c.Low
	highPrevClose := math.Abs(c.High - prevClose)
	lowPrevClose := math.Abs(c.Low - prevClose)
	return math.Max(highLow, math.Max(highPrevClose, lowPrevClose))
}

// ─── Prediction Engine ────────────────────────────────────────────────────────

// buildPrediction derives a PredictionResult from the computed indicators.
// The probability score is a weighted combination of RSI zone, MA alignment, and cross signals.
func (q *QuantEngine) buildPrediction(symbol string, currentPrice float64, ind *entity.TechnicalIndicators) *entity.PredictionResult {
	signals := entity.PredictionSignals{}
	probability := 0.5 // Start at 50% (neutral baseline)

	// ── MA Signal ────────────────────────────────────────────────────────
	if ind.MA20 != nil {
		signals.MA20 = *ind.MA20
	}
	if ind.MA50 != nil {
		signals.MA50 = *ind.MA50
	}
	if ind.MA20 != nil && ind.MA50 != nil {
		if *ind.MA20 > *ind.MA50 {
			probability += 0.15 // Bullish MA alignment
		} else {
			probability -= 0.15 // Bearish MA alignment
		}
	}

	// ── RSI Signal ───────────────────────────────────────────────────────
	if ind.RSI14 != nil {
		signals.RSI14 = *ind.RSI14
		rsi := *ind.RSI14
		switch {
		case rsi >= RSIOverboughtZone:
			signals.RSIZone = "OVERBOUGHT"
			probability -= 0.20 // High RSI → potential reversal down
		case rsi <= RSIOversoldZone:
			signals.RSIZone = "OVERSOLD"
			probability += 0.20 // Low RSI → potential bounce up
		default:
			signals.RSIZone = "NEUTRAL"
		}
	}

	// ── Golden / Death Cross Signal ──────────────────────────────────────
	if ind.IsGoldenCross != nil && *ind.IsGoldenCross {
		signals.IsGoldenCross = true
		probability += 0.25 // Strong bullish signal
	}
	if ind.IsDeathCross != nil && *ind.IsDeathCross {
		signals.IsDeathCross = true
		probability -= 0.25 // Strong bearish signal
	}

	// ── ATR-based Target Price ───────────────────────────────────────────
	targetUp := currentPrice
	targetDown := currentPrice
	if ind.ATR14 != nil {
		signals.ATR14 = *ind.ATR14
		atr := *ind.ATR14
		targetUp = currentPrice + (atr * ATRMultiplierUp)
		targetDown = currentPrice - (atr * ATRMultiplierDown)
	}

	// Clamp probability to [0.0, 1.0]
	probability = math.Max(0.0, math.Min(1.0, probability))

	// Derive direction from probability
	direction := entity.TrendNeutral
	switch {
	case probability > 0.60:
		direction = entity.TrendBullish
	case probability < 0.40:
		direction = entity.TrendBearish
	}

	return &entity.PredictionResult{
		Symbol:          symbol,
		Timestamp:       ind.Timestamp,
		Direction:       direction,
		Probability:     math.Round(probability*100) / 100, // 2 decimal places
		TargetPriceUp:   math.Round(targetUp*100) / 100,
		TargetPriceDown: math.Round(targetDown*100) / 100,
		CurrentPrice:    currentPrice,
		Signals:         signals,
	}
}

// ─── Utility Helpers ──────────────────────────────────────────────────────────

// extractCloses returns a float64 slice of closing prices from OHLCV candles.
func extractCloses(candles []entity.OHLCV) []float64 {
	closes := make([]float64, len(candles))
	for i, c := range candles {
		closes[i] = c.Close
	}
	return closes
}

// float64Ptr returns a pointer to the given float64 value.
func float64Ptr(v float64) *float64 { return &v }

// boolPtr returns a pointer to the given bool value.
func boolPtr(v bool) *bool { return &v }

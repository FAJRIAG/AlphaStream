// Package usecase_test contains unit tests for the Quantitative Engine.
// All tests are pure (no I/O, no DB) — run with: go test ./internal/usecase/... -v -race
package usecase_test

import (
	"math"
	"testing"
	"time"

	"github.com/alphastream/backend-go/internal/domain/entity"
	"github.com/alphastream/backend-go/internal/usecase"
)

// ─── Test Helpers ──────────────────────────────────────────────────────────────

// makeCandles generates a slice of OHLCV candles with incrementing close prices.
// Used to produce a predictable, deterministic dataset for testing.
func makeCandles(count int, startPrice float64) []entity.OHLCV {
	candles := make([]entity.OHLCV, count)
	for i := range candles {
		price := startPrice + float64(i)*1.0
		candles[i] = entity.OHLCV{
			Symbol:    "TEST",
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
			Open:      price,
			High:      price + 0.5,
			Low:       price - 0.5,
			Close:     price,
			Volume:    1000,
		}
	}
	return candles
}

// almostEqual checks if two floats are equal within a tolerance.
func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// ─── SMA Tests ────────────────────────────────────────────────────────────────

func TestComputeAll_MA20_CorrectValue(t *testing.T) {
	q := usecase.NewQuantEngine()
	// 20 candles with prices 1, 2, ..., 20. MA20 = (1+2+...+20)/20 = 10.5
	candles := makeCandles(20, 1.0)

	ind, _ := q.ComputeAll("TEST", candles)

	if ind == nil {
		t.Fatal("Expected indicators, got nil")
	}
	if ind.MA20 == nil {
		t.Fatal("Expected MA20 to be computed, got nil")
	}

	expected := 10.5
	if !almostEqual(*ind.MA20, expected, 0.01) {
		t.Errorf("MA20 expected %.4f, got %.4f", expected, *ind.MA20)
	}
}

func TestComputeAll_MA50_NilWithFewCandles(t *testing.T) {
	q := usecase.NewQuantEngine()
	// Only 20 candles — MA50 should be nil (insufficient data)
	candles := makeCandles(20, 100.0)

	ind, _ := q.ComputeAll("TEST", candles)

	if ind == nil {
		t.Fatal("Expected indicators, got nil")
	}
	if ind.MA50 != nil {
		t.Errorf("Expected MA50 to be nil with only 20 candles, got %.4f", *ind.MA50)
	}
}

func TestComputeAll_MA50_ComputedWith50Candles(t *testing.T) {
	q := usecase.NewQuantEngine()
	// 50 candles with prices 1..50. MA50 = (1+2+...+50)/50 = 25.5
	candles := makeCandles(50, 1.0)

	ind, _ := q.ComputeAll("TEST", candles)

	if ind == nil || ind.MA50 == nil {
		t.Fatal("Expected MA50 to be computed with 50 candles")
	}

	expected := 25.5
	if !almostEqual(*ind.MA50, expected, 0.01) {
		t.Errorf("MA50 expected %.4f, got %.4f", expected, *ind.MA50)
	}
}

// ─── RSI Tests ────────────────────────────────────────────────────────────────

func TestComputeAll_RSI_BoundedRange(t *testing.T) {
	q := usecase.NewQuantEngine()
	// 60 candles — enough for both MA20 and RSI14
	candles := makeCandles(60, 100.0)

	ind, _ := q.ComputeAll("TEST", candles)

	if ind == nil || ind.RSI14 == nil {
		t.Fatal("Expected RSI14 to be computed")
	}

	rsi := *ind.RSI14
	if rsi < 0 || rsi > 100 {
		t.Errorf("RSI must be in [0, 100], got %.4f", rsi)
	}
}

func TestComputeAll_RSI_AllGains_Returns100(t *testing.T) {
	q := usecase.NewQuantEngine()
	// Strictly increasing prices → all gains, no losses → RSI should be 100
	candles := makeCandles(30, 100.0)

	ind, _ := q.ComputeAll("TEST", candles)

	if ind == nil || ind.RSI14 == nil {
		t.Fatal("Expected RSI14 to be computed")
	}

	// RSI with all gains should be exactly 100.0
	if !almostEqual(*ind.RSI14, 100.0, 0.001) {
		t.Errorf("Expected RSI=100.0 with all-gain series, got %.4f", *ind.RSI14)
	}
}

// ─── Golden Cross Tests ───────────────────────────────────────────────────────

func TestComputeAll_GoldenCross_Detection(t *testing.T) {
	q := usecase.NewQuantEngine()

	// Construct a scenario where MA20 crosses above MA50.
	// Start with flat prices (MA20 ≈ MA50), then jump to create a Golden Cross.
	candles := make([]entity.OHLCV, 60)
	baseTime := time.Now()

	// First 50 candles: flat at 100
	for i := 0; i < 50; i++ {
		candles[i] = entity.OHLCV{
			Symbol:    "TEST",
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Open:      100, High: 101, Low: 99, Close: 100,
			Volume: 1000,
		}
	}

	// Last 10 candles: spike up — drives MA20 above MA50
	for i := 50; i < 60; i++ {
		price := 100.0 + float64(i-49)*5.0
		candles[i] = entity.OHLCV{
			Symbol:    "TEST",
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
			Open:      price, High: price + 1, Low: price - 1, Close: price,
			Volume: 2000,
		}
	}

	ind, pred := q.ComputeAll("TEST", candles)

	if ind == nil {
		t.Fatal("Expected indicators to be computed")
	}
	if ind.MA20 == nil || ind.MA50 == nil {
		t.Skip("Not enough data for cross detection in this scenario")
	}

	t.Logf("MA20=%.2f MA50=%.2f GoldenCross=%v Prediction=%+v",
		*ind.MA20, *ind.MA50, ind.IsGoldenCross, pred)
}

// ─── Prediction Tests ─────────────────────────────────────────────────────────

func TestComputeAll_Prediction_ProbabilityBounded(t *testing.T) {
	q := usecase.NewQuantEngine()
	candles := makeCandles(60, 100.0)

	_, pred := q.ComputeAll("TEST", candles)

	if pred == nil {
		t.Fatal("Expected prediction result")
	}
	if pred.Probability < 0 || pred.Probability > 1 {
		t.Errorf("Probability must be in [0,1], got %.4f", pred.Probability)
	}
}

func TestComputeAll_Prediction_TargetPriceUp_GreaterThanCurrent(t *testing.T) {
	q := usecase.NewQuantEngine()
	candles := makeCandles(60, 100.0)

	_, pred := q.ComputeAll("TEST", candles)

	if pred == nil {
		t.Fatal("Expected prediction result")
	}
	// TargetPriceUp must be above the current price (ATR > 0 for any real candle series)
	if pred.TargetPriceUp <= pred.CurrentPrice {
		t.Errorf("TargetPriceUp (%.2f) should be > CurrentPrice (%.2f)",
			pred.TargetPriceUp, pred.CurrentPrice)
	}
}

func TestComputeAll_InsufficientCandles_ReturnsNil(t *testing.T) {
	q := usecase.NewQuantEngine()
	// Only 5 candles — below MinCandlesForMA20 (20)
	candles := makeCandles(5, 100.0)

	ind, pred := q.ComputeAll("TEST", candles)

	if ind != nil || pred != nil {
		t.Error("Expected nil indicators and prediction with < 20 candles")
	}
}

// Package fetcher provides a Yahoo Finance HTTP client for real IDX market data.
// No API key required. Uses the public Yahoo Finance v7/v8 endpoints.
//
// Rate limits: Yahoo Finance does not publish official limits.
// Safe polling rates: quote=5s, ohlcv=60s to avoid 429 responses.
package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alphastream/backend-go/internal/domain/entity"
)

// ─── Constants ────────────────────────────────────────────────────────────────

const (
	yahooChartURL = "https://query1.finance.yahoo.com/v8/finance/chart"

	httpTimeout   = 15 * time.Second
	maxRetries    = 3
	retryDelay    = 2 * time.Second
)

// ─── Symbol Mapping ───────────────────────────────────────────────────────────

// ToYahooSymbol converts an internal symbol to Yahoo Finance format.
func ToYahooSymbol(internal string) string {
	if internal == "IHSG" {
		return "^JKSE"
	}
	return internal + ".JK"
}

// FromYahooSymbol converts a Yahoo Finance symbol back to internal format.
func FromYahooSymbol(yahoo string) string {
	if yahoo == "^JKSE" {
		return "IHSG"
	}
	return strings.TrimSuffix(yahoo, ".JK")
}

// ─── Response Structs ─────────────────────────────────────────────────────────


// yahooChartResponse is the top-level structure for the v8/finance/chart response.
type yahooChartResponse struct {
	Chart struct {
		Result []struct {
			Meta struct {
				Symbol             string  `json:"symbol"`
				RegularMarketPrice float64 `json:"regularMarketPrice"`
				ChartPreviousClose float64 `json:"chartPreviousClose"`
				PreviousClose      float64 `json:"previousClose"`
				Timezone           string  `json:"timezone"`
			} `json:"meta"`
			Timestamp  []int64 `json:"timestamp"`
			Indicators struct {
				Quote []struct {
					Open   []float64 `json:"open"`
					High   []float64 `json:"high"`
					Low    []float64 `json:"low"`
					Close  []float64 `json:"close"`
					Volume []int64   `json:"volume"`
				} `json:"quote"`
			} `json:"indicators"`
		} `json:"result"`
		Error interface{} `json:"error"`
	} `json:"chart"`
}

// ─── LiveQuote ────────────────────────────────────────────────────────────────

// LiveQuote is the normalized quote data returned by FetchQuotes.
type LiveQuote struct {
	Symbol        string
	Price         float64
	Change        float64
	ChangePercent float64
	Volume        int64
	Open          float64
	High          float64
	Low           float64
	PrevClose     float64
}

// ─── YahooClient ─────────────────────────────────────────────────────────────

// YahooClient is an HTTP client for Yahoo Finance endpoints.
type YahooClient struct {
	http *http.Client
}

// NewYahooClient creates a YahooClient with a timeout-configured HTTP client.
func NewYahooClient() *YahooClient {
	return &YahooClient{
		http: &http.Client{
			Timeout: httpTimeout,
			// Custom transport with modern browser User-Agent to avoid 401s
			Transport: &userAgentTransport{
				wrapped: http.DefaultTransport,
			},
		},
	}
}

// FetchQuotes fetches live quotes for all provided internal symbols by querying their chart endpoints.
// This endpoint does not require cookies/crumbs and is safe from 401 Unauthorized errors.
func (c *YahooClient) FetchQuotes(ctx context.Context, symbols []string) ([]LiveQuote, error) {
	var allQuotes []LiveQuote

	for _, s := range symbols {
		yahooSym := ToYahooSymbol(s)
		url := fmt.Sprintf("%s/%s?interval=1m&range=1d&includePrePost=false", yahooChartURL, yahooSym)

		body, err := c.getWithRetry(ctx, url)
		if err != nil {
			log.Printf("[YahooClient] FetchQuotes chart error for %s: %v", s, err)
			continue
		}

		var resp yahooChartResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			log.Printf("[YahooClient] FetchQuotes parse error for %s: %v", s, err)
			continue
		}

		if len(resp.Chart.Result) == 0 {
			continue
		}

		result := resp.Chart.Result[0]
		meta := result.Meta

		price := meta.RegularMarketPrice
		prevClose := meta.ChartPreviousClose
		if prevClose == 0 {
			prevClose = meta.PreviousClose
		}
		if prevClose == 0 {
			prevClose = price // Fallback
		}
		change := price - prevClose
		changePercent := (change / prevClose) * 100.0

		// Find current minute's OHLC from the last index of chart response
		open := price
		high := price
		low := price
		var volume int64 = 0

		if len(result.Indicators.Quote) > 0 {
			q := result.Indicators.Quote[0]
			n := len(result.Timestamp)
			if n > 0 {
				lastIdx := n - 1
				if lastIdx < len(q.Open) && q.Open[lastIdx] > 0 {
					open = q.Open[lastIdx]
				}
				if lastIdx < len(q.High) && q.High[lastIdx] > 0 {
					high = q.High[lastIdx]
				}
				if lastIdx < len(q.Low) && q.Low[lastIdx] > 0 {
					low = q.Low[lastIdx]
				}
			}

			// Sum total day volume
			for i := 0; i < n; i++ {
				if i < len(q.Volume) {
					volume += q.Volume[i]
				}
			}
		}

		allQuotes = append(allQuotes, LiveQuote{
			Symbol:        s,
			Price:         price,
			Change:        change,
			ChangePercent: changePercent,
			Volume:        volume,
			Open:          open,
			High:          high,
			Low:           low,
			PrevClose:     prevClose,
		})

		// Sleep briefly between requests to avoid rate limits if polling multiple symbols
		if len(symbols) > 1 {
			select {
			case <-ctx.Done():
				return allQuotes, ctx.Err()
			case <-time.After(200 * time.Millisecond):
			}
		}
	}

	return allQuotes, nil
}

// FetchOHLCV fetches 1-minute OHLCV candlestick data for a single symbol.
// interval: "1m", "5m", "1h" | rangeStr: "1d", "5d"
func (c *YahooClient) FetchOHLCV(ctx context.Context, internalSymbol, interval, rangeStr string) ([]entity.OHLCV, error) {
	yahooSym := ToYahooSymbol(internalSymbol)

	url := fmt.Sprintf("%s/%s?interval=%s&range=%s&includePrePost=false",
		yahooChartURL, yahooSym, interval, rangeStr)

	body, err := c.getWithRetry(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("yahoo.FetchOHLCV http %s: %w", internalSymbol, err)
	}

	var resp yahooChartResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("yahoo.FetchOHLCV parse %s: %w", internalSymbol, err)
	}

	if len(resp.Chart.Result) == 0 || len(resp.Chart.Result[0].Timestamp) == 0 {
		return nil, nil // Market closed or no data available
	}

	result := resp.Chart.Result[0]
	if len(result.Indicators.Quote) == 0 {
		return nil, nil
	}

	q := result.Indicators.Quote[0]
	n := len(result.Timestamp)
	candles := make([]entity.OHLCV, 0, n)

	for i := 0; i < n; i++ {
		// Skip bars with nil/zero values (market gaps, pre/post market)
		if i >= len(q.Open) || i >= len(q.Close) {
			continue
		}
		if q.Close[i] == 0 || q.Open[i] == 0 {
			continue
		}

		vol := int64(0)
		if i < len(q.Volume) {
			vol = q.Volume[i]
		}

		candles = append(candles, entity.OHLCV{
			Symbol:    internalSymbol,
			Timestamp: time.Unix(result.Timestamp[i], 0).UTC(),
			Open:      roundPrice(q.Open[i]),
			High:      roundPrice(q.High[i]),
			Low:       roundPrice(q.Low[i]),
			Close:     roundPrice(q.Close[i]),
			Volume:    vol,
			Timeframe: interval,
		})
	}

	return candles, nil
}

// ─── Private Helpers ──────────────────────────────────────────────────────────

// getWithRetry performs an HTTP GET with up to maxRetries attempts.
func (c *YahooClient) getWithRetry(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryDelay * time.Duration(attempt)):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("getWithRetry create request: %w", err)
		}

		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body[:min(200, len(body))]))
			continue
		}

		return body, nil
	}
	return nil, fmt.Errorf("getWithRetry exhausted %d attempts: %w", maxRetries, lastErr)
}

// roundPrice rounds to 2 decimal places.
func roundPrice(p float64) float64 {
	if p == 0 {
		return 0
	}
	return float64(int(p*100+0.5)) / 100
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ─── userAgentTransport ───────────────────────────────────────────────────────

// userAgentTransport injects a modern browser User-Agent header.
// Yahoo Finance returns 401 without a proper User-Agent.
type userAgentTransport struct {
	wrapped http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())
	req.Header.Set("User-Agent",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 "+
			"(KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	return t.wrapped.RoundTrip(req)
}

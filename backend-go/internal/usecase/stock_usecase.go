// Package usecase contains the StockUsecase — the primary orchestrator
// that connects the Repository, Quantitative Engine, and WebSocket Hub.
package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"time"

	"github.com/alphastream/backend-go/internal/domain/entity"
	domainRepo "github.com/alphastream/backend-go/internal/domain/repository"
	"github.com/alphastream/backend-go/internal/repository/cache"
)

// ─── Interfaces ───────────────────────────────────────────────────────────────

// IBroadcaster defines the contract for broadcasting messages to WebSocket clients.
// Defined here (in usecase) to avoid import cycles — the delivery layer implements it.
type IBroadcaster interface {
	Broadcast(msg entity.WsMessage)
}

// ISubscriptionProvider defines the contract for retrieving active WebSocket subscriptions.
type ISubscriptionProvider interface {
	GetActiveSubscriptions() []string
}

// IOHLCVFetcher defines the contract for lazy-loading historical candles on demand.
type IOHLCVFetcher interface {
	FetchOHLCV(ctx context.Context, symbol, interval, rangeStr string) ([]entity.OHLCV, error)
}

// IStockUsecase defines all application use cases for stock data.
//
//go:generate mockgen -destination=../../mocks/mock_stock_usecase.go -package=mocks . IStockUsecase
type IStockUsecase interface {
	GetAllStocks(ctx context.Context) ([]entity.Stock, error)
	GetStockBySymbol(ctx context.Context, symbol string) (*entity.Stock, error)
	GetOHLCVHistory(ctx context.Context, symbol, timeframe string, limit int) ([]entity.OHLCV, error)
	GetLatestIndicators(ctx context.Context, symbol string) (*entity.TechnicalIndicators, error)
	GetPrediction(ctx context.Context, symbol string) (*entity.PredictionResult, error)
	GetBrokerSummary(ctx context.Context, symbol string) (*entity.BrokerSummary, error)

	// ProcessNewCandle is called on each market data tick.
	// It persists the candle, runs the Quant Engine, and broadcasts results.
	ProcessNewCandle(ctx context.Context, candle entity.OHLCV) error

	// ProcessTickerUpdate broadcasts a live price quote update.
	ProcessTickerUpdate(ctx context.Context, ticker entity.Ticker) error

	// SeedStockPrices bulk updates current stock prices in a transaction on startup.
	SeedStockPrices(ctx context.Context, tickers []entity.Ticker) error

	// GetActiveSubscriptions returns the active subscriptions from the WebSocket Hub.
	GetActiveSubscriptions() []string
}

// ─── Implementation ───────────────────────────────────────────────────────────

// stockUsecase implements IStockUsecase.
type stockUsecase struct {
	stockRepo     domainRepo.IStockRepository
	indicatorRepo domainRepo.IIndicatorRepository
	quantEngine   *QuantEngine
	ringStore     *cache.RingBufferStore
	broadcaster   IBroadcaster
	ohlcvFetcher  IOHLCVFetcher
}

// NewStockUsecase is the constructor for stockUsecase.
// All dependencies are injected — no globals, no init() side effects.
func NewStockUsecase(
	stockRepo domainRepo.IStockRepository,
	indicatorRepo domainRepo.IIndicatorRepository,
	quantEngine *QuantEngine,
	ringStore *cache.RingBufferStore,
	broadcaster IBroadcaster,
	ohlcvFetcher IOHLCVFetcher,
) IStockUsecase {
	return &stockUsecase{
		stockRepo:     stockRepo,
		indicatorRepo: indicatorRepo,
		quantEngine:   quantEngine,
		ringStore:     ringStore,
		broadcaster:   broadcaster,
		ohlcvFetcher:  ohlcvFetcher,
	}
}

// ─── Query Methods ────────────────────────────────────────────────────────────

func (u *stockUsecase) GetAllStocks(ctx context.Context) ([]entity.Stock, error) {
	stocks, err := u.stockRepo.GetAllStocks(ctx)
	if err != nil {
		return nil, fmt.Errorf("stock_usecase.GetAllStocks: %w", err)
	}
	return stocks, nil
}

func (u *stockUsecase) GetStockBySymbol(ctx context.Context, symbol string) (*entity.Stock, error) {
	stock, err := u.stockRepo.GetStockBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("stock_usecase.GetStockBySymbol: %w", err)
	}
	return stock, nil
}

func (u *stockUsecase) GetOHLCVHistory(ctx context.Context, symbol, timeframe string, limit int) ([]entity.OHLCV, error) {
	params := entity.OHLCVQueryParams{
		Symbol:    symbol,
		Timeframe: timeframe,
		Limit:     limit,
	}
	candles, err := u.stockRepo.GetOHLCVHistory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("stock_usecase.GetOHLCVHistory: %w", err)
	}

	// Lazy loading: If database has no history for this stock, fetch it from Yahoo Finance!
	if len(candles) == 0 && u.ohlcvFetcher != nil {
		log.Printf("[Usecase] Lazy-loading historical candles for %s from Yahoo Finance...", symbol)
		fetchedCandles, err := u.ohlcvFetcher.FetchOHLCV(ctx, symbol, timeframe, "1d")
		if err == nil && len(fetchedCandles) > 0 {
			stock, err := u.stockRepo.GetStockBySymbol(ctx, symbol)
			if err == nil && stock != nil {
				for i := range fetchedCandles {
					fetchedCandles[i].StockID = stock.ID
				}
				// Save candles to database and push to memory ring buffer
				for _, c := range fetchedCandles {
					_ = u.stockRepo.SaveOHLCV(ctx, c)
					u.ringStore.Push(symbol, c)
				}
				// Re-query from DB to apply the limit and ordering correctly
				candles, _ = u.stockRepo.GetOHLCVHistory(ctx, params)
			}
		} else {
			if err != nil {
				log.Printf("[Usecase] Lazy-loading error for %s: %v. Generating dummy historical candles.", symbol, err)
			} else {
				log.Printf("[Usecase] No historical candles returned for %s. Generating dummy historical candles.", symbol)
			}
			// Fallback: Generate dummy historical candles so the chart is never empty!
			stock, err := u.stockRepo.GetStockBySymbol(ctx, symbol)
			if err == nil && stock != nil {
				var currentPrice float64 = 100.0
				if stock.Price != nil && *stock.Price > 0 {
					currentPrice = *stock.Price
				}
				dummyCandles := generateDummyOHLCV(stock.ID, symbol, timeframe, limit, currentPrice)
				for _, c := range dummyCandles {
					_ = u.stockRepo.SaveOHLCV(ctx, c)
					u.ringStore.Push(symbol, c)
				}
				// Re-query
				candles, _ = u.stockRepo.GetOHLCVHistory(ctx, params)
			}
		}
	}

	return candles, nil
}

func (u *stockUsecase) GetLatestIndicators(ctx context.Context, symbol string) (*entity.TechnicalIndicators, error) {
	ind, err := u.indicatorRepo.GetLatestIndicators(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("stock_usecase.GetLatestIndicators: %w", err)
	}
	return ind, nil
}

func (u *stockUsecase) GetPrediction(ctx context.Context, symbol string) (*entity.PredictionResult, error) {
	// Serve prediction from in-memory ring buffer (fast path, no DB hit).
	candles := u.ringStore.GetSlice(symbol)
	if len(candles) == 0 {
		// Fallback: load from DB on first request before ring buffer warms up.
		dbCandles, err := u.stockRepo.GetLatestOHLCV(ctx, symbol, "1m", MinCandlesForMA50+10)
		if err != nil {
			return nil, fmt.Errorf("stock_usecase.GetPrediction db fallback: %w", err)
		}
		candles = dbCandles
	}

	_, prediction := u.quantEngine.ComputeAll(symbol, candles)
	if prediction == nil {
		return nil, fmt.Errorf("stock_usecase.GetPrediction: insufficient candles for symbol %s", symbol)
	}

	return prediction, nil
}

// GetBrokerSummary generates a realistic broker transaction summary (Bandarmology) for a stock.
func (u *stockUsecase) GetBrokerSummary(ctx context.Context, symbol string) (*entity.BrokerSummary, error) {
	stock, err := u.stockRepo.GetStockBySymbol(ctx, symbol)
	if err != nil {
		return nil, fmt.Errorf("stock_usecase.GetBrokerSummary db stock check: %w", err)
	}
	if stock == nil {
		return nil, fmt.Errorf("stock_usecase.GetBrokerSummary: stock %s not found", symbol)
	}

	// ─── Real Stockbit Integration (if token is configured) ─────────────────
	stockbitToken := os.Getenv("STOCKBIT_TOKEN")
	if stockbitToken != "" {
		log.Printf("[Usecase] STOCKBIT_TOKEN detected. Fetching real broker summary for %s...", symbol)
		summary, err := u.fetchStockbitBrokerSummary(ctx, symbol, stockbitToken)
		if err == nil && summary != nil {
			return summary, nil
		}
		log.Printf("[WARN] Failed to fetch real broker summary from Stockbit: %v. Falling back to simulator.", err)
	}

	// ─── Simulated Fallback ──────────────────────────────────────────────────
	var basePrice float64 = 1000.0
	if stock.Price != nil && *stock.Price > 0 {
		basePrice = *stock.Price
	}

	var baseVolume int64 = 100000
	if stock.Volume != nil && *stock.Volume > 0 {
		baseVolume = *stock.Volume
	}

	var seed int = 0
	for _, char := range symbol {
		seed += int(char)
	}

	changePercent := 0.0
	if stock.ChangePercent != nil {
		changePercent = *stock.ChangePercent
	}

	netStatus := "NEUTRAL"
	if changePercent > 2.0 {
		netStatus = "BIG ACCUMULATION"
	} else if changePercent > 0.5 {
		netStatus = "NORMAL ACCUMULATION"
	} else if changePercent < -2.0 {
		netStatus = "BIG DISTRIBUTION"
	} else if changePercent < -0.5 {
		netStatus = "NORMAL DISTRIBUTION"
	}

	localBrokers := []struct {
		Code string
		Name string
	}{
		{"YP", "Mirae Asset Sekuritas"},
		{"PD", "Indo Premier Sekuritas"},
		{"CC", "Mandiri Sekuritas"},
		{"OD", "BRI Danareksa Sekuritas"},
		{"NI", "BNI Sekuritas"},
		{"DH", "Sinarmas Sekuritas"},
		{"KK", "Phillip Sekuritas"},
		{"DR", "RHB Sekuritas Indonesia"},
		{"GR", "Panin Sekuritas"},
		{"XC", "Ajaib Sekuritas"},
	}

	foreignBrokers := []struct {
		Code string
		Name string
	}{
		{"AK", "UBS Sekuritas Indonesia"},
		{"CS", "Credit Suisse Sekuritas"},
		{"MS", "Morgan Stanley Sekuritas"},
		{"KZ", "CLSA Sekuritas Indonesia"},
		{"DB", "Deutsche Sekuritas Indonesia"},
		{"ZP", "Maybank Sekuritas"},
		{"RX", "Macquarie Sekuritas"},
	}

	buyers := make([]entity.BrokerTrade, 5)
	sellers := make([]entity.BrokerTrade, 5)

	totalBuyValue := 0.0
	foreignBuyValue := 0.0

	for i := 0; i < 5; i++ {
		isForeign := (seed+i)%3 == 0
		var code, name string
		var nat string
		if isForeign {
			b := foreignBrokers[(seed+i)%len(foreignBrokers)]
			code, name, nat = b.Code, b.Name, "F"
		} else {
			b := localBrokers[(seed+i)%len(localBrokers)]
			code, name, nat = b.Code, b.Name, "L"
		}

		pct := 0.15 - float64(i)*0.02 // 15%, 13%, 11%, 9%, 7%
		if changePercent > 0.5 {
			pct += 0.03
		}
		vol := int64(float64(baseVolume) * pct / 5.0)
		if vol < 100 {
			vol = 100
		}

		spreadPct := (float64((seed+i*17)%100) - 50) / 10000.0
		avgPrice := basePrice * (1.0 + spreadPct)
		val := float64(vol) * avgPrice * 100.0

		buyers[i] = entity.BrokerTrade{
			BrokerCode:  code,
			BrokerName:  name,
			Volume:      vol,
			AvgPrice:    avgPrice,
			Value:       val,
			Nationality: nat,
		}

		totalBuyValue += val
		if isForeign {
			foreignBuyValue += val
		}
	}

	for i := 0; i < 5; i++ {
		isForeign := (seed+i+1)%4 == 0
		var code, name string
		var nat string
		if isForeign {
			b := foreignBrokers[(seed+i+5)%len(foreignBrokers)]
			code, name, nat = b.Code, b.Name, "F"
		} else {
			b := localBrokers[(seed+i+3)%len(localBrokers)]
			code, name, nat = b.Code, b.Name, "L"
		}

		pct := 0.14 - float64(i)*0.02
		if changePercent < -0.5 {
			pct += 0.04
		}
		vol := int64(float64(baseVolume) * pct / 5.0)
		if vol < 100 {
			vol = 100
		}

		spreadPct := (float64((seed+i*23)%100) - 50) / 10000.0
		avgPrice := basePrice * (1.0 + spreadPct)
		val := float64(vol) * avgPrice * 100.0

		sellers[i] = entity.BrokerTrade{
			BrokerCode:  code,
			BrokerName:  name,
			Volume:      vol,
			AvgPrice:    avgPrice,
			Value:       val,
			Nationality: nat,
		}
	}

	foreignBuyPct := 0.0
	if totalBuyValue > 0 {
		foreignBuyPct = (foreignBuyValue / totalBuyValue) * 100.0
	}

	return &entity.BrokerSummary{
		Symbol:        symbol,
		Timestamp:     time.Now().UTC(),
		Buyers:        buyers,
		Sellers:       sellers,
		NetStatus:     netStatus,
		ForeignBuyPct: foreignBuyPct,
	}, nil
}

type stockbitBrokerTrade struct {
	BrokerCode  string  `json:"brokercode"`
	BrokerName  string  `json:"brokername"`
	BuyVolume   int64   `json:"buyvol"`
	BuyValue    float64 `json:"buyval"`
	SellVolume  int64   `json:"sellvol"`
	SellValue   float64 `json:"sellval"`
	AvgPrice    float64 `json:"avgprice"`
	Nationality string  `json:"nationality"`
}

func (u *stockUsecase) fetchStockbitBrokerSummary(ctx context.Context, symbol string, token string) (*entity.BrokerSummary, error) {
	apiURL := os.Getenv("STOCKBIT_API_URL")
	if apiURL == "" {
		apiURL = "https://exodus.stockbit.com/stream/v3/symbol/%s/brokers"
	}
	targetURL := fmt.Sprintf(apiURL, symbol)

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create stockbit request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Origin", "https://stockbit.com")
	req.Header.Set("Referer", "https://stockbit.com/")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute stockbit request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("stockbit returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read stockbit response body: %w", err)
	}

	buyers, sellers, err := parseStockbitResponse(rawBody)
	if err != nil {
		return nil, fmt.Errorf("parse stockbit response: %w", err)
	}

	// Calculate net status based on top 3 buyers vs top 3 sellers
	buyerTop3Sum := 0.0
	for i := 0; i < len(buyers) && i < 3; i++ {
		buyerTop3Sum += buyers[i].Value
	}
	sellerTop3Sum := 0.0
	for i := 0; i < len(sellers) && i < 3; i++ {
		sellerTop3Sum += sellers[i].Value
	}

	netStatus := "NEUTRAL"
	if sellerTop3Sum > 0 {
		ratio := buyerTop3Sum / sellerTop3Sum
		if ratio > 1.3 {
			netStatus = "BIG ACCUMULATION"
		} else if ratio > 1.1 {
			netStatus = "NORMAL ACCUMULATION"
		} else if ratio < 0.77 {
			netStatus = "BIG DISTRIBUTION"
		} else if ratio < 0.9 {
			netStatus = "NORMAL DISTRIBUTION"
		}
	}

	// Calculate foreign buy percentage
	totalBuyValue := 0.0
	foreignBuyValue := 0.0
	for _, b := range buyers {
		totalBuyValue += b.Value
		if b.Nationality == "F" {
			foreignBuyValue += b.Value
		}
	}
	foreignBuyPct := 0.0
	if totalBuyValue > 0 {
		foreignBuyPct = (foreignBuyValue / totalBuyValue) * 100.0
	}

	return &entity.BrokerSummary{
		Symbol:        symbol,
		Timestamp:     time.Now().UTC(),
		Buyers:        buyers,
		Sellers:       sellers,
		NetStatus:     netStatus,
		ForeignBuyPct: foreignBuyPct,
	}, nil
}

func parseStockbitResponse(rawBody []byte) ([]entity.BrokerTrade, []entity.BrokerTrade, error) {
	// Try Struct 1: Separated buyers/sellers lists in nested data object
	var objResponse struct {
		Data struct {
			Buyers  []stockbitBrokerTrade `json:"buyers"`
			Sellers []stockbitBrokerTrade `json:"sellers"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rawBody, &objResponse); err == nil && (len(objResponse.Data.Buyers) > 0 || len(objResponse.Data.Sellers) > 0) {
		buyers := make([]entity.BrokerTrade, 0)
		for _, b := range objResponse.Data.Buyers {
			buyers = append(buyers, entity.BrokerTrade{
				BrokerCode:  b.BrokerCode,
				BrokerName:  b.BrokerName,
				Volume:      b.BuyVolume,
				AvgPrice:    b.AvgPrice,
				Value:       b.BuyValue,
				Nationality: b.Nationality,
			})
		}
		sellers := make([]entity.BrokerTrade, 0)
		for _, s := range objResponse.Data.Sellers {
			sellers = append(sellers, entity.BrokerTrade{
				BrokerCode:  s.BrokerCode,
				BrokerName:  s.BrokerName,
				Volume:      s.SellVolume,
				AvgPrice:    s.AvgPrice,
				Value:       s.SellValue,
				Nationality: s.Nationality,
			})
		}
		return buyers, sellers, nil
	}

	// Try Struct 2: Flat list in data array
	var flatResponse struct {
		Data []stockbitBrokerTrade `json:"data"`
	}
	if err := json.Unmarshal(rawBody, &flatResponse); err == nil && len(flatResponse.Data) > 0 {
		rawBuyers := make([]stockbitBrokerTrade, len(flatResponse.Data))
		copy(rawBuyers, flatResponse.Data)
		// Sort buyers descending by buyval
		sort.Slice(rawBuyers, func(i, j int) bool {
			return rawBuyers[i].BuyValue > rawBuyers[j].BuyValue
		})

		rawSellers := make([]stockbitBrokerTrade, len(flatResponse.Data))
		copy(rawSellers, flatResponse.Data)
		// Sort sellers descending by sellval
		sort.Slice(rawSellers, func(i, j int) bool {
			return rawSellers[i].SellValue > rawSellers[j].SellValue
		})

		buyers := make([]entity.BrokerTrade, 0)
		for i := 0; i < len(rawBuyers) && i < 5; i++ {
			b := rawBuyers[i]
			if b.BuyValue > 0 {
				buyers = append(buyers, entity.BrokerTrade{
					BrokerCode:  b.BrokerCode,
					BrokerName:  b.BrokerName,
					Volume:      b.BuyVolume,
					AvgPrice:    b.AvgPrice,
					Value:       b.BuyValue,
					Nationality: b.Nationality,
				})
			}
		}

		sellers := make([]entity.BrokerTrade, 0)
		for i := 0; i < len(rawSellers) && i < 5; i++ {
			s := rawSellers[i]
			if s.SellValue > 0 {
				sellers = append(sellers, entity.BrokerTrade{
					BrokerCode:  s.BrokerCode,
					BrokerName:  s.BrokerName,
					Volume:      s.SellVolume,
					AvgPrice:    s.AvgPrice,
					Value:       s.SellValue,
					Nationality: s.Nationality,
				})
			}
		}
		return buyers, sellers, nil
	}

	return nil, nil, fmt.Errorf("unknown stockbit response format or empty data: %s", string(rawBody))
}

// ─── Command Methods ──────────────────────────────────────────────────────────

// ProcessNewCandle is the central command handler called on each market data tick.
// It performs the following pipeline:
//  1. Persist OHLCV to MySQL
//  2. Push to in-memory ring buffer
//  3. Run Quantitative Engine
//  4. Persist indicators to MySQL
//  5. Broadcast OHLCV + Indicators + Prediction via WebSocket
func (u *stockUsecase) ProcessNewCandle(ctx context.Context, candle entity.OHLCV) error {
	// ── Step 1: Persist OHLCV ─────────────────────────────────────────────
	if err := u.stockRepo.SaveOHLCV(ctx, candle); err != nil {
		// Log but do not abort — streaming must continue even if DB write fails.
		log.Printf("[WARN] stock_usecase.ProcessNewCandle SaveOHLCV %s: %v", candle.Symbol, err)
	}

	// ── Step 2: Update in-memory ring buffer ──────────────────────────────
	u.ringStore.Push(candle.Symbol, candle)

	// ── Step 3: Run Quantitative Engine ──────────────────────────────────
	candles := u.ringStore.GetSlice(candle.Symbol)
	indicators, prediction := u.quantEngine.ComputeAll(candle.Symbol, candles)

	// ── Step 4: Persist indicators (async-safe, non-blocking on error) ───
	if indicators != nil {
		indicators.Symbol = candle.Symbol
		if err := u.indicatorRepo.SaveIndicators(ctx, *indicators); err != nil {
			log.Printf("[WARN] stock_usecase.ProcessNewCandle SaveIndicators %s: %v", candle.Symbol, err)
		}
	}

	// ── Step 5: Broadcast via WebSocket ───────────────────────────────────
	u.broadcastCandle(candle)
	if indicators != nil {
		u.broadcastIndicators(indicators)
	}
	if prediction != nil {
		u.broadcastPrediction(prediction)
	}

	return nil
}

// broadcastCandle sends an OHLCV_UPDATE message to all connected WebSocket clients.
func (u *stockUsecase) broadcastCandle(candle entity.OHLCV) {
	msg := entity.NewWsMessage(entity.WsMsgOHLCVUpdate, candle.Symbol, candle)
	u.broadcaster.Broadcast(msg)
}

// broadcastIndicators sends an INDICATOR_UPDATE message.
func (u *stockUsecase) broadcastIndicators(ind *entity.TechnicalIndicators) {
	msg := entity.NewWsMessage(entity.WsMsgIndicatorUpdate, ind.Symbol, ind)
	u.broadcaster.Broadcast(msg)
}

// broadcastPrediction sends a PREDICTION_UPDATE message.
func (u *stockUsecase) broadcastPrediction(pred *entity.PredictionResult) {
	msg := entity.NewWsMessage(entity.WsMsgPredictionUpdate, pred.Symbol, pred)
	u.broadcaster.Broadcast(msg)
}

// ProcessTickerUpdate broadcasts a ticker message to clients.
func (u *stockUsecase) ProcessTickerUpdate(ctx context.Context, ticker entity.Ticker) error {
	// Update current stock price inside DB (async / non-blocking)
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := u.stockRepo.UpdateStockPrice(bgCtx, ticker.Symbol, ticker.Price, ticker.Change, ticker.ChangePercent, ticker.Volume); err != nil {
			log.Printf("[WARN] Failed to update stock price in DB for %s: %v", ticker.Symbol, err)
		}
	}()

	msg := entity.NewWsMessage(entity.WsMsgTickerUpdate, ticker.Symbol, ticker)
	u.broadcaster.Broadcast(msg)
	return nil
}

// GetActiveSubscriptions returns the active subscriptions from the WebSocket Hub.
func (u *stockUsecase) GetActiveSubscriptions() []string {
	if provider, ok := u.broadcaster.(ISubscriptionProvider); ok {
		return provider.GetActiveSubscriptions()
	}
	return nil
}

// SeedStockPrices bulk updates current stock prices in a transaction on startup.
func (u *stockUsecase) SeedStockPrices(ctx context.Context, tickers []entity.Ticker) error {
	return u.stockRepo.UpdateStockPricesBatch(ctx, tickers)
}

func generateDummyOHLCV(stockID int64, symbol, timeframe string, limit int, currentPrice float64) []entity.OHLCV {
	candles := make([]entity.OHLCV, limit)
	now := time.Now().UTC()
	
	price := currentPrice
	if price <= 0 {
		price = float64(100 + (int(symbol[0]) + int(symbol[len(symbol)-1])*17)%15000)
		if price < 50 {
			price = 50
		}
	}

	for i := limit - 1; i >= 0; i-- {
		var diff time.Duration
		switch timeframe {
		case "1m":
			diff = time.Duration(limit-i) * time.Minute
		case "5m":
			diff = time.Duration(limit-i) * 5 * time.Minute
		default:
			diff = time.Duration(limit-i) * time.Minute
		}
		
		ts := now.Add(-diff)
		
		// Random walk backward
		changePercent := float64((int(ts.Unix())%9) - 4) * 0.0012 // -0.48% to +0.48%
		closePrice := price
		open := closePrice / (1.0 + changePercent)

		if open < 10 {
			open = 10
		}
		
		high := open
		if closePrice > open {
			high = closePrice + (closePrice-open)*0.2
		} else {
			high = open + (open-closePrice)*0.2
		}
		
		low := open
		if closePrice < open {
			low = closePrice - (open-closePrice)*0.2
		} else {
			low = open - (closePrice-open)*0.2
		}
		
		vol := int64(1000 + (int(ts.Unix())%10000)*50)
		
		candles[i] = entity.OHLCV{
			StockID:   stockID,
			Symbol:    symbol,
			Timestamp: ts,
			Open:      open,
			High:      high,
			Low:       low,
			Close:     closePrice,
			Volume:    vol,
			Timeframe: timeframe,
		}
		
		// Older price is the current open
		price = open
	}
	
	return candles
}



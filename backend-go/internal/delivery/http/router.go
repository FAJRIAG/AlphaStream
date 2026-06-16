// Package http provides the Gin router setup for AlphaStream's REST API.
package http

import (
	"github.com/gin-gonic/gin"

	"github.com/alphastream/backend-go/internal/delivery/http/handler"
	"github.com/alphastream/backend-go/internal/delivery/http/middleware"
	wsDelivery "github.com/alphastream/backend-go/internal/delivery/websocket"
)

// RouterConfig holds all handler dependencies needed to build the router.
type RouterConfig struct {
	StockHandler     *handler.StockHandler
	IndicatorHandler *handler.IndicatorHandler
	WsHandler        *wsDelivery.WsHandler
	AppEnv           string
}

// NewRouter constructs and returns a configured *gin.Engine.
// In production mode, Gin's debug logging is disabled automatically.
func NewRouter(cfg RouterConfig) *gin.Engine {
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New() // gin.New() instead of gin.Default() — we supply our own middleware

	// ── Global Middleware ────────────────────────────────────────────────
	router.Use(gin.Recovery())          // Recover from panics, return 500
	router.Use(middleware.CORS())       // CORS headers for frontend
	router.Use(middleware.Logger())     // Structured request logging

	// ── Health Check ─────────────────────────────────────────────────────
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "alphastream-backend"})
	})

	// ── WebSocket ─────────────────────────────────────────────────────────
	router.GET("/ws", cfg.WsHandler.ServeWS)

	// ── REST API v1 ───────────────────────────────────────────────────────
	v1 := router.Group("/api/v1")
	{
		// Stock endpoints
		stocks := v1.Group("/stocks")
		{
			stocks.GET("", cfg.StockHandler.GetAllStocks)
			stocks.GET("/recommendations", cfg.StockHandler.GetBuyRecommendations)
			stocks.GET("/:symbol", cfg.StockHandler.GetStockBySymbol)
			stocks.GET("/:symbol/ohlcv", cfg.StockHandler.GetOHLCVHistory)
			stocks.GET("/:symbol/prediction", cfg.StockHandler.GetPrediction)
			stocks.GET("/:symbol/broker-summary", cfg.StockHandler.GetBrokerSummary)
		}

		// Indicator endpoints
		indicators := v1.Group("/indicators")
		{
			indicators.GET("/:symbol", cfg.IndicatorHandler.GetLatestIndicators)
			indicators.GET("/:symbol/summary", cfg.IndicatorHandler.GetIndicatorsSummary)
		}
	}

	return router
}

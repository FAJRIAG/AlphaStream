// Package handler provides the HTTP handler for stock-related REST endpoints.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alphastream/backend-go/internal/usecase"
)

// ─── Response Shapes ──────────────────────────────────────────────────────────

// apiResponse is the standard JSON envelope for all API responses.
type apiResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func successResponse(data interface{}) apiResponse {
	return apiResponse{Success: true, Data: data}
}

func errorResponse(msg string) apiResponse {
	return apiResponse{Success: false, Error: msg}
}

// ─── StockHandler ─────────────────────────────────────────────────────────────

// StockHandler handles all /api/v1/stocks endpoints.
type StockHandler struct {
	usecase usecase.IStockUsecase
}

// NewStockHandler creates a StockHandler with injected usecase.
func NewStockHandler(uc usecase.IStockUsecase) *StockHandler {
	return &StockHandler{usecase: uc}
}

// GetAllStocks handles GET /api/v1/stocks
// Returns all active stocks.
func (h *StockHandler) GetAllStocks(c *gin.Context) {
	stocks, err := h.usecase.GetAllStocks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to retrieve stocks"))
		return
	}

	c.JSON(http.StatusOK, successResponse(stocks))
}

// GetStockBySymbol handles GET /api/v1/stocks/:symbol
// Returns a single stock detail.
func (h *StockHandler) GetStockBySymbol(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, errorResponse("Symbol parameter is required"))
		return
	}

	stock, err := h.usecase.GetStockBySymbol(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to retrieve stock"))
		return
	}
	if stock == nil {
		c.JSON(http.StatusNotFound, errorResponse("Stock not found"))
		return
	}

	c.JSON(http.StatusOK, successResponse(stock))
}

// GetOHLCVHistory handles GET /api/v1/stocks/:symbol/ohlcv
// Query params: timeframe (default "1m"), limit (default 200, max 1000)
func (h *StockHandler) GetOHLCVHistory(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, errorResponse("Symbol parameter is required"))
		return
	}

	timeframe := c.DefaultQuery("timeframe", "1m")
	limitStr := c.DefaultQuery("limit", "200")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse("limit must be a positive integer"))
		return
	}
	if limit > 1000 {
		limit = 1000 // Hard cap to prevent runaway queries
	}

	candles, err := h.usecase.GetOHLCVHistory(c.Request.Context(), symbol, timeframe, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to retrieve OHLCV history"))
		return
	}

	c.JSON(http.StatusOK, successResponse(candles))
}

// GetPrediction handles GET /api/v1/stocks/:symbol/prediction
// Returns the latest prediction result from the Quantitative Engine.
func (h *StockHandler) GetPrediction(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, errorResponse("Symbol parameter is required"))
		return
	}

	prediction, err := h.usecase.GetPrediction(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(prediction))
}

// GetBrokerSummary handles GET /api/v1/stocks/:symbol/broker-summary
// Returns the simulated/generated broker summary for a stock.
func (h *StockHandler) GetBrokerSummary(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, errorResponse("Symbol parameter is required"))
		return
	}

	summary, err := h.usecase.GetBrokerSummary(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}

	c.JSON(http.StatusOK, successResponse(summary))
}

// GetBuyRecommendations handles GET /api/v1/stocks/recommendations
// Returns a list of recommended buy stocks.
func (h *StockHandler) GetBuyRecommendations(c *gin.Context) {
	recommendations, err := h.usecase.GetBuyRecommendations(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to retrieve buy recommendations"))
		return
	}

	c.JSON(http.StatusOK, successResponse(recommendations))
}



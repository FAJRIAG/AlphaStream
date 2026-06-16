// Package handler provides the HTTP handler for technical indicator REST endpoints.
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/alphastream/backend-go/internal/usecase"
)

// IndicatorHandler handles all /api/v1/indicators endpoints.
type IndicatorHandler struct {
	usecase usecase.IStockUsecase
}

// NewIndicatorHandler creates an IndicatorHandler with injected usecase.
func NewIndicatorHandler(uc usecase.IStockUsecase) *IndicatorHandler {
	return &IndicatorHandler{usecase: uc}
}

// GetLatestIndicators handles GET /api/v1/indicators/:symbol
// Returns the most recently computed MA, RSI, ATR, and cross signal values.
func (h *IndicatorHandler) GetLatestIndicators(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, errorResponse("Symbol parameter is required"))
		return
	}

	indicators, err := h.usecase.GetLatestIndicators(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to retrieve indicators"))
		return
	}
	if indicators == nil {
		c.JSON(http.StatusNotFound, errorResponse("No indicator data found for symbol. Engine may still be warming up."))
		return
	}

	c.JSON(http.StatusOK, successResponse(indicators))
}

// GetIndicatorsSummary handles GET /api/v1/indicators/:symbol/summary
// Returns the prediction result along with the latest indicators in one response.
func (h *IndicatorHandler) GetIndicatorsSummary(c *gin.Context) {
	symbol := c.Param("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, errorResponse("Symbol parameter is required"))
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, errorResponse("limit must be a positive integer"))
		return
	}

	indicators, err := h.usecase.GetLatestIndicators(c.Request.Context(), symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse("Failed to retrieve indicators"))
		return
	}

	prediction, err := h.usecase.GetPrediction(c.Request.Context(), symbol)
	if err != nil {
		// Prediction may fail if engine hasn't warmed up — return partial response.
		c.JSON(http.StatusOK, successResponse(map[string]interface{}{
			"indicators": indicators,
			"prediction": nil,
		}))
		return
	}

	c.JSON(http.StatusOK, successResponse(map[string]interface{}{
		"indicators": indicators,
		"prediction": prediction,
	}))
}

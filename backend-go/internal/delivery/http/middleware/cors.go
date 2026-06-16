// Package middleware provides reusable Gin middleware for the AlphaStream HTTP server.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that adds Cross-Origin Resource Sharing headers.
// This allows the Next.js frontend (typically on port 3000) to call the Go API (port 8080).
//
// Security note: AllowOrigins is intentionally permissive for development.
// For production, restrict to the specific deployed frontend URL.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Max-Age", "86400") // Cache preflight for 24h

		// Handle preflight OPTIONS request immediately without hitting any handler.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

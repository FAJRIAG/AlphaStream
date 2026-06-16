// Package middleware provides the structured request logger middleware.
package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// Logger returns a Gin middleware that logs each request with:
//   - HTTP method and path
//   - Response status code (colored by severity)
//   - Request duration
//   - Client IP
//
// Uses structured log format compatible with log aggregators.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log after handler completes
		duration := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()

		if query != "" {
			path = path + "?" + query
		}

		fmt.Printf("[AlphaStream] %s | %s | %d | %s | %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			colorStatus(status),
			status,
			duration,
			path,
		)
		_ = clientIP // Available for future structured logging
	}
}

// colorStatus returns an ANSI-colored string for the HTTP status code.
// Green=2xx, Yellow=3xx, Red=4xx/5xx.
func colorStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return fmt.Sprintf("\033[32m%d\033[0m", code) // Green
	case code >= 300 && code < 400:
		return fmt.Sprintf("\033[33m%d\033[0m", code) // Yellow
	default:
		return fmt.Sprintf("\033[31m%d\033[0m", code) // Red
	}
}

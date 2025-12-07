package middleware

import (
	"time"

	"cinemaos-backend/internal/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponseTimeMiddleware logs the response time for each request
// This is the implementation for Exercise 3 in LEARNING_GUIDE.md
func ResponseTimeMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		start := time.Now()

		// Process request - this calls other handlers and middleware
		c.Next()

		// Calculate duration after request is complete
		duration := time.Since(start)

		// Log the response time with request details
		log.Info("request completed",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
		)

		// Also set as header so clients can see it
		c.Header("X-Response-Time", duration.String())
	}
}

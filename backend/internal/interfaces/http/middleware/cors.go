package middleware

import (
	"net/http"
	"time"

	"cinemaos-backend/config"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Check if origin is allowed
		allowed := false
		for _, o := range cfg.AllowOrigins {
			if o == "*" || o == origin {
				allowed = true
				break
			}
		}

		if allowed {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Set CORS headers
		c.Writer.Header().Set("Access-Control-Allow-Methods", joinStrings(cfg.AllowMethods))
		c.Writer.Header().Set("Access-Control-Allow-Headers", joinStrings(cfg.AllowHeaders))
		c.Writer.Header().Set("Access-Control-Expose-Headers", joinStrings(cfg.ExposeHeaders))
		
		if cfg.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if cfg.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", itoa(cfg.MaxAge))
		}

		// Handle preflight request
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware provides basic rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimit returns a rate limiting middleware
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// Clean old requests
		if timestamps, exists := rl.requests[clientIP]; exists {
			var valid []time.Time
			for _, t := range timestamps {
				if now.Sub(t) < rl.window {
					valid = append(valid, t)
				}
			}
			rl.requests[clientIP] = valid
		}

		// Check limit
		if len(rl.requests[clientIP]) >= rl.limit {
			c.Header("Retry-After", itoa(int(rl.window.Seconds())))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "TOO_MANY_REQUESTS",
					"message": "Rate limit exceeded. Please try again later.",
				},
			})
			return
		}

		// Record request
		rl.requests[clientIP] = append(rl.requests[clientIP], now)

		// Set rate limit headers
		c.Header("X-RateLimit-Limit", itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", itoa(rl.limit-len(rl.requests[clientIP])))

		c.Next()
	}
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Note: For production, consider using context with timeout
		// ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		// defer cancel()
		// c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// SecureHeadersMiddleware adds security headers
func SecureHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")
		
		c.Next()
	}
}

// Helper functions
func joinStrings(strs []string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += ", " + strs[i]
	}
	return result
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

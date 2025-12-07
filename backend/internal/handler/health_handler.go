package handler

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"cinemaos-backend/internal/config"

	"github.com/gin-gonic/gin"
)

// HealthChecker interface for health checks
type HealthChecker interface {
	Health(ctx context.Context) error
}

// HealthHandler handles health check endpoints
type HealthHandler struct {
	cfg      *config.Config
	db       HealthChecker
	redis    HealthChecker
	startTime time.Time
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config, db, redis HealthChecker) *HealthHandler {
	return &HealthHandler{
		cfg:       cfg,
		db:        db,
		redis:     redis,
		startTime: time.Now(),
	}
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status      string                 `json:"status"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Uptime      string                 `json:"uptime"`
	Checks      map[string]CheckStatus `json:"checks,omitempty"`
}

// CheckStatus represents individual health check status
type CheckStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// Health godoc
// @Summary Health check
// @Description Basic health check endpoint
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status:      "healthy",
		Version:     h.cfg.App.Version,
		Environment: h.cfg.App.Environment,
		Uptime:      time.Since(h.startTime).String(),
	})
}

// HealthDetailed godoc
// @Summary Detailed health check
// @Description Health check with dependency status
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health/ready [get]
func (h *HealthHandler) HealthDetailed(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]CheckStatus)
	overallStatus := "healthy"

	// Check database
	if h.db != nil {
		if err := h.db.Health(ctx); err != nil {
			checks["database"] = CheckStatus{Status: "unhealthy", Message: err.Error()}
			overallStatus = "unhealthy"
		} else {
			checks["database"] = CheckStatus{Status: "healthy"}
		}
	}

	// Check Redis
	if h.redis != nil {
		if err := h.redis.Health(ctx); err != nil {
			checks["redis"] = CheckStatus{Status: "unhealthy", Message: err.Error()}
			overallStatus = "degraded" // Redis might be optional
		} else {
			checks["redis"] = CheckStatus{Status: "healthy"}
		}
	}

	resp := HealthResponse{
		Status:      overallStatus,
		Version:     h.cfg.App.Version,
		Environment: h.cfg.App.Environment,
		Uptime:      time.Since(h.startTime).String(),
		Checks:      checks,
	}

	status := http.StatusOK
	if overallStatus == "unhealthy" {
		status = http.StatusServiceUnavailable
	}

	c.JSON(status, resp)
}

// Live godoc
// @Summary Liveness probe
// @Description Kubernetes liveness probe
// @Tags health
// @Success 200 {string} string "OK"
// @Router /health/live [get]
func (h *HealthHandler) Live(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}

// Info godoc
// @Summary Application info
// @Description Get application information
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /info [get]
func (h *HealthHandler) Info(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"app": gin.H{
			"name":        h.cfg.App.Name,
			"version":     h.cfg.App.Version,
			"environment": h.cfg.App.Environment,
		},
		"runtime": gin.H{
			"go_version":   runtime.Version(),
			"num_cpu":      runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
		},
		"uptime": time.Since(h.startTime).String(),
	})
}

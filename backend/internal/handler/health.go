package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"backend/internal/jobs"
)

// HealthHandler handles health check endpoints
// Reference: BACKEND-IMPLEMENTATION-ANALYSIS.md - Recommendation #5
type HealthHandler struct {
	db        *gorm.DB
	redis     *redis.Client
	scheduler *jobs.Scheduler
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *gorm.DB, redis *redis.Client, scheduler *jobs.Scheduler) *HealthHandler {
	return &HealthHandler{
		db:        db,
		redis:     redis,
		scheduler: scheduler,
	}
}

// Liveness checks if the application is alive
// GET /health
// Used by Kubernetes liveness probe
func (h *HealthHandler) Liveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
		"time":   time.Now().Unix(),
	})
}

// Readiness checks if the application can serve traffic
// GET /ready
// Used by Kubernetes readiness probe
// Checks database and Redis connectivity
func (h *HealthHandler) Readiness(c *gin.Context) {
	checks := make(map[string]string)
	overallHealthy := true

	// Check database connection
	sqlDB, err := h.db.DB()
	if err != nil {
		checks["database"] = "unhealthy: failed to get DB instance"
		overallHealthy = false
	} else if err := sqlDB.Ping(); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		overallHealthy = false
	} else {
		checks["database"] = "healthy"
	}

	// Check Redis connection (if provided)
	if h.redis != nil {
		ctx := c.Request.Context()
		if err := h.redis.Ping(ctx).Err(); err != nil {
			checks["redis"] = "degraded: " + err.Error()
			// Redis is non-critical, don't fail readiness
		} else {
			checks["redis"] = "healthy"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	// Check job scheduler status (if provided)
	schedulerInfo := make(map[string]interface{})
	if h.scheduler != nil {
		if h.scheduler.IsRunning() {
			checks["scheduler"] = "healthy"
			schedulerInfo["status"] = "running"

			// Check if scheduler is stale (no cleanup in last 2 hours)
			lastCleanup := h.scheduler.GetLastCleanupTime()
			if !lastCleanup.IsZero() {
				schedulerInfo["last_cleanup"] = lastCleanup.Unix()
				timeSinceCleanup := time.Since(lastCleanup)
				schedulerInfo["hours_since_cleanup"] = timeSinceCleanup.Hours()

				// If no cleanup in last 2 hours, mark as degraded
				if timeSinceCleanup > 2*time.Hour {
					checks["scheduler"] = "degraded: no cleanup in last 2 hours"
				}
			} else {
				schedulerInfo["last_cleanup"] = "never"
			}
		} else {
			checks["scheduler"] = "unhealthy: not running"
			schedulerInfo["status"] = "stopped"
			// Scheduler stopped is non-critical for readiness, but worth noting
		}
	} else {
		checks["scheduler"] = "not_configured"
		schedulerInfo["status"] = "disabled"
	}

	status := "ready"
	statusCode := http.StatusOK
	if !overallHealthy {
		status = "not_ready"
		statusCode = http.StatusServiceUnavailable
	}

	response := gin.H{
		"status": status,
		"checks": checks,
		"time":   time.Now().Unix(),
	}

	// Add scheduler info if available
	if h.scheduler != nil {
		response["scheduler"] = schedulerInfo
	}

	c.JSON(statusCode, response)
}

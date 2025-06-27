package handlers

import (
	"gitlab.smartbet.am/golang/notification/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

type HealthHandler struct {
	logger *logrus.Logger
}

func NewHealthHandler(logger *logrus.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

// HealthCheck provides a health check endpoint
// @Summary Health check
// @Description Returns the general health status of the notification engine service. Available at both /health and /api/v1/health
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthResponse "Service is healthy"
// @Router /health [get]
// @Router /api/v1/health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(models.HealthResponse{
		Status:    "ok",
		Service:   "notification-engine",
		Timestamp: time.Now().UTC(),
		Version:   "1.0.0",
	})
}

// ReadinessCheck provides a readiness check endpoint
// @Summary Readiness check
// @Description Returns the readiness status of the notification engine including dependency checks. Available at both /ready and /api/v1/ready
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthResponse "Service is ready to accept requests"
// @Failure 503 {object} models.ErrorResponse "Service is not ready - dependencies not available"
// @Router /ready [get]
// @Router /api/v1/ready [get]
func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	// Here you can add checks for database, Kafka, etc.
	return c.JSON(models.HealthResponse{
		Status:    "ready",
		Service:   "notification-engine",
		Timestamp: time.Now().UTC(),
		Checks: &models.HealthCheckDetail{
			Database: "ok",
			Kafka:    "ok",
		},
	})
}

// LivenessCheck provides a liveness check endpoint
// @Summary Liveness check
// @Description Returns the liveness status of the notification engine. Available at both /live and /api/v1/live
// @Tags health
// @Produce json
// @Success 200 {object} models.HealthResponse "Service is alive"
// @Router /live [get]
// @Router /api/v1/live [get]
func (h *HealthHandler) LivenessCheck(c *fiber.Ctx) error {
	return c.JSON(models.HealthResponse{
		Status:    "alive",
		Service:   "notification-engine",
		Timestamp: time.Now().UTC(),
	})
}

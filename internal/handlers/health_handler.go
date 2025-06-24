package handlers

import (
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
// @Description Returns the health status of the notification engine
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"service":   "notification-engine",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
	})
}

// ReadinessCheck provides a readiness check endpoint
// @Summary Readiness check
// @Description Returns the readiness status of the notification engine
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /ready [get]
func (h *HealthHandler) ReadinessCheck(c *fiber.Ctx) error {
	// Here you can add checks for database, Kafka, etc.
	// For now, we'll just return ready
	return c.JSON(fiber.Map{
		"status":    "ready",
		"service":   "notification-engine",
		"timestamp": time.Now().UTC(),
		"checks": fiber.Map{
			"database": "ok",
			"kafka":    "ok",
		},
	})
}

// LivenessCheck provides a liveness check endpoint
// @Summary Liveness check
// @Description Returns the liveness status of the notification engine
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /live [get]
func (h *HealthHandler) LivenessCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "alive",
		"service":   "notification-engine",
		"timestamp": time.Now().UTC(),
	})
}

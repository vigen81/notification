package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"go.uber.org/zap"
)

type ConfigHandler struct {
	configRepo *repository.PartnerConfigRepository
	logger     *zap.Logger
}

func NewConfigHandler(configRepo *repository.PartnerConfigRepository, logger *zap.Logger) *ConfigHandler {
	return &ConfigHandler{
		configRepo: configRepo,
		logger:     logger,
	}
}

func (h *ConfigHandler) GetConfig(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(int64)

	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		h.logger.Error("failed to get config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve configuration",
			"code":  "CONFIG_ERROR",
		})
	}

	return c.JSON(config)
}

func (h *ConfigHandler) UpdateConfig(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(int64)

	var config models.PartnerConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	config.TenantID = tenantID

	if err := h.configRepo.Save(context.Background(), &config); err != nil {
		h.logger.Error("failed to save config", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
			"code":  "SAVE_ERROR",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Configuration updated successfully",
	})
}

func (h *ConfigHandler) AddEmailProvider(c *fiber.Ctx) error {
	// Implementation for adding email provider
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *ConfigHandler) AddSMSProvider(c *fiber.Ctx) error {
	// Implementation for adding SMS provider
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *ConfigHandler) AddPushProvider(c *fiber.Ctx) error {
	// Implementation for adding push provider
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *ConfigHandler) RemoveProvider(c *fiber.Ctx) error {
	// Implementation for removing provider
	return c.JSON(fiber.Map{"status": "ok"})
}

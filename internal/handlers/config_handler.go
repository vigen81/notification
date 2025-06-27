package handlers

import (
	"context"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
)

type ConfigHandler struct {
	configRepo *repository.PartnerConfigRepository
	logger     *logrus.Logger
}

func NewConfigHandler(configRepo *repository.PartnerConfigRepository, logger *logrus.Logger) *ConfigHandler {
	return &ConfigHandler{
		configRepo: configRepo,
		logger:     logger,
	}
}

// GetConfig retrieves the configuration for a specific tenant
// @Summary Get partner configuration
// @Description Get the configuration for a specific tenant by tenant_id
// @Tags configuration
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Success 200 {object} models.PartnerConfig
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /config/{tenant_id} [get]
func (h *ConfigHandler) GetConfig(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	if tenantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant ID is required",
			"code":  "MISSING_TENANT_ID",
		})
	}

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tenant ID",
			"code":  "INVALID_TENANT_ID",
		})
	}

	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		logger.WithTenant(tenantID).Error("Failed to get config", err, map[string]interface{}{})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve configuration",
			"code":  "CONFIG_ERROR",
		})
	}

	return c.JSON(config)
}

// UpdateConfig updates the configuration for a specific tenant
// @Summary Update partner configuration
// @Description Update the configuration for a specific tenant
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param config body models.PartnerConfigRequest true "Configuration update request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /config/{tenant_id} [put]
func (h *ConfigHandler) UpdateConfig(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	if tenantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Tenant ID is required",
			"code":  "MISSING_TENANT_ID",
		})
	}

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tenant ID",
			"code":  "INVALID_TENANT_ID",
		})
	}

	var req models.PartnerConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Get existing config or create new one
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		// Create new config if not found
		config = &models.PartnerConfig{
			TenantID: tenantID,
		}
	}

	// Update config fields
	config.EmailProviders = req.EmailProviders
	config.SMSProviders = req.SMSProviders
	config.PushProviders = req.PushProviders
	config.BatchConfig = req.BatchConfig
	config.RateLimits = req.RateLimits
	config.Enabled = req.Enabled

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		logger.WithTenant(tenantID).Error("Failed to save config", err, map[string]interface{}{})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
			"code":  "SAVE_ERROR",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Configuration updated successfully",
		"status":    "success",
		"tenant_id": tenantID,
	})
}

// AddEmailProvider adds a new email provider to a tenant configuration
// @Summary Add email provider
// @Description Add a new email provider to a specific tenant configuration
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param provider body models.AddProviderRequest true "Email provider request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/email [post]
func (h *ConfigHandler) AddEmailProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tenant ID",
			"code":  "INVALID_TENANT_ID",
		})
	}

	var req models.AddProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get configuration",
			"code":  "CONFIG_ERROR",
		})
	}

	// Add new provider
	newProvider := models.ProviderConfig{
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Enabled:  req.Enabled,
		Config:   req.Config,
	}

	config.EmailProviders = append(config.EmailProviders, newProvider)

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
			"code":  "SAVE_ERROR",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Email provider added successfully",
		"status":    "success",
		"tenant_id": tenantID,
	})
}

// AddSMSProvider adds a new SMS provider to a tenant configuration
// @Summary Add SMS provider
// @Description Add a new SMS provider to a specific tenant configuration
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param provider body models.AddProviderRequest true "SMS provider request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/sms [post]
func (h *ConfigHandler) AddSMSProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tenant ID",
			"code":  "INVALID_TENANT_ID",
		})
	}

	var req models.AddProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get configuration",
			"code":  "CONFIG_ERROR",
		})
	}

	// Add new provider
	newProvider := models.ProviderConfig{
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Enabled:  req.Enabled,
		Config:   req.Config,
	}

	config.SMSProviders = append(config.SMSProviders, newProvider)

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
			"code":  "SAVE_ERROR",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "SMS provider added successfully",
		"status":    "success",
		"tenant_id": tenantID,
	})
}

// AddPushProvider adds a new push provider to a tenant configuration
// @Summary Add push provider
// @Description Add a new push provider to a specific tenant configuration
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID"
// @Param provider body models.AddProviderRequest true "Push provider request"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/push [post]
func (h *ConfigHandler) AddPushProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tenant ID",
			"code":  "INVALID_TENANT_ID",
		})
	}

	var req models.AddProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get configuration",
			"code":  "CONFIG_ERROR",
		})
	}

	// Add new provider
	newProvider := models.ProviderConfig{
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Enabled:  req.Enabled,
		Config:   req.Config,
	}

	config.PushProviders = append(config.PushProviders, newProvider)

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
			"code":  "SAVE_ERROR",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":   "Push provider added successfully",
		"status":    "success",
		"tenant_id": tenantID,
	})
}

// RemoveProvider removes a provider from a tenant configuration
// @Summary Remove provider
// @Description Remove a provider from a specific tenant configuration
// @Tags configuration
// @Param tenant_id path int true "Tenant ID"
// @Param type path string true "Provider type (email, sms, push)"
// @Param name path string true "Provider name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/{type}/{name} [delete]
func (h *ConfigHandler) RemoveProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid tenant ID",
			"code":  "INVALID_TENANT_ID",
		})
	}

	providerType := c.Params("type")
	providerName := c.Params("name")

	if providerType == "" || providerName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Provider type and name are required",
			"code":  "MISSING_PARAMS",
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get configuration",
			"code":  "CONFIG_ERROR",
		})
	}

	// Remove provider based on type
	found := false
	switch providerType {
	case "email":
		for i, provider := range config.EmailProviders {
			if provider.Name == providerName {
				config.EmailProviders = append(config.EmailProviders[:i], config.EmailProviders[i+1:]...)
				found = true
				break
			}
		}
	case "sms":
		for i, provider := range config.SMSProviders {
			if provider.Name == providerName {
				config.SMSProviders = append(config.SMSProviders[:i], config.SMSProviders[i+1:]...)
				found = true
				break
			}
		}
	case "push":
		for i, provider := range config.PushProviders {
			if provider.Name == providerName {
				config.PushProviders = append(config.PushProviders[:i], config.PushProviders[i+1:]...)
				found = true
				break
			}
		}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid provider type",
			"code":  "INVALID_TYPE",
		})
	}

	if !found {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Provider not found",
			"code":  "NOT_FOUND",
		})
	}

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save configuration",
			"code":  "SAVE_ERROR",
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Provider removed successfully",
		"status":    "success",
		"tenant_id": tenantID,
	})
}

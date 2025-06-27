package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent/schema"
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
// @Description Get the complete configuration for a specific tenant including email, SMS, push providers, batch settings, and rate limits
// @Tags configuration
// @Produce json
// @Param tenant_id path int true "Tenant ID" minimum(1)
// @Success 200 {object} models.PartnerConfig "Configuration retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid tenant ID"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 404 {object} models.ErrorResponse "Not Found - Tenant configuration not found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Failed to retrieve configuration"
// @Security BearerAuth
// @Router /config/{tenant_id} [get]
func (h *ConfigHandler) GetConfig(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	if tenantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Tenant ID is required",
			Code:      "MISSING_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid tenant ID",
			Code:      "INVALID_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		logger.WithTenant(tenantID).Error("Failed to get config", err, map[string]interface{}{})
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to retrieve configuration",
			Code:      "CONFIG_ERROR",
			Timestamp: time.Now(),
		})
	}

	return c.JSON(config)
}

// UpdateConfig updates the configuration for a specific tenant
// @Summary Update partner configuration
// @Description Update the complete configuration for a specific tenant. This replaces the entire configuration.
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID" minimum(1)
// @Param config body models.PartnerConfigRequest true "Configuration update request"
// @Success 200 {object} models.ConfigSuccessResponse "Configuration updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid tenant ID or request body"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Failed to save configuration"
// @Security BearerAuth
// @Router /config/{tenant_id} [put]
func (h *ConfigHandler) UpdateConfig(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	if tenantIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Tenant ID is required",
			Code:      "MISSING_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid tenant ID",
			Code:      "INVALID_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	var req models.PartnerConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid request body",
			Code:      "INVALID_REQUEST",
			Timestamp: time.Now(),
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
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to save configuration",
			Code:      "SAVE_ERROR",
			Timestamp: time.Now(),
		})
	}

	return c.JSON(models.ConfigSuccessResponse{
		Message:   "Configuration updated successfully",
		Status:    "success",
		TenantID:  tenantID,
		Timestamp: time.Now(),
	})
}

// AddEmailProvider adds a new email provider to a tenant configuration
// @Summary Add email provider
// @Description Add a new email provider to a specific tenant configuration. Supports SMTP, SendGrid, SendX providers.
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID" minimum(1)
// @Param provider body models.AddProviderRequest true "Email provider request"
// @Success 201 {object} models.ConfigSuccessResponse "Email provider added successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid tenant ID or provider configuration"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Failed to save configuration"
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/email [post]
func (h *ConfigHandler) AddEmailProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid tenant ID",
			Code:      "INVALID_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	var req models.AddProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid request body",
			Code:      "INVALID_REQUEST",
			Timestamp: time.Now(),
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to get configuration",
			Code:      "CONFIG_ERROR",
			Timestamp: time.Now(),
		})
	}

	// Add new provider using the correct schema type
	newProvider := schema.ProviderConfig{
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Enabled:  req.Enabled,
		Config:   req.Config,
	}

	config.EmailProviders = append(config.EmailProviders, newProvider)

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to save configuration",
			Code:      "SAVE_ERROR",
			Timestamp: time.Now(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.ConfigSuccessResponse{
		Message:   "Email provider added successfully",
		Status:    "success",
		TenantID:  tenantID,
		Timestamp: time.Now(),
	})
}

// AddSMSProvider adds a new SMS provider to a tenant configuration
// @Summary Add SMS provider
// @Description Add a new SMS provider to a specific tenant configuration. Supports Twilio, Nexmo providers.
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID" minimum(1)
// @Param provider body models.AddProviderRequest true "SMS provider request"
// @Success 201 {object} models.ConfigSuccessResponse "SMS provider added successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid tenant ID or provider configuration"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Failed to save configuration"
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/sms [post]
func (h *ConfigHandler) AddSMSProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid tenant ID",
			Code:      "INVALID_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	var req models.AddProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid request body",
			Code:      "INVALID_REQUEST",
			Timestamp: time.Now(),
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to get configuration",
			Code:      "CONFIG_ERROR",
			Timestamp: time.Now(),
		})
	}

	// Add new provider using the correct schema type
	newProvider := schema.ProviderConfig{
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Enabled:  req.Enabled,
		Config:   req.Config,
	}

	config.SMSProviders = append(config.SMSProviders, newProvider)

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to save configuration",
			Code:      "SAVE_ERROR",
			Timestamp: time.Now(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.ConfigSuccessResponse{
		Message:   "SMS provider added successfully",
		Status:    "success",
		TenantID:  tenantID,
		Timestamp: time.Now(),
	})
}

// AddPushProvider adds a new push provider to a tenant configuration
// @Summary Add push provider
// @Description Add a new push provider to a specific tenant configuration. Supports FCM (Firebase Cloud Messaging).
// @Tags configuration
// @Accept json
// @Produce json
// @Param tenant_id path int true "Tenant ID" minimum(1)
// @Param provider body models.AddProviderRequest true "Push provider request"
// @Success 201 {object} models.ConfigSuccessResponse "Push provider added successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid tenant ID or provider configuration"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Failed to save configuration"
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/push [post]
func (h *ConfigHandler) AddPushProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid tenant ID",
			Code:      "INVALID_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	var req models.AddProviderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid request body",
			Code:      "INVALID_REQUEST",
			Timestamp: time.Now(),
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to get configuration",
			Code:      "CONFIG_ERROR",
			Timestamp: time.Now(),
		})
	}

	// Add new provider using the correct schema type
	newProvider := schema.ProviderConfig{
		Name:     req.Name,
		Type:     req.Type,
		Priority: req.Priority,
		Enabled:  req.Enabled,
		Config:   req.Config,
	}

	config.PushProviders = append(config.PushProviders, newProvider)

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to save configuration",
			Code:      "SAVE_ERROR",
			Timestamp: time.Now(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.ConfigSuccessResponse{
		Message:   "Push provider added successfully",
		Status:    "success",
		TenantID:  tenantID,
		Timestamp: time.Now(),
	})
}

// RemoveProvider removes a provider from a tenant configuration
// @Summary Remove provider
// @Description Remove a specific provider from a tenant configuration by provider type and name
// @Tags configuration
// @Param tenant_id path int true "Tenant ID" minimum(1)
// @Param type path string true "Provider type" Enums(email,sms,push)
// @Param name path string true "Provider name"
// @Success 200 {object} models.ConfigSuccessResponse "Provider removed successfully"
// @Failure 400 {object} models.ErrorResponse "Bad Request - Invalid parameters"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - Invalid or missing JWT token"
// @Failure 404 {object} models.ErrorResponse "Not Found - Provider not found"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error - Failed to save configuration"
// @Security BearerAuth
// @Router /config/{tenant_id}/providers/{type}/{name} [delete]
func (h *ConfigHandler) RemoveProvider(c *fiber.Ctx) error {
	tenantIDStr := c.Params("tenant_id")
	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid tenant ID",
			Code:      "INVALID_TENANT_ID",
			Timestamp: time.Now(),
		})
	}

	providerType := c.Params("type")
	providerName := c.Params("name")

	if providerType == "" || providerName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Provider type and name are required",
			Code:      "MISSING_PARAMS",
			Timestamp: time.Now(),
		})
	}

	// Get existing config
	config, err := h.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to get configuration",
			Code:      "CONFIG_ERROR",
			Timestamp: time.Now(),
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
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:     "Invalid provider type",
			Code:      "INVALID_TYPE",
			Timestamp: time.Now(),
		})
	}

	if !found {
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:     "Provider not found",
			Code:      "NOT_FOUND",
			Timestamp: time.Now(),
		})
	}

	if err := h.configRepo.Save(context.Background(), config); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:     "Failed to save configuration",
			Code:      "SAVE_ERROR",
			Timestamp: time.Now(),
		})
	}

	return c.JSON(models.ConfigSuccessResponse{
		Message:   "Provider removed successfully",
		Status:    "success",
		TenantID:  tenantID,
		Timestamp: time.Now(),
	})
}

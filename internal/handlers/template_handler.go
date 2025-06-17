package handlers

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/services"
	"go.uber.org/zap"
)

type TemplateHandler struct {
	templateRepo *repository.TemplateRepository
	templateSvc  *services.TemplateService
	logger       *zap.Logger
}

func NewTemplateHandler(
	templateRepo *repository.TemplateRepository,
	templateSvc *services.TemplateService,
	logger *zap.Logger,
) *TemplateHandler {
	return &TemplateHandler{
		templateRepo: templateRepo,
		templateSvc:  templateSvc,
		logger:       logger,
	}
}

func (h *TemplateHandler) ListTemplates(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(int64)

	templates, err := h.templateRepo.ListByTenant(context.Background(), tenantID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve templates",
			"code":  "FETCH_ERROR",
		})
	}

	return c.JSON(templates)
}

func (h *TemplateHandler) GetTemplate(c *fiber.Ctx) error {
	templateID := c.Params("id")
	tenantID, _ := c.Locals("tenant_id").(int64)

	template, err := h.templateRepo.GetByID(context.Background(), tenantID, templateID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Template not found",
			"code":  "NOT_FOUND",
		})
	}

	return c.JSON(template)
}

func (h *TemplateHandler) CreateTemplate(c *fiber.Ctx) error {
	tenantID, _ := c.Locals("tenant_id").(int64)

	var req models.TemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	template, err := h.templateSvc.CreateTemplate(context.Background(), tenantID, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  "CREATE_ERROR",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(template)
}

func (h *TemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	// Implementation for updating template
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *TemplateHandler) DeleteTemplate(c *fiber.Ctx) error {
	// Implementation for deleting template
	return c.JSON(fiber.Map{"status": "ok"})
}

func (h *TemplateHandler) PreviewTemplate(c *fiber.Ctx) error {
	// Implementation for previewing template
	return c.JSON(fiber.Map{"status": "ok"})
}

package handlers

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	publisher *kafka.Publisher
	notifRepo *repository.NotificationRepository
	logger    *zap.Logger
}

func NewNotificationHandler(
	publisher *kafka.Publisher,
	notifRepo *repository.NotificationRepository,
	logger *zap.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		publisher: publisher,
		notifRepo: notifRepo,
		logger:    logger,
	}
}

// SendNotification handles HTTP requests and publishes to Kafka
func (h *NotificationHandler) SendNotification(c *fiber.Ctx) error {
	var req models.NotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Get tenant ID from context (set by auth middleware)
	tenantID, ok := c.Locals("tenant_id").(int64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid tenant",
			"code":  "INVALID_TENANT",
		})
	}
	req.TenantID = tenantID

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Validate request
	if err := h.validateRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
			"code":  "VALIDATION_ERROR",
		})
	}

	// Publish to Kafka for async processing
	data, err := json.Marshal(req)
	if err != nil {
		h.logger.Error("failed to marshal notification request", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	}

	if err := h.publisher.Publish(context.Background(), "notifications", req.RequestID, data); err != nil {
		h.logger.Error("failed to publish to kafka", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to queue notification",
			"code":  "QUEUE_ERROR",
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"request_id": req.RequestID,
		"status":     "queued",
		"message":    "Notification queued for processing",
	})
}

// SendBatchNotification handles batch notification requests
func (h *NotificationHandler) SendBatchNotification(c *fiber.Ctx) error {
	var req struct {
		Type       models.NotificationType `json:"type"`
		TemplateID string                  `json:"template_id,omitempty"`
		Recipients []string                `json:"recipients"`
		Data       map[string]interface{}  `json:"data,omitempty"`
		Locale     string                  `json:"locale,omitempty"`
		ScheduleTS *int64                  `json:"schedule_ts,omitempty"`
		From       string                  `json:"from,omitempty"`
		ReplyTo    string                  `json:"reply_to,omitempty"`
		Tag        string                  `json:"tag,omitempty"`
		Body       string                  `json:"body,omitempty"`
		Headline   string                  `json:"headline,omitempty"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	tenantID, _ := c.Locals("tenant_id").(int64)
	batchID := uuid.New().String()

	// Split recipients into chunks and publish to Kafka
	chunkSize := 100
	var publishedCount int

	for i := 0; i < len(req.Recipients); i += chunkSize {
		end := i + chunkSize
		if end > len(req.Recipients) {
			end = len(req.Recipients)
		}

		notifReq := models.NotificationRequest{
			RequestID:  uuid.New().String(),
			TenantID:   tenantID,
			Type:       req.Type,
			TemplateID: req.TemplateID,
			Recipients: req.Recipients[i:end],
			Data:       req.Data,
			BatchID:    batchID,
			Locale:     req.Locale,
			ScheduleTS: req.ScheduleTS,
			From:       req.From,
			ReplyTo:    req.ReplyTo,
			Tag:        req.Tag,
			Body:       req.Body,
			Headline:   req.Headline,
		}

		data, _ := json.Marshal(notifReq)
		if err := h.publisher.Publish(context.Background(), "notifications", notifReq.RequestID, data); err != nil {
			h.logger.Error("failed to publish batch request",
				zap.String("request_id", notifReq.RequestID),
				zap.Error(err),
			)
		} else {
			publishedCount += len(notifReq.Recipients)
		}
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"batch_id":          batchID,
		"total_recipients":  len(req.Recipients),
		"queued_recipients": publishedCount,
		"status":            "processing",
	})
}

// GetNotificationStatus retrieves notification status by request ID
func (h *NotificationHandler) GetNotificationStatus(c *fiber.Ctx) error {
	requestID := c.Params("request_id")
	if requestID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Request ID is required",
			"code":  "MISSING_REQUEST_ID",
		})
	}

	tenantID, _ := c.Locals("tenant_id").(int64)

	notification, err := h.notifRepo.GetByRequestID(context.Background(), requestID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Notification not found",
			"code":  "NOT_FOUND",
		})
	}

	// Verify tenant ownership
	if notification.TenantID != tenantID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied",
			"code":  "ACCESS_DENIED",
		})
	}

	response := fiber.Map{
		"request_id": notification.RequestID,
		"status":     notification.Status,
		"type":       notification.Type,
		"created_at": notification.CreateTime,
		"updated_at": notification.UpdateTime,
	}

	if notification.ErrorMessage != nil {
		response["error_message"] = *notification.ErrorMessage
	}

	if notification.ScheduleTs != nil {
		response["schedule_ts"] = *notification.ScheduleTs
	}

	return c.JSON(response)
}

// PublishToKafka handles direct Kafka publishing
func (h *NotificationHandler) PublishToKafka(c *fiber.Ctx) error {
	var req models.NotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Publish to Kafka
	data, err := json.Marshal(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process request",
			"code":  "PROCESSING_ERROR",
		})
	}

	if err := h.publisher.Publish(context.Background(), "notifications", req.RequestID, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to publish to Kafka",
			"code":  "PUBLISH_ERROR",
		})
	}

	return c.JSON(fiber.Map{
		"request_id": req.RequestID,
		"status":     "published",
	})
}

func (h *NotificationHandler) validateRequest(req *models.NotificationRequest) error {
	if len(req.Recipients) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Recipients list cannot be empty")
	}

	if req.Type == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Notification type is required")
	}

	if req.TemplateID == "" && req.Body == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Either template_id or body is required")
	}

	return nil
}

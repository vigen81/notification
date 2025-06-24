package handlers

import (
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
)

type NotificationHandler struct {
	publisher *kafka.Publisher
	notifRepo *repository.NotificationRepository
	logger    *logrus.Logger
}

func NewNotificationHandler(
	publisher *kafka.Publisher,
	notifRepo *repository.NotificationRepository,
	logger *logrus.Logger,
) *NotificationHandler {
	return &NotificationHandler{
		publisher: publisher,
		notifRepo: notifRepo,
		logger:    logger,
	}
}

// SendNotification handles HTTP requests and publishes to Kafka
// @Summary Send a notification
// @Description Send a single notification via HTTP API
// @Tags notifications
// @Accept json
// @Produce json
// @Param notification body models.NotificationRequest true "Notification request"
// @Success 202 {object} models.NotificationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notifications/send [post]
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
		logger.WithRequest(req.RequestID).Error("Failed to marshal notification request", err, map[string]interface{}{
			"tenant_id": tenantID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
			"code":  "INTERNAL_ERROR",
		})
	}

	if err := h.publisher.Publish(context.Background(), "notifications", req.RequestID, data); err != nil {
		logger.WithRequest(req.RequestID).Error("Failed to publish to Kafka", err, map[string]interface{}{
			"tenant_id": tenantID,
		})
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to queue notification",
			"code":  "QUEUE_ERROR",
		})
	}

	response := models.NotificationResponse{
		RequestID: req.RequestID,
		Status:    "queued",
		Message:   "Notification queued for processing",
	}

	return c.Status(fiber.StatusAccepted).JSON(response)
}

// SendBatchNotification handles batch notification requests
// @Summary Send batch notifications
// @Description Send multiple notifications in a batch
// @Tags notifications
// @Accept json
// @Produce json
// @Param batch body models.BatchNotificationRequest true "Batch notification request"
// @Success 202 {object} models.BatchNotificationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notifications/batch [post]
func (h *NotificationHandler) SendBatchNotification(c *fiber.Ctx) error {
	var req models.BatchNotificationRequest
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
			RequestID:   uuid.New().String(),
			TenantID:    tenantID,
			Type:        req.Type,
			Recipients:  req.Recipients[i:end],
			Body:        req.Body,
			Headline:    req.Headline,
			From:        req.From,
			ReplyTo:     req.ReplyTo,
			Tag:         req.Tag,
			ScheduleTS:  req.ScheduleTS,
			Data:        req.Data,
			BatchID:     batchID,
			MessageType: req.MessageType,
		}

		data, _ := json.Marshal(notifReq)
		if err := h.publisher.Publish(context.Background(), "notifications", notifReq.RequestID, data); err != nil {
			logger.WithRequest(notifReq.RequestID).Error("Failed to publish batch request", err, map[string]interface{}{
				"batch_id":  batchID,
				"tenant_id": tenantID,
			})
		} else {
			publishedCount += len(notifReq.Recipients)
		}
	}

	response := models.BatchNotificationResponse{
		BatchID:          batchID,
		TotalRecipients:  len(req.Recipients),
		QueuedRecipients: publishedCount,
		Status:           "processing",
	}

	return c.Status(fiber.StatusAccepted).JSON(response)
}

// GetNotificationStatus retrieves notification status by request ID
// @Summary Get notification status
// @Description Get the status of a notification by request ID
// @Tags notifications
// @Produce json
// @Param request_id path string true "Request ID"
// @Success 200 {object} models.NotificationStatusResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /notifications/status/{request_id} [get]
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

	response := models.NotificationStatusResponse{
		RequestID: notification.RequestID,
		Status:    string(notification.Status),
		Type:      string(notification.Type),
		CreatedAt: notification.CreateTime,
		UpdatedAt: notification.UpdateTime,
	}

	if notification.ErrorMessage != nil {
		response.ErrorMessage = notification.ErrorMessage
	}

	if notification.ScheduleTs != nil {
		response.ScheduleTS = notification.ScheduleTs
	}

	return c.JSON(response)
}

// PublishToKafka handles direct Kafka publishing
// @Summary Publish to Kafka
// @Description Directly publish a notification to Kafka
// @Tags kafka
// @Accept json
// @Produce json
// @Param notification body models.KafkaNotificationRequest true "Kafka notification request"
// @Success 200 {object} models.KafkaResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /kafka/publish [post]
func (h *NotificationHandler) PublishToKafka(c *fiber.Ctx) error {
	var req models.KafkaNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"code":  "INVALID_REQUEST",
		})
	}

	// Convert to standard notification request
	notifReq := models.NotificationRequest{
		RequestID:   uuid.New().String(),
		TenantID:    req.TenantID,
		Type:        req.Type,
		Recipients:  req.Recipients,
		Body:        req.Body,
		Headline:    req.Headline,
		MessageType: req.MessageType,
		Data:        req.Data,
	}

	// Publish to Kafka
	data, err := json.Marshal(notifReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to process request",
			"code":  "PROCESSING_ERROR",
		})
	}

	if err := h.publisher.Publish(context.Background(), "notifications", notifReq.RequestID, data); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to publish to Kafka",
			"code":  "PUBLISH_ERROR",
		})
	}

	response := models.KafkaResponse{
		RequestID: notifReq.RequestID,
		Status:    "published",
	}

	return c.JSON(response)
}

func (h *NotificationHandler) validateRequest(req *models.NotificationRequest) error {
	if len(req.Recipients) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "Recipients list cannot be empty")
	}

	if req.Type == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Notification type is required")
	}

	if req.Body == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Body is required")
	}

	// Validate notification type
	switch req.Type {
	case models.TypeEmail, models.TypeSMS, models.TypePush:
		// Valid types
	default:
		return fiber.NewError(fiber.StatusBadRequest, "Invalid notification type")
	}

	return nil
}

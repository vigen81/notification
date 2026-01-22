package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/services"
)

type SendGridEvent struct {
	Email     string `json:"email"`
	Timestamp int64  `json:"timestamp"`
	Event     string `json:"event"` // open, click, delivered, bounce, dropped, etc.
	SGEventID string `json:"sg_event_id"`
	SGMsgID   string `json:"sg_message_id"`
	RequestID string `json:"request_id,omitempty"` // custom arg we pass when sending
	TenantID  int64  `json:"tenant_id,omitempty"`
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"useragent,omitempty"`
	URL       string `json:"url,omitempty"` // for click events
}

type WebhookHandler struct {
	notificationService *services.NotificationService
	logger              *logrus.Logger
}

func NewWebhookHandler(ns *services.NotificationService, logger *logrus.Logger) *WebhookHandler {
	return &WebhookHandler{
		notificationService: ns,
		logger:              logger,
	}
}

// HandleSendGridWebhook processes SendGrid event webhooks
// @Summary Handle SendGrid webhook events
// @Tags Webhooks
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /api/v1/webhooks/sendgrid [post]
func (h *WebhookHandler) HandleSendGridWebhook(c *fiber.Ctx) error {
	var events []SendGridEvent
	if err := c.BodyParser(&events); err != nil {
		h.logger.WithError(err).Error("failed to parse sendgrid webhook payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	for _, event := range events {
		h.logger.WithFields(logrus.Fields{
			"event":      event.Event,
			"email":      event.Email,
			"request_id": event.RequestID,
			"sg_msg_id":  event.SGMsgID,
		}).Info("received sendgrid webhook event")

		if event.RequestID != "" {
			if err := h.notificationService.UpdateStatusFromWebhook(c.Context(), event.RequestID, event.Event); err != nil {
				h.logger.WithError(err).Error("failed to update notification status")
			}
		}
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

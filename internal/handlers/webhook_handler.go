package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/services"
)

const (
	MsgStatusFailed    = 0
	MsgStatusSent      = 1
	MsgStatusDelivered = 2
	MsgStatusOpened    = 3
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
// @Description Receives email events from SendGrid (open, click, delivered, bounce, etc.)
// @Tags webhooks
// @Accept json
// @Produce json
// @Param events body []SendGridEvent true "SendGrid events array"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /public/api/v1/webhooks/sendgrid [post]
func (h *WebhookHandler) HandleSendGridWebhook(c *fiber.Ctx) error {
	var events []SendGridEvent
	if err := c.BodyParser(&events); err != nil {
		h.logger.WithError(err).Error("failed to parse sendgrid webhook payload")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, event := range events {
		if dotIndex := strings.Index(event.SGMsgID, "."); dotIndex != -1 {
			event.SGMsgID = event.SGMsgID[:dotIndex]
		}

		h.logger.WithFields(logrus.Fields{
			"event":        event.Event,
			"email":        event.Email,
			"request_id":   event.RequestID,
			"sg_msg_id":    event.SGMsgID,
			"event_object": event,
		}).Info("received sendgrid webhook event")

		/*
			if event.RequestID != "" {
				if err := h.notificationService.UpdateStatusFromWebhook(c.Context(), event.RequestID, event.Event); err != nil {
					h.logger.WithError(err).Error("failed to update notification status")
				}
			}
		*/

		// Map SendGrid event to our status
		status := mapEventToStatus(event.Event)
		if status == -1 {
			continue // Skip unknown events
		}

		// Prepare request to external service
		payload := map[string]interface{}{
			"provider":   "sendgrid",
			"message_id": event.SGMsgID,
			"status":     status,
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			h.logger.WithError(err).Error("failed to marshal webhook payload")
			continue
		}

		resp, err := client.Post("http://platform/notification/change_status", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			h.logger.WithError(err).Error("failed to send status update to platform")
			continue
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			h.logger.WithField("status_code", resp.StatusCode).Error("platform returned non-OK status")
			continue
		}

		if readErr != nil {
			h.logger.WithError(readErr).Error("failed to read platform response body")
			continue
		}

		var platformResponse struct {
			Status int `json:"status"`
		}
		if err := json.Unmarshal(body, &platformResponse); err != nil {
			h.logger.WithError(err).Error("failed to parse platform response body")
			continue
		}

		if platformResponse.Status == 0 {
			h.logger.Error("platform response status indicates failure")
		}
	}

	return c.JSON(fiber.Map{"status": "ok"})
}

func mapEventToStatus(event string) int {
	switch event {
	case "processed", "sent":
		return MsgStatusSent
	case "delivered":
		return MsgStatusDelivered
	case "open":
		return MsgStatusOpened
	case "bounce", "dropped":
		return MsgStatusFailed
	default:
		return -1
	}
}

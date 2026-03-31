package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/models"
)

// CustomProvider implements the SMSProvider interface for Nikita SMS API
type CustomProvider struct {
	config CustomConfig
	client *http.Client
}

// CustomConfig represents your SMS provider configuration
type CustomConfig struct {
	URLMRK          string `json:"url_mrk"`
	URLTrans        string `json:"url_trans"`
	PasswordMRK     string `json:"password_mrk"`
	UsernameMRK     string `json:"username_mrk"`
	OriginatorMRK   string `json:"originator_mrk"`
	PasswordTrans   string `json:"password_trans"`
	UsernameTrans   string `json:"username_trans"`
	OriginatorTrans string `json:"originator_trans"`

	// Message type specific originators (optional)
	MSGBonusOriginator   string `json:"MSGBonusOriginator"`
	MSGPromoOriginator   string `json:"MSGPromoOriginator"`
	MSGSystemOriginator  string `json:"MSGSystemOriginator"`
	MSGReportOriginator  string `json:"MSGReportOriginator"`
	MSGPaymentOriginator string `json:"MSGPaymentOriginator"`
	MSGSupportOriginator string `json:"MSGSupportOriginator"`
}

// SMSRequest represents the request structure for Nikita SMS API
type SMSRequest struct {
	Messages []SMSMessage `json:"messages"`
}

type SMSMessage struct {
	Recipient string     `json:"recipient"`
	Priority  string     `json:"priority"`
	SMS       SMSContent `json:"sms"`
	MessageID string     `json:"message-id"`
}

type SMSContent struct {
	Originator string  `json:"originator"`
	Content    SMSText `json:"content"`
}

type SMSText struct {
	Text string `json:"text"`
}

// SMSResponse represents the response from Nikita SMS API
type SMSResponse struct {
	ErrorDescription string `json:"error-description,omitempty"`
}

// NewCustomProvider creates a new Custom SMS provider
func NewCustomProvider(config map[string]interface{}) (*CustomProvider, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Custom SMS config: %w", err)
	}

	var customConfig CustomConfig
	if err := json.Unmarshal(configBytes, &customConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Custom SMS config: %w", err)
	}

	provider := &CustomProvider{
		config: customConfig,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid Custom SMS config: %w", err)
	}

	return provider, nil
}

// Send sends a single SMS via Nikita API
func (c *CustomProvider) Send(ctx context.Context, notification *ent.Notification, messageType models.MessageType) error {
	log := logger.WithRequest(notification.RequestID)

	// Determine which endpoint and credentials to use based on message type
	apiURL, username, password, originator := c.getEndpointConfig(messageType)

	// Clean phone number (remove + if present)
	phoneNumber := strings.TrimPrefix(string(notification.Address), "+")

	// Generate message ID
	messageID := fmt.Sprintf("%s:%d:%d", phoneNumber, time.Now().Unix(), time.Now().Nanosecond()/1000000)

	// Send SMS
	response, err := c.sendSMS(ctx, apiURL, username, password, originator, phoneNumber, notification.Body, messageID)
	if err != nil {
		log.Error("Failed to send Nikita SMS", err, map[string]interface{}{
			"notification_id": notification.ID,
			"to":              phoneNumber,
			"originator":      originator,
			"api_url":         apiURL,
			"message_id":      messageID,
		})
		return fmt.Errorf("failed to send Nikita SMS: %w", err)
	}

	log.Info("Nikita SMS sent successfully", map[string]interface{}{
		"notification_id": notification.ID,
		"to":              phoneNumber,
		"originator":      originator,
		"message_id":      messageID,
		"response":        response,
	})

	return nil
}

// SendBatch sends multiple SMS messages
func (c *CustomProvider) SendBatch(ctx context.Context, notifications []*ent.Notification, messageType models.MessageType) error {
	log := logger.To("nikita_sms_batch")

	// Group notifications by message type to use appropriate endpoint
	typeGroups := make(map[models.MessageType][]*ent.Notification)
	for _, notif := range notifications {
		// Determine message type from notification meta if available
		msgType := messageType
		if notif.Meta != nil && notif.Meta.Params != nil {
			if mt, exists := notif.Meta.Params["message_type"]; exists {
				if mtStr, ok := mt.(string); ok {
					msgType = models.MessageType(mtStr)
				}
			}
		}
		typeGroups[msgType] = append(typeGroups[msgType], notif)
	}

	var errors []error
	successCount := 0

	// Process each group with appropriate endpoint
	for msgType, groupNotifications := range typeGroups {
		// Determine endpoint config for this message type
		apiURL, username, password, originator := c.getEndpointConfig(msgType)

		// Build batch request
		var messages []SMSMessage
		for _, notif := range groupNotifications {
			phoneNumber := strings.TrimPrefix(string(notif.Address), "+")
			messageID := fmt.Sprintf("%s:%d:%d", phoneNumber, time.Now().Unix(), time.Now().Nanosecond()/1000000)

			messages = append(messages, SMSMessage{
				Recipient: phoneNumber,
				Priority:  "4",
				SMS: SMSContent{
					Originator: originator,
					Content:    SMSText{Text: notif.Body},
				},
				MessageID: messageID,
			})
		}

		// Send batch
		_, err := c.sendBatchSMS(ctx, apiURL, username, password, messages)
		if err != nil {
			errors = append(errors, err)
			log.Error("Failed to send SMS batch", err, map[string]interface{}{
				"message_type": msgType,
				"batch_size":   len(groupNotifications),
			})
		} else {
			successCount += len(groupNotifications)
			log.Info("SMS batch sent successfully", map[string]interface{}{
				"message_type": msgType,
				"batch_size":   len(groupNotifications),
			})
		}
	}

	log.Info("Nikita SMS batch processing completed", map[string]interface{}{
		"total_notifications": len(notifications),
		"success_count":       successCount,
		"error_count":         len(errors),
	})

	if len(errors) > 0 {
		return fmt.Errorf("batch sending completed with %d errors out of %d SMS messages", len(errors), len(notifications))
	}

	return nil
}

// ValidateConfig validates the Custom SMS configuration
func (c *CustomProvider) ValidateConfig() error {
	if c.config.URLMRK == "" {
		return fmt.Errorf("Nikita SMS URL MRK is required")
	}
	if c.config.URLTrans == "" {
		return fmt.Errorf("Nikita SMS URL Trans is required")
	}
	if c.config.UsernameMRK == "" {
		return fmt.Errorf("Nikita SMS Username MRK is required")
	}
	if c.config.UsernameTrans == "" {
		return fmt.Errorf("Nikita SMS Username Trans is required")
	}
	if c.config.PasswordMRK == "" {
		return fmt.Errorf("Nikita SMS Password MRK is required")
	}
	if c.config.PasswordTrans == "" {
		return fmt.Errorf("Nikita SMS Password Trans is required")
	}
	return nil
}

// GetType returns the provider type
func (c *CustomProvider) GetType() string {
	return "custom"
}

// getEndpointConfig returns the appropriate endpoint and credentials based on message type
func (c *CustomProvider) getEndpointConfig(messageType models.MessageType) (string, string, string, string) {
	// Use MRK endpoint for promotional messages, Trans for everything else
	if messageType == models.MessageTypePromo || messageType == models.MessageTypeBonus {
		originator := c.getOriginator(messageType, c.config.OriginatorMRK)
		return c.config.URLMRK, c.config.UsernameMRK, c.config.PasswordMRK, originator
	}

	originator := c.getOriginator(messageType, c.config.OriginatorTrans)
	return c.config.URLTrans, c.config.UsernameTrans, c.config.PasswordTrans, originator
}

// getOriginator returns the appropriate originator based on message type
func (c *CustomProvider) getOriginator(messageType models.MessageType, defaultOriginator string) string {
	switch messageType {
	case models.MessageTypeBonus:
		if c.config.MSGBonusOriginator != "" {
			return c.config.MSGBonusOriginator
		}
	case models.MessageTypePromo:
		if c.config.MSGPromoOriginator != "" {
			return c.config.MSGPromoOriginator
		}
	case models.MessageTypeSystem:
		if c.config.MSGSystemOriginator != "" {
			return c.config.MSGSystemOriginator
		}
	case models.MessageTypeReport:
		if c.config.MSGReportOriginator != "" {
			return c.config.MSGReportOriginator
		}
	case models.MessageTypePayment:
		if c.config.MSGPaymentOriginator != "" {
			return c.config.MSGPaymentOriginator
		}
	case models.MessageTypeSupport:
		if c.config.MSGSupportOriginator != "" {
			return c.config.MSGSupportOriginator
		}
	}
	return defaultOriginator
}

// sendSMS sends a single SMS via Nikita API
func (c *CustomProvider) sendSMS(ctx context.Context, apiURL, username, password, originator, destination, message, messageID string) (interface{}, error) {
	messages := []SMSMessage{
		{
			Recipient: destination,
			Priority:  "4",
			SMS: SMSContent{
				Originator: originator,
				Content:    SMSText{Text: message},
			},
			MessageID: messageID,
		},
	}

	return c.sendBatchSMS(ctx, apiURL, username, password, messages)
}

// sendBatchSMS sends multiple SMS messages via Nikita API
func (c *CustomProvider) sendBatchSMS(ctx context.Context, apiURL, username, password string, messages []SMSMessage) (interface{}, error) {
	// Build request URL - Fixed to use port 80 and correct path
	fullURL := fmt.Sprintf("%s/broker-api/send", apiURL)

	// Build request body
	requestData := SMSRequest{
		Messages: messages,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("charset", "utf-8")

	// Set basic auth
	req.SetBasicAuth(username, password)

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	responseStr := string(body)

	// Check status code - API returns 200 with "OK" for success
	if resp.StatusCode != 200 {
		var errorResp SMSResponse
		json.Unmarshal(body, &errorResp)

		errorMsg := fmt.Sprintf("SMS API error %d", resp.StatusCode)
		if errorResp.ErrorDescription != "" {
			errorMsg = errorResp.ErrorDescription
		}

		return responseStr, fmt.Errorf("Nikita SMS API error: %s, Response: %s", errorMsg, responseStr)
	}

	// Check if response indicates success (your curl returned "OK")
	if !strings.Contains(responseStr, "OK") {
		return responseStr, fmt.Errorf("SMS API error: unexpected response: %s", responseStr)
	}

	return responseStr, nil
}

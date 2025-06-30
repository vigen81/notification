package sms

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/models"
)

// TwilioProvider implements the SMSProvider interface for Twilio API
type TwilioProvider struct {
	accountSID string
	authToken  string
	fromNumber string
	baseURL    string
	client     *http.Client
	config     TwilioConfig
}

// TwilioConfig represents Twilio configuration
type TwilioConfig struct {
	AccountSID string `json:"account_sid"`
	AuthToken  string `json:"auth_token"`
	FromNumber string `json:"from_number"`
	BaseURL    string `json:"base_url"`

	// Message type specific from numbers (optional)
	MSGBonusFrom   string `json:"MSGBonusFrom"`
	MSGPromoFrom   string `json:"MSGPromoFrom"`
	MSGReportFrom  string `json:"MSGReportFrom"`
	MSGSystemFrom  string `json:"MSGSystemFrom"`
	MSGPaymentFrom string `json:"MSGPaymentFrom"`
	MSGSupportFrom string `json:"MSGSupportFrom"`
}

// TwilioRequest represents the request payload for Twilio API
type TwilioRequest struct {
	To   string `json:"To"`
	From string `json:"From"`
	Body string `json:"Body"`
}

// TwilioResponse represents the response from Twilio API
type TwilioResponse struct {
	SID          string      `json:"sid"`
	AccountSID   string      `json:"account_sid"`
	To           string      `json:"to"`
	From         string      `json:"from"`
	Body         string      `json:"body"`
	Status       string      `json:"status"`
	Direction    string      `json:"direction"`
	DateCreated  string      `json:"date_created"`
	DateUpdated  string      `json:"date_updated"`
	DateSent     interface{} `json:"date_sent"`
	URI          string      `json:"uri"`
	ErrorCode    interface{} `json:"error_code"`
	ErrorMessage interface{} `json:"error_message"`
	Price        interface{} `json:"price"`
	PriceUnit    string      `json:"price_unit"`
}

// TwilioErrorResponse represents error response from Twilio
type TwilioErrorResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	MoreInfo string `json:"more_info"`
	Status   int    `json:"status"`
}

// NewTwilioProvider creates a new Twilio SMS provider
func NewTwilioProvider(config map[string]interface{}) (*TwilioProvider, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Twilio config: %w", err)
	}

	var twilioConfig TwilioConfig
	if err := json.Unmarshal(configBytes, &twilioConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Twilio config: %w", err)
	}

	// Set default base URL if not provided
	if twilioConfig.BaseURL == "" {
		twilioConfig.BaseURL = "https://api.twilio.com/2010-04-01"
	}

	provider := &TwilioProvider{
		accountSID: twilioConfig.AccountSID,
		authToken:  twilioConfig.AuthToken,
		fromNumber: twilioConfig.FromNumber,
		baseURL:    twilioConfig.BaseURL,
		config:     twilioConfig,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid Twilio config: %w", err)
	}

	return provider, nil
}

// Send sends a single SMS via Twilio API
func (t *TwilioProvider) Send(ctx context.Context, notification *ent.Notification, messageType models.MessageType) error {
	log := logger.WithRequest(notification.RequestID)

	// Get appropriate from number based on message type
	fromNumber := t.getFromNumber(messageType)

	// Prepare form data for Twilio API
	data := url.Values{
		"To":   {string(notification.Address)},
		"From": {fromNumber},
		"Body": {notification.Body},
	}

	// Send SMS
	response, err := t.sendSMS(ctx, data)
	if err != nil {
		log.Error("Failed to send Twilio SMS", err, map[string]interface{}{
			"notification_id": notification.ID,
			"to":              string(notification.Address),
			"from":            fromNumber,
		})
		return fmt.Errorf("failed to send Twilio SMS: %w", err)
	}

	log.Info("Twilio SMS sent successfully", map[string]interface{}{
		"notification_id": notification.ID,
		"to":              string(notification.Address),
		"from":            fromNumber,
		"twilio_sid":      response.SID,
		"status":          response.Status,
	})

	return nil
}

// SendBatch sends multiple SMS messages via Twilio API
func (t *TwilioProvider) SendBatch(ctx context.Context, notifications []*ent.Notification, messageType models.MessageType) error {
	log := logger.To("twilio_batch")

	// Twilio doesn't have a native batch SMS API, so we send individually
	// We could optimize this with goroutines, but keeping it simple for now
	var errors []error
	successCount := 0

	for _, notification := range notifications {
		if err := t.Send(ctx, notification, messageType); err != nil {
			errors = append(errors, err)
			log.Error("Failed to send SMS in batch", err, map[string]interface{}{
				"notification_id": notification.ID,
				"batch_size":      len(notifications),
			})
		} else {
			successCount++
		}
	}

	log.Info("Twilio batch processing completed", map[string]interface{}{
		"batch_size":    len(notifications),
		"success_count": successCount,
		"error_count":   len(errors),
	})

	if len(errors) > 0 {
		return fmt.Errorf("batch sending completed with %d errors out of %d SMS messages", len(errors), len(notifications))
	}

	return nil
}

// ValidateConfig validates the Twilio configuration
func (t *TwilioProvider) ValidateConfig() error {
	if t.accountSID == "" {
		return fmt.Errorf("Twilio Account SID is required")
	}
	if t.authToken == "" {
		return fmt.Errorf("Twilio Auth Token is required")
	}
	if t.fromNumber == "" {
		return fmt.Errorf("Twilio From Number is required")
	}

	// Validate phone number format (basic check)
	if !strings.HasPrefix(t.fromNumber, "+") {
		return fmt.Errorf("Twilio From Number must be in E.164 format (e.g., +1234567890)")
	}

	return nil
}

// GetType returns the provider type
func (t *TwilioProvider) GetType() string {
	return "twilio"
}

// sendSMS sends the SMS via Twilio API
func (t *TwilioProvider) sendSMS(ctx context.Context, data url.Values) (*TwilioResponse, error) {
	// Create request URL
	url := fmt.Sprintf("%s/Accounts/%s/Messages.json", t.baseURL, t.accountSID)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set basic auth
	auth := base64.StdEncoding.EncodeToString([]byte(t.accountSID + ":" + t.authToken))
	req.Header.Set("Authorization", "Basic "+auth)

	// Send request
	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var errorResp TwilioErrorResponse
		if err := json.Unmarshal(body, &errorResp); err == nil {
			return nil, fmt.Errorf("Twilio API error: %s (code: %d)", errorResp.Message, errorResp.Code)
		}
		return nil, fmt.Errorf("Twilio API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var twilioResp TwilioResponse
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &twilioResp, nil
}

// getFromNumber returns the appropriate from number based on message type
func (t *TwilioProvider) getFromNumber(messageType models.MessageType) string {
	switch messageType {
	case models.MessageTypeBonus:
		if t.config.MSGBonusFrom != "" {
			return t.config.MSGBonusFrom
		}
	case models.MessageTypePromo:
		if t.config.MSGPromoFrom != "" {
			return t.config.MSGPromoFrom
		}
	case models.MessageTypeReport:
		if t.config.MSGReportFrom != "" {
			return t.config.MSGReportFrom
		}
	case models.MessageTypeSystem:
		if t.config.MSGSystemFrom != "" {
			return t.config.MSGSystemFrom
		}
	case models.MessageTypePayment:
		if t.config.MSGPaymentFrom != "" {
			return t.config.MSGPaymentFrom
		}
	case models.MessageTypeSupport:
		if t.config.MSGSupportFrom != "" {
			return t.config.MSGSupportFrom
		}
	}

	// Default to main from number
	return t.fromNumber
}

package providers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"gitlab.smartbet.am/golang/notification/ent"
)

type SendGridProvider struct {
	client *sendgrid.Client
	config SendGridConfig
}

type SendGridConfig struct {
	APIKey      string `json:"api_key"`
	FromEmail   string `json:"from_email"`
	FromName    string `json:"from_name"`
	SandboxMode bool   `json:"sandbox_mode"`
}

func NewSendGridProvider(config map[string]interface{}) (EmailProvider, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var cfg SendGridConfig
	if err := json.Unmarshal(configBytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	client := sendgrid.NewSendClient(cfg.APIKey)

	return &SendGridProvider{
		client: client,
		config: cfg,
	}, nil
}

func (p *SendGridProvider) Send(ctx context.Context, notification *ent.Notification, templateData map[string]interface{}) error {
	fromEmail := p.config.FromEmail
	if notification.From != nil && *notification.From != "" {
		fromEmail = *notification.From
	}

	from := mail.NewEmail(p.config.FromName, fromEmail)
	to := mail.NewEmail("", string(notification.Address))

	subject := ""
	if notification.Headline != nil {
		subject = *notification.Headline
	}

	message := mail.NewSingleEmail(from, subject, to, notification.Body, notification.Body)

	if p.config.SandboxMode {
		message.MailSettings = &mail.MailSettings{
			SandboxMode: &mail.Setting{Enable: true},
		}
	}

	if notification.ReplyTo != nil && *notification.ReplyTo != "" {
		message.ReplyTo = &mail.Email{
			Address: *notification.ReplyTo,
		}
	}

	// Handle attachments if present
	if notification.Meta != nil && notification.Meta.Attachment != nil {
		attachment := mail.NewAttachment()
		attachment.SetContent(notification.Meta.Attachment.Content)
		attachment.SetType(notification.Meta.Attachment.Type)
		attachment.SetFilename(notification.Meta.Attachment.Filename)
		attachment.SetDisposition(notification.Meta.Attachment.Disposition)
		message.AddAttachment(attachment)
	}

	response, err := p.client.Send(message)
	if err != nil {
		return fmt.Errorf("sendgrid error: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid returned status %d: %s", response.StatusCode, response.Body)
	}

	return nil
}

func (p *SendGridProvider) SendBatch(ctx context.Context, notifications []*ent.Notification, templateData map[string]interface{}) error {
	message := mail.NewV3Mail()

	fromEmail := p.config.FromEmail
	if len(notifications) > 0 && notifications[0].From != nil {
		fromEmail = *notifications[0].From
	}

	message.SetFrom(mail.NewEmail(p.config.FromName, fromEmail))

	if len(notifications) > 0 && notifications[0].Headline != nil {
		message.Subject = *notifications[0].Headline
	}

	for _, notification := range notifications {
		personalization := mail.NewPersonalization()
		personalization.AddTos(mail.NewEmail("", string(notification.Address)))

		// Add dynamic template data
		if notification.Meta != nil && notification.Meta.Params != nil {
			for key, value := range notification.Meta.Params {
				personalization.SetDynamicTemplateData(key, value)
			}
		}

		message.AddPersonalizations(personalization)
	}

	if len(notifications) > 0 {
		message.AddContent(mail.NewContent("text/plain", notifications[0].Body))
		message.AddContent(mail.NewContent("text/html", notifications[0].Body))
	}

	response, err := p.client.Send(message)
	if err != nil {
		return fmt.Errorf("sendgrid batch error: %w", err)
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("sendgrid batch returned status %d: %s", response.StatusCode, response.Body)
	}

	return nil
}

func (p *SendGridProvider) ValidateConfig() error {
	if p.config.APIKey == "" {
		return fmt.Errorf("sendgrid api key is required")
	}
	if p.config.FromEmail == "" {
		return fmt.Errorf("sendgrid from email is required")
	}
	return nil
}

// Stub implementations for other providers
func NewSendXProvider(config map[string]interface{}) (EmailProvider, error) {
	// Implementation for SendX provider
	return nil, fmt.Errorf("sendx provider not implemented")
}

func NewTwilioProvider(config map[string]interface{}) (SMSProvider, error) {
	// Implementation for Twilio provider
	return nil, fmt.Errorf("twilio provider not implemented")
}

func NewFCMProvider(config map[string]interface{}) (PushProvider, error) {
	// Implementation for FCM provider
	return nil, fmt.Errorf("fcm provider not implemented")
}

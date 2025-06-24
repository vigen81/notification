package email

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strconv"
	"strings"

	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/models"
)

// SMTPProvider implements the EmailProvider interface for SMTP
type SMTPProvider struct {
	config models.SMTPConfig
}

// NewSMTPProvider creates a new SMTP provider
func NewSMTPProvider(config map[string]interface{}) (*SMTPProvider, error) {
	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal SMTP config: %w", err)
	}

	var smtpConfig models.SMTPConfig
	if err := json.Unmarshal(configBytes, &smtpConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal SMTP config: %w", err)
	}

	provider := &SMTPProvider{
		config: smtpConfig,
	}

	if err := provider.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid SMTP config: %w", err)
	}

	return provider, nil
}

// Send sends a single email via SMTP
func (s *SMTPProvider) Send(ctx context.Context, notification *ent.Notification, messageType models.MessageType) error {
	log := logger.WithRequest(notification.RequestID)

	// Get appropriate from address and name based on message type
	fromAddr := s.config.GetFromAddress(messageType)
	fromName := s.config.GetFromName(messageType)

	// Override with notification's from if provided
	if notification.From != "" {
		fromAddr = notification.From
	}

	// Build email message
	to := string(notification.Address)
	subject := ""
	if notification.Headline != "" {
		subject = notification.Headline
	}

	var replyTo *string
	if notification.ReplyTo != "" {
		replyTo = &notification.ReplyTo
	}

	message := s.buildEmailMessage(fromAddr, fromName, to, subject, notification.Body, replyTo)

	// Send email
	if err := s.sendEmail(fromAddr, []string{to}, message); err != nil {
		log.Error("Failed to send SMTP email", err, map[string]interface{}{
			"notification_id": notification.ID,
			"to":              to,
			"from":            fromAddr,
		})
		return fmt.Errorf("failed to send SMTP email: %w", err)
	}

	log.Info("SMTP email sent successfully", map[string]interface{}{
		"notification_id": notification.ID,
		"to":              to,
		"from":            fromAddr,
	})

	return nil
}

// SendBatch sends multiple emails via SMTP
func (s *SMTPProvider) SendBatch(ctx context.Context, notifications []*ent.Notification, messageType models.MessageType) error {
	log := logger.To("smtp_batch")

	for _, notification := range notifications {
		if err := s.Send(ctx, notification, messageType); err != nil {
			log.Error("Failed to send email in batch", err, map[string]interface{}{
				"notification_id": notification.ID,
				"batch_size":      len(notifications),
			})
			// Continue with other emails even if one fails
		}
	}

	log.Info("SMTP batch processing completed", map[string]interface{}{
		"batch_size": len(notifications),
	})

	return nil
}

// ValidateConfig validates the SMTP configuration
func (s *SMTPProvider) ValidateConfig() error {
	if s.config.Host == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if s.config.Port == "" {
		return fmt.Errorf("SMTP port is required")
	}
	if s.config.Username == "" {
		return fmt.Errorf("SMTP username is required")
	}
	if s.config.Password == "" {
		return fmt.Errorf("SMTP password is required")
	}
	return nil
}

// GetType returns the provider type
func (s *SMTPProvider) GetType() string {
	return "smtp"
}

// buildEmailMessage constructs the email message
func (s *SMTPProvider) buildEmailMessage(from, fromName, to, subject, body string, replyTo *string) []byte {
	var message strings.Builder

	// Headers
	if fromName != "" {
		message.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, from))
	} else {
		message.WriteString(fmt.Sprintf("From: %s\r\n", from))
	}

	message.WriteString(fmt.Sprintf("To: %s\r\n", to))

	if replyTo != nil && *replyTo != "" {
		message.WriteString(fmt.Sprintf("Reply-To: %s\r\n", *replyTo))
	}

	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("\r\n")

	// Body
	message.WriteString(body)

	return []byte(message.String())
}

// sendEmail sends the email using SMTP
func (s *SMTPProvider) sendEmail(from string, to []string, message []byte) error {
	// Parse port
	port, err := strconv.Atoi(s.config.Port)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	// Server address
	addr := fmt.Sprintf("%s:%d", s.config.Host, port)

	// Setup authentication if required
	var auth smtp.Auth
	if s.config.SMTPAuth == "1" {
		auth = smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)
	}

	// Handle SSL/TLS
	if s.config.SMTPSecure == "ssl" || s.config.SMTPSecure == "tls" {
		return s.sendEmailTLS(addr, auth, from, to, message)
	}

	// Plain SMTP
	return smtp.SendMail(addr, auth, from, to, message)
}

// sendEmailTLS sends email using TLS connection
func (s *SMTPProvider) sendEmailTLS(addr string, auth smtp.Auth, from string, to []string, message []byte) error {
	// Create TLS config
	tlsConfig := &tls.Config{
		ServerName:         s.config.Host,
		InsecureSkipVerify: false,
	}

	// Connect to server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect with TLS: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate if required
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	if _, err := writer.Write(message); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return writer.Close()
}

package providers

import (
	"context"

	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/internal/models"
)

// EmailProvider defines the interface for email providers
type EmailProvider interface {
	Send(ctx context.Context, notification *ent.Notification, messageType models.MessageType) error
	SendBatch(ctx context.Context, notifications []*ent.Notification, messageType models.MessageType) error
	ValidateConfig() error
	GetType() string
}

// SMSProvider defines the interface for SMS providers
type SMSProvider interface {
	Send(ctx context.Context, notification *ent.Notification, messageType models.MessageType) error
	SendBatch(ctx context.Context, notifications []*ent.Notification, messageType models.MessageType) error
	ValidateConfig() error
	GetType() string
}

// PushProvider defines the interface for push notification providers
type PushProvider interface {
	Send(ctx context.Context, notification *ent.Notification, messageType models.MessageType) error
	SendBatch(ctx context.Context, notifications []*ent.Notification, messageType models.MessageType) error
	ValidateConfig() error
	GetType() string
}

// ProviderFactory defines factory functions for creating providers
type EmailProviderFactory func(config map[string]interface{}) (EmailProvider, error)
type SMSProviderFactory func(config map[string]interface{}) (SMSProvider, error)
type PushProviderFactory func(config map[string]interface{}) (PushProvider, error)

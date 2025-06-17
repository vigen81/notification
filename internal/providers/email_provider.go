package providers

import (
	"context"

	"gitlab.smartbet.am/golang/notification/ent"
)

type EmailProvider interface {
	Send(ctx context.Context, notification *ent.Notification, templateData map[string]interface{}) error
	SendBatch(ctx context.Context, notifications []*ent.Notification, templateData map[string]interface{}) error
	ValidateConfig() error
}

type SMSProvider interface {
	Send(ctx context.Context, notification *ent.Notification, templateData map[string]interface{}) error
	SendBatch(ctx context.Context, notifications []*ent.Notification, templateData map[string]interface{}) error
	ValidateConfig() error
}

type PushProvider interface {
	Send(ctx context.Context, notification *ent.Notification, templateData map[string]interface{}) error
	SendBatch(ctx context.Context, notifications []*ent.Notification, templateData map[string]interface{}) error
	ValidateConfig() error
}

package models

import (
	"encoding/json"
	"time"
)

type NotificationType string

const (
	TypeEmail NotificationType = "EMAIL"
	TypeSMS   NotificationType = "SMS"
	TypePush  NotificationType = "PUSH"
)

type NotificationStatus string

const (
	StatusActive    NotificationStatus = "ACTIVE"
	StatusCompleted NotificationStatus = "COMPLETED"
	StatusCancel    NotificationStatus = "CANCEL"
	StatusPending   NotificationStatus = "PENDING"
	StatusFailed    NotificationStatus = "FAILED"
)

type NotificationRequest struct {
	RequestID  string                 `json:"request_id"`
	TenantID   int64                  `json:"tenant_id"`
	Type       NotificationType       `json:"type"`
	TemplateID string                 `json:"template_id,omitempty"`
	Recipients []string               `json:"recipients"`
	Body       string                 `json:"body,omitempty"`
	Headline   string                 `json:"headline,omitempty"`
	From       string                 `json:"from,omitempty"`
	ReplyTo    string                 `json:"reply_to,omitempty"`
	Tag        string                 `json:"tag,omitempty"`
	ScheduleTS *int64                 `json:"schedule_ts,omitempty"`
	Meta       *NotificationMeta      `json:"meta,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Locale     string                 `json:"locale,omitempty"`
	BatchID    string                 `json:"batch_id,omitempty"`
}

type NotificationMeta struct {
	Service    string                 `json:"service,omitempty"`
	TemplateID string                 `json:"template_id,omitempty"`
	Params     map[string]interface{} `json:"params,omitempty"`
	Attachment *Attachment            `json:"attachment,omitempty"`
	Data       json.RawMessage        `json:"data,omitempty"`
}

type Attachment struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	Disposition string `json:"disposition"`
	Type        string `json:"type"`
}

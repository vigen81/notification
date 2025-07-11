package models

import (
	"encoding/json"
	"time"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	TypeEmail NotificationType = "EMAIL"
	TypeSMS   NotificationType = "SMS"
	TypePush  NotificationType = "PUSH"
)

// NotificationStatus represents the status of a notification
type NotificationStatus string

const (
	StatusActive    NotificationStatus = "ACTIVE"
	StatusCompleted NotificationStatus = "COMPLETED"
	StatusCancel    NotificationStatus = "CANCEL"
	StatusPending   NotificationStatus = "PENDING"
	StatusFailed    NotificationStatus = "FAILED"
)

// MessageType represents the category of message for routing
type MessageType string

const (
	MessageTypeBonus   MessageType = "bonus"
	MessageTypePromo   MessageType = "promo"
	MessageTypeReport  MessageType = "report"
	MessageTypeSystem  MessageType = "system"
	MessageTypePayment MessageType = "payment"
	MessageTypeSupport MessageType = "support"
)

// NotificationRequest represents the incoming notification request
type NotificationRequest struct {
	// Public fields for API
	TenantID    int64                  `json:"tenant_id" example:"1001"`
	Type        NotificationType       `json:"type" example:"EMAIL"`
	Recipients  []string               `json:"recipients" example:"user@example.com"`
	Body        string                 `json:"body" example:"Hello World! This is your notification."`
	Headline    string                 `json:"headline,omitempty" example:"Important Notification"`
	From        string                 `json:"from,omitempty" example:"noreply@example.com"`
	ReplyTo     string                 `json:"reply_to,omitempty" example:"support@example.com"`
	Tag         string                 `json:"tag,omitempty" example:"marketing"`
	ScheduleTS  *int64                 `json:"schedule_ts,omitempty" example:"1640995200"`
	MessageType MessageType            `json:"message_type,omitempty" example:"bonus"`
	Data        map[string]interface{} `json:"data,omitempty"`

	// Internal fields (not exposed in API)
	RequestID string            `json:"request_id,omitempty" swaggerignore:"true"`
	BatchID   string            `json:"batch_id,omitempty" swaggerignore:"true"`
	Meta      *NotificationMeta `json:"meta,omitempty" swaggerignore:"true"`
}

// BatchNotificationRequest represents a batch notification request
type BatchNotificationRequest struct {
	TenantID    int64                  `json:"tenant_id" example:"1001"`
	Type        NotificationType       `json:"type" example:"EMAIL"`
	Recipients  []string               `json:"recipients" example:"user1@example.com,user2@example.com"`
	Body        string                 `json:"body" example:"Hello! This is a batch notification."`
	Headline    string                 `json:"headline,omitempty" example:"Batch Notification"`
	From        string                 `json:"from,omitempty" example:"noreply@example.com"`
	ReplyTo     string                 `json:"reply_to,omitempty" example:"support@example.com"`
	Tag         string                 `json:"tag,omitempty" example:"batch"`
	ScheduleTS  *int64                 `json:"schedule_ts,omitempty" example:"1640995200"`
	MessageType MessageType            `json:"message_type,omitempty" example:"promo"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// KafkaNotificationRequest represents the Kafka message structure
type KafkaNotificationRequest struct {
	TenantID    int64                  `json:"tenant_id" example:"1001"`
	Type        NotificationType       `json:"type" example:"EMAIL"`
	Recipients  []string               `json:"recipients" example:"user@example.com"`
	Body        string                 `json:"body" example:"Direct Kafka notification"`
	Headline    string                 `json:"headline,omitempty" example:"Kafka Notification"`
	MessageType MessageType            `json:"message_type,omitempty" example:"system"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// NotificationResponse represents the response for notification requests
type NotificationResponse struct {
	RequestID string `json:"request_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    string `json:"status" example:"queued"`
	Message   string `json:"message" example:"Notification queued for processing"`
}

// BatchNotificationResponse represents the response for batch notification requests
type BatchNotificationResponse struct {
	BatchID          string `json:"batch_id" example:"batch_123"`
	TotalRecipients  int    `json:"total_recipients" example:"100"`
	QueuedRecipients int    `json:"queued_recipients" example:"100"`
	Status           string `json:"status" example:"processing"`
}

// NotificationStatusResponse represents the status response for a notification
type NotificationStatusResponse struct {
	RequestID    string    `json:"request_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status       string    `json:"status" example:"COMPLETED"`
	Type         string    `json:"type" example:"EMAIL"`
	TenantID     int64     `json:"tenant_id" example:"1001"`
	CreatedAt    time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt    time.Time `json:"updated_at" example:"2023-01-01T00:01:00Z"`
	ErrorMessage *string   `json:"error_message,omitempty" example:"SMTP connection failed"`
	ScheduleTS   *int64    `json:"schedule_ts,omitempty" example:"1640995200"`
}

// BatchNotificationStatusResponse represents the status response for a batch of notifications
type BatchNotificationStatusResponse struct {
	RequestID      string    `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	BatchID        string    `json:"batch_id,omitempty" example:"batch_123"`
	Status         string    `json:"status" example:"COMPLETED"`
	Type           string    `json:"type" example:"EMAIL"`
	TenantID       int64     `json:"tenant_id" example:"1001"`
	CreatedAt      time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt      time.Time `json:"updated_at" example:"2023-01-01T00:01:00Z"`
	TotalCount     int       `json:"total_count" example:"100"`
	CompletedCount int       `json:"completed_count" example:"95"`
	FailedCount    int       `json:"failed_count" example:"3"`
	PendingCount   int       `json:"pending_count" example:"2"`
}

// KafkaResponse represents the response for Kafka publishing
type KafkaResponse struct {
	RequestID string `json:"request_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Status    string `json:"status" example:"published"`
}

// ErrorResponse represents error responses
type ErrorResponse struct {
	Error     string      `json:"error" example:"Validation failed"`
	Code      string      `json:"code" example:"VALIDATION_ERROR"`
	Message   string      `json:"message,omitempty" example:"Additional error details"`
	RequestID string      `json:"request_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
	Timestamp time.Time   `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Details   interface{} `json:"details,omitempty"`
}

// HealthResponse represents health check responses
type HealthResponse struct {
	Status    string             `json:"status" example:"ok"`
	Service   string             `json:"service" example:"notification-engine"`
	Timestamp time.Time          `json:"timestamp" example:"2023-01-01T00:00:00Z"`
	Version   string             `json:"version,omitempty" example:"1.0.0"`
	Checks    *HealthCheckDetail `json:"checks,omitempty"`
}

// HealthCheckDetail represents detailed health check information
type HealthCheckDetail struct {
	Database string `json:"database" example:"ok"`
	Kafka    string `json:"kafka" example:"ok"`
}

// Internal models (used internally but not directly exposed in Swagger)

// NotificationMeta contains additional metadata for the notification
type NotificationMeta struct {
	Service    string                 `json:"service,omitempty" example:"user-service"`
	TemplateID string                 `json:"template_id,omitempty" example:"welcome_template"`
	Params     map[string]interface{} `json:"params,omitempty"`
	Attachment *Attachment            `json:"attachment,omitempty"`
	Data       json.RawMessage        `json:"data,omitempty"`
}

// Attachment represents file attachments for notifications
type Attachment struct {
	Filename    string `json:"filename" example:"document.pdf"`
	Content     string `json:"content" example:"base64_encoded_content"`
	Disposition string `json:"disposition" example:"attachment"`
	Type        string `json:"type" example:"application/pdf"`
}

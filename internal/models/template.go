package models

import (
	"time"
)

type Template struct {
	ID        int                    `json:"id"`
	TenantID  int64                  `json:"tenant_id"`
	Name      string                 `json:"name"`
	Type      NotificationType       `json:"type"`
	Subject   string                 `json:"subject,omitempty"`
	Body      string                 `json:"body"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type TemplateRequest struct {
	Name     string                 `json:"name"`
	Type     NotificationType       `json:"type"`
	Subject  string                 `json:"subject,omitempty"`
	Body     string                 `json:"body"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

package models

import (
	"time"

	"gitlab.smartbet.am/golang/notification/ent/schema"
)

// PartnerConfig represents the configuration for a specific tenant/partner
type PartnerConfig struct {
	ID             string                      `json:"id"`
	TenantID       int64                       `json:"tenant_id" example:"123"`
	EmailProviders []schema.ProviderConfig     `json:"email_providers"`
	SMSProviders   []schema.ProviderConfig     `json:"sms_providers"`
	PushProviders  []schema.ProviderConfig     `json:"push_providers"`
	BatchConfig    *schema.BatchConfig         `json:"batch_config"`
	RateLimits     map[string]schema.RateLimit `json:"rate_limits"`
	Enabled        bool                        `json:"enabled" example:"true"`
	CreatedAt      time.Time                   `json:"created_at"`
	UpdatedAt      time.Time                   `json:"updated_at"`
}

// Legacy type aliases for backward compatibility (if needed elsewhere in codebase)
type ProviderConfig = schema.ProviderConfig
type BatchConfig = schema.BatchConfig
type RateLimit = schema.RateLimit
type SMTPConfig = schema.SMTPConfig

// PartnerConfigRequest represents the request to update partner configuration
type PartnerConfigRequest struct {
	EmailProviders []schema.ProviderConfig     `json:"email_providers"`
	SMSProviders   []schema.ProviderConfig     `json:"sms_providers"`
	PushProviders  []schema.ProviderConfig     `json:"push_providers"`
	BatchConfig    *schema.BatchConfig         `json:"batch_config"`
	RateLimits     map[string]schema.RateLimit `json:"rate_limits"`
	Enabled        bool                        `json:"enabled" example:"true"`
}

// AddProviderRequest represents the request to add a new provider
type AddProviderRequest struct {
	Name     string                 `json:"name" example:"secondary"`
	Type     string                 `json:"type" example:"sendx"`
	Priority int                    `json:"priority" example:"2"`
	Enabled  bool                   `json:"enabled" example:"true"`
	Config   map[string]interface{} `json:"config"`
}

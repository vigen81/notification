package models

import (
	"time"

	"gitlab.smartbet.am/golang/notification/ent/schema"
)

// ProviderConfig represents a provider configuration
type ProviderConfig struct {
	Name     string                 `json:"name" example:"primary"`
	Type     string                 `json:"type" example:"smtp"`
	Priority int                    `json:"priority" example:"1"`
	Enabled  bool                   `json:"enabled" example:"true"`
	Config   map[string]interface{} `json:"config"`
}

// BatchConfig represents batch processing configuration
type BatchConfig struct {
	Enabled              bool `json:"enabled" example:"true"`
	MaxBatchSize         int  `json:"max_batch_size" example:"100"`
	FlushIntervalSeconds int  `json:"flush_interval_seconds" example:"10"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	Limit    int    `json:"limit" example:"1000"`
	Window   string `json:"window" example:"1h"`
	Strategy string `json:"strategy" example:"sliding"`
}

// PartnerConfig represents the configuration for a specific tenant/partner
type PartnerConfig struct {
	ID             string                      `json:"id" example:"goodwin-casino-1001"`
	TenantID       int64                       `json:"tenant_id" example:"1001"`
	EmailProviders []schema.ProviderConfig     `json:"email_providers"`
	SMSProviders   []schema.ProviderConfig     `json:"sms_providers"`
	PushProviders  []schema.ProviderConfig     `json:"push_providers"`
	BatchConfig    *schema.BatchConfig         `json:"batch_config"`
	RateLimits     map[string]schema.RateLimit `json:"rate_limits"`
	Enabled        bool                        `json:"enabled" example:"true"`
	CreatedAt      time.Time                   `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt      time.Time                   `json:"updated_at" example:"2023-01-01T00:01:00Z"`
}

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

// ConfigSuccessResponse represents successful configuration operations
type ConfigSuccessResponse struct {
	Message   string    `json:"message" example:"Configuration updated successfully"`
	Status    string    `json:"status" example:"success"`
	TenantID  int64     `json:"tenant_id" example:"1001"`
	Timestamp time.Time `json:"timestamp,omitempty" example:"2023-01-01T00:00:00Z"`
}

// Example configurations for documentation

// SMTPProviderConfig shows SMTP configuration structure for documentation
type SMTPProviderConfig struct {
	Host               string `json:"Host" example:"smtp.sendgrid.net"`
	Port               string `json:"Port" example:"465"`
	Username           string `json:"Username" example:"apikey"`
	Password           string `json:"Password" example:"SG.your_api_key"`
	SMTPAuth           string `json:"SMTPAuth" example:"1"`
	SMTPSecure         string `json:"SMTPSecure" example:"ssl"`
	MSGBonusFrom       string `json:"MSGBonusFrom" example:"bonus@goodwin.am"`
	MSGPromoFrom       string `json:"MSGPromoFrom" example:"promo@goodwin.am"`
	MSGSystemFrom      string `json:"MSGSystemFrom" example:"noreply@goodwin.am"`
	MSGBonusFromName   string `json:"MSGBonusFromName" example:"Goodwin Bonus Team"`
	MSGPromoFromName   string `json:"MSGPromoFromName" example:"Goodwin Promotions"`
	MSGSystemFromName  string `json:"MSGSystemFromName" example:"Goodwin System"`
	MSGReportFrom      string `json:"MSGReportFrom" example:"reports@goodwin.am"`
	MSGPaymentFrom     string `json:"MSGPaymentFrom" example:"payments@goodwin.am"`
	MSGSupportFrom     string `json:"MSGSupportFrom" example:"support@goodwin.am"`
	MSGReportFromName  string `json:"MSGReportFromName" example:"Goodwin Reports"`
	MSGPaymentFromName string `json:"MSGPaymentFromName" example:"Goodwin Payments"`
	MSGSupportFromName string `json:"MSGSupportFromName" example:"Goodwin Support"`
}

// TwilioProviderConfig shows Twilio configuration structure for documentation
type TwilioProviderConfig struct {
	AccountSID     string `json:"account_sid" example:"AC_your_account_sid"`
	AuthToken      string `json:"auth_token" example:"your_auth_token"`
	FromNumber     string `json:"from_number" example:"+1234567890"`
	MSGBonusFrom   string `json:"MSGBonusFrom,omitempty" example:"+1234567891"`
	MSGPromoFrom   string `json:"MSGPromoFrom,omitempty" example:"+1234567892"`
	MSGSystemFrom  string `json:"MSGSystemFrom,omitempty" example:"+1234567893"`
	MSGReportFrom  string `json:"MSGReportFrom,omitempty" example:"+1234567894"`
	MSGPaymentFrom string `json:"MSGPaymentFrom,omitempty" example:"+1234567895"`
	MSGSupportFrom string `json:"MSGSupportFrom,omitempty" example:"+1234567896"`
}

// FCMProviderConfig shows FCM configuration structure for documentation
type FCMProviderConfig struct {
	ServerKey string `json:"server_key" example:"fcm_server_key_123"`
	ProjectID string `json:"project_id" example:"goodwin-casino"`
}

// Legacy type aliases for backward compatibility (if needed elsewhere in codebase)
type ProviderConfigLegacy = schema.ProviderConfig
type BatchConfigLegacy = schema.BatchConfig
type RateLimitLegacy = schema.RateLimit
type SMTPConfigLegacy = schema.SMTPConfig

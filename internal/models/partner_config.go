package models

import (
	"time"
)

// PartnerConfig represents the configuration for a specific tenant/partner
type PartnerConfig struct {
	ID             string               `json:"id"`
	TenantID       int64                `json:"tenant_id" example:"123"`
	EmailProviders []ProviderConfig     `json:"email_providers"`
	SMSProviders   []ProviderConfig     `json:"sms_providers"`
	PushProviders  []ProviderConfig     `json:"push_providers"`
	BatchConfig    BatchConfig          `json:"batch_config"`
	RateLimits     map[string]RateLimit `json:"rate_limits"`
	Enabled        bool                 `json:"enabled" example:"true"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// ProviderConfig represents the configuration for a specific provider
type ProviderConfig struct {
	Name     string                 `json:"name" example:"primary"`
	Type     string                 `json:"type" example:"sendgrid"`
	Priority int                    `json:"priority" example:"1"`
	Enabled  bool                   `json:"enabled" example:"true"`
	Config   map[string]interface{} `json:"config"`
}

// SMTPConfig represents SMTP configuration with multiple from addresses
type SMTPConfig struct {
	Host       string `json:"Host" example:"smtp.sendgrid.net"`
	Port       string `json:"Port" example:"465"`
	Username   string `json:"Username" example:"apikey"`
	Password   string `json:"Password" example:"your_api_key"`
	SMTPAuth   string `json:"SMTPAuth" example:"1"`
	SMTPDebug  string `json:"SMTPDebug" example:"0"`
	SMTPSecure string `json:"SMTPSecure" example:"ssl"`

	// Message type specific from addresses
	MSGBonusFrom   string `json:"MSGBonusFrom" example:"bonus@goodwin.am"`
	MSGPromoFrom   string `json:"MSGPromoFrom" example:"noreply@goodwin.am"`
	MSGReportFrom  string `json:"MSGReportFrom" example:"report@goodwin.am"`
	MSGSystemFrom  string `json:"MSGSystemFrom" example:"noreply@goodwin.am"`
	MSGPaymentFrom string `json:"MSGPaymentFrom" example:"noreply@goodwin.am"`
	MSGSupportFrom string `json:"MSGSupportFrom" example:"support@goodwin.am"`

	// Message type specific from names
	MSGBonusFromName   string `json:"MSGBonusFromName" example:"Goodwin Bonus"`
	MSGPromoFromName   string `json:"MSGPromoFromName" example:"Goodwin promo"`
	MSGReportFromName  string `json:"MSGReportFromName" example:"Smartbet Report"`
	MSGSystemFromName  string `json:"MSGSystemFromName" example:"Goodwin system"`
	MSGPaymentFromName string `json:"MSGPaymentFromName" example:"Goodwin payment"`
	MSGSupportFromName string `json:"MSGSupportFromName" example:"Goodwin support"`
}

// BatchConfig represents batch processing configuration
type BatchConfig struct {
	Enabled       bool `json:"enabled" example:"true"`
	MaxBatchSize  int  `json:"max_batch_size" example:"100"`
	FlushInterval int  `json:"flush_interval_seconds" example:"10"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	Limit    int    `json:"limit" example:"1000"`
	Window   string `json:"window" example:"1h"`
	Strategy string `json:"strategy" example:"sliding"`
}

// PartnerConfigRequest represents the request to update partner configuration
type PartnerConfigRequest struct {
	EmailProviders []ProviderConfig     `json:"email_providers"`
	SMSProviders   []ProviderConfig     `json:"sms_providers"`
	PushProviders  []ProviderConfig     `json:"push_providers"`
	BatchConfig    BatchConfig          `json:"batch_config"`
	RateLimits     map[string]RateLimit `json:"rate_limits"`
	Enabled        bool                 `json:"enabled" example:"true"`
}

// AddProviderRequest represents the request to add a new provider
type AddProviderRequest struct {
	Name     string                 `json:"name" example:"secondary"`
	Type     string                 `json:"type" example:"sendx"`
	Priority int                    `json:"priority" example:"2"`
	Enabled  bool                   `json:"enabled" example:"true"`
	Config   map[string]interface{} `json:"config"`
}

// GetFromAddress returns the appropriate from address based on message type
func (s *SMTPConfig) GetFromAddress(messageType MessageType) string {
	switch messageType {
	case MessageTypeBonus:
		return s.MSGBonusFrom
	case MessageTypePromo:
		return s.MSGPromoFrom
	case MessageTypeReport:
		return s.MSGReportFrom
	case MessageTypeSystem:
		return s.MSGSystemFrom
	case MessageTypePayment:
		return s.MSGPaymentFrom
	case MessageTypeSupport:
		return s.MSGSupportFrom
	default:
		return s.MSGSystemFrom
	}
}

// GetFromName returns the appropriate from name based on message type
func (s *SMTPConfig) GetFromName(messageType MessageType) string {
	switch messageType {
	case MessageTypeBonus:
		return s.MSGBonusFromName
	case MessageTypePromo:
		return s.MSGPromoFromName
	case MessageTypeReport:
		return s.MSGReportFromName
	case MessageTypeSystem:
		return s.MSGSystemFromName
	case MessageTypePayment:
		return s.MSGPaymentFromName
	case MessageTypeSupport:
		return s.MSGSupportFromName
	default:
		return s.MSGSystemFromName
	}
}

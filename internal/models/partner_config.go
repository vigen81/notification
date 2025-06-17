package models

import (
	"time"
)

type PartnerConfig struct {
	ID             string               `json:"id"`
	TenantID       int64                `json:"tenant_id"`
	EmailProviders []ProviderConfig     `json:"email_providers"`
	SMSProviders   []ProviderConfig     `json:"sms_providers"`
	PushProviders  []ProviderConfig     `json:"push_providers"`
	BatchConfig    BatchConfig          `json:"batch_config"`
	RateLimits     map[string]RateLimit `json:"rate_limits"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type ProviderConfig struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Priority int                    `json:"priority"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
}

type BatchConfig struct {
	Enabled       bool `json:"enabled"`
	MaxBatchSize  int  `json:"max_batch_size"`
	FlushInterval int  `json:"flush_interval_seconds"`
}

type RateLimit struct {
	Limit    int    `json:"limit"`
	Window   string `json:"window"`
	Strategy string `json:"strategy"`
}

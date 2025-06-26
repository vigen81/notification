package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"entgo.io/ent/schema/mixin"
)

// ProviderConfig represents a single provider configuration
type ProviderConfig struct {
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`
	Priority int                    `json:"priority"`
	Enabled  bool                   `json:"enabled"`
	Config   map[string]interface{} `json:"config"`
}

// BatchConfig represents batch processing configuration
type BatchConfig struct {
	Enabled              bool `json:"enabled"`
	MaxBatchSize         int  `json:"max_batch_size"`
	FlushIntervalSeconds int  `json:"flush_interval_seconds"`
}

// RateLimit represents rate limiting configuration
type RateLimit struct {
	Limit    int    `json:"limit"`
	Window   string `json:"window"`
	Strategy string `json:"strategy"`
}

// SMTPConfig represents SMTP configuration with multiple from addresses
type SMTPConfig struct {
	Host       string `json:"Host"`
	Port       string `json:"Port"`
	Username   string `json:"Username"`
	Password   string `json:"Password"`
	SMTPAuth   string `json:"SMTPAuth"`
	SMTPDebug  string `json:"SMTPDebug"`
	SMTPSecure string `json:"SMTPSecure"`

	// Message type specific from addresses
	MSGBonusFrom   string `json:"MSGBonusFrom"`
	MSGPromoFrom   string `json:"MSGPromoFrom"`
	MSGReportFrom  string `json:"MSGReportFrom"`
	MSGSystemFrom  string `json:"MSGSystemFrom"`
	MSGPaymentFrom string `json:"MSGPaymentFrom"`
	MSGSupportFrom string `json:"MSGSupportFrom"`

	// Message type specific from names
	MSGBonusFromName   string `json:"MSGBonusFromName"`
	MSGPromoFromName   string `json:"MSGPromoFromName"`
	MSGReportFromName  string `json:"MSGReportFromName"`
	MSGSystemFromName  string `json:"MSGSystemFromName"`
	MSGPaymentFromName string `json:"MSGPaymentFromName"`
	MSGSupportFromName string `json:"MSGSupportFromName"`
}

// GetFromAddress returns the appropriate from address based on message type
func (s *SMTPConfig) GetFromAddress(messageType string) string {
	switch messageType {
	case "bonus":
		return s.MSGBonusFrom
	case "promo":
		return s.MSGPromoFrom
	case "report":
		return s.MSGReportFrom
	case "system":
		return s.MSGSystemFrom
	case "payment":
		return s.MSGPaymentFrom
	case "support":
		return s.MSGSupportFrom
	default:
		return s.MSGSystemFrom
	}
}

// GetFromName returns the appropriate from name based on message type
func (s *SMTPConfig) GetFromName(messageType string) string {
	switch messageType {
	case "bonus":
		return s.MSGBonusFromName
	case "promo":
		return s.MSGPromoFromName
	case "report":
		return s.MSGReportFromName
	case "system":
		return s.MSGSystemFromName
	case "payment":
		return s.MSGPaymentFromName
	case "support":
		return s.MSGSupportFromName
	default:
		return s.MSGSystemFromName
	}
}

// PartnerConfig holds the schema definition for the PartnerConfig entity.
type PartnerConfig struct {
	ent.Schema
}

// Fields of the PartnerConfig.
func (PartnerConfig) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").Default(""),
		field.Int64("tenant_id").Unique(),

		// Use proper Go types instead of []byte{}
		field.JSON("email_providers", []ProviderConfig{}).Optional(),
		field.JSON("sms_providers", []ProviderConfig{}).Optional(),
		field.JSON("push_providers", []ProviderConfig{}).Optional(),
		field.JSON("batch_config", &BatchConfig{}).Optional(),
		field.JSON("rate_limits", map[string]RateLimit{}).Optional(),

		field.Bool("enabled").Default(true),
	}
}

func (PartnerConfig) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.Time{},
	}
}

// Edges of the PartnerConfig.
func (PartnerConfig) Edges() []ent.Edge {
	return nil
}

func (PartnerConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id"),
		index.Fields("enabled"),
	}
}

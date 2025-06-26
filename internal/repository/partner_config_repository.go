package repository

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/ent/partnerconfig"
	"gitlab.smartbet.am/golang/notification/ent/schema"
	"gitlab.smartbet.am/golang/notification/internal/models"
)

type PartnerConfigRepository struct {
	client *ent.Client
	logger *logrus.Logger
}

func NewPartnerConfigRepository(client *ent.Client, logger *logrus.Logger) *PartnerConfigRepository {
	return &PartnerConfigRepository{
		client: client,
		logger: logger,
	}
}

func (r *PartnerConfigRepository) GetByTenantID(ctx context.Context, tenantID int64) (*models.PartnerConfig, error) {
	config, err := r.client.PartnerConfig.Query().
		Where(partnerconfig.TenantID(tenantID)).
		First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			// Return default config if not found
			return r.getDefaultConfig(tenantID), nil
		}
		return nil, err
	}

	return r.entToModel(config), nil
}

func (r *PartnerConfigRepository) Save(ctx context.Context, config *models.PartnerConfig) error {
	// Check if exists
	exists, err := r.client.PartnerConfig.Query().
		Where(partnerconfig.TenantID(config.TenantID)).
		Exist(ctx)

	if err != nil {
		return err
	}

	if exists {
		// Update existing
		return r.client.PartnerConfig.Update().
			Where(partnerconfig.TenantID(config.TenantID)).
			SetEmailProviders(config.EmailProviders).
			SetSmsProviders(config.SMSProviders).
			SetPushProviders(config.PushProviders).
			SetBatchConfig(config.BatchConfig).
			SetRateLimits(config.RateLimits).
			SetEnabled(config.Enabled).
			Exec(ctx)
	}

	// Create new
	_, err = r.client.PartnerConfig.Create().
		SetID(config.ID).
		SetTenantID(config.TenantID).
		SetEmailProviders(config.EmailProviders).
		SetSmsProviders(config.SMSProviders).
		SetPushProviders(config.PushProviders).
		SetBatchConfig(config.BatchConfig).
		SetRateLimits(config.RateLimits).
		SetEnabled(config.Enabled).
		Save(ctx)

	return err
}

func (r *PartnerConfigRepository) entToModel(config *ent.PartnerConfig) *models.PartnerConfig {
	return &models.PartnerConfig{
		ID:             config.ID,
		TenantID:       config.TenantID,
		EmailProviders: config.EmailProviders,
		SMSProviders:   config.SmsProviders,
		PushProviders:  config.PushProviders,
		BatchConfig:    config.BatchConfig,
		RateLimits:     config.RateLimits,
		Enabled:        config.Enabled,
		CreatedAt:      config.CreateTime,
		UpdatedAt:      config.UpdateTime,
	}
}

func (r *PartnerConfigRepository) getDefaultConfig(tenantID int64) *models.PartnerConfig {
	return &models.PartnerConfig{
		TenantID: tenantID,
		EmailProviders: []schema.ProviderConfig{
			{
				Name:     "default",
				Type:     "smtp",
				Priority: 1,
				Enabled:  true,
				Config:   map[string]interface{}{},
			},
		},
		SMSProviders: []schema.ProviderConfig{
			{
				Name:     "default",
				Type:     "twilio",
				Priority: 1,
				Enabled:  true,
				Config:   map[string]interface{}{},
			},
		},
		PushProviders: []schema.ProviderConfig{
			{
				Name:     "default",
				Type:     "fcm",
				Priority: 1,
				Enabled:  true,
				Config:   map[string]interface{}{},
			},
		},
		BatchConfig: &schema.BatchConfig{
			Enabled:              true,
			MaxBatchSize:         100,
			FlushIntervalSeconds: 10,
		},
		RateLimits: map[string]schema.RateLimit{
			"email": {Limit: 1000, Window: "1h", Strategy: "sliding"},
			"sms":   {Limit: 500, Window: "1h", Strategy: "sliding"},
			"push":  {Limit: 5000, Window: "1h", Strategy: "sliding"},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

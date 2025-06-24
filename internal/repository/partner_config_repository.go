package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/ent/partnerconfig"
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
	// Convert model to JSON for storage
	emailProviders, _ := json.Marshal(config.EmailProviders)
	smsProviders, _ := json.Marshal(config.SMSProviders)
	pushProviders, _ := json.Marshal(config.PushProviders)
	batchConfig, _ := json.Marshal(config.BatchConfig)
	rateLimits, _ := json.Marshal(config.RateLimits)

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
			SetEmailProviders(emailProviders).
			SetSmsProviders(smsProviders).
			SetPushProviders(pushProviders).
			SetBatchConfig(batchConfig).
			SetRateLimits(rateLimits).
			SetEnabled(config.Enabled).
			Exec(ctx)
	}

	// Create new
	_, err = r.client.PartnerConfig.Create().
		SetTenantID(config.TenantID).
		SetEmailProviders(emailProviders).
		SetSmsProviders(smsProviders).
		SetPushProviders(pushProviders).
		SetBatchConfig(batchConfig).
		SetRateLimits(rateLimits).
		SetEnabled(config.Enabled).
		Save(ctx)

	return err
}

func (r *PartnerConfigRepository) entToModel(config *ent.PartnerConfig) *models.PartnerConfig {
	model := &models.PartnerConfig{
		ID:        config.ID,
		TenantID:  config.TenantID,
		Enabled:   config.Enabled,
		CreatedAt: config.CreateTime,
		UpdatedAt: config.UpdateTime,
	}

	// Unmarshal JSON fields
	json.Unmarshal(config.EmailProviders, &model.EmailProviders)
	json.Unmarshal(config.SmsProviders, &model.SMSProviders)
	json.Unmarshal(config.PushProviders, &model.PushProviders)
	json.Unmarshal(config.BatchConfig, &model.BatchConfig)
	json.Unmarshal(config.RateLimits, &model.RateLimits)

	return model
}

func (r *PartnerConfigRepository) getDefaultConfig(tenantID int64) *models.PartnerConfig {
	return &models.PartnerConfig{
		TenantID: tenantID,
		EmailProviders: []models.ProviderConfig{
			{
				Name:     "default",
				Type:     "smtp",
				Priority: 1,
				Enabled:  true,
				Config:   map[string]interface{}{},
			},
		},
		SMSProviders: []models.ProviderConfig{
			{
				Name:     "default",
				Type:     "twilio",
				Priority: 1,
				Enabled:  true,
				Config:   map[string]interface{}{},
			},
		},
		PushProviders: []models.ProviderConfig{
			{
				Name:     "default",
				Type:     "fcm",
				Priority: 1,
				Enabled:  true,
				Config:   map[string]interface{}{},
			},
		},
		BatchConfig: models.BatchConfig{
			Enabled:       true,
			MaxBatchSize:  100,
			FlushInterval: 10,
		},
		RateLimits: map[string]models.RateLimit{
			"email": {Limit: 1000, Window: "1h", Strategy: "sliding"},
			"sms":   {Limit: 500, Window: "1h", Strategy: "sliding"},
			"push":  {Limit: 5000, Window: "1h", Strategy: "sliding"},
		},
		Enabled:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

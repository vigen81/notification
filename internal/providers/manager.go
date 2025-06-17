package providers

import (
	"context"
	"fmt"
	"sync"

	"gitlab.smartbet.am/golang/notification/internal/repository"
	"go.uber.org/zap"
)

type EmailProviderManager struct {
	registry   *ProviderRegistry
	configRepo *repository.PartnerConfigRepository
	providers  map[int64]EmailProvider
	mu         sync.RWMutex
	logger     *zap.Logger
}

func NewEmailProviderManager(
	registry *ProviderRegistry,
	configRepo *repository.PartnerConfigRepository,
	logger *zap.Logger,
) *EmailProviderManager {
	return &EmailProviderManager{
		registry:   registry,
		configRepo: configRepo,
		providers:  make(map[int64]EmailProvider),
		logger:     logger,
	}
}

func (m *EmailProviderManager) GetProvider(tenantID int64) (EmailProvider, error) {
	m.mu.RLock()
	provider, exists := m.providers[tenantID]
	m.mu.RUnlock()

	if exists {
		return provider, nil
	}

	// Load provider configuration
	config, err := m.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		return nil, err
	}

	// Find enabled email provider
	for _, providerConfig := range config.EmailProviders {
		if providerConfig.Enabled {
			provider, err := m.registry.CreateEmailProvider(providerConfig)
			if err != nil {
				continue
			}

			m.mu.Lock()
			m.providers[tenantID] = provider
			m.mu.Unlock()

			return provider, nil
		}
	}

	return nil, fmt.Errorf("no enabled email provider found for tenant %d", tenantID)
}

type SMSProviderManager struct {
	registry   *ProviderRegistry
	configRepo *repository.PartnerConfigRepository
	providers  map[int64]SMSProvider
	mu         sync.RWMutex
	logger     *zap.Logger
}

func NewSMSProviderManager(
	registry *ProviderRegistry,
	configRepo *repository.PartnerConfigRepository,
	logger *zap.Logger,
) *SMSProviderManager {
	return &SMSProviderManager{
		registry:   registry,
		configRepo: configRepo,
		providers:  make(map[int64]SMSProvider),
		logger:     logger,
	}
}

func (m *SMSProviderManager) GetProvider(tenantID int64) (SMSProvider, error) {
	// Similar implementation to EmailProviderManager
	return nil, fmt.Errorf("no enabled SMS provider found for tenant %d", tenantID)
}

type PushProviderManager struct {
	registry   *ProviderRegistry
	configRepo *repository.PartnerConfigRepository
	providers  map[int64]PushProvider
	mu         sync.RWMutex
	logger     *zap.Logger
}

func NewPushProviderManager(
	registry *ProviderRegistry,
	configRepo *repository.PartnerConfigRepository,
	logger *zap.Logger,
) *PushProviderManager {
	return &PushProviderManager{
		registry:   registry,
		configRepo: configRepo,
		providers:  make(map[int64]PushProvider),
		logger:     logger,
	}
}

func (m *PushProviderManager) GetProvider(tenantID int64) (PushProvider, error) {
	// Similar implementation to EmailProviderManager
	return nil, fmt.Errorf("no enabled push provider found for tenant %d", tenantID)
}

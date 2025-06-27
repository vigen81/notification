package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	_ "gitlab.smartbet.am/golang/notification/ent/schema"
	"gitlab.smartbet.am/golang/notification/internal/repository"
)

type EmailProviderManager struct {
	registry   *ProviderRegistry
	configRepo *repository.PartnerConfigRepository
	providers  map[int64]EmailProvider
	mu         sync.RWMutex
	logger     *logrus.Logger
}

func NewEmailProviderManager(
	registry *ProviderRegistry,
	configRepo *repository.PartnerConfigRepository,
	logger *logrus.Logger,
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
			// Convert schema.ProviderConfig to the format expected by registry
			provider, err := m.registry.CreateEmailProvider(providerConfig.Config, providerConfig.Type)
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
	logger     *logrus.Logger
}

func NewSMSProviderManager(
	registry *ProviderRegistry,
	configRepo *repository.PartnerConfigRepository,
	logger *logrus.Logger,
) *SMSProviderManager {
	return &SMSProviderManager{
		registry:   registry,
		configRepo: configRepo,
		providers:  make(map[int64]SMSProvider),
		logger:     logger,
	}
}

func (m *SMSProviderManager) GetProvider(tenantID int64) (SMSProvider, error) {
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

	// Find enabled SMS provider
	for _, providerConfig := range config.SMSProviders {
		if providerConfig.Enabled {
			// Convert schema.ProviderConfig to the format expected by registry
			provider, err := m.registry.CreateSMSProvider(providerConfig.Config, providerConfig.Type)
			if err != nil {
				continue
			}

			m.mu.Lock()
			m.providers[tenantID] = provider
			m.mu.Unlock()

			return provider, nil
		}
	}

	return nil, fmt.Errorf("no enabled SMS provider found for tenant %d", tenantID)
}

package providers

import (
	"fmt"
	"sync"

	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/providers/email"
)

type ProviderRegistry struct {
	emailFactories map[string]EmailProviderFactory
	smsFactories   map[string]SMSProviderFactory
	pushFactories  map[string]PushProviderFactory
	mu             sync.RWMutex
}

func NewProviderRegistry() *ProviderRegistry {
	registry := &ProviderRegistry{
		emailFactories: make(map[string]EmailProviderFactory),
		smsFactories:   make(map[string]SMSProviderFactory),
		pushFactories:  make(map[string]PushProviderFactory),
	}

	// Register default providers
	registry.RegisterEmailProvider("smtp", func(config map[string]interface{}) (EmailProvider, error) {
		provider, err := email.NewSMTPProvider(config)
		if err != nil {
			return nil, err
		}
		return provider, nil
	})

	return registry
}

func (r *ProviderRegistry) RegisterEmailProvider(name string, factory EmailProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emailFactories[name] = factory
}

func (r *ProviderRegistry) RegisterSMSProvider(name string, factory SMSProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.smsFactories[name] = factory
}

func (r *ProviderRegistry) RegisterPushProvider(name string, factory PushProviderFactory) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pushFactories[name] = factory
}

func (r *ProviderRegistry) CreateEmailProvider(config models.ProviderConfig) (EmailProvider, error) {
	r.mu.RLock()
	factory, exists := r.emailFactories[config.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("email provider %s not registered", config.Type)
	}

	return factory(config.Config)
}

func (r *ProviderRegistry) CreateSMSProvider(config models.ProviderConfig) (SMSProvider, error) {
	r.mu.RLock()
	factory, exists := r.smsFactories[config.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sms provider %s not registered", config.Type)
	}

	return factory(config.Config)
}

func (r *ProviderRegistry) CreatePushProvider(config models.ProviderConfig) (PushProvider, error) {
	r.mu.RLock()
	factory, exists := r.pushFactories[config.Type]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("push provider %s not registered", config.Type)
	}

	return factory(config.Config)
}

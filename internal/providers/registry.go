package providers

import (
	"fmt"
	"sync"

	"gitlab.smartbet.am/golang/notification/internal/models"
)

type ProviderRegistry struct {
	emailFactories map[string]func(config map[string]interface{}) (EmailProvider, error)
	smsFactories   map[string]func(config map[string]interface{}) (SMSProvider, error)
	pushFactories  map[string]func(config map[string]interface{}) (PushProvider, error)
	mu             sync.RWMutex
}

func NewProviderRegistry() *ProviderRegistry {
	registry := &ProviderRegistry{
		emailFactories: make(map[string]func(config map[string]interface{}) (EmailProvider, error)),
		smsFactories:   make(map[string]func(config map[string]interface{}) (SMSProvider, error)),
		pushFactories:  make(map[string]func(config map[string]interface{}) (PushProvider, error)),
	}

	// Register default providers
	registry.RegisterEmailProvider("sendgrid", NewSendGridProvider)
	registry.RegisterEmailProvider("sendx", NewSendXProvider)
	registry.RegisterSMSProvider("twilio", NewTwilioProvider)
	registry.RegisterPushProvider("fcm", NewFCMProvider)

	return registry
}

func (r *ProviderRegistry) RegisterEmailProvider(name string, factory func(config map[string]interface{}) (EmailProvider, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.emailFactories[name] = factory
}

func (r *ProviderRegistry) RegisterSMSProvider(name string, factory func(config map[string]interface{}) (SMSProvider, error)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.smsFactories[name] = factory
}

func (r *ProviderRegistry) RegisterPushProvider(name string, factory func(config map[string]interface{}) (PushProvider, error)) {
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

package providers

import (
	"fmt"
	"sync"

	"gitlab.smartbet.am/golang/notification/internal/providers/email"
	"gitlab.smartbet.am/golang/notification/internal/providers/sms"
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

	// Register email providers
	registry.RegisterEmailProvider("smtp", func(config map[string]interface{}) (EmailProvider, error) {
		provider, err := email.NewSMTPProvider(config)
		if err != nil {
			return nil, err
		}
		return provider, nil
	})

	// Register SMS providers
	registry.RegisterSMSProvider("twilio", func(config map[string]interface{}) (SMSProvider, error) {
		provider, err := sms.NewTwilioProvider(config)
		if err != nil {
			return nil, err
		}
		return provider, nil
	})

	// TODO: Register push providers when implemented
	// registry.RegisterPushProvider("fcm", func(config map[string]interface{}) (PushProvider, error) {
	//     provider, err := push.NewFCMProvider(config)
	//     if err != nil {
	//         return nil, err
	//     }
	//     return provider, nil
	// })

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

// Updated methods to accept config and type separately
func (r *ProviderRegistry) CreateEmailProvider(config map[string]interface{}, providerType string) (EmailProvider, error) {
	r.mu.RLock()
	factory, exists := r.emailFactories[providerType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("email provider %s not registered", providerType)
	}

	return factory(config)
}

func (r *ProviderRegistry) CreateSMSProvider(config map[string]interface{}, providerType string) (SMSProvider, error) {
	r.mu.RLock()
	factory, exists := r.smsFactories[providerType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("sms provider %s not registered", providerType)
	}

	return factory(config)
}

func (r *ProviderRegistry) CreatePushProvider(config map[string]interface{}, providerType string) (PushProvider, error) {
	r.mu.RLock()
	factory, exists := r.pushFactories[providerType]
	r.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("push provider %s not registered", providerType)
	}

	return factory(config)
}

// Legacy methods for backward compatibility (if needed elsewhere)
// These can be removed if not used anywhere else
func (r *ProviderRegistry) CreateEmailProviderLegacy(providerConfig interface{}) (EmailProvider, error) {
	// This would need type assertion and conversion logic
	// Only implement if needed for backward compatibility
	return nil, fmt.Errorf("legacy method not implemented")
}

func (r *ProviderRegistry) CreateSMSProviderLegacy(providerConfig interface{}) (SMSProvider, error) {
	// This would need type assertion and conversion logic
	// Only implement if needed for backward compatibility
	return nil, fmt.Errorf("legacy method not implemented")
}

// GetRegisteredEmailProviders returns a list of registered email provider types
func (r *ProviderRegistry) GetRegisteredEmailProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]string, 0, len(r.emailFactories))
	for name := range r.emailFactories {
		providers = append(providers, name)
	}
	return providers
}

// GetRegisteredSMSProviders returns a list of registered SMS provider types
func (r *ProviderRegistry) GetRegisteredSMSProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]string, 0, len(r.smsFactories))
	for name := range r.smsFactories {
		providers = append(providers, name)
	}
	return providers
}

// GetRegisteredPushProviders returns a list of registered push provider types
func (r *ProviderRegistry) GetRegisteredPushProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	providers := make([]string, 0, len(r.pushFactories))
	for name := range r.pushFactories {
		providers = append(providers, name)
	}
	return providers
}

package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go.uber.org/zap"
	"golang.org/x/text/language"
)

type LocalizationService struct {
	bundle     *i18n.Bundle
	localizers map[string]*i18n.Localizer
	mu         sync.RWMutex
	logger     *zap.Logger
}

func NewLocalizationService(logger *zap.Logger) *LocalizationService {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	svc := &LocalizationService{
		bundle:     bundle,
		localizers: make(map[string]*i18n.Localizer),
		logger:     logger,
	}

	// Load translation files
	if err := svc.loadTranslations(); err != nil {
		logger.Error("failed to load translations", zap.Error(err))
	}

	return svc
}

func (s *LocalizationService) loadTranslations() error {
	translationsDir := "./translations"

	return filepath.Walk(translationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(path) == ".json" {
			if _, err := s.bundle.LoadMessageFile(path); err != nil {
				s.logger.Warn("failed to load translation file",
					zap.String("file", path),
					zap.Error(err),
				)
			}
		}

		return nil
	})
}

func (s *LocalizationService) GetLocalizer(locale string) *i18n.Localizer {
	s.mu.RLock()
	localizer, exists := s.localizers[locale]
	s.mu.RUnlock()

	if exists {
		return localizer
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check again after acquiring write lock
	if localizer, exists := s.localizers[locale]; exists {
		return localizer
	}

	// Create new localizer
	localizer = i18n.NewLocalizer(s.bundle, locale)
	s.localizers[locale] = localizer

	return localizer
}

func (s *LocalizationService) LocalizeData(locale string, data map[string]interface{}) map[string]interface{} {
	localizer := s.GetLocalizer(locale)
	localizedData := make(map[string]interface{})

	// Copy original data
	for k, v := range data {
		localizedData[k] = v
	}

	// Add localized strings
	for key, value := range data {
		if strValue, ok := value.(string); ok && strings.HasPrefix(strValue, "i18n:") {
			messageID := strings.TrimPrefix(strValue, "i18n:")
			localized, err := localizer.Localize(&i18n.LocalizeConfig{
				MessageID: messageID,
			})
			if err == nil {
				localizedData[key] = localized
			}
		}
	}

	return localizedData
}

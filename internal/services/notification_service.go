package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.smartbet.am/golang/notification/ent"
	//"gitlab.smartbet.am/golang/notification/ent/notification"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/providers"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"go.uber.org/zap"
)

type NotificationService struct {
	notifRepo       *repository.NotificationRepository
	configRepo      *repository.PartnerConfigRepository
	templateSvc     *TemplateService
	localizationSvc *LocalizationService
	emailManager    *providers.EmailProviderManager
	smsManager      *providers.SMSProviderManager
	pushManager     *providers.PushProviderManager
	logger          *zap.Logger
}

func NewNotificationService(
	notifRepo *repository.NotificationRepository,
	configRepo *repository.PartnerConfigRepository,
	templateSvc *TemplateService,
	localizationSvc *LocalizationService,
	emailManager *providers.EmailProviderManager,
	smsManager *providers.SMSProviderManager,
	pushManager *providers.PushProviderManager,
	logger *zap.Logger,
) *NotificationService {
	return &NotificationService{
		notifRepo:       notifRepo,
		configRepo:      configRepo,
		templateSvc:     templateSvc,
		localizationSvc: localizationSvc,
		emailManager:    emailManager,
		smsManager:      smsManager,
		pushManager:     pushManager,
		logger:          logger,
	}
}

func (s *NotificationService) ProcessNotification(ctx context.Context, req *models.NotificationRequest) error {
	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	// Get partner configuration
	config, err := s.configRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get partner config: %w", err)
	}

	// Process template if template ID is provided
	if req.TemplateID != "" {
		template, err := s.templateSvc.GetTemplate(ctx, req.TenantID, req.TemplateID)
		if err != nil {
			return fmt.Errorf("failed to get template: %w", err)
		}

		// Process template with localization
		body, headline, err := s.templateSvc.ProcessTemplate(ctx, template, req.Data, req.Locale)
		if err != nil {
			return fmt.Errorf("failed to process template: %w", err)
		}

		req.Body = body
		if headline != "" {
			req.Headline = headline
		}
	}

	// Get notifications from database
	notifications, err := s.notifRepo.GetByRequestIDAndStatus(ctx, req.RequestID, notification.StatusPENDING)
	if err != nil {
		return fmt.Errorf("failed to get notifications: %w", err)
	}

	// Check if batch processing is enabled
	if config.BatchConfig.Enabled && len(notifications) > 1 {
		return s.processBatch(ctx, notifications, config)
	}

	// Process individual notifications
	for _, notif := range notifications {
		if err := s.sendNotification(ctx, notif, config); err != nil {
			errorMsg := err.Error()
			s.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusFAILED, &errorMsg)
			s.logger.Error("failed to send notification",
				zap.Int("notification_id", notif.ID),
				zap.Error(err),
			)
		} else {
			s.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusCOMPLETED, nil)
		}
	}

	return nil
}

func (s *NotificationService) ProcessStoredNotification(ctx context.Context, notif *ent.Notification) error {
	// Get partner configuration
	config, err := s.configRepo.GetByTenantID(ctx, notif.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get partner config: %w", err)
	}

	// Send the notification
	if err := s.sendNotification(ctx, notif, config); err != nil {
		errorMsg := err.Error()
		s.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusFAILED, &errorMsg)
		return err
	}

	s.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusCOMPLETED, nil)
	return nil
}

func (s *NotificationService) sendNotification(ctx context.Context, notif *ent.Notification, config *models.PartnerConfig) error {
	var templateData map[string]interface{}
	if notif.Meta != nil && notif.Meta.Params != nil {
		templateData = notif.Meta.Params
	}

	switch notif.Type {
	case notification.TypeEMAIL:
		provider, err := s.emailManager.GetProvider(notif.TenantID)
		if err != nil {
			return err
		}
		return provider.Send(ctx, notif, templateData)

	case notification.TypeSMS:
		provider, err := s.smsManager.GetProvider(notif.TenantID)
		if err != nil {
			return err
		}
		return provider.Send(ctx, notif, templateData)

	case notification.TypePUSH:
		provider, err := s.pushManager.GetProvider(notif.TenantID)
		if err != nil {
			return err
		}
		return provider.Send(ctx, notif, templateData)

	default:
		return fmt.Errorf("unsupported notification type: %s", notif.Type)
	}
}

func (s *NotificationService) processBatch(ctx context.Context, notifications []*ent.Notification, config *models.PartnerConfig) error {
	// Group notifications by type
	grouped := make(map[notification.Type][]*ent.Notification)
	for _, notif := range notifications {
		grouped[notif.Type] = append(grouped[notif.Type], notif)
	}

	// Process each group in batches
	for notifType, group := range grouped {
		batchSize := config.BatchConfig.MaxBatchSize
		if batchSize == 0 {
			batchSize = 100
		}

		for i := 0; i < len(group); i += batchSize {
			end := i + batchSize
			if end > len(group) {
				end = len(group)
			}

			batch := group[i:end]
			if err := s.sendBatch(ctx, batch, notifType, config); err != nil {
				// Update status for failed batch
				for _, notif := range batch {
					errorMsg := err.Error()
					s.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusFAILED, &errorMsg)
				}
			} else {
				// Update status for successful batch
				for _, notif := range batch {
					s.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusCOMPLETED, nil)
				}
			}
		}
	}

	return nil
}

func (s *NotificationService) sendBatch(ctx context.Context, notifications []*ent.Notification, notifType notification.Type, config *models.PartnerConfig) error {
	if len(notifications) == 0 {
		return nil
	}

	tenantID := notifications[0].TenantID
	var templateData map[string]interface{}

	switch notifType {
	case notification.TypeEMAIL:
		provider, err := s.emailManager.GetProvider(tenantID)
		if err != nil {
			return err
		}
		return provider.SendBatch(ctx, notifications, templateData)

	case notification.TypeSMS:
		provider, err := s.smsManager.GetProvider(tenantID)
		if err != nil {
			return err
		}
		return provider.SendBatch(ctx, notifications, templateData)

	case notification.TypePUSH:
		provider, err := s.pushManager.GetProvider(tenantID)
		if err != nil {
			return err
		}
		return provider.SendBatch(ctx, notifications, templateData)

	default:
		return fmt.Errorf("unsupported notification type for batch: %s", notifType)
	}
}

func (s *NotificationService) GetNotification(ctx context.Context, tenantID int64, requestID string) (*ent.Notification, error) {
	notif, err := s.notifRepo.GetByRequestID(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if notif.TenantID != tenantID {
		return nil, fmt.Errorf("notification not found")
	}

	return notif, nil
}

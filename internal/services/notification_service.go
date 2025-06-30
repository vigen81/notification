package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/ent/notification"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/providers"
	"gitlab.smartbet.am/golang/notification/internal/repository"
)

type NotificationService struct {
	notifRepo    *repository.NotificationRepository
	configRepo   *repository.PartnerConfigRepository
	emailManager *providers.EmailProviderManager
	smsManager   *providers.SMSProviderManager
	logger       *logrus.Logger
}

func NewNotificationService(
	notifRepo *repository.NotificationRepository,
	configRepo *repository.PartnerConfigRepository,
	emailManager *providers.EmailProviderManager,
	smsManager *providers.SMSProviderManager,
	logger *logrus.Logger,
) *NotificationService {
	return &NotificationService{
		notifRepo:    notifRepo,
		configRepo:   configRepo,
		emailManager: emailManager,
		smsManager:   smsManager,
		logger:       logger,
	}
}

// ProcessNotification processes a notification request
func (s *NotificationService) ProcessNotification(ctx context.Context, req *models.NotificationRequest) error {
	log := logger.WithRequest(req.RequestID)

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	log.Info("Processing notification request", map[string]interface{}{
		"tenant_id":  req.TenantID,
		"type":       req.Type,
		"recipients": len(req.Recipients),
		"batch_id":   req.BatchID,
		"scheduled":  req.ScheduleTS != nil,
	})

	// Get partner configuration
	config, err := s.configRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		log.Error("Failed to get partner config", err, map[string]interface{}{
			"tenant_id": req.TenantID,
		})
		return fmt.Errorf("failed to get partner config: %w", err)
	}

	// ALWAYS store notifications in database first (whether immediate or scheduled)
	var notifications []*ent.Notification
	for _, recipient := range req.Recipients {
		notif, err := s.notifRepo.Create(ctx, req, recipient)
		if err != nil {
			log.Error("Failed to store notification in database", err, map[string]interface{}{
				"recipient": recipient,
				"tenant_id": req.TenantID,
			})
			continue // Continue with other recipients
		}
		notifications = append(notifications, notif)
	}

	if len(notifications) == 0 {
		return fmt.Errorf("failed to store any notifications in database")
	}

	// Check if scheduled for future - if so, just return (scheduler will handle)
	if req.ScheduleTS != nil && *req.ScheduleTS > time.Now().Unix() {
		log.Info("Notifications stored and scheduled for future processing", map[string]interface{}{
			"schedule_ts": *req.ScheduleTS,
			"count":       len(notifications),
		})
		return nil // Scheduler worker will handle it
	}

	// Process notifications immediately
	if config.BatchConfig.Enabled && len(notifications) > 1 {
		return s.processBatch(ctx, notifications, config, req.MessageType)
	}

	// Process individual notifications
	for _, notif := range notifications {
		if err := s.sendNotification(ctx, notif, config, req.MessageType); err != nil {
			s.updateNotificationStatus(ctx, notif.ID, notification.StatusFAILED, err.Error())
			log.Error("Failed to send notification", err, map[string]interface{}{
				"notification_id": notif.ID,
				"recipient":       string(notif.Address),
			})
		} else {
			s.updateNotificationStatus(ctx, notif.ID, notification.StatusCOMPLETED, "")
			log.Info("Notification sent successfully", map[string]interface{}{
				"notification_id": notif.ID,
				"recipient":       string(notif.Address),
			})
		}
	}

	return nil
}

// ProcessStoredNotification processes a notification that's already stored in database
func (s *NotificationService) ProcessStoredNotification(ctx context.Context, notif *ent.Notification) error {
	log := logger.WithRequest(notif.RequestID)

	// Get partner configuration
	config, err := s.configRepo.GetByTenantID(ctx, notif.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get partner config: %w", err)
	}

	// Determine message type from notification meta or default to system
	messageType := models.MessageTypeSystem
	if notif.Meta != nil && notif.Meta.Params != nil {
		if mt, exists := notif.Meta.Params["message_type"]; exists {
			if mtStr, ok := mt.(string); ok {
				messageType = models.MessageType(mtStr)
			}
		}
	}

	// Send the notification
	if err := s.sendNotification(ctx, notif, config, messageType); err != nil {
		s.updateNotificationStatus(ctx, notif.ID, notification.StatusFAILED, err.Error())
		log.Error("Failed to send stored notification", err, map[string]interface{}{
			"notification_id": notif.ID,
		})
		return err
	}

	s.updateNotificationStatus(ctx, notif.ID, notification.StatusCOMPLETED, "")
	log.Info("Stored notification sent successfully", map[string]interface{}{
		"notification_id": notif.ID,
	})

	return nil
}

// sendNotification sends a single notification using the appropriate provider
func (s *NotificationService) sendNotification(ctx context.Context, notif *ent.Notification, config *models.PartnerConfig, messageType models.MessageType) error {
	switch notif.Type {
	case notification.TypeEMAIL:
		provider, err := s.emailManager.GetProvider(notif.TenantID)
		if err != nil {
			return fmt.Errorf("failed to get email provider: %w", err)
		}
		return provider.Send(ctx, notif, messageType)

	case notification.TypeSMS:
		provider, err := s.smsManager.GetProvider(notif.TenantID)
		if err != nil {
			return fmt.Errorf("failed to get SMS provider: %w", err)
		}
		return provider.Send(ctx, notif, messageType)

	case notification.TypePUSH:
		// TODO: Implement push provider when available
		return fmt.Errorf("push notifications not yet implemented")

	default:
		return fmt.Errorf("unsupported notification type: %s", notif.Type)
	}
}

// processBatch processes multiple notifications as a batch
func (s *NotificationService) processBatch(ctx context.Context, notifications []*ent.Notification, config *models.PartnerConfig, messageType models.MessageType) error {
	log := logger.To("batch_processor")

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

		log.Info("Processing batch", map[string]interface{}{
			"type":       notifType,
			"total":      len(group),
			"batch_size": batchSize,
		})

		for i := 0; i < len(group); i += batchSize {
			end := i + batchSize
			if end > len(group) {
				end = len(group)
			}

			batch := group[i:end]
			if err := s.sendBatch(ctx, batch, notifType, messageType); err != nil {
				// Update status for failed batch
				for _, notif := range batch {
					s.updateNotificationStatus(ctx, notif.ID, notification.StatusFAILED, err.Error())
				}
				log.Error("Failed to send batch", err, map[string]interface{}{
					"batch_size": len(batch),
					"type":       notifType,
				})
			} else {
				// Update status for successful batch
				for _, notif := range batch {
					s.updateNotificationStatus(ctx, notif.ID, notification.StatusCOMPLETED, "")
				}
				log.Info("Batch sent successfully", map[string]interface{}{
					"batch_size": len(batch),
					"type":       notifType,
				})
			}
		}
	}

	return nil
}

// sendBatch sends a batch of notifications
func (s *NotificationService) sendBatch(ctx context.Context, notifications []*ent.Notification, notifType notification.Type, messageType models.MessageType) error {
	if len(notifications) == 0 {
		return nil
	}

	tenantID := notifications[0].TenantID

	switch notifType {
	case notification.TypeEMAIL:
		provider, err := s.emailManager.GetProvider(tenantID)
		if err != nil {
			return fmt.Errorf("failed to get email provider: %w", err)
		}
		return provider.SendBatch(ctx, notifications, messageType)

	case notification.TypeSMS:
		provider, err := s.smsManager.GetProvider(tenantID)
		if err != nil {
			return fmt.Errorf("failed to get SMS provider: %w", err)
		}
		return provider.SendBatch(ctx, notifications, messageType)

	case notification.TypePUSH:
		// TODO: Implement push provider batch sending
		return fmt.Errorf("push notification batching not yet implemented")

	default:
		return fmt.Errorf("unsupported notification type for batch: %s", notifType)
	}
}

// updateNotificationStatus updates the status of a notification
func (s *NotificationService) updateNotificationStatus(ctx context.Context, notificationID int, status notification.Status, errorMsg string) {
	var errorMsgPtr *string
	if errorMsg != "" {
		errorMsgPtr = &errorMsg
	}

	if err := s.notifRepo.UpdateStatus(ctx, notificationID, status, errorMsgPtr); err != nil {
		s.logger.WithField("notification_id", notificationID).
			WithError(err).
			Error("Failed to update notification status")
	}
}

// GetNotification retrieves a notification by request ID for a specific tenant
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

package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent/schema"
	"gitlab.smartbet.am/golang/notification/internal/buffer"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
)

type BufferedNotificationService struct {
	notificationSvc *NotificationService
	configRepo      *repository.PartnerConfigRepository
	buffers         map[string]*buffer.Service[*models.NotificationRequest]
	mu              sync.RWMutex
	logger          *logrus.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewBufferedNotificationService(
	notificationSvc *NotificationService,
	configRepo *repository.PartnerConfigRepository,
	logger *logrus.Logger,
) *BufferedNotificationService {
	ctx, cancel := context.WithCancel(context.Background())

	service := &BufferedNotificationService{
		notificationSvc: notificationSvc,
		configRepo:      configRepo,
		buffers:         make(map[string]*buffer.Service[*models.NotificationRequest]),
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}

	return service
}

func (bns *BufferedNotificationService) getBufferKey(tenantID int64, notifType models.NotificationType) string {
	return fmt.Sprintf("tenant-%d-%s", tenantID, notifType)
}

func (bns *BufferedNotificationService) getOrCreateBuffer(tenantID int64, notifType models.NotificationType) *buffer.Service[*models.NotificationRequest] {
	bufferKey := bns.getBufferKey(tenantID, notifType)

	bns.mu.RLock()
	buf, exists := bns.buffers[bufferKey]
	bns.mu.RUnlock()

	if exists {
		return buf
	}

	bns.mu.Lock()
	defer bns.mu.Unlock()

	// Double-check after acquiring write lock
	if buf, exists = bns.buffers[bufferKey]; exists {
		return buf
	}

	// Get tenant config
	config, err := bns.configRepo.GetByTenantID(context.Background(), tenantID)
	if err != nil {
		bns.logger.WithError(err).WithField("tenant_id", tenantID).Warn("Failed to get tenant config, using defaults")
		// Create default config if tenant config doesn't exist
		config = bns.createDefaultConfig(tenantID)
	}

	// Use existing BatchConfig or create default
	batchConfig := config.BatchConfig
	if batchConfig == nil {
		batchConfig = &schema.BatchConfig{
			Enabled:              true,
			MaxBatchSize:         50,
			FlushIntervalSeconds: 30,
		}
	}

	maxSize := batchConfig.MaxBatchSize
	if maxSize <= 0 {
		maxSize = 50
	}

	flushPeriod := time.Duration(batchConfig.FlushIntervalSeconds) * time.Second
	if flushPeriod <= 0 {
		flushPeriod = 30 * time.Second
	}

	// Create buffer
	buf = buffer.NewService[*models.NotificationRequest](maxSize, flushPeriod)

	// Start processor for this buffer
	go bns.startBufferProcessor(buf, tenantID, notifType)

	bns.buffers[bufferKey] = buf
	return buf
}

func (bns *BufferedNotificationService) createDefaultConfig(tenantID int64) *models.PartnerConfig {
	return &models.PartnerConfig{
		TenantID: tenantID,
		BatchConfig: &schema.BatchConfig{
			Enabled:              true,
			MaxBatchSize:         50,
			FlushIntervalSeconds: 30,
		},
		EmailProviders: []schema.ProviderConfig{},
		SMSProviders:   []schema.ProviderConfig{},
		PushProviders:  []schema.ProviderConfig{},
		RateLimits:     map[string]schema.RateLimit{},
		Enabled:        true,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func (bns *BufferedNotificationService) startBufferProcessor(
	buf *buffer.Service[*models.NotificationRequest],
	tenantID int64,
	notifType models.NotificationType,
) {
	log := logger.WithTenant(tenantID).WithField("type", notifType)
	log.Info("Starting buffer processor")

	for {
		select {
		case <-bns.ctx.Done():
			log.Info("Buffer processor stopping")
			return
		case <-buf.Ready():
			notifications := buf.Pop()
			if len(notifications) > 0 {
				log.WithField("batch_size", len(notifications)).Info("Processing buffered notifications")
				bns.processBatch(notifications)
			}
		}
	}
}

func (bns *BufferedNotificationService) processBatch(requests []*models.NotificationRequest) {
	if len(requests) == 0 {
		return
	}

	// Group by tenant for processing
	tenantGroups := make(map[int64][]*models.NotificationRequest)
	for _, req := range requests {
		tenantGroups[req.TenantID] = append(tenantGroups[req.TenantID], req)
	}

	ctx := context.Background()

	for tenantID, tenantRequests := range tenantGroups {
		log := logger.WithTenant(tenantID).WithField("batch_size", len(tenantRequests))

		for _, req := range tenantRequests {
			if err := bns.notificationSvc.ProcessNotification(ctx, req); err != nil {
				log.WithError(err).WithField("request_id", req.RequestID).Error("Failed to process notification in batch")
			}
		}
	}
}

// ProcessNotification decides whether to buffer or process immediately
func (bns *BufferedNotificationService) ProcessNotification(ctx context.Context, req *models.NotificationRequest) error {
	// For scheduled notifications, process immediately (they'll be stored and scheduled)
	if req.ScheduleTS != nil && *req.ScheduleTS > time.Now().Unix() {
		return bns.notificationSvc.ProcessNotification(ctx, req)
	}

	// Check if batching is enabled for this tenant
	config, err := bns.configRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		// If can't get config, process immediately
		bns.logger.WithError(err).WithField("tenant_id", req.TenantID).Warn("Failed to get tenant config, processing immediately")
		return bns.notificationSvc.ProcessNotification(ctx, req)
	}

	// Check if batching is enabled
	if config.BatchConfig == nil || !config.BatchConfig.Enabled {
		// Batching disabled, process immediately
		return bns.notificationSvc.ProcessNotification(ctx, req)
	}

	// Add to buffer
	buf := bns.getOrCreateBuffer(req.TenantID, req.Type)
	if !buf.Push(req) {
		// Buffer full, process immediately
		bns.logger.WithFields(logrus.Fields{
			"tenant_id":  req.TenantID,
			"type":       req.Type,
			"request_id": req.RequestID,
		}).Warn("Buffer full, processing immediately")
		return bns.notificationSvc.ProcessNotification(ctx, req)
	}

	bns.logger.WithFields(logrus.Fields{
		"tenant_id":   req.TenantID,
		"type":        req.Type,
		"request_id":  req.RequestID,
		"buffer_size": buf.Size(),
	}).Debug("Added notification to buffer")

	return nil
}

func (bns *BufferedNotificationService) Start() error {
	bns.logger.Info("Buffered notification service started")
	return nil
}

func (bns *BufferedNotificationService) Stop() {
	bns.logger.Info("Stopping buffered notification service")
	bns.cancel()

	bns.mu.Lock()
	defer bns.mu.Unlock()

	for key, buf := range bns.buffers {
		bns.logger.WithField("buffer", key).Info("Closing buffer")
		buf.Close()
	}
	bns.buffers = make(map[string]*buffer.Service[*models.NotificationRequest])
}

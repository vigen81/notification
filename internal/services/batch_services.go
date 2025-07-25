package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
)

type BatchService struct {
	notificationSvc *NotificationService
	configRepo      *repository.PartnerConfigRepository
	batches         map[string]*Batch
	mu              sync.RWMutex
	logger          *logrus.Logger
}

type Batch struct {
	ID          string
	TenantID    int64
	Requests    []*models.NotificationRequest
	CreatedAt   time.Time
	LastAddedAt time.Time
}

func NewBatchService(
	notificationSvc *NotificationService,
	configRepo *repository.PartnerConfigRepository,
	logger *logrus.Logger,
) *BatchService {
	svc := &BatchService{
		notificationSvc: notificationSvc,
		configRepo:      configRepo,
		batches:         make(map[string]*Batch),
		logger:          logger,
	}

	// Start batch processor
	go svc.processBatches()

	return svc
}

func (s *BatchService) ProcessBatch(ctx context.Context, req *models.NotificationRequest) error {
	return s.notificationSvc.ProcessNotification(ctx, req)
}

func (s *BatchService) AddToBatch(ctx context.Context, req *models.NotificationRequest) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	batchKey := fmt.Sprintf("%d-%s", req.TenantID, req.Type)

	batch, exists := s.batches[batchKey]
	if !exists {
		batch = &Batch{
			ID:        batchKey,
			TenantID:  req.TenantID,
			Requests:  []*models.NotificationRequest{},
			CreatedAt: time.Now(),
		}
		s.batches[batchKey] = batch
	}

	batch.Requests = append(batch.Requests, req)
	batch.LastAddedAt = time.Now()

	// Check if batch should be flushed
	config, err := s.configRepo.GetByTenantID(ctx, req.TenantID)
	if err != nil {
		return err
	}

	if len(batch.Requests) >= config.BatchConfig.MaxBatchSize {
		return s.flushBatch(ctx, batchKey)
	}

	return nil
}

func (s *BatchService) flushBatch(ctx context.Context, batchKey string) error {
	batch, exists := s.batches[batchKey]
	if !exists {
		return nil
	}

	// Process all requests in the batch
	for _, req := range batch.Requests {
		if err := s.notificationSvc.ProcessNotification(ctx, req); err != nil {
			s.logger.WithFields(logrus.Fields{
				"batch_id":   batch.ID,
				"request_id": req.RequestID,
			}).WithError(err).Error("Failed to process batch notification")
		}
	}

	// Remove batch
	delete(s.batches, batchKey)
	return nil
}

func (s *BatchService) processBatches() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()

		for batchKey, batch := range s.batches {
			// Flush batches older than configured interval
			if now.Sub(batch.LastAddedAt) > 30*time.Second {
				err := s.flushBatch(context.Background(), batchKey)
				if err != nil {
					return
				}
			}
		}

		s.mu.Unlock()
	}
}

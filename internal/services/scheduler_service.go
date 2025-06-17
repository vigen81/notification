package services

import (
	"context"
	"time"

	"gitlab.smartbet.am/golang/notification/internal/repository"
	"go.uber.org/zap"
)

type SchedulerService struct {
	notifRepo *repository.NotificationRepository
	logger    *zap.Logger
}

func NewSchedulerService(
	notifRepo *repository.NotificationRepository,
	logger *zap.Logger,
) *SchedulerService {
	return &SchedulerService{
		notifRepo: notifRepo,
		logger:    logger,
	}
}

func (s *SchedulerService) GetDueNotifications(ctx context.Context) ([]int, error) {
	timestamp := time.Now().Unix()
	notifications, err := s.notifRepo.GetPendingScheduled(ctx, timestamp)
	if err != nil {
		return nil, err
	}

	ids := make([]int, len(notifications))
	for i, n := range notifications {
		ids[i] = n.ID
	}

	return ids, nil
}

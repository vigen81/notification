package workers

import (
	"context"
	"time"

	"gitlab.smartbet.am/golang/notification/ent/notification"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/services"
	"go.uber.org/zap"
)

type SchedulerWorker struct {
	notifRepo       *repository.NotificationRepository
	notificationSvc *services.NotificationService
	logger          *zap.Logger
	ticker          *time.Ticker
	stopChan        chan struct{}
}

func NewSchedulerWorker(
	notifRepo *repository.NotificationRepository,
	notificationSvc *services.NotificationService,
	logger *zap.Logger,
) *SchedulerWorker {
	return &SchedulerWorker{
		notifRepo:       notifRepo,
		notificationSvc: notificationSvc,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

func (w *SchedulerWorker) Start(ctx context.Context) error {
	w.ticker = time.NewTicker(30 * time.Second)

	go w.run(ctx)

	w.logger.Info("scheduler worker started")
	return nil
}

func (w *SchedulerWorker) Stop() {
	if w.ticker != nil {
		w.ticker.Stop()
	}
	close(w.stopChan)
}

func (w *SchedulerWorker) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("scheduler worker stopping due to context cancellation")
			return
		case <-w.stopChan:
			w.logger.Info("scheduler worker stopping")
			return
		case <-w.ticker.C:
			w.processScheduledNotifications(ctx)
		}
	}
}

func (w *SchedulerWorker) processScheduledNotifications(ctx context.Context) {
	now := time.Now().Unix()

	// Get all pending scheduled notifications that are due
	notifications, err := w.notifRepo.GetPendingScheduled(ctx, now)
	if err != nil {
		w.logger.Error("failed to get scheduled notifications", zap.Error(err))
		return
	}

	if len(notifications) == 0 {
		return
	}

	w.logger.Info("processing scheduled notifications", zap.Int("count", len(notifications)))

	for _, notif := range notifications {
		// Update status to ACTIVE to prevent duplicate processing
		if err := w.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusACTIVE, nil); err != nil {
			w.logger.Error("failed to update notification status",
				zap.Int("notification_id", notif.ID),
				zap.Error(err),
			)
			continue
		}

		// Process the notification
		if err := w.notificationSvc.ProcessStoredNotification(ctx, notif); err != nil {
			errorMsg := err.Error()
			w.notifRepo.UpdateStatus(ctx, notif.ID, notification.StatusFAILED, &errorMsg)
			w.logger.Error("failed to process scheduled notification",
				zap.Int("notification_id", notif.ID),
				zap.Error(err),
			)
		}
	}
}

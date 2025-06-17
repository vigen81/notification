package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/services"
	"go.uber.org/zap"
)

type NotificationWorker struct {
	subscriber      *kafka.Subscriber
	notifRepo       *repository.NotificationRepository
	notificationSvc *services.NotificationService
	batchSvc        *services.BatchService
	logger          *zap.Logger
	stopChan        chan struct{}
}

func NewNotificationWorker(
	subscriber *kafka.Subscriber,
	notifRepo *repository.NotificationRepository,
	notificationSvc *services.NotificationService,
	batchSvc *services.BatchService,
	logger *zap.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		subscriber:      subscriber,
		notifRepo:       notifRepo,
		notificationSvc: notificationSvc,
		batchSvc:        batchSvc,
		logger:          logger,
		stopChan:        make(chan struct{}),
	}
}

func (w *NotificationWorker) Start(ctx context.Context) error {
	messages, err := w.subscriber.Subscribe(ctx, "notifications")
	if err != nil {
		return fmt.Errorf("failed to subscribe to notifications topic: %w", err)
	}

	go w.processMessages(ctx, messages)

	w.logger.Info("notification worker started")
	return nil
}

func (w *NotificationWorker) Stop() {
	close(w.stopChan)
}

func (w *NotificationWorker) processMessages(ctx context.Context, messages <-chan *message.Message) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("notification worker stopping due to context cancellation")
			return
		case <-w.stopChan:
			w.logger.Info("notification worker stopping")
			return
		case msg := <-messages:
			startTime := time.Now()

			if err := w.processMessage(ctx, msg); err != nil {
				w.logger.Error("failed to process message",
					zap.String("message_id", msg.UUID),
					zap.Error(err),
					zap.Duration("duration", time.Since(startTime)),
				)
				msg.Nack()
			} else {
				w.logger.Info("message processed successfully",
					zap.String("message_id", msg.UUID),
					zap.Duration("duration", time.Since(startTime)),
				)
				msg.Ack()
			}
		}
	}
}

func (w *NotificationWorker) processMessage(ctx context.Context, msg *message.Message) error {
	var req models.NotificationRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal notification request: %w", err)
	}

	// First, store notifications in database
	for _, recipient := range req.Recipients {
		if _, err := w.notifRepo.Create(ctx, &req, recipient); err != nil {
			w.logger.Error("failed to store notification in database",
				zap.String("request_id", req.RequestID),
				zap.String("recipient", recipient),
				zap.Error(err),
			)
			// Continue with other recipients even if one fails
		}
	}

	// Check if it's scheduled for future
	if req.ScheduleTS != nil && *req.ScheduleTS > time.Now().Unix() {
		// Skip processing for now, scheduler worker will handle it
		w.logger.Info("notification scheduled for future",
			zap.String("request_id", req.RequestID),
			zap.Int64("schedule_ts", *req.ScheduleTS),
		)
		return nil
	}

	// Process notification immediately
	return w.notificationSvc.ProcessNotification(ctx, &req)
}

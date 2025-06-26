package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/services"
)

type NotificationWorker struct {
	subscriber      *kafka.Subscriber
	notifRepo       *repository.NotificationRepository
	notificationSvc *services.NotificationService
	batchSvc        *services.BatchService
	logger          *logrus.Logger
	stopChan        chan struct{}
}

func NewNotificationWorker(
	subscriber *kafka.Subscriber,
	notifRepo *repository.NotificationRepository,
	notificationSvc *services.NotificationService,
	batchSvc *services.BatchService,
	logger *logrus.Logger,
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

	w.logger.Info("Notification worker started")
	return nil
}

func (w *NotificationWorker) Stop() {
	close(w.stopChan)
}

func (w *NotificationWorker) processMessages(ctx context.Context, messages <-chan *message.Message) {
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Notification worker stopping due to context cancellation")
			return
		case <-w.stopChan:
			w.logger.Info("Notification worker stopping")
			return
		case msg := <-messages:
			startTime := time.Now()

			if err := w.processMessage(ctx, msg); err != nil {
				w.logger.WithFields(logrus.Fields{
					"message_id": msg.UUID,
					"duration":   time.Since(startTime),
				}).WithError(err).Error("Failed to process message")
				msg.Nack()
			} else {
				w.logger.WithFields(logrus.Fields{
					"message_id": msg.UUID,
					"duration":   time.Since(startTime),
				}).Info("Message processed successfully")
				msg.Ack()
			}
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (w *NotificationWorker) processMessage(ctx context.Context, msg *message.Message) error {
	var req models.NotificationRequest
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		return fmt.Errorf("failed to unmarshal notification request: %w", err)
	}

	w.logger.WithFields(logrus.Fields{
		"request_id": req.RequestID,
		"tenant_id":  req.TenantID,
		"type":       req.Type,
		"recipients": len(req.Recipients),
	}).Info("Processing notification request")

	// Check if it's scheduled for future
	if req.ScheduleTS != nil && *req.ScheduleTS > time.Now().Unix() {
		// Skip processing for now, scheduler worker will handle it
		w.logger.WithFields(logrus.Fields{
			"request_id":  req.RequestID,
			"schedule_ts": *req.ScheduleTS,
		}).Info("Notification scheduled for future")
		return nil
	}

	// Process notification immediately
	return w.notificationSvc.ProcessNotification(ctx, &req)
}

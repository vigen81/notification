package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/services"
)

type NotificationWorker struct {
	subscriber  *kafka.Subscriber
	notifRepo   *repository.NotificationRepository
	bufferedSvc *services.BufferedNotificationService // Use buffered service
	logger      *logrus.Logger
	stopChan    chan struct{}
}

func NewNotificationWorker(
	subscriber *kafka.Subscriber,
	notifRepo *repository.NotificationRepository,
	bufferedSvc *services.BufferedNotificationService,
	logger *logrus.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		subscriber:  subscriber,
		notifRepo:   notifRepo,
		bufferedSvc: bufferedSvc,
		logger:      logger,
		stopChan:    make(chan struct{}),
	}
}

func (w *NotificationWorker) Start(ctx context.Context) error {
	messages, err := w.subscriber.Subscribe(ctx, "notifications")
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") ||
			strings.Contains(err.Error(), "unknown topic") ||
			strings.Contains(err.Error(), "UNKNOWN_TOPIC_OR_PARTITION") {
			return fmt.Errorf("topic 'notifications' does not exist - please create it first: %w", err)
		}
		return fmt.Errorf("failed to subscribe to notifications topic: %w", err)
	}

	if err := w.bufferedSvc.Start(); err != nil {
		return fmt.Errorf("failed to start buffered service: %w", err)
	}

	go w.processMessages(ctx, messages)

	w.logger.Info("Notification worker started with buffering")
	return nil
}

func (w *NotificationWorker) Stop() {
	w.bufferedSvc.Stop()
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
				if strings.Contains(err.Error(), "failed to get partner config") ||
					strings.Contains(err.Error(), "provider") ||
					strings.Contains(err.Error(), "timeout") ||
					strings.Contains(err.Error(), "connection") {
					w.logger.WithFields(logrus.Fields{
						"message_id": msg.UUID,
						"duration":   time.Since(startTime),
					}).WithError(err).Error("Failed to process message - will retry")
					msg.Nack()
				} else {
					w.logger.WithFields(logrus.Fields{
						"message_id": msg.UUID,
						"duration":   time.Since(startTime),
					}).WithError(err).Error("Failed to process message - unrecoverable error")
					msg.Ack()
				}
			} else {
				w.logger.WithFields(logrus.Fields{
					"message_id": msg.UUID,
					"duration":   time.Since(startTime),
				}).Debug("Message processed successfully")
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
		w.logger.WithFields(logrus.Fields{
			"message_id": msg.UUID,
			"error":      err.Error(),
			"payload":    string(msg.Payload),
		}).Error("Failed to unmarshal message - skipping")
		return nil
	}

	if req.TenantID == 0 || req.Type == "" || len(req.Recipients) == 0 || req.Body == "" {
		w.logger.WithFields(logrus.Fields{
			"message_id": msg.UUID,
			"tenant_id":  req.TenantID,
			"type":       req.Type,
			"recipients": len(req.Recipients),
		}).Error("Invalid notification request - skipping")
		return nil
	}

	// Simply pass to buffered service - it will decide whether to buffer or process immediately
	return w.bufferedSvc.ProcessNotification(ctx, &req)
}

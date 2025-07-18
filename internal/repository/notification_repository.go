// File: internal/repository/notification_repository.go

package repository

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/ent/notification"
	"gitlab.smartbet.am/golang/notification/ent/schema"
	"gitlab.smartbet.am/golang/notification/internal/models"
	"gitlab.smartbet.am/golang/notification/types"
)

type NotificationRepository struct {
	client *ent.Client
	logger *logrus.Logger
}

func NewNotificationRepository(client *ent.Client, logger *logrus.Logger) *NotificationRepository {
	return &NotificationRepository{
		client: client,
		logger: logger,
	}
}

func (r *NotificationRepository) Create(ctx context.Context, req *models.NotificationRequest, address string) (*ent.Notification, error) {
	// Generate a unique request_id for each individual notification record
	// Even in batch processing, each recipient gets their own database record with unique request_id
	uniqueRequestID := uuid.New().String()

	create := r.client.Notification.Create().
		SetRequestID(uniqueRequestID). // Use unique ID for each record
		SetTenantID(req.TenantID).
		SetType(notification.Type(req.Type)).
		SetBody(req.Body).
		SetAddress(types.Address(address)).
		SetStatus(notification.StatusPENDING)

	if req.Headline != "" {
		create.SetHeadline(req.Headline)
	}
	if req.From != "" {
		create.SetFrom(req.From)
	}
	if req.ReplyTo != "" {
		create.SetReplyTo(req.ReplyTo)
	}
	if req.Tag != "" {
		create.SetTag(req.Tag)
	}
	if req.ScheduleTS != nil {
		create.SetScheduleTs(*req.ScheduleTS)
	}
	if req.BatchID != "" {
		create.SetBatchID(req.BatchID)
	}
	if req.Meta != nil {
		// Use the schema structs
		meta := &schema.NotificationMeta{
			Service:    req.Meta.Service,
			TemplateID: req.Meta.TemplateID,
			Params:     req.Meta.Params,
			Data:       req.Meta.Data,
		}
		if req.Meta.Attachment != nil {
			meta.Attachment = &schema.Attachment{
				Filename:    req.Meta.Attachment.Filename,
				Content:     req.Meta.Attachment.Content,
				Disposition: req.Meta.Attachment.Disposition,
				Type:        req.Meta.Attachment.Type,
			}
		}

		// Add the original request_id to meta for tracking
		if meta.Params == nil {
			meta.Params = make(map[string]interface{})
		}
		meta.Params["original_request_id"] = req.RequestID
		meta.Params["message_type"] = string(req.MessageType)

		create.SetMeta(meta)
	} else {
		// Create meta with original request_id even if no other meta exists
		meta := &schema.NotificationMeta{
			Params: map[string]interface{}{
				"original_request_id": req.RequestID,
				"message_type":        string(req.MessageType),
			},
		}
		create.SetMeta(meta)
	}

	return create.Save(ctx)
}

func (r *NotificationRepository) CreateBatch(ctx context.Context, req *models.NotificationRequest) ([]*ent.Notification, error) {
	if len(req.Recipients) == 0 {
		return nil, fmt.Errorf("no recipients provided")
	}

	// Build the base meta once
	var baseMeta *schema.NotificationMeta
	if req.Meta != nil {
		baseMeta = &schema.NotificationMeta{
			Service:    req.Meta.Service,
			TemplateID: req.Meta.TemplateID,
			Params:     req.Meta.Params,
			Data:       req.Meta.Data,
		}
		if req.Meta.Attachment != nil {
			baseMeta.Attachment = &schema.Attachment{
				Filename:    req.Meta.Attachment.Filename,
				Content:     req.Meta.Attachment.Content,
				Disposition: req.Meta.Attachment.Disposition,
				Type:        req.Meta.Attachment.Type,
			}
		}
	} else {
		baseMeta = &schema.NotificationMeta{}
	}

	if baseMeta.Params == nil {
		baseMeta.Params = make(map[string]interface{})
	}
	baseMeta.Params["original_request_id"] = req.RequestID
	baseMeta.Params["message_type"] = string(req.MessageType)

	builders := make([]*ent.NotificationCreate, 0, len(req.Recipients))

	for _, recipient := range req.Recipients {
		uniqueRequestID := uuid.New().String()

		create := r.client.Notification.Create().
			SetRequestID(uniqueRequestID).
			SetTenantID(req.TenantID).
			SetType(notification.Type(req.Type)).
			SetBody(req.Body).
			SetAddress(types.Address(recipient)).
			SetStatus(notification.StatusPENDING).
			SetMeta(baseMeta)

		if req.Headline != "" {
			create.SetHeadline(req.Headline)
		}
		if req.From != "" {
			create.SetFrom(req.From)
		}
		if req.ReplyTo != "" {
			create.SetReplyTo(req.ReplyTo)
		}
		if req.Tag != "" {
			create.SetTag(req.Tag)
		}
		if req.ScheduleTS != nil {
			create.SetScheduleTs(*req.ScheduleTS)
		}
		if req.BatchID != "" {
			create.SetBatchID(req.BatchID)
		}

		builders = append(builders, create)
	}

	notifications, err := r.client.Notification.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to bulk create notifications: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"tenant_id": req.TenantID,
		"count":     len(notifications),
		"batch_id":  req.BatchID,
	}).Debug("Batch created notifications")

	return notifications, nil
}

func (r *NotificationRepository) GetByRequestID(ctx context.Context, requestID string) (*ent.Notification, error) {
	return r.client.Notification.Query().
		Where(notification.RequestID(requestID)).
		First(ctx)
}

// Add method to get notifications by original request ID (for batch tracking)
func (r *NotificationRepository) GetByOriginalRequestID(ctx context.Context, originalRequestID string) ([]*ent.Notification, error) {
	// Use raw SQL to query JSON field - this works with MySQL JSON functions
	return r.client.Notification.Query().
		Where(func(s *sql.Selector) {
			s.Where(sql.P(func(b *sql.Builder) {
				b.WriteString("JSON_EXTRACT(meta, '$.params.original_request_id') = ?")
				b.Arg(originalRequestID)
			}))
		}).
		All(ctx)
}

// Alternative method using batch_id if original_request_id lookup fails
func (r *NotificationRepository) GetByBatchIDOrRequestID(ctx context.Context, id string) ([]*ent.Notification, error) {
	// First try batch_id
	notifications, err := r.client.Notification.Query().
		Where(notification.BatchID(id)).
		All(ctx)

	if err == nil && len(notifications) > 0 {
		return notifications, nil
	}

	// If not found by batch_id, try original_request_id in meta
	return r.GetByOriginalRequestID(ctx, id)
}

func (r *NotificationRepository) GetByRequestIDAndStatus(ctx context.Context, requestID string, status notification.Status) ([]*ent.Notification, error) {
	return r.client.Notification.Query().
		Where(
			notification.RequestID(requestID),
			notification.StatusEQ(status),
		).
		All(ctx)
}

func (r *NotificationRepository) GetPendingScheduled(ctx context.Context, timestamp int64) ([]*ent.Notification, error) {
	return r.client.Notification.Query().
		Where(
			notification.StatusEQ(notification.StatusPENDING),
			notification.ScheduleTsLTE(timestamp),
			notification.ScheduleTsNotNil(),
		).
		All(ctx)
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id int, status notification.Status, errorMsg *string) error {
	update := r.client.Notification.UpdateOneID(id).
		SetStatus(status)

	if errorMsg != nil {
		update.SetErrorMessage(*errorMsg)
	}

	return update.Exec(ctx)
}

func (r *NotificationRepository) GetByTenantAndStatus(ctx context.Context, tenantID int64, status notification.Status, limit int) ([]*ent.Notification, error) {
	return r.client.Notification.Query().
		Where(
			notification.TenantID(tenantID),
			notification.StatusEQ(status),
		).
		Limit(limit).
		All(ctx)
}

// Add method to get batch notifications by batch_id
func (r *NotificationRepository) GetByBatchID(ctx context.Context, batchID string) ([]*ent.Notification, error) {
	return r.client.Notification.Query().
		Where(notification.BatchID(batchID)).
		All(ctx)
}

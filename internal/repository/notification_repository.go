package repository

import (
	"context"

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
	create := r.client.Notification.Create().
		SetRequestID(req.RequestID).
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
		create.SetMeta(meta)
	}

	return create.Save(ctx)
}

func (r *NotificationRepository) GetByRequestID(ctx context.Context, requestID string) (*ent.Notification, error) {
	return r.client.Notification.Query().
		Where(notification.RequestID(requestID)).
		First(ctx)
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

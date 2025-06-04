package handler

import (
	"context"
	"encoding/json"
	"google.golang.org/protobuf/types/known/emptypb"

	"gitlab.smartbet.am/golang/notification/ent/notification"
	"gitlab.smartbet.am/golang/notification/ent/schema"
	pb "gitlab.smartbet.am/golang/notification/pb/notification"
	"gitlab.smartbet.am/golang/notification/service/db"
	"gitlab.smartbet.am/golang/notification/service/processing"
	"gitlab.smartbet.am/golang/notification/types"
)

type NotificationService struct{}

func (s *NotificationService) CancelPrefix(ctx context.Context, request *pb.CancelPrefixRequest, empty *emptypb.Empty) error {
	err := db.Client().Notification.Update().
		Where(
			notification.RequestIDHasPrefix(request.Prefix),
			notification.StatusEQ(notification.StatusPENDING),
		).
		SetStatus(notification.StatusCANCEL).
		Exec(context.Background())
	return err
}

func (s *NotificationService) ListNotifications(ctx context.Context, request *pb.ListNotificationsRequest, list *pb.NotificationList) error {
	notifications, err := db.Client().Notification.Query().Where(
		notification.RequestIDHasPrefix(request.Prefix),
	).All(ctx)

	if nil != err {
		return err
	}

	for _, item := range notifications {
		list.Notifications = append(list.Notifications, &pb.Notification{
			From:         item.From,
			Headline:     item.Headline,
			Name:         item.Name,
			Subject:      item.Headline,
			Status:       item.Status.String(),
			ErrorMessage: item.ErrorMessage,
			Info: &pb.NotificationInfo{
				RequestId: item.RequestID,
				Tag:       item.Tag,
			},
		})
	}

	return nil
}

func (s *NotificationService) Retry(ctx context.Context, info *pb.NotificationInfo, empty *emptypb.Empty) error {
	message, err := db.Client().
		Notification.
		Query().
		Where(
			notification.RequestID(info.RequestId),
		).Only(context.Background())
	if nil != err {
		return err
	}
	return processing.Processors.Handle(message)
}

func (s *NotificationService) Cancel(ctx context.Context, request *pb.CancelRequest, empty *emptypb.Empty) error {
	_, err := db.Client().Notification.Update().
		Where(notification.RequestIDIn(request.RequestIds...)).
		SetStatus(notification.StatusCANCEL).
		Save(context.Background())
	return err
}

func checkRetry(requestID string) bool {
	exits, _ := db.Client().Notification.Query().Where(notification.RequestID(requestID)).Exist(context.Background())
	return exits
}

func (*NotificationService) Email(c context.Context, request *pb.EmailRequest, _ *emptypb.Empty) error {
	return notificationServiceHandler{}.Email(c, request)
}
func (*NotificationService) Sms(c context.Context, request *pb.SmsRequest, _ *emptypb.Empty) error {
	return notificationServiceHandler{}.Sms(c, request)
}
func (*NotificationService) Notification(c context.Context, request *pb.PushRequest, response *emptypb.Empty) error {
	return notificationServiceHandler{}.Notification(c, request, response)
}

type notificationServiceHandler struct{}

func (n notificationServiceHandler) Email(_ context.Context, request *pb.EmailRequest) error {

	if checkRetry(request.Info.RequestId) {
		return nil
	}

	item := db.Client().Notification.Create().
		SetType(notification.TypeEMAIL).
		SetFrom(request.From).
		SetReplyTo(request.ReplyTo).
		SetTag(request.Info.Tag).
		SetTenantID(request.TenantId).
		SetAddress(types.Address(request.Address)).
		SetBody(request.Body).
		SetHeadline(request.Subject).
		SetName(request.Name).
		SetStatus(notification.StatusACTIVE).
		SetRequestID(request.Info.RequestId)

	if nil != request.Meta {
		meta := &schema.NotificationMeta{
			TemplateID: request.Meta.TemplateId,
			Params:     request.Meta.Params.AsMap(),
		}
		if nil != request.Meta.Attachment {
			meta.Attachment = &schema.Attachment{
				Filename:    request.Meta.Attachment.Filename,
				Content:     request.Meta.Attachment.Content,
				Disposition: request.Meta.Attachment.Disposition,
				Type:        request.Meta.Attachment.Type,
			}
		}

		item.SetMeta(meta)
	}

	if nil != request.Schedule {
		item.
			SetScheduleTs(request.Schedule.Time.GetSeconds()).
			SetStatus(notification.StatusPENDING)
	}

	result, err := item.Save(context.Background())

	if nil != err {
		return err
	}

	if nil == request.Schedule {
		err := processing.Processors.Handle(result)
		if nil != err {
			return err
		}
	}

	return err
}

func (n notificationServiceHandler) Sms(_ context.Context, request *pb.SmsRequest) error {
	// TODO implement me

	if checkRetry(request.Info.RequestId) {
		return nil
	}

	item := db.Client().Notification.Create().
		SetAddress(types.Address(request.PhoneNumber)).
		SetBody(request.Body).
		SetTag(request.Info.Tag).
		SetTenantID(request.TenantId).
		SetType(notification.TypeSMS).
		SetStatus(notification.StatusACTIVE).
		SetMeta(&schema.NotificationMeta{Service: request.Service}).
		SetRequestID(request.Info.RequestId)

	if nil != request.Schedule {
		item.
			SetScheduleTs(request.Schedule.Time.GetSeconds()).
			SetStatus(notification.StatusPENDING)
	}

	result, err := item.Save(context.Background())

	if nil != err {
		return err
	}

	if nil == request.Schedule {
		err := processing.Processors.Handle(result)
		if nil != err {
			return err
		}
	}
	return err
}

func (n notificationServiceHandler) Notification(_ context.Context, request *pb.PushRequest, _ *emptypb.Empty) error {

	if checkRetry(request.Info.RequestId) {
		return nil
	}

	item := db.Client().Notification.Create().
		SetAddress(types.Address(request.Address)).
		SetBody(request.Body).
		SetTag(request.Info.Tag).
		SetHeadline(request.Headline).
		SetType(notification.TypePUSH).
		SetStatus(notification.StatusACTIVE).
		SetMeta(&schema.NotificationMeta{
			Data: json.RawMessage(request.Meta.Data),
		}).
		SetRequestID(request.Info.RequestId)

	if nil != request.Schedule {
		item.
			SetScheduleTs(request.Schedule.Time.GetSeconds()).
			SetStatus(notification.StatusPENDING)
	}

	result, err := item.Save(context.Background())

	if nil != err {
		return err
	}

	if nil == request.Schedule {
		err := processing.Processors.Handle(result)
		if nil != err {
			return err
		}
	}
	return err
}

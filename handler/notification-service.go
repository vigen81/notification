package handler

import (
	`context`
	`encoding/json`
	
	`gitlab.com/healthcare-integration/golang/notification-service/ent/notification`
	`gitlab.com/healthcare-integration/golang/notification-service/ent/schema`
	pb `gitlab.com/healthcare-integration/golang/notification-service/pb/notification`
	`gitlab.com/healthcare-integration/golang/notification-service/service/db`
	`gitlab.com/healthcare-integration/golang/notification-service/service/processing`
	`gitlab.com/healthcare-integration/golang/notification-service/types`
)

type NotificationService struct{}

func checkRetry(requestID *string) bool {
	if nil == requestID {
		return false
	}
	exits, _ := db.Client().Notification.Query().Where(notification.RequestID(*requestID)).Exist(context.Background())
	return exits
}

func (s *NotificationService) Retry(ctx context.Context, request *pb.RetryRequest, response *pb.RetryResponse) error {
	
	message, err := db.Client().
		Notification.
		Query().
		Where(
			notification.ID(int(request.MessageId)),
		).Only(context.Background())
	if nil != err {
		return err
	}
	return processing.Processors.Handle(message)
	
}

func (*NotificationService) Email(c context.Context, request *pb.EmailRequest, response *pb.EmailResponse) error {
	return notificationServiceHandler{}.Email(c, request, response)
}
func (*NotificationService) Sms(c context.Context, request *pb.SmsRequest, response *pb.SmsResponse) error {
	return notificationServiceHandler{}.Sms(c, request, response)
}
func (*NotificationService) Notification(c context.Context, request *pb.PushRequest, response *pb.PushResponse) error {
	return notificationServiceHandler{}.Notification(c, request, response)
}

type notificationServiceHandler struct{}

func (n notificationServiceHandler) Email(_ context.Context, request *pb.EmailRequest, response *pb.EmailResponse) error {
	
	if checkRetry(request.RequestId) {
		return nil
	}
	
	item := db.Client().Notification.Create().
		SetType(notification.TypeEMAIL).
		SetAddress(types.Address(request.Address)).
		SetBody(request.Body).
		SetHeadline(request.Subject).
		SetName(request.Name).
		SetStatus(notification.StatusACTIVE).
		SetNillableRequestID(request.RequestId)
	
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

func (n notificationServiceHandler) Sms(_ context.Context, request *pb.SmsRequest, response *pb.SmsResponse) error {
	// TODO implement me
	
	if checkRetry(request.RequestId) {
		return nil
	}
	
	item := db.Client().Notification.Create().
		SetAddress(types.Address(request.PhoneNumber)).
		SetBody(request.Body).
		SetType(notification.TypeSMS).
		SetStatus(notification.StatusACTIVE).
		SetMeta(&schema.NotificationMeta{Service: request.Service}).
		SetNillableRequestID(request.RequestId)
	
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

func (n notificationServiceHandler) Notification(_ context.Context, request *pb.PushRequest, response *pb.PushResponse) error {
	
	if checkRetry(request.RequestId) {
		return nil
	}
	
	item := db.Client().Notification.Create().
		SetAddress(types.Address(request.Address)).
		SetBody(request.Body).
		SetHeadline(request.Headline).
		SetType(notification.TypePUSH).
		SetStatus(notification.StatusACTIVE).
		SetMeta(&schema.NotificationMeta{
			Data: json.RawMessage(request.Meta.Data),
		}).
		SetNillableRequestID(request.RequestId)
	
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

package processing

import (
	`context`
	`errors`
	
	`go-micro.dev/v4/logger`
	
	`gitlab.com/healthcare-integration/golang/notification-service/ent`
	`gitlab.com/healthcare-integration/golang/notification-service/ent/notification`
	`gitlab.com/healthcare-integration/golang/notification-service/service/processing/email`
)

type IHandle interface {
	Do(item *ent.Notification) error
}

type handler struct {
	enum map[notification.Type]IHandle
}

func (h *handler) GetProcessor(name notification.Type) (handle IHandle, err error) {
	handle, ok := h.enum[name]
	if false == ok {
		return nil, errors.New("unknown notification type")
	}
	return
}

func (h *handler) Handle(item *ent.Notification) error {
	executor, err := h.GetProcessor(item.Type)
	if nil != err {
		return err
	}
	
	err = item.Update().SetStatus(notification.StatusACTIVE).Exec(context.Background())
	if nil != err {
		return err
	}
	processError := executor.Do(item)
	status := notification.StatusCOMPLETED
	var errorMessage *string
	if nil != processError {
		status = notification.StatusFAILED
		e := processError.Error()
		errorMessage = &e
		
	}
	_, err = item.Update().SetStatus(status).SetNillableErrorMessage(errorMessage).Save(context.Background())
	
	if nil != err {
		logger.Errorf("can't update notification status: %s", err.Error())
	}
	return processError
	
}

var Processors *handler

func init() {
	Processors = &handler{
		enum: map[notification.Type]IHandle{
			notification.TypePUSH:  NewNotification(),
			notification.TypeSMS:   NewSms(),
			notification.TypeEMAIL: email.NewMailer(),
		},
	}
}

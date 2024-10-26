package processing

import (
	"gitlab.com/healthcare-integration/golang/notification-service/ent"
	"gitlab.com/healthcare-integration/golang/notification-service/service/processing/twilio_client"
)

type Sms struct {
}

func NewSms() *Sms {
	return &Sms{}
}

func (s *Sms) Do(notification *ent.Notification) (err error) {
	addr := notification.Address.String()
	if nil != err {
		return err
	}
	meta := notification.Meta

	_, err = twilio_client.Sms(addr, notification.Body, meta.Service)

	if nil != err {
		return err
	}

	return
}

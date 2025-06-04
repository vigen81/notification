package processing

import (
	"gitlab.smartbet.am/golang/notification/ent"
	"gitlab.smartbet.am/golang/notification/service/processing/twilio_client"
)

type Sms struct {
}

func NewSms() *Sms {
	return &Sms{}
}

func (s *Sms) Do(notification *ent.Notification) (err error) {
	addr := notification.Address.String()

	_, err = twilio_client.Sms(notification.TenantID, addr, notification.Body)

	if nil != err {
		return err
	}

	return
}

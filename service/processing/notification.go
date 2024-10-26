package processing

import (
	`gitlab.com/healthcare-integration/golang/notification-service/ent`
	`gitlab.com/healthcare-integration/golang/notification-service/service/processing/fcm`
)

type Notification struct {
}

func NewNotification() *Notification {
	return &Notification{}
}

func (s *Notification) Do(notification *ent.Notification) error {
	addr := notification.Address
	
	err := fcm.Send(fcm.Message{
		Address: addr.String(),
		Title:   notification.Headline,
		Body:    notification.Body,
		Data:    notification.Meta.Data,
	})
	return err
}

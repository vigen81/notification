package task

import (
	`context`
	`github.com/gammazero/workerpool`
	`github.com/jasonlvhit/gocron`
	`gitlab.com/healthcare-integration/golang/notification-service/ent/notification`
	`gitlab.com/healthcare-integration/golang/notification-service/service/db`
	`gitlab.com/healthcare-integration/golang/notification-service/service/processing`
	`go-micro.dev/v4/errors`
	`go-micro.dev/v4/logger`
	`time`
)

var wp = workerpool.New(16)

func NotificationTask() {
	items, err := db.Client().
		Notification.
		Query().
		Where(
			notification.StatusEQ(notification.StatusPENDING),
			notification.ScheduleTsLTE(time.Now().Unix()),
			notification.ScheduleTsNotNil(),
		).All(context.Background())
	
	if nil != err {
		logger.Error(err.Error())
	}
	for _, message := range items {
		wp.Submit(func() {
			err := processing.Processors.Handle(message)
			if err != nil {
				logger.Errorf("can't process message: %s", err.Error())
			}
		})
	}
}

func Background() error {
	err := gocron.Every(1).Minute().Do(NotificationTask)
	if err != nil {
		return errors.New("cannot run task %s", err.Error(), -1)
	}
	gocron.Start()
	
	return err
	
}

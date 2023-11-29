package consume

import (
	`context`
	`time`
	
	`gitlab.com/healthcare-integration/common/bootstrap/bootstrap`
	`gitlab.com/healthcare-integration/stream`
	`go-micro.dev/v4/config/reader`
	`go-micro.dev/v4/logger`
	
	`gitlab.com/healthcare-integration/golang/notification-service/handler`
	pb `gitlab.com/healthcare-integration/golang/notification-service/pb/notification`
)

const (
	Subject    = "ns.*"
	StreamName = "notifications"
)

func Configure(value reader.Value) error {
	var addr bootstrap.Address
	err := value.Scan(&addr)
	if nil != err {
		return err
	}
	return Init(addr.DSN())
}

var ctx stream.Stream

func Init(address string) error {
	ctx = stream.New(
		stream.WithCreateStream(true),
		stream.WithName(StreamName),
		stream.WithSubject(Subject),
		stream.WithAddress(address),
		stream.WithRetentionPolicy(stream.WorkingQue),
		stream.WithMaxAge(time.Hour*4),
	)
	
	err := ctx.Init()
	if nil != err {
		return err
	}
	return Background()
	
}

func Background() error {
	h := &handler.NotificationService{}
	return ctx.Subscriber().Subscribe("ns.>", func(msg *stream.IncomingMessage) error {
		
		logger.Info(`consume.Background`, string(msg.Body))
		switch msg.Subject {
		case "ns.sms":
			var item = &pb.SmsRequest{}
			err := msg.Unmarshal(item)
			if nil != err {
				return err
			}
			err = h.Sms(context.Background(), item, &pb.SmsResponse{})
			if nil != err {
				return err
			}
			return nil
		case "ns.email":
			var item = &pb.EmailRequest{}
			err := msg.Unmarshal(item)
			if nil != err {
				return err
			}
			err = h.Email(context.Background(), item, &pb.EmailResponse{})
			if nil != err {
				return err
			}
			
			return nil
		case "ns.push":
			logger.Infof("message ns.push %s", string(msg.Body))
			var item = &pb.PushRequest{}
			err := msg.Unmarshal(item)
			if nil != err {
				return err
			}
			err = h.Notification(context.Background(), item, &pb.PushResponse{})
			if nil != err {
				return err
			}
		}
		return nil
	})
}

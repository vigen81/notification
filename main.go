package main

import (
	`time`
	_ "time/tzdata"
	
	`gitlab.com/healthcare-integration/common/bootstrap/bootstrap`
	`go-micro.dev/v4`
	`go-micro.dev/v4/config/reader`
	`go-micro.dev/v4/logger`
	
	`gitlab.com/healthcare-integration/golang/notification-service/handler`
	pb `gitlab.com/healthcare-integration/golang/notification-service/pb/notification`
	`gitlab.com/healthcare-integration/golang/notification-service/service/consume`
	`gitlab.com/healthcare-integration/golang/notification-service/service/crypt`
	`gitlab.com/healthcare-integration/golang/notification-service/service/db`
	`gitlab.com/healthcare-integration/golang/notification-service/service/processing/email`
	`gitlab.com/healthcare-integration/golang/notification-service/service/processing/fcm`
	`gitlab.com/healthcare-integration/golang/notification-service/service/processing/twilio_client`
	`gitlab.com/healthcare-integration/golang/notification-service/service/task`
	
	`go-micro.dev/v4/config`
	
	gs "github.com/asim/go-micro/plugins/server/grpc/v4"
)

var (
	serviceName = "code.codify.notification.service"
	version     = "latest"
)

func main() {
	
	loc, err := time.LoadLocation("America/New_York")
	time.Local = loc
	if nil != err {
		logger.Fatal(err)
	}
	
	if nil != err {
		logger.Fatalf("error loading config: %s", err.Error())
	}
	
	// fmt.Println(xx)
	if nil != err {
		logger.Fatal(err)
	}
	
	err = bootstrap.Init(
		bootstrap.WithService(serviceName),
		bootstrap.WithLogger(true),
		bootstrap.WithRegistry(),
	)
	if err != nil {
		logger.Fatal(err.Error())
	}
	// Create service
	service := micro.NewService(
		micro.Server(gs.NewServer()),
		micro.Address(":8160"),
		micro.Name(serviceName),
		micro.Version(version),
		micro.Registry(bootstrap.Registry()),
		micro.WrapHandler(bootstrap.TcpWrapper),
	)
	
	service.Init(
		micro.BeforeStart(func() error {
			return configure()
		}),
		micro.AfterStart(func() error {
			return task.Background()
		}),
	)
	
	logger.Fields(map[string]interface{}{"x": 2}).Log(logger.InfoLevel, "debug")
	
	// Register handler
	err = pb.RegisterNotificationServiceHandler(service.Server(), new(handler.NotificationService))
	if err != nil {
		logger.Fatal("ERR ", err)
	}
	
	// Run service
	if err := service.Run(); err != nil {
		logger.Fatal("RUN ", err)
	}
}

type configFunc func(value reader.Value) error

func configure() (err error) {
	configurators := map[string]configFunc{
		"database": db.Configure,
		"twilio":   twilio_client.Configure,
		"crypt":    crypt.Configure,
		"fcm":      fcm.Configure,
		"sendgrid": email.Configure,
		"nats":     consume.Configure,
	}
	
	for name, call := range configurators {
		err = call(config.Get(name))
		if nil != err {
			return err
		}
	}
	return
}

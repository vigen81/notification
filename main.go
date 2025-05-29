package main

import (
	"gitlab.com/healthcare-integration/golang/notification-service/service/broker"
	"gitlab.com/healthcare-integration/golang/notification-service/web/routes"
	"go-micro.dev/v4"
	"go-micro.dev/v4/server"
	"os"
	"time"
	_ "time/tzdata"

	"gitlab.com/healthcare-integration/common/bootstrap"
	"go-micro.dev/v4/config/reader"
	"go-micro.dev/v4/logger"

	httpServer "github.com/go-micro/plugins/v4/server/http"
	"gitlab.com/healthcare-integration/golang/notification-service/handler"
	pb "gitlab.com/healthcare-integration/golang/notification-service/pb/notification"
	"gitlab.com/healthcare-integration/golang/notification-service/service/crypt"
	"gitlab.com/healthcare-integration/golang/notification-service/service/db"
	"gitlab.com/healthcare-integration/golang/notification-service/service/processing/email"
	"gitlab.com/healthcare-integration/golang/notification-service/service/processing/twilio_client"
	"gitlab.com/healthcare-integration/golang/notification-service/service/task"

	"go-micro.dev/v4/config"

	gs "github.com/go-micro/plugins/v4/server/grpc"
)

var (
	serviceName     = "core.codify.notification.service"
	version         = "latest"
	httpServiceName = "core.codify.notification.http.service"
)

func main() {

	loc, err := time.LoadLocation("America/New_York")
	time.Local = loc
	if nil != err {
		logger.Fatal(err)
	}

	err = bootstrap.Init(
		bootstrap.WithService(serviceName),
		bootstrap.WithConfigServiceDSN("config-service:8010"),
		bootstrap.WithLogger(true),
		bootstrap.WithRegistry(),
	)
	if err != nil {
		logger.Fatal(err.Error())
	}
	var port string = ":50050"
	if os.Getenv("DEV") == "true" {
		port = ":50058"
	}

	webServer := httpServer.NewServer(
		server.Address(":8080"),
		server.Name(httpServiceName),
		server.Version(version),
		server.RegisterInterval(time.Second*30),
		server.RegisterTTL(time.Second*15),
		server.Registry(bootstrap.Registry()),
	)

	service := micro.NewService(
		micro.Server(gs.NewServer()),
		micro.Address(port),
		micro.Name(serviceName),
		micro.Version(version),
		micro.Registry(bootstrap.Registry()),
		micro.RegisterTTL(time.Second*60),
		micro.RegisterInterval(time.Second*60),
		micro.BeforeStart(func() error {
			return configure()
		}),
		micro.WrapHandler(bootstrap.TcpWrapper),
		micro.AfterStart(func() error {
			err = task.Background()

			if err != nil {
				return err
			}
			err = webServer.Start()
			//
			if nil != err {
				return err
			}
			//return nil
			//err := webServer.Start()
			//if nil != err {
			//	return err
			//}
			return nil
		}),
	)

	//service.Init()

	hd := webServer.NewHandler(routes.Server())

	if err := webServer.Handle(hd); err != nil {
		logger.Fatal(err)
	}

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
		"rabbitmq": broker.Configure,
		"sendgrid": email.Configure,
	}

	for name, call := range configurators {
		err = call(config.Get(name))
		if nil != err {
			return err
		}
	}
	return
}

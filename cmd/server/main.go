package main

import (
	"context"

	"gitlab.smartbet.am/golang/notification/internal/app"
	"gitlab.smartbet.am/golang/notification/internal/config"
	"gitlab.smartbet.am/golang/notification/internal/db"
	"gitlab.smartbet.am/golang/notification/internal/handlers"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/logger"
	"gitlab.smartbet.am/golang/notification/internal/providers"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/server"
	"gitlab.smartbet.am/golang/notification/internal/services"
	"gitlab.smartbet.am/golang/notification/internal/workers"

	"go.uber.org/fx"
)

// @title Notification Engine API
// @version 1.0
// @description Multi-tenant notification engine supporting Email, SMS, and Push notifications
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	app := fx.New(
		// Configuration
		fx.Provide(config.NewConfig),
		fx.Provide(logger.NewLogger),

		// Database
		fx.Provide(db.NewDatabase),
		fx.Provide(db.NewEntClient),

		// Kafka
		fx.Provide(kafka.NewKafkaConfig),
		fx.Provide(kafka.NewPublisher),
		fx.Provide(kafka.NewSubscriber),

		// Repositories
		fx.Provide(repository.NewNotificationRepository),
		fx.Provide(repository.NewPartnerConfigRepository),

		// Provider System
		fx.Provide(providers.NewProviderRegistry),
		fx.Provide(providers.NewEmailProviderManager),
		fx.Provide(providers.NewSMSProviderManager),

		// Services
		fx.Provide(services.NewNotificationService),
		fx.Provide(services.NewBatchService),

		// Handlers
		fx.Provide(handlers.NewNotificationHandler),
		fx.Provide(handlers.NewConfigHandler),
		fx.Provide(handlers.NewHealthHandler),

		// Workers
		fx.Provide(workers.NewNotificationWorker),
		fx.Provide(workers.NewSchedulerWorker),

		// Server
		fx.Provide(server.NewFiberServer),

		// Application
		fx.Provide(app.NewApplication),

		// Lifecycle
		fx.Invoke(func(lifecycle fx.Lifecycle, app *app.Application) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return app.Start(ctx)
				},
				OnStop: func(ctx context.Context) error {
					return app.Stop(ctx)
				},
			})
		}),
	)

	app.Run()
}

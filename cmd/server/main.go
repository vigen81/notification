package main

import (
	"context"
	"gitlab.smartbet.am/golang/notification/internal/app"
	"gitlab.smartbet.am/golang/notification/internal/config"
	"gitlab.smartbet.am/golang/notification/internal/db"
	"gitlab.smartbet.am/golang/notification/internal/handlers"
	"gitlab.smartbet.am/golang/notification/internal/kafka"
	"gitlab.smartbet.am/golang/notification/internal/providers"
	"gitlab.smartbet.am/golang/notification/internal/repository"
	"gitlab.smartbet.am/golang/notification/internal/server"
	"gitlab.smartbet.am/golang/notification/internal/services"
	"gitlab.smartbet.am/golang/notification/internal/workers"

	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		// Configuration
		fx.Provide(config.NewConfig),
		fx.Provide(config.NewLogger),

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
		fx.Provide(repository.NewTemplateRepository),

		// Provider Factory and Managers
		fx.Provide(providers.NewProviderRegistry),
		fx.Provide(providers.NewEmailProviderManager),
		fx.Provide(providers.NewSMSProviderManager),
		fx.Provide(providers.NewPushProviderManager),

		// Services
		fx.Provide(services.NewNotificationService),
		fx.Provide(services.NewTemplateService),
		fx.Provide(services.NewLocalizationService),
		fx.Provide(services.NewBatchService),
		fx.Provide(services.NewSchedulerService),

		// Handlers
		fx.Provide(handlers.NewNotificationHandler),
		fx.Provide(handlers.NewConfigHandler),
		fx.Provide(handlers.NewTemplateHandler),

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

package main

import (
	"context"
	"database/sql"

	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/ent"
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

	// Import generated docs for Swagger
	_ "gitlab.smartbet.am/golang/notification/docs"

	"go.uber.org/fx"
)

// @title Notification Engine API
// @version 1.0
// @description A high-performance, multi-tenant notification engine supporting Email, SMS, and Push notifications with per-partner configurations and batch processing capabilities.
// @description
// @description ## Features
// @description - **Multi-tenant Architecture**: Per-partner configurations with isolated data
// @description - **Multiple Notification Types**: Email, SMS, and Push notifications
// @description - **Provider Flexibility**: Support for multiple providers per channel
// @description - **Dual API Support**: HTTP REST API and Kafka messaging
// @description - **Batch Processing**: Efficient batch sending with configurable thresholds
// @description - **Scheduled Notifications**: Support for future-dated notifications
// @description - **Message Type Based Routing**: Different from addresses based on message type
// @description - **Global Authentication**: Manage any tenant from a single authenticated session
// @description
// @description ## Authentication
// @description All API endpoints require a JWT Bearer token. The token should contain admin-level permissions to access any tenant.
// @description For Kafka endpoints, an additional X-Kafka-API-Key header is required.
// @description
// @description ## Message Types
// @description - `bonus`: Bonus-related notifications
// @description - `promo`: Promotional messages
// @description - `report`: Report and analytics notifications
// @description - `system`: System and account notifications
// @description - `payment`: Payment-related notifications
// @description - `support`: Customer support messages
// @description
// @description ## Scheduling
// @description Notifications can be scheduled for future delivery by providing a `schedule_ts` timestamp (Unix epoch).
// @description Immediate notifications are processed right away, while scheduled ones are handled by the scheduler worker.
// @description
// @description ## Rate Limits
// @description Each tenant can configure rate limits per notification type. Default limits apply if not configured.

// @termsOfService http://swagger.io/terms/

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token. Example: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

// @tag.name notifications
// @tag.description Notification sending and status operations

// @tag.name configuration
// @tag.description Partner configuration management

// @tag.name kafka
// @tag.description Direct Kafka operations

// @tag.name health
// @tag.description Health and readiness checks

func main() {
	fx.New(
		// Configuration
		fx.Provide(func() (*config.Config, error) {
			return config.NewConfig()
		}),
		fx.Provide(func() *logrus.Logger {
			return logger.NewLogger()
		}),

		// Database
		fx.Provide(func(cfg *config.Config) (*sql.DB, error) {
			return db.NewDatabase(cfg)
		}),
		fx.Provide(func(database *sql.DB, logger *logrus.Logger) (*ent.Client, error) {
			return db.NewEntClient(database, logger)
		}),

		// Kafka
		fx.Provide(func(cfg *config.Config) *kafka.KafkaConfig {
			return kafka.NewKafkaConfig(cfg)
		}),
		fx.Provide(func(cfg *config.Config) (*kafka.Publisher, error) {
			return kafka.NewPublisher(cfg)
		}),
		fx.Provide(func(cfg *config.Config) (*kafka.Subscriber, error) {
			return kafka.NewSubscriber(cfg)
		}),

		// Repositories
		fx.Provide(func(client *ent.Client, logger *logrus.Logger) *repository.NotificationRepository {
			return repository.NewNotificationRepository(client, logger)
		}),
		fx.Provide(func(client *ent.Client, logger *logrus.Logger) *repository.PartnerConfigRepository {
			return repository.NewPartnerConfigRepository(client, logger)
		}),

		// Provider System
		fx.Provide(func() *providers.ProviderRegistry {
			return providers.NewProviderRegistry()
		}),
		fx.Provide(func(registry *providers.ProviderRegistry, configRepo *repository.PartnerConfigRepository, logger *logrus.Logger) *providers.EmailProviderManager {
			return providers.NewEmailProviderManager(registry, configRepo, logger)
		}),
		fx.Provide(func(registry *providers.ProviderRegistry, configRepo *repository.PartnerConfigRepository, logger *logrus.Logger) *providers.SMSProviderManager {
			return providers.NewSMSProviderManager(registry, configRepo, logger)
		}),

		// Services
		fx.Provide(func(
			notifRepo *repository.NotificationRepository,
			configRepo *repository.PartnerConfigRepository,
			emailManager *providers.EmailProviderManager,
			smsManager *providers.SMSProviderManager,
			logger *logrus.Logger,
		) *services.NotificationService {
			return services.NewNotificationService(notifRepo, configRepo, emailManager, smsManager, logger)
		}),
		fx.Provide(func(
			notificationSvc *services.NotificationService,
			configRepo *repository.PartnerConfigRepository,
			logger *logrus.Logger,
		) *services.BatchService {
			return services.NewBatchService(notificationSvc, configRepo, logger)
		}),

		// Handlers
		fx.Provide(func(
			publisher *kafka.Publisher,
			notifRepo *repository.NotificationRepository,
			logger *logrus.Logger,
		) *handlers.NotificationHandler {
			return handlers.NewNotificationHandler(publisher, notifRepo, logger)
		}),
		fx.Provide(func(configRepo *repository.PartnerConfigRepository, logger *logrus.Logger) *handlers.ConfigHandler {
			return handlers.NewConfigHandler(configRepo, logger)
		}),
		fx.Provide(func(logger *logrus.Logger) *handlers.HealthHandler {
			return handlers.NewHealthHandler(logger)
		}),

		// Workers
		fx.Provide(func(
			subscriber *kafka.Subscriber,
			notifRepo *repository.NotificationRepository,
			notificationSvc *services.NotificationService,
			batchSvc *services.BatchService,
			logger *logrus.Logger,
		) *workers.NotificationWorker {
			return workers.NewNotificationWorker(subscriber, notifRepo, notificationSvc, batchSvc, logger)
		}),
		fx.Provide(func(
			notifRepo *repository.NotificationRepository,
			notificationSvc *services.NotificationService,
			logger *logrus.Logger,
		) *workers.SchedulerWorker {
			return workers.NewSchedulerWorker(notifRepo, notificationSvc, logger)
		}),

		// Server
		fx.Provide(func(
			cfg *config.Config,
			notifHandler *handlers.NotificationHandler,
			configHandler *handlers.ConfigHandler,
			healthHandler *handlers.HealthHandler,
			logger *logrus.Logger,
		) *server.FiberServer {
			return server.NewFiberServer(cfg, notifHandler, configHandler, healthHandler, logger)
		}),

		// Lifecycle
		fx.Invoke(func(
			lifecycle fx.Lifecycle,
			fiberServer *server.FiberServer,
			notificationWorker *workers.NotificationWorker,
			schedulerWorker *workers.SchedulerWorker,
			logger *logrus.Logger,
		) {
			lifecycle.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					logger.Info("Starting notification engine application")

					workerCtx := context.Background()
					// Start workers
					if err := notificationWorker.Start(workerCtx); err != nil {
						return err
					}
					if err := schedulerWorker.Start(workerCtx); err != nil {
						return err
					}

					// Start HTTP server in goroutine
					go func() {
						if err := fiberServer.Start(":8080"); err != nil {
							logger.WithError(err).Fatal("Failed to start server")
						}
					}()

					logger.Info("Notification engine started successfully")
					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("Stopping notification engine application")

					// Shutdown server
					if err := fiberServer.Shutdown(ctx); err != nil {
						logger.WithError(err).Error("Error shutting down server")
					}

					// Stop workers
					notificationWorker.Stop()
					schedulerWorker.Stop()

					logger.Info("Notification engine stopped")
					return nil
				},
			})
		}),
	).Run()
}

package app

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/config"
	"gitlab.smartbet.am/golang/notification/internal/server"
	"gitlab.smartbet.am/golang/notification/internal/workers"
)

type Application struct {
	config             *config.Config
	server             *server.FiberServer
	notificationWorker *workers.NotificationWorker
	schedulerWorker    *workers.SchedulerWorker
	logger             *logrus.Logger
}

func NewApplication(
	config *config.Config,
	server *server.FiberServer,
	notificationWorker *workers.NotificationWorker,
	schedulerWorker *workers.SchedulerWorker,
	logger *logrus.Logger,
) *Application {
	return &Application{
		config:             config,
		server:             server,
		notificationWorker: notificationWorker,
		schedulerWorker:    schedulerWorker,
		logger:             logger,
	}
}

func (a *Application) Start(ctx context.Context) error {
	a.logger.Info("Starting notification engine application")

	// Start workers
	if err := a.notificationWorker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start notification worker: %w", err)
	}

	if err := a.schedulerWorker.Start(ctx); err != nil {
		return fmt.Errorf("failed to start scheduler worker: %w", err)
	}

	// Start HTTP server
	go func() {
		if err := a.server.Start(a.config.Server.Port); err != nil {
			a.logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	a.logger.Info("Notification engine started successfully")
	return nil
}

func (a *Application) Stop(ctx context.Context) error {
	a.logger.Info("Stopping notification engine application")

	// Shutdown server
	if err := a.server.Shutdown(ctx); err != nil {
		a.logger.WithError(err).Error("Error shutting down server")
	}

	// Stop workers
	a.notificationWorker.Stop()
	a.schedulerWorker.Stop()

	a.logger.Info("Notification engine stopped")
	return nil
}

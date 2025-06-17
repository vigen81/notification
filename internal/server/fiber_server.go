package server

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"gitlab.smartbet.am/golang/notification/internal/config"
	"gitlab.smartbet.am/golang/notification/internal/handlers"
	"gitlab.smartbet.am/golang/notification/internal/middleware"
	"go.uber.org/zap"
)

type FiberServer struct {
	app             *fiber.App
	config          *config.Config
	notifHandler    *handlers.NotificationHandler
	configHandler   *handlers.ConfigHandler
	templateHandler *handlers.TemplateHandler
	logger          *zap.Logger
}

func NewFiberServer(
	config *config.Config,
	notifHandler *handlers.NotificationHandler,
	configHandler *handlers.ConfigHandler,
	templateHandler *handlers.TemplateHandler,
	logger *zap.Logger,
) *FiberServer {
	app := fiber.New(fiber.Config{
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Tenant-ID",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	server := &FiberServer{
		app:             app,
		config:          config,
		notifHandler:    notifHandler,
		configHandler:   configHandler,
		templateHandler: templateHandler,
		logger:          logger,
	}

	server.setupRoutes()
	return server
}

func (s *FiberServer) setupRoutes() {
	// Health check
	s.app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "notification-engine",
		})
	})

	// API v1 routes
	v1 := s.app.Group("/api/v1")

	// Apply auth middleware
	v1.Use(middleware.AuthMiddleware())
	v1.Use(middleware.TenantMiddleware())

	// Notification routes
	notifications := v1.Group("/notifications")
	notifications.Post("/send", s.notifHandler.SendNotification)
	notifications.Post("/batch", s.notifHandler.SendBatchNotification)
	notifications.Get("/status/:request_id", s.notifHandler.GetNotificationStatus)

	// Partner configuration routes
	configs := v1.Group("/config")
	configs.Get("/", s.configHandler.GetConfig)
	configs.Put("/", s.configHandler.UpdateConfig)
	configs.Post("/providers/email", s.configHandler.AddEmailProvider)
	configs.Post("/providers/sms", s.configHandler.AddSMSProvider)
	configs.Post("/providers/push", s.configHandler.AddPushProvider)
	configs.Delete("/providers/:type/:name", s.configHandler.RemoveProvider)

	// Template routes
	templates := v1.Group("/templates")
	templates.Get("/", s.templateHandler.ListTemplates)
	templates.Get("/:id", s.templateHandler.GetTemplate)
	templates.Post("/", s.templateHandler.CreateTemplate)
	templates.Put("/:id", s.templateHandler.UpdateTemplate)
	templates.Delete("/:id", s.templateHandler.DeleteTemplate)
	templates.Post("/:id/preview", s.templateHandler.PreviewTemplate)

	// Kafka API endpoints (for direct Kafka publishing)
	kafkaAPI := v1.Group("/kafka")
	kafkaAPI.Use(middleware.KafkaAuthMiddleware()) // Additional security for Kafka endpoints
	kafkaAPI.Post("/publish", s.notifHandler.PublishToKafka)
}

func (s *FiberServer) Start(addr string) error {
	s.logger.Info("starting fiber server", zap.String("address", addr))
	return s.app.Listen(addr)
}

func (s *FiberServer) Shutdown(ctx context.Context) error {
	s.logger.Info("shutting down fiber server")
	return s.app.ShutdownWithContext(ctx)
}

func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal Server Error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error":      message,
		"code":       code,
		"request_id": c.Locals("requestid"),
	})
}

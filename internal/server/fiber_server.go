package server

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/gofiber/swagger"
	"github.com/sirupsen/logrus"
	"gitlab.smartbet.am/golang/notification/internal/config"
	"gitlab.smartbet.am/golang/notification/internal/handlers"
	"gitlab.smartbet.am/golang/notification/internal/middleware"
)

type FiberServer struct {
	app           *fiber.App
	config        *config.Config
	notifHandler  *handlers.NotificationHandler
	configHandler *handlers.ConfigHandler
	healthHandler *handlers.HealthHandler
	logger        *logrus.Logger
}

func NewFiberServer(
	config *config.Config,
	notifHandler *handlers.NotificationHandler,
	configHandler *handlers.ConfigHandler,
	healthHandler *handlers.HealthHandler,
	logger *logrus.Logger,
) *FiberServer {
	app := fiber.New(fiber.Config{
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(requestid.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Tenant-ID, X-Kafka-API-Key",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	server := &FiberServer{
		app:           app,
		config:        config,
		notifHandler:  notifHandler,
		configHandler: configHandler,
		healthHandler: healthHandler,
		logger:        logger,
	}

	server.setupRoutes()
	return server
}

func (s *FiberServer) setupRoutes() {
	// Health check endpoints (no auth required)
	s.app.Get("/health", s.healthHandler.HealthCheck)
	s.app.Get("/ready", s.healthHandler.ReadinessCheck)
	s.app.Get("/live", s.healthHandler.LivenessCheck)

	// Swagger documentation
	if s.config.Swagger.Enabled {
		s.app.Get("/swagger/*", swagger.HandlerDefault)
	}

	// API v1 routes
	v1 := s.app.Group("/api/v1")
	v1.Use(func(c *fiber.Ctx) error {
		c.Locals("tenant_id", int64(1001)) // Set test tenant
		return c.Next()
	})

	// Apply auth middleware for all API routes
	//v1.Use(middleware.AuthMiddleware())
	//v1.Use(middleware.TenantMiddleware())

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

	// Kafka API endpoints (for direct Kafka publishing)
	kafkaAPI := v1.Group("/kafka")
	kafkaAPI.Use(middleware.KafkaAuthMiddleware()) // Additional security for Kafka endpoints
	kafkaAPI.Post("/publish", s.notifHandler.PublishToKafka)

	// Add a catch-all route for undefined endpoints
	s.app.Use("*", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Endpoint not found",
			"code":    "NOT_FOUND",
			"path":    c.Path(),
			"method":  c.Method(),
			"message": "The requested endpoint does not exist",
		})
	})
}

func (s *FiberServer) Start(addr string) error {
	s.logger.WithField("address", addr).Info("Starting Fiber server")
	return s.app.Listen(addr)
}

func (s *FiberServer) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down Fiber server")
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
		"timestamp":  time.Now(),
	})
}

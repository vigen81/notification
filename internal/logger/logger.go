package logger

import (
	"context"
	"os"

	graylog "github.com/gemnasium/logrus-graylog-hook/v3"
	log "github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

var (
	Log *log.Logger
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		DisableQuote:     true,
		PadLevelText:     true,
		QuoteEmptyFields: true,
		ForceColors:      true,
	})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)

	// Initialize Graylog hook
	graylogAddr := os.Getenv("GRAYLOG_ADDR")
	if graylogAddr == "" {
		graylogAddr = "gelf-udp-service:12222"
	}

	hook := graylog.NewGraylogHook(graylogAddr, map[string]interface{}{
		"service": "notification-engine",
	})

	Log = log.StandardLogger()
	Log.AddHook(hook)
}

// NewLogger creates a new logger instance for Uber FX
func NewLogger() *log.Logger {
	return Log
}

// To creates a logger entry with a specific type field
func To(name string) *log.Entry {
	return Log.WithField("type", name)
}

// WithTenant creates a logger entry with tenant information
func WithTenant(tenantID int64) *log.Entry {
	return Log.WithField("tenant_id", tenantID)
}

// WithRequest creates a logger entry with request information
func WithRequest(requestID string) *log.Entry {
	return Log.WithField("request_id", requestID)
}

// NotificationLogger provides structured logging for notifications
type NotificationLogger struct {
	logger *log.Logger
}

func NewNotificationLogger() *NotificationLogger {
	return &NotificationLogger{
		logger: Log,
	}
}

func (nl *NotificationLogger) Info(msg string, fields map[string]interface{}) {
	entry := nl.logger.WithFields(log.Fields(fields))
	entry.Info(msg)
}

func (nl *NotificationLogger) Error(msg string, err error, fields map[string]interface{}) {
	entry := nl.logger.WithFields(log.Fields(fields)).WithError(err)
	entry.Error(msg)
}

func (nl *NotificationLogger) Debug(msg string, fields map[string]interface{}) {
	entry := nl.logger.WithFields(log.Fields(fields))
	entry.Debug(msg)
}

func (nl *NotificationLogger) Warn(msg string, fields map[string]interface{}) {
	entry := nl.logger.WithFields(log.Fields(fields))
	entry.Warn(msg)
}

// FxLogger wraps the logger for use with Uber FX
type FxLogger struct {
	*log.Logger
}

func (l FxLogger) Printf(format string, v ...interface{}) {
	l.Logger.Printf(format, v...)
}

// ProvideLogger provides the logger for dependency injection
func ProvideLogger(lc fx.Lifecycle) *log.Logger {
	logger := Log

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			logger.Info("Logger initialized")
			return nil
		},
		OnStop: func(context.Context) error {
			logger.Info("Logger shutting down")
			return nil
		},
	})

	return logger
}

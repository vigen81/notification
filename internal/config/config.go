package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"gitlab.smartbet.am/golang/notification/internal/plugin/ams"
	"go.uber.org/fx"
)

const mockConfig = `{
	"server": {
		"port": ":8080",
		"read_timeout": "15s",
		"write_timeout": "15s",
		"idle_timeout": "120s"
	},
	"database": {
		"host": "localhost",
		"port": "3306",
		"user": "notification_user",
		"password": "notification_pass",
		"database": "notification_db",
		"max_open_conns": 25,
		"max_idle_conns": 25,
		"conn_max_lifetime": 300
	},
	"kafka": {
		"brokers": ["localhost:9092"],
		"consumer_group": "notification-engine",
		"topics": {
			"notifications": "notifications",
			"events": "notification-events",
			"dead_letter": "notifications-dlq"
		}
	},
	"providers": {
		"sendgrid": {
			"sandbox_mode": false
		},
		"twilio": {
			"edge_location": "sydney"
		},
		"fcm": {
			"validate_only": false
		}
	},
	"batch_defaults": {
		"max_batch_size": 100,
		"flush_interval": "10s",
		"max_retries": 3,
		"retry_backoff": "5s"
	},
	"swagger": {
		"enabled": true,
		"host": "localhost:8080",
		"title": "Notification Engine API",
		"version": "1.0"
	},
	"logging": {
		"graylog_addr": "gelf-udp-service:12222",
		"service_name": "notification-engine"
	},
	"auth": {
		"jwt_secret": "your-secret-key",
		"kafka_api_key": "your-kafka-api-key"
	}
}`

type Config struct {
	Server        ServerConfig        `json:"server"`
	Database      DatabaseConfig      `json:"database"`
	Kafka         KafkaConfig         `json:"kafka"`
	Providers     ProvidersConfig     `json:"providers"`
	BatchDefaults BatchDefaultsConfig `json:"batch_defaults"`
	Swagger       SwaggerConfig       `json:"swagger"`
	Logging       LoggingConfig       `json:"logging"`
	Auth          AuthConfig          `json:"auth"`
}

type ServerConfig struct {
	Port         string `json:"port"`
	ReadTimeout  string `json:"read_timeout"`
	WriteTimeout string `json:"write_timeout"`
	IdleTimeout  string `json:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            string `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	Database        string `json:"database"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

type KafkaConfig struct {
	Brokers       []string     `json:"brokers"`
	ConsumerGroup string       `json:"consumer_group"`
	Topics        TopicsConfig `json:"topics"`
}

type TopicsConfig struct {
	Notifications string `json:"notifications"`
	Events        string `json:"events"`
	DeadLetter    string `json:"dead_letter"`
}

type ProvidersConfig struct {
	SendGrid SendGridDefaults `json:"sendgrid"`
	Twilio   TwilioDefaults   `json:"twilio"`
	FCM      FCMDefaults      `json:"fcm"`
}

type SendGridDefaults struct {
	SandboxMode bool `json:"sandbox_mode"`
}

type TwilioDefaults struct {
	EdgeLocation string `json:"edge_location"`
}

type FCMDefaults struct {
	ValidateOnly bool `json:"validate_only"`
}

type BatchDefaultsConfig struct {
	MaxBatchSize  int    `json:"max_batch_size"`
	FlushInterval string `json:"flush_interval"`
	MaxRetries    int    `json:"max_retries"`
	RetryBackoff  string `json:"retry_backoff"`
}

type SwaggerConfig struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Title   string `json:"title"`
	Version string `json:"version"`
}

type LoggingConfig struct {
	GraylogAddr string `json:"graylog_addr"`
	ServiceName string `json:"service_name"`
}

type AuthConfig struct {
	JWTSecret   string `json:"jwt_secret"`
	KafkaAPIKey string `json:"kafka_api_key"`
}

// Helper methods to parse duration strings
func (c *Config) GetServerReadTimeout() time.Duration {
	if d, err := time.ParseDuration(c.Server.ReadTimeout); err == nil {
		return d
	}
	return 15 * time.Second
}

func (c *Config) GetServerWriteTimeout() time.Duration {
	if d, err := time.ParseDuration(c.Server.WriteTimeout); err == nil {
		return d
	}
	return 15 * time.Second
}

func (c *Config) GetServerIdleTimeout() time.Duration {
	if d, err := time.ParseDuration(c.Server.IdleTimeout); err == nil {
		return d
	}
	return 120 * time.Second
}

func (c *Config) GetBatchFlushInterval() time.Duration {
	if d, err := time.ParseDuration(c.BatchDefaults.FlushInterval); err == nil {
		return d
	}
	return 10 * time.Second
}

func (c *Config) GetBatchRetryBackoff() time.Duration {
	if d, err := time.ParseDuration(c.BatchDefaults.RetryBackoff); err == nil {
		return d
	}
	return 5 * time.Second
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
	)
}

func (cnf *Config) run(serviceName string) error {
	var data []byte
	var err error

	if os.Getenv("POD_ENV") == "local" {
		data = []byte(mockConfig)
	} else {
		secretName := fmt.Sprintf("/%s/%s", os.Getenv("POD_ENV"), serviceName)
		r := ams.NewSource(ams.WithSecretName(secretName))
		data, err = r.Read()
		if err != nil {
			return fmt.Errorf("failed to read config from AWS Parameter Store: %w", err)
		}
	}

	if err := json.Unmarshal(data, cnf); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override with environment variables if present
	cnf.overrideWithEnv()

	return nil
}

func (c *Config) overrideWithEnv() {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		c.Server.Port = port
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if dbPass := os.Getenv("DB_PASSWORD"); dbPass != "" {
		c.Database.Password = dbPass
	}
	if kafkaBrokers := os.Getenv("KAFKA_BROKERS"); kafkaBrokers != "" {
		c.Kafka.Brokers = strings.Split(kafkaBrokers, ",")
	}
	if graylogAddr := os.Getenv("GRAYLOG_ADDR"); graylogAddr != "" {
		c.Logging.GraylogAddr = graylogAddr
	}
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		c.Auth.JWTSecret = jwtSecret
	}
	if kafkaKey := os.Getenv("KAFKA_API_KEY"); kafkaKey != "" {
		c.Auth.KafkaAPIKey = kafkaKey
	}
}

// Provider creates a new config instance using Uber FX lifecycle
func Provider(lifecycle fx.Lifecycle, serviceName string) (*Config, error) {
	c := &Config{}
	err := c.run(serviceName)

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})

	return c, err
}

// NewConfig creates and loads configuration (legacy function for compatibility)
func NewConfig() (*Config, error) {
	c := &Config{}
	serviceName := "notification-engine"
	return c, c.run(serviceName)
}

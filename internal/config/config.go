package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Database      DatabaseConfig      `yaml:"database"`
	Kafka         KafkaConfig         `yaml:"kafka"`
	Providers     ProvidersConfig     `yaml:"providers"`
	BatchDefaults BatchDefaultsConfig `yaml:"batch_defaults"`
	Swagger       SwaggerConfig       `yaml:"swagger"`
}

type ServerConfig struct {
	Port         string        `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

type DatabaseConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type KafkaConfig struct {
	Brokers       []string     `yaml:"brokers"`
	ConsumerGroup string       `yaml:"consumer_group"`
	Topics        TopicsConfig `yaml:"topics"`
}

type TopicsConfig struct {
	Notifications string `yaml:"notifications"`
	Events        string `yaml:"events"`
	DeadLetter    string `yaml:"dead_letter"`
}

type ProvidersConfig struct {
	SendGrid SendGridDefaults `yaml:"sendgrid"`
	Twilio   TwilioDefaults   `yaml:"twilio"`
	FCM      FCMDefaults      `yaml:"fcm"`
}

type SendGridDefaults struct {
	SandboxMode bool `yaml:"sandbox_mode"`
}

type TwilioDefaults struct {
	EdgeLocation string `yaml:"edge_location"`
}

type FCMDefaults struct {
	ValidateOnly bool `yaml:"validate_only"`
}

type BatchDefaultsConfig struct {
	MaxBatchSize  int           `yaml:"max_batch_size"`
	FlushInterval time.Duration `yaml:"flush_interval"`
	MaxRetries    int           `yaml:"max_retries"`
	RetryBackoff  time.Duration `yaml:"retry_backoff"`
}

type SwaggerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Title   string `yaml:"title"`
	Version string `yaml:"version"`
}

func NewConfig() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	cfg.setDefaults()
	cfg.overrideWithEnv()

	return &cfg, nil
}

func (c *Config) setDefaults() {
	if c.Server.Port == "" {
		c.Server.Port = ":8080"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 15 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 15 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 120 * time.Second
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 25
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 25
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 300
	}
	if c.Kafka.ConsumerGroup == "" {
		c.Kafka.ConsumerGroup = "notification-engine"
	}
	if c.BatchDefaults.MaxBatchSize == 0 {
		c.BatchDefaults.MaxBatchSize = 100
	}
	if c.BatchDefaults.FlushInterval == 0 {
		c.BatchDefaults.FlushInterval = 10 * time.Second
	}
	if c.BatchDefaults.MaxRetries == 0 {
		c.BatchDefaults.MaxRetries = 3
	}
	if c.BatchDefaults.RetryBackoff == 0 {
		c.BatchDefaults.RetryBackoff = 5 * time.Second
	}
	if c.Swagger.Title == "" {
		c.Swagger.Title = "Notification Engine API"
	}
	if c.Swagger.Version == "" {
		c.Swagger.Version = "1.0"
	}
	if c.Swagger.Host == "" {
		c.Swagger.Host = "localhost:8080"
	}
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

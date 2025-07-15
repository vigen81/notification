package kafka

import (
	"gitlab.smartbet.am/golang/notification/internal/config"
)

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	Topics  Topics   `json:"topics"`
}

type Topics struct {
	Notifications string `json:"notifications"`
	Events        string `json:"events"`
	DeadLetter    string `json:"dead_letter"`
}

func NewKafkaConfig(cfg *config.Config) *KafkaConfig {
	return &KafkaConfig{
		Brokers: cfg.Kafka.Brokers,
		Topics: Topics{
			Notifications: cfg.Kafka.Topics.Notifications,
			Events:        cfg.Kafka.Topics.Events,
			DeadLetter:    cfg.Kafka.Topics.DeadLetter,
		},
	}
}

package kafka

import (
	"gitlab.smartbet.am/golang/notification/internal/config"
)

type KafkaConfig struct {
	Brokers []string
	Topics  Topics
}

type Topics struct {
	Notifications string
	Events        string
	DeadLetter    string
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

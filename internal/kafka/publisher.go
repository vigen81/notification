package kafka

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"gitlab.smartbet.am/golang/notification/internal/config"
)

type Publisher struct {
	publisher message.Publisher
	logger    watermill.LoggerAdapter
}

func NewPublisher(cfg *config.Config) (*Publisher, error) {
	logger := watermill.NewStdLogger(false, false)

	saramaConfig := kafka.DefaultSaramaSyncPublisherConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true

	publisherConfig := kafka.PublisherConfig{
		Brokers:               cfg.Kafka.Brokers,
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: saramaConfig,
	}

	dd, _ := json.Marshal(cfg.Kafka)
	fmt.Println(string(dd))
	publisher, err := kafka.NewPublisher(publisherConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka publisher: %w", err)
	}

	return &Publisher{
		publisher: publisher,
		logger:    logger,
	}, nil
}

func (p *Publisher) Publish(ctx context.Context, topic string, key string, payload []byte) error {
	msg := message.NewMessage(watermill.NewUUID(), payload)
	msg.Metadata.Set("key", key)

	if err := p.publisher.Publish(topic, msg); err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (p *Publisher) Close() error {
	return p.publisher.Close()
}

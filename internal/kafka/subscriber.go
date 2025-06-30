package kafka

import (
	"context"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"gitlab.smartbet.am/golang/notification/internal/config"
)

type Subscriber struct {
	subscriber message.Subscriber
	logger     watermill.LoggerAdapter
}

func NewSubscriber(cfg *config.Config) (*Subscriber, error) {
	logger := watermill.NewStdLogger(false, false)

	saramaConfig := kafka.DefaultSaramaSubscriberConfig()
	saramaConfig.Consumer.Offsets.Initial = -1

	subscriberConfig := kafka.SubscriberConfig{
		Brokers:               cfg.Kafka.Brokers,
		Unmarshaler:           kafka.DefaultMarshaler{},
		ConsumerGroup:         cfg.Kafka.ConsumerGroup,
		OverwriteSaramaConfig: saramaConfig,
	}

	subscriber, err := kafka.NewSubscriber(subscriberConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka subscriber: %w", err)
	}

	return &Subscriber{
		subscriber: subscriber,
		logger:     logger,
	}, nil
}

func (s *Subscriber) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	return s.subscriber.Subscribe(ctx, topic)
}

func (s *Subscriber) Close() error {
	return s.subscriber.Close()
}

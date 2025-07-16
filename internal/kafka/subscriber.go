package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Shopify/sarama"
	"os"

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

	saramaConfig := sarama.NewConfig()

	saramaConfig.Consumer.Offsets.Initial = -1

	if os.Getenv("LOCAL") != "true" {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		saramaConfig.Net.SASL.TokenProvider = &MSKAccessTokenProvider{Region: "eu-central-1"}
		saramaConfig.Net.TLS.Enable = true // Replace with your region
		saramaConfig.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: true, // This is not recommended for production use
		}
	}

	saramaConfig.Version = sarama.V2_1_0_0
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.ClientID = "watermill"

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

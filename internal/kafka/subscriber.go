package kafka

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"gitlab.smartbet.am/golang/notification/internal/config"
)

type Subscriber struct {
	subscriber message.Subscriber
	logger     watermill.LoggerAdapter
}

func NewSubscriber(cfg *config.Config) (*Subscriber, error) {
	logger := watermill.NewStdLogger(true, true)

	for _, broker := range cfg.Kafka.Brokers {
		conn, err := net.DialTimeout("tcp", broker, 5*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to kafka broker %s: %w", broker, err)
		}
		conn.Close()
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Net.TLS.Enable = true
	saramaConfig.Net.TLS.Config = &tls.Config{
		InsecureSkipVerify: true, // Equivalent to ssl.endpoint.identification.algorithm=
	}
	saramaConfig.Net.SASL.Enable = true
	saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeOAuth
	saramaConfig.Net.SASL.TokenProvider = &MSKAccessTokenProvider{Region: "eu-central-1"}

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

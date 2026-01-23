package kafka

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/IBM/sarama"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aws/aws-msk-iam-sasl-signer-go/signer"
	"gitlab.smartbet.am/golang/notification/internal/config"
)

type Publisher struct {
	publisher message.Publisher
	logger    watermill.LoggerAdapter
}

type MSKAccessTokenProvider struct {
	Region string
}

func (m *MSKAccessTokenProvider) Token() (*sarama.AccessToken, error) {
	token, _, err := signer.GenerateAuthToken(context.TODO(), m.Region)
	return &sarama.AccessToken{Token: token}, err
}

func NewPublisher(cfg *config.Config) (*Publisher, error) {
	logger := watermill.NewStdLogger(true, true)

	for _, broker := range cfg.Kafka.Brokers {
		conn, err := net.DialTimeout("tcp", broker, 5*time.Second)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to kafka broker %s: %w", broker, err)
		}
		conn.Close()
		fmt.Println("Connected to broker:", broker)
	}

	saramaConfig := sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_0_0
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Net.TLS.Enable = false
	saramaConfig.Net.SASL.Enable = true
	saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeOAuth
	saramaConfig.Net.SASL.TokenProvider = &MSKAccessTokenProvider{Region: "eu-central-1"}

	publisherConfig := kafka.PublisherConfig{
		Brokers:               cfg.Kafka.Brokers,
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: saramaConfig,
	}

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

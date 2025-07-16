package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"log"
	"os"

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

	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Return.Errors = true
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)

	if os.Getenv("LOCAL") != "true" {
		saramaConfig.Net.SASL.Enable = true
		saramaConfig.Net.SASL.Mechanism = sarama.SASLTypeOAuth
		saramaConfig.Net.SASL.TokenProvider = &MSKAccessTokenProvider{Region: "eu-central-1"}
		saramaConfig.Net.TLS.Enable = true
		saramaConfig.Net.TLS.Config = &tls.Config{
			InsecureSkipVerify: true, // This is not recommended for production use
		}
	}
	saramaConfig.Version = sarama.V2_1_0_0
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.ClientID = "watermill"

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

package broker

import (
	"context"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

type SubscriberService interface {
	Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error)
}

type subscriberService struct {
	*kafka.Subscriber
	ConsumerGroup string
	Brokers       []string
}

type Option func(*subscriberService)

func WithConsumerGroup(consumerGroup string) Option {
	return func(s *subscriberService) {
		s.ConsumerGroup = consumerGroup
	}
}

func WithBrokers(brokers []string) Option {
	return func(s *subscriberService) {
		s.Brokers = brokers
	}
}

func (s *subscriberService) Subscribe(ctx context.Context, topic string) (<-chan *message.Message, error) {
	messages, err := s.Subscriber.Subscribe(ctx, topic)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func NewSubscriberService(cnf *Config) SubscriberService {
	//conf := kafka.DefaultSaramaSubscriberConfig()
	//conf.Consumer.Offsets.Initial = sarama.OffsetOldest
	var subConfig = kafka.SubscriberConfig{
		Brokers:       cnf.Brokers,
		Unmarshaler:   kafka.DefaultMarshaler{},
		ConsumerGroup: "questix3",
		//OverwriteSaramaConfig: conf,
	}

	sub, err := kafka.NewSubscriber(subConfig, watermill.NewStdLogger(true, false))
	if err != nil {
		panic(err)
	}
	return &subscriberService{Subscriber: sub}
}

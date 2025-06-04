package broker

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
)

type PublisherService interface {
	Publish(string, ...*message.Message) error
}

type publisherService struct {
	*kafka.Publisher
}

func NewPublisherService(cnf *Config) PublisherService {

	var config = kafka.PublisherConfig{
		Brokers:   cnf.Brokers,
		Marshaler: kafka.DefaultMarshaler{},
	}

	kafkaPublisher, err := kafka.NewPublisher(config, watermill.NewStdLogger(true, false))
	if err != nil {
		panic(err)
	}
	return &publisherService{Publisher: kafkaPublisher}
}

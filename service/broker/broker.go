package broker

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.comsmarbet/internal/topic"
)

type Item interface {
}
type Broker[T Item] struct {
	publisher  PublisherService
	subscriber SubscriberService
}

func NewBroker[T Item](publisher PublisherService, subscriber SubscriberService, cnf *Config) *Broker[T] {
	return &Broker[T]{
		publisher:  publisher,
		subscriber: subscriber,
	}
}

func (b *Broker[T]) Publish(topic topic.Topic, data ...T) error {
	var messages []*message.Message
	for _, item := range data {
		msg, err := json.Marshal(item)
		if err != nil {
			return err
		}
		messages = append(messages, message.NewMessage(watermill.NewUUID(), msg))

	}

	return b.publisher.Publish(string(topic), messages...)
}

func (b *Broker[T]) Subscribe(topic topic.Topic) (<-chan *message.Message, error) {
	return b.subscriber.Subscribe(context.Background(), string(topic))
}

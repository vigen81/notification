package buffer

import (
	"encoding/json"
	"fmt"
	"time"
)

type BufItem interface {
}
type Handler interface {
	Handle(any) error
}

type Service[T BufItem] struct {
	buffer      chan T
	ready       chan struct{}
	broker      *broker.Broker[T]
	maxSize     int
	topic       topic.Topic
	flushPeriod time.Duration
}

const maxBufferSize = 50

// NewService creates and initializes a new Service instance with the given broker, topic name, and flush period.
func NewService[T BufItem](broker *broker.Broker[T], topicName topic.Topic, flushPeriod time.Duration) *Service[T] {
	return &Service[T]{
		buffer:      make(chan T, maxBufferSize),
		ready:       make(chan struct{}),
		maxSize:     maxBufferSize,
		broker:      broker,
		topic:       topicName,
		flushPeriod: flushPeriod,
	}
}

func (s *Service[T]) Push(msg T) {
	s.buffer <- msg
	if s.Size() == maxBufferSize-1 {
		s.ready <- struct{}{}
	}
}

func (s *Service[T]) Pop() []T {
	result := make([]T, 0)
	for r := range s.buffer {
		result = append(result, r)
		if len(s.buffer) == 0 {
			break
		}
	}
	return result
}

func (s *Service[T]) IsEmpty() bool {
	return len(s.buffer) == 0
}

func (s *Service[T]) Ready() <-chan struct{} {
	return s.ready
}

func (s *Service[T]) Size() int {
	return len(s.buffer)
}

func (s *Service[T]) Close() {
	close(s.buffer)
}

func (s *Service[T]) StartBufferService() {
	go func() {
		messages, err := s.broker.Subscribe(s.topic)
		if err != nil {
			fmt.Println("Error subscribing to topic:", err)
			return
		}
		for msg := range messages {
			var action T
			err := json.Unmarshal(msg.Payload, &action)
			if err != nil {
				fmt.Println("Error unmarshalling message:", err)
				continue
			}
			s.Push(action)
			msg.Ack()
		}
	}()

	go func() {
		for range time.Tick(s.flushPeriod) {
			if s.IsEmpty() {
				continue
			}
			s.ready <- struct{}{}
		}
	}()

}

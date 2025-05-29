package broker

import (
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-amqp/v2/pkg/amqp"
	"go-micro.dev/v4/config/reader"
	"sync"
)

type config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

var params config

func Configure(r reader.Value) error {
	err := r.Scan(&params)
	if nil != err {
		return err
	}
	return connect()

}

type Publishers struct {
	sync.Mutex
	publishers map[string]*amqp.Publisher
}

var publishers Publishers
var subscriber *amqp.Subscriber
var amqpConfig amqp.Config

func init() {
	publishers = Publishers{
		publishers: make(map[string]*amqp.Publisher),
	}
}

type ConfigOption func(publisher *amqp.Config)

func WithType(t string) ConfigOption {
	return func(conf *amqp.Config) {
		conf.Exchange.Type = t
	}
}

var Direct = WithType("direct")
var Fanout = WithType("fanout")

func Publisher(topic string, exchange string, opts ...ConfigOption) (*amqp.Publisher, error) {
	publishers.Lock()
	defer publishers.Unlock()

	if publisher, ok := publishers.publishers[topic]; ok {
		return publisher, nil
	}
	amqpConfig.Exchange.GenerateName = func(topic string) string {
		return exchange
	}
	//amqpConfig.Publish.GenerateRoutingKey = func(topic string) string {
	//	return fmt.Sprintf("%s_rk", topic)
	//}
	for _, opt := range opts {
		opt(&amqpConfig)
	}
	publisher, err := amqp.NewPublisher(amqpConfig, watermill.NewStdLogger(true, true))

	if err != nil {
		return nil, err
	}

	publishers.publishers[topic] = publisher
	return publisher, nil
}

func connect() error {
	amqpURI := fmt.Sprintf("amqp://%s:%s@%s:%d/", params.Username, params.Password, params.Host, params.Port)
	amqpConfig = amqp.NewDurablePubSubConfig(amqpURI, func(topic string) string {
		return topic
	})

	//amqpConfig.Exchange.GenerateName = func(topic string) string {
	//	return "storage-changes_x"
	//}

	return amqpConfig.ValidateSubscriberWithConnection()

}

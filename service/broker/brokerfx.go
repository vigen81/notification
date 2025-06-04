package broker

import (
	"go.uber.org/fx"
)

type TType struct {
}

var Module = fx.Module("broker", fx.Provide(
	NewBroker[TType],
	NewPublisherService,
	NewSubscriberService,
))

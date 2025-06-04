package broker

import (
	"github.comsmarbet/internal/dto"
	"go.uber.org/fx"
)

var Module = fx.Module("broker", fx.Provide(
	NewBroker[dto.ActionDto],
	NewPublisherService,
	NewSubscriberService,
))

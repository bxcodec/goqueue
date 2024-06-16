package consumer

import (
	"github.com/bxcodec/goqueue/internal/consumer"
	"github.com/bxcodec/goqueue/internal/consumer/rabbitmq"
	"github.com/bxcodec/goqueue/options"
	consumerOpts "github.com/bxcodec/goqueue/options/consumer"
)

func NewConsumer(platform options.Platform, opts ...consumerOpts.ConsumerOptionFunc) consumer.Consumer {
	switch platform {
	case consumerOpts.ConsumerPlatformRabbitMQ:
		return rabbitmq.NewConsumer(opts...)
	case consumerOpts.ConsumerPlatformGooglePubSub:
		// TODO (bxcodec): implement google pubsub publisher
	case consumerOpts.ConsumerPlatformSQS:
		// TODO (bxcodec): implement sns publisher
	default:
		panic("unknown publisher platform")
	}
	return nil
}

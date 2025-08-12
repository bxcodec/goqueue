package consumer

import (
	"github.com/bxcodec/goqueue/internal/consumer"
	"github.com/bxcodec/goqueue/internal/consumer/rabbitmq"
	_ "github.com/bxcodec/goqueue/internal/shared" // Auto-setup logging
	"github.com/bxcodec/goqueue/options"
	consumerOpts "github.com/bxcodec/goqueue/options/consumer"
)

// NewConsumer creates a new consumer based on the specified platform.
// It accepts a platform option and additional consumer option functions.
// It returns a consumer.Consumer interface implementation.
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

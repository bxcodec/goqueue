package publisher

import (
	"github.com/bxcodec/goqueue/internal/publisher"
	"github.com/bxcodec/goqueue/internal/publisher/rabbitmq"
	_ "github.com/bxcodec/goqueue/internal/shared" // Auto-setup logging
	"github.com/bxcodec/goqueue/options"
	publisherOpts "github.com/bxcodec/goqueue/options/publisher"
)

// NewPublisher creates a new publisher based on the specified platform.
// It accepts the platform type and optional publisher options.
// Returns the corresponding publisher implementation based on the platform.
func NewPublisher(platform options.Platform, opts ...publisherOpts.PublisherOptionFunc) publisher.Publisher {
	switch platform {
	case publisherOpts.PublisherPlatformRabbitMQ:
		return rabbitmq.NewPublisher(opts...)
	case publisherOpts.PublisherPlatformGooglePubSub:
		// TODO (bxcodec): implement google pubsub publisher
	case publisherOpts.PublisherPlatformSNS:
		// TODO (bxcodec): implement sns publisher
	default:
		panic("unknown publisher platform")
	}
	return nil
}

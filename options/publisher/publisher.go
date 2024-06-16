package publisher

import (
	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/options"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DefaultContentType = headerVal.ContentTypeJSON
)

// Option define the option property
type PublisherOption struct {
	PublisherID             string
	Middlewares             []interfaces.PublisherMiddlewareFunc
	RabbitMQPublisherConfig *RabbitMQPublisherConfig // optional, only if using RabbitMQ
}

// PublisherOptionFunc used for option chaining
type PublisherOptionFunc func(opt *PublisherOption)

// WithPublisherID sets the publisher ID for the PublisherOption.
// It returns a PublisherOptionFunc that can be used to configure the PublisherOption.
func WithPublisherID(id string) PublisherOptionFunc {
	return func(opt *PublisherOption) {
		opt.PublisherID = id
	}
}

// WithMiddlewares sets the middlewares for the publisher.
// It accepts one or more PublisherMiddlewareFunc functions as arguments.
// These functions are used to modify the behavior of the publisher.
// The middlewares are applied in the order they are provided.
func WithMiddlewares(middlewares ...interfaces.PublisherMiddlewareFunc) PublisherOptionFunc {
	return func(opt *PublisherOption) {
		opt.Middlewares = middlewares
	}
}

// DefaultPublisherOption returns the default options for a publisher.
var DefaultPublisherOption = func() *PublisherOption {
	return &PublisherOption{
		Middlewares: []interfaces.PublisherMiddlewareFunc{},
		PublisherID: uuid.New().String(),
	}
}

// RabbitMQPublisherConfig represents the configuration options for a RabbitMQ publisher.
type RabbitMQPublisherConfig struct {
	// PublisherChannelPoolSize specifies the size of the channel pool for publishing messages.
	PublisherChannelPoolSize int

	// Conn is the RabbitMQ connection to be used by the publisher.
	Conn *amqp.Connection
}

// WithRabbitMQPublisherConfig sets the RabbitMQ publisher configuration for the publisher option.
// It takes a RabbitMQPublisherConfig struct as input and returns a PublisherOptionFunc.
// The returned PublisherOptionFunc sets the RabbitMQPublisherConfig field of the PublisherOption struct.
func WithRabbitMQPublisherConfig(rabbitMQOption *RabbitMQPublisherConfig) PublisherOptionFunc {
	return func(opt *PublisherOption) {
		opt.RabbitMQPublisherConfig = rabbitMQOption
	}
}

const (
	PublisherPlatformRabbitMQ     = options.PlatformRabbitMQ
	PublisherPlatformGooglePubSub = options.PlatformGooglePubSub
	PublisherPlatformSNS          = options.PlatformSNS
)

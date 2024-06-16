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

func WithPublisherID(id string) PublisherOptionFunc {
	return func(opt *PublisherOption) {
		opt.PublisherID = id
	}
}

func WithMiddlewares(middlewares ...interfaces.PublisherMiddlewareFunc) PublisherOptionFunc {
	return func(opt *PublisherOption) {
		opt.Middlewares = middlewares
	}
}

var DefaultPublisherOption = func() *PublisherOption {
	return &PublisherOption{
		Middlewares: []interfaces.PublisherMiddlewareFunc{},
		PublisherID: uuid.New().String(),
	}
}

type RabbitMQPublisherConfig struct {
	PublisherChannelPoolSize int
	Conn                     *amqp.Connection
}

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

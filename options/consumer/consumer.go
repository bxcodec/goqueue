package consumer

import (
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/options"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DefaultMaxRetryFailedMessage = 3
	DefaultBatchMessageSize      = 1
)

// ConsumerOption represents the configuration options for the consumer.
// ConsumerOption represents the options for configuring a consumer.
type ConsumerOption struct {
	// BatchMessageSize specifies the maximum number of messages to be processed in a single batch.
	BatchMessageSize int

	// QueueName specifies the name of the queue to consume messages from.
	QueueName string

	// Middlewares is a list of middleware functions to be applied to the inbound message handler.
	Middlewares []interfaces.InboundMessageHandlerMiddlewareFunc

	// ActionsPatternSubscribed specifies the list of action patterns that the consumer is subscribed to.
	ActionsPatternSubscribed []string

	// TopicName specifies the name of the topic to consume messages from.
	TopicName string

	// MaxRetryFailedMessage specifies the maximum number of times a failed message should be retried.
	MaxRetryFailedMessage int64

	// ConsumerID specifies the unique identifier for the consumer.
	ConsumerID string

	// RabbitMQConsumerConfig specifies the configuration for RabbitMQ consumer (optional, only if using RabbitMQ).
	RabbitMQConsumerConfig *RabbitMQConsumerConfig
}

// ConsumerOptionFunc is a function type that takes an `opt` parameter of type `*ConsumerOption`.
// It is used as an option for configuring behavior in the `ConsumerOption` struct.
type ConsumerOptionFunc func(opt *ConsumerOption)

// WithBatchMessageSize sets the batch message size option for the consumer.
// It takes an integer value 'n' and returns an ConsumerOptionFunc that sets the
// BatchMessageSize field of the ConsumerOption struct to 'n'.
func WithBatchMessageSize(n int) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.BatchMessageSize = n
	}
}

// WithQueueName sets the queue name for the consumer option.
func WithQueueName(name string) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.QueueName = name
	}
}

// WithMiddlewares is an ConsumerOptionFunc that sets the provided middlewares for the consumer.
// Middlewares are used to process inbound messages before they are handled by the consumer.
// The middlewares are applied in the order they are provided.
func WithMiddlewares(middlewares ...interfaces.InboundMessageHandlerMiddlewareFunc) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.Middlewares = middlewares
	}
}

// WithActionsPatternSubscribed sets the actions that the consumer will subscribe to.
// It takes a variadic parameter `actions` which represents the actions to be subscribed.
// The actions are stored in the `ActionsPatternSubscribed` field of the `ConsumerOption` struct.
func WithActionsPatternSubscribed(actions ...string) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.ActionsPatternSubscribed = actions
	}
}

// WithTopicName sets the topic name for the consumer option.
func WithTopicName(name string) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.TopicName = name
	}
}

// WithMaxRetryFailedMessage sets the maximum number of retries for failed messages.
// It takes an integer parameter 'n' and returns an ConsumerOptionFunc.
// The ConsumerOptionFunc updates the 'MaxRetryFailedMessage' field of the ConsumerOption struct.
func WithMaxRetryFailedMessage(n int64) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.MaxRetryFailedMessage = n
	}
}

// WithConsumerID sets the consumer ID for the consumer option.
func WithConsumerID(id string) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.ConsumerID = id
	}
}

// WithRabbitMQConsumerConfig sets the RabbitMQ consumer configuration for the consumer option.
func WithRabbitMQConsumerConfig(rabbitMQOption *RabbitMQConsumerConfig) ConsumerOptionFunc {
	return func(opt *ConsumerOption) {
		opt.RabbitMQConsumerConfig = rabbitMQOption
	}
}

// DefaultConsumerOption returns the default consumer option.
var DefaultConsumerOption = func() *ConsumerOption {
	return &ConsumerOption{
		Middlewares:           []interfaces.InboundMessageHandlerMiddlewareFunc{},
		BatchMessageSize:      DefaultBatchMessageSize,
		MaxRetryFailedMessage: DefaultMaxRetryFailedMessage,
	}
}

// RabbitMQConsumerConfig represents the configuration for a RabbitMQ consumer.
type RabbitMQConsumerConfig struct {
	// ConsumerChannel is the channel used for consuming messages from RabbitMQ.
	ConsumerChannel *amqp.Channel
	// ReQueueChannel is the channel used for re-queuing messages in RabbitMQ.
	ReQueueChannel *amqp.Channel
}

const (
	ConsumerPlatformRabbitMQ     = options.PlatformRabbitMQ
	ConsumerPlatformGooglePubSub = options.PlatformGooglePubSub
	ConsumerPlatformSQS          = options.PlatformSQS
)

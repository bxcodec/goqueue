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

	// QueueDeclareConfig specifies the configuration for declaring a RabbitMQ queue.
	QueueDeclareConfig *RabbitMQQueueDeclareConfig

	// QueueBindConfig specifies the configuration for binding a queue to an exchange in RabbitMQ.
	QueueBindConfig *RabbitMQQueueBindConfig
}

// RabbitMQQueueDeclareConfig represents the configuration options for declaring a RabbitMQ queue.
// * Durable and Non-Auto-Deleted queues will survive server restarts and remain
// when there are no remaining consumers or bindings.  Persistent publishings will
// be restored in this queue on server restart.  These queues are only able to be
// bound to durable exchanges.
// * Non-Durable and Auto-Deleted queues will not be redeclared on server restart
// and will be deleted by the server after a short time when the last consumer is
// canceled or the last consumer's channel is closed.  Queues with this lifetime
// can also be deleted normally with QueueDelete.  These durable queues can only
// be bound to non-durable exchanges.
// * Non-Durable and Non-Auto-Deleted queues will remain declared as long as the
// server is running regardless of how many consumers.  This lifetime is useful
// for temporary topologies that may have long delays between consumer activity.
// These queues can only be bound to non-durable exchanges.
// * Durable and Auto-Deleted queues will be restored on server restart, but without
// active consumers will not survive and be removed.  This Lifetime is unlikely
// to be useful.
// * Exclusive queues are only accessible by the connection that declares them and
// will be deleted when the connection closes.  Channels on other connections
// will receive an error when attempting  to declare, bind, consume, purge or
// delete a queue with the same name.
// * When noWait is true, the queue will assume to be declared on the server.  A
// channel exception will arrive if the conditions are met for existing queues
// or attempting to modify an existing queue from a different connection.
// When the error return value is not nil, you can assume the queue could not be
// declared with these parameters, and the channel will be closed.
type RabbitMQQueueDeclareConfig struct {
	Durable    bool       // Whether the queue should survive a broker restart.
	AutoDelete bool       // Whether the queue should be deleted when there are no more consumers.
	Exclusive  bool       // Whether the queue should be exclusive to the connection that declares it.
	NoWait     bool       // Whether to wait for a response from the server after declaring the queue.
	Args       amqp.Table // Additional arguments to be passed when declaring the queue.
}

// RabbitMQQueueBindConfig represents the configuration options for binding a queue to an exchange in RabbitMQ.
type RabbitMQQueueBindConfig struct {
	RoutingKeys  []string   // The routing key to use for the binding.
	ExchangeName string     // The name of the exchange to bind to.
	NoWait       bool       // Whether to wait for a response from the server.
	Args         amqp.Table // Additional arguments for the binding.
}

const (
	ConsumerPlatformRabbitMQ     = options.PlatformRabbitMQ
	ConsumerPlatformGooglePubSub = options.PlatformGooglePubSub
	ConsumerPlatformSQS          = options.PlatformSQS
)

// RabbitMQConfigWithDefaultTopicFanOutPattern returns a RabbitMQConsumerConfig with default configuration for topic fanout pattern.
// It takes the queueName, exchangeName, and routingKeys as parameters and returns a pointer to RabbitMQConsumerConfig.
// The default configuration includes a durable queue that is not auto-deleted, exclusive, and no-wait.
// The queue is bound to the exchange with the provided routing keys.
// the exchange is must be a fanout exchange. The exchange must be declared before using this configuration.
// Read more about fanout exchange: https://www.rabbitmq.com/tutorials/amqp-concepts.html#exchange-fanout
//
//	the routing keys can be in pattern format.
//	e.g. "a.*.b.#" will match "a.b", "a.c.b", "a.c.d.b", etc.
//	For more information on pattern matching, see https://www.rabbitmq.com/tutorials/tutorial-five-go.html
func RabbitMQConfigWithDefaultTopicFanOutPattern(consumerChannel, requeueChannel *amqp.Channel,
	exchangeName string, routingKeys []string) *RabbitMQConsumerConfig {
	return &RabbitMQConsumerConfig{
		ConsumerChannel: consumerChannel,
		ReQueueChannel:  requeueChannel,
		QueueDeclareConfig: &RabbitMQQueueDeclareConfig{
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
		QueueBindConfig: &RabbitMQQueueBindConfig{
			RoutingKeys:  routingKeys,
			ExchangeName: exchangeName,
			NoWait:       false,
			Args:         nil,
		},
	}
}

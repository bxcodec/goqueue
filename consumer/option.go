package consumer

import "github.com/bxcodec/goqueue"

const (
	DefaultMaxRetryFailedMessage = 3
	DefaultBatchMessageSize      = 1
)

// Option represents the configuration options for the consumer.
type Option struct {
	// BatchMessageSize specifies the maximum number of messages to be processed in a single batch.
	BatchMessageSize int
	// QueueName specifies the name of the queue to consume messages from.
	QueueName string
	// Middlewares is a list of middleware functions to be applied to the inbound message handler.
	Middlewares              []goqueue.InboundMessageHandlerMiddlewareFunc
	ActionsPatternSubscribed []string
	TopicName                string
	MaxRetryFailedMessage    int64
}

// OptionFunc is a function type that takes an `opt` parameter of type `*Option`.
// It is used as an option for configuring behavior in the `Option` struct.
type OptionFunc func(opt *Option)

// WithBatchMessageSize sets the batch message size option for the consumer.
// It takes an integer value 'n' and returns an OptionFunc that sets the
// BatchMessageSize field of the Option struct to 'n'.
func WithBatchMessageSize(n int) OptionFunc {
	return func(opt *Option) {
		opt.BatchMessageSize = n
	}
}

// WithQueueName sets the queue name for the consumer option.
func WithQueueName(name string) OptionFunc {
	return func(opt *Option) {
		opt.QueueName = name
	}
}

// WithMiddlewares is an OptionFunc that sets the provided middlewares for the consumer.
// Middlewares are used to process inbound messages before they are handled by the consumer.
// The middlewares are applied in the order they are provided.
func WithMiddlewares(middlewares ...goqueue.InboundMessageHandlerMiddlewareFunc) OptionFunc {
	return func(opt *Option) {
		opt.Middlewares = middlewares
	}
}

// WithActionsPatternSubscribed sets the actions that the consumer will subscribe to.
// It takes a variadic parameter `actions` which represents the actions to be subscribed.
// The actions are stored in the `ActionsPatternSubscribed` field of the `Option` struct.
func WithActionsPatternSubscribed(actions ...string) OptionFunc {
	return func(opt *Option) {
		opt.ActionsPatternSubscribed = actions
	}
}

// WithTopicName sets the topic name for the consumer option.
func WithTopicName(name string) OptionFunc {
	return func(opt *Option) {
		opt.TopicName = name
	}
}

// WithMaxRetryFailedMessage sets the maximum number of retries for failed messages.
// It takes an integer parameter 'n' and returns an OptionFunc.
// The OptionFunc updates the 'MaxRetryFailedMessage' field of the Option struct.
func WithMaxRetryFailedMessage(n int64) OptionFunc {
	return func(opt *Option) {
		opt.MaxRetryFailedMessage = n
	}
}

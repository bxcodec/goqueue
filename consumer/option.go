package consumer

import "github.com/bxcodec/goqu"

// Option represents the configuration options for the consumer.
type Option struct {
	// BatchMessageSize specifies the maximum number of messages to be processed in a single batch.
	BatchMessageSize int
	// QueueName specifies the name of the queue to consume messages from.
	QueueName string
	// Middlewares is a list of middleware functions to be applied to the inbound message handler.
	Middlewares []goqu.InboundMessageHandlerMiddlewareFunc
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
func WithMiddlewares(middlewares ...goqu.InboundMessageHandlerMiddlewareFunc) OptionFunc {
	return func(opt *Option) {
		opt.Middlewares = middlewares
	}
}

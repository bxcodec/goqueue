package goqueue

import "context"

// Consumer represents an entity that consumes messages from a queue.
type Consumer interface {
	// Consume consumes messages from the queue and passes them to the provided handler.
	// It takes a context, an InboundMessageHandler, and a map of metadata as parameters.
	// It returns an error if there was a problem consuming the messages.
	Consume(ctx context.Context, handler InboundMessageHandler, meta map[string]interface{}) (err error)

	// Stop stops the consumer from consuming messages.
	// It takes a context as a parameter and returns an error if there was a problem stopping the consumer.
	Stop(ctx context.Context) (err error)
}

type InboundMessageHandler interface {
	HandleMessage(ctx context.Context, m InboundMessage) (err error)
}

type InboundMessageHandlerFunc func(ctx context.Context, m InboundMessage) (err error)

func (mhf InboundMessageHandlerFunc) HandleMessage(ctx context.Context, m InboundMessage) (err error) {
	return mhf(ctx, m)
}

type InboundMessageHandlerMiddlewareFunc func(next InboundMessageHandlerFunc) InboundMessageHandlerFunc

type InboundMessage struct {
	Message
	RetryCount int64                  `json:"retryCount"`
	Metadata   map[string]interface{} `json:"metadata"`
	// Ack is used for confirming the message. It will drop the message from the queue.
	Ack func(ctx context.Context) (err error) `json:"-"`
	// Nack is used for rejecting the message. It will requeue the message to be re-delivered again.
	Nack func(ctx context.Context) (err error) `json:"-"`
	// MoveToDeadLetterQueue is used for rejecting the message same with Nack, but instead of requeueing the message,
	// Read how to configure dead letter queue in each queue provider.
	// eg RabbitMQ: https://www.rabbitmq.com/docs/dlx
	MoveToDeadLetterQueue func(ctx context.Context) (err error) `json:"-"`
	// Requeue is used to put the message back to the tail of the queue after a delay.
	Requeue func(ctx context.Context, delayFn DelayFn) (err error) `json:"-"`
}

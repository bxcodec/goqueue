package interfaces

import "context"

//go:generate mockery --name InboundMessageHandler
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
	// RetryCount is the number of times the message has been retried.
	// This is set by the library to identify the number of times the message has been retried.
	RetryCount int64 `json:"retryCount"`
	// Metadata is the metadata of the message.
	// This is set by the library to identify the metadata of the message.
	Metadata map[string]any `json:"metadata"`
	// Ack is used for confirming the message. It will drop the message from the queue.
	Ack func(ctx context.Context) (err error) `json:"-"`
	// Nack is used for rejecting the message. It will requeue the message to be re-delivered again.
	Nack func(ctx context.Context) (err error) `json:"-"`
	// MoveToDeadLetterQueue is used for rejecting the message same with Nack, but instead of requeueing the message,
	// Read how to configure dead letter queue in each queue provider.
	// eg RabbitMQ: https://www.rabbitmq.com/docs/dlx
	MoveToDeadLetterQueue func(ctx context.Context) (err error) `json:"-"`
	// Requeue is used to put the message back to the tail of the queue after a delay.
	RetryWithDelayFn func(ctx context.Context, delayFn DelayFn) (err error) `json:"-"`
}

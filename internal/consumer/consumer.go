package consumer

import (
	"context"

	"github.com/bxcodec/goqueue/interfaces"
)

// Consumer represents an entity that consumes messages from a queue.
//
//go:generate mockery --name Consumer
type Consumer interface {
	// Consume consumes messages from the queue and passes them to the provided handler.
	// It takes a context, an InboundMessageHandler, and a map of metadata as parameters.
	// It returns an error if there was a problem consuming the messages.
	Consume(ctx context.Context, handler interfaces.InboundMessageHandler, meta map[string]any) (err error)

	// Stop stops the consumer from consuming messages.
	// It takes a context as a parameter and returns an error if there was a problem stopping the consumer.
	Stop(ctx context.Context) (err error)
}

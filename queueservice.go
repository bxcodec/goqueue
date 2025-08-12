package goqueue

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/internal/consumer"
	"github.com/bxcodec/goqueue/internal/publisher"
	"github.com/bxcodec/goqueue/options"
)

// QueueService represents a service that handles message queuing operations.
type QueueService struct {
	consumer         consumer.Consumer                // The consumer responsible for consuming messages from the queue.
	handler          interfaces.InboundMessageHandler // The handler responsible for processing incoming messages.
	publisher        publisher.Publisher              // The publisher responsible for publishing messages to the queue.
	NumberOfConsumer int                              // The number of consumers to process messages concurrently.
}

// NewQueueService creates a new instance of QueueService with the provided options.
// It accepts a variadic parameter `opts` which allows configuring the QueueService.
// The options are applied in the order they are provided.
// Returns a pointer to the created QueueService.
func NewQueueService(opts ...options.GoQueueOptionFunc) *QueueService {
	opt := options.DefaultGoQueueOption()
	for _, o := range opts {
		o(opt)
	}
	return &QueueService{
		consumer:         opt.Consumer,
		handler:          opt.MessageHandler,
		publisher:        opt.Publisher,
		NumberOfConsumer: opt.NumberOfConsumer,
	}
}

// Start starts the queue service by spawning multiple consumers to process messages.
// It returns an error if the consumer or handler is not defined.
// The method uses the provided context to manage the lifecycle of the consumers.
func (qs *QueueService) Start(ctx context.Context) (err error) {
	if qs.consumer == nil {
		return errors.New("consumer is not defined")
	}
	if qs.handler == nil {
		return errors.New("handler is not defined")
	}

	g, ctx := errgroup.WithContext(ctx)
	for i := range qs.NumberOfConsumer {
		meta := map[string]any{
			"consumer_id":  i,
			"started_time": time.Now(),
		}
		g.Go(func() error {
			return qs.consumer.Consume(ctx, qs.handler, meta)
		})
	}

	return g.Wait()
}

// Stop stops the queue service by stopping the consumer and closing the publisher.
// It returns an error if there was an issue stopping the consumer or closing the publisher.
func (qs *QueueService) Stop(ctx context.Context) error {
	if qs.consumer != nil {
		err := qs.consumer.Stop(ctx)
		if err != nil {
			return err
		}
	}

	if qs.publisher != nil {
		err := qs.publisher.Close(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Publish publishes a message to the queue.
// It returns an error if the publisher is not defined or if there was an error while publishing the message.
func (qs *QueueService) Publish(ctx context.Context, m interfaces.Message) (err error) {
	if qs.publisher == nil {
		return errors.New("publisher is not defined")
	}
	return qs.publisher.Publish(ctx, m)
}

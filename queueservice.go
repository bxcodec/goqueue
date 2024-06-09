package goqueue

import (
	"context"
	"errors"
	"time"

	"golang.org/x/sync/errgroup"
)

// QueueService represents a queue service that handles incoming messages.
type QueueService struct {
	Consumer         Consumer              // The consumer responsible for consuming messages from the queue.
	Handler          InboundMessageHandler // The handler responsible for processing incoming messages.
	Publisher        Publisher             // The publisher responsible for publishing messages to the queue.
	NumberOfConsumer int                   // The number of consumers to process messages concurrently.
}

// NewQueueService creates a new instance of QueueService with the provided options.
// It accepts zero or more OptionFunc functions to customize the behavior of the QueueService.
// Returns a pointer to the created QueueService.
func NewQueueService(opts ...OptionFunc) *QueueService {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	return &QueueService{
		Consumer:         opt.Consumer,
		Handler:          opt.MessageHandler,
		Publisher:        opt.Publisher,
		NumberOfConsumer: opt.NumberOfConsumer,
	}
}

// Start starts the queue service by spawning multiple consumers to process messages.
// It returns an error if the consumer or handler is not defined.
// The method uses the provided context to manage the lifecycle of the consumers.
// Each consumer is assigned a unique consumer ID and the start time is recorded in the meta data.
// The method uses the errgroup package to manage the goroutines and waits for all consumers to finish.
func (qs *QueueService) Start(ctx context.Context) (err error) {
	if qs.Consumer == nil {
		return errors.New("consumer is not defined")
	}
	if qs.Handler == nil {
		return errors.New("handler is not defined")
	}

	g, ctx := errgroup.WithContext(ctx)
	for i := 0; i < qs.NumberOfConsumer; i++ {
		meta := map[string]interface{}{
			"consumer_id":  i,
			"started_time": time.Now(),
		}
		g.Go(func() error {
			return qs.Consumer.Consume(ctx, qs.Handler, meta)
		})
	}

	return g.Wait()
}

// Stop stops the queue service by stopping the consumer and closing the publisher.
// It returns an error if there is an issue stopping the consumer or closing the publisher.
func (qs *QueueService) Stop(ctx context.Context) error {
	if qs.Consumer == nil {
		return errors.New("consumer is not defined")
	}
	err := qs.Consumer.Stop(ctx)
	if err != nil {
		return err
	}

	err = qs.Publisher.Close(ctx)
	if err != nil {
		return err
	}
	return nil
}

// Publish sends a message to the queue using the defined publisher.
// It returns an error if the publisher is not defined or if there was an error while publishing the message.
func (qs *QueueService) Publish(ctx context.Context, m Message) (err error) {
	if qs.Publisher == nil {
		return errors.New("publisher is not defined")
	}
	return qs.Publisher.Publish(ctx, m)
}

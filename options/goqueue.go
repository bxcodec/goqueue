package options

import (
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/internal/consumer"
	"github.com/bxcodec/goqueue/internal/publisher"
)

// GoQueueOption represents the options for configuring a GoQueue instance.
type GoQueueOption struct {
	// NumberOfConsumer is the number of consumer/worker/goroutine that will be spawned in one GoQueue instance.
	NumberOfConsumer int
	// Consumer is the consumer implementation that will be used by the GoQueue instance.
	Consumer consumer.Consumer
	// Publisher is the publisher implementation that will be used by the GoQueue instance.
	Publisher publisher.Publisher
	// MessageHandler is the inbound message handler that will be used by the GoQueue instance.
	MessageHandler interfaces.InboundMessageHandler
}

// GoQueueOptionFunc used for option chaining
type GoQueueOptionFunc func(opt *GoQueueOption)

func DefaultGoQueueOption() *GoQueueOption {
	return &GoQueueOption{
		NumberOfConsumer: 1,
	}
}

// WithNumberOfConsumer sets the number of consumers for the GoQueue.
// It takes an integer value 'n' as input and updates the NumberOfConsumer field of the GoQueueOption struct.
// This option determines how many goroutines will be created to consume items from the queue concurrently.
func WithNumberOfConsumer(n int) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.NumberOfConsumer = n
	}
}

// WithConsumer sets the consumer for the GoQueueOption.
// It takes a consumer.Consumer as a parameter and returns a GoQueueOptionFunc.
// The GoQueueOptionFunc sets the Consumer field of the GoQueueOption.
func WithConsumer(c consumer.Consumer) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.Consumer = c
	}
}

// WithPublisher sets the publisher for the GoQueueOption.
// It takes a publisher.Publisher as a parameter and returns a GoQueueOptionFunc.
// The returned GoQueueOptionFunc sets the Publisher field of the GoQueueOption to the provided publisher.
func WithPublisher(p publisher.Publisher) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.Publisher = p
	}
}

// WithMessageHandler sets the inbound message handler for the GoQueueOption.
// The inbound message handler is responsible for processing incoming messages.
// It takes an instance of the interfaces.InboundMessageHandler interface as a parameter.
// Example usage:
//
//	WithMessageHandler(func(message interfaces.InboundMessage) {
//	  // Process the incoming message
//	})
func WithMessageHandler(h interfaces.InboundMessageHandler) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.MessageHandler = h
	}
}

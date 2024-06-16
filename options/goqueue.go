package options

import (
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/internal/consumer"
	"github.com/bxcodec/goqueue/internal/publisher"
)

type GoQueueOption struct {
	// number of consumer/worker/goroutine that will be spawned in one goqueue instance
	NumberOfConsumer int
	Consumer         consumer.Consumer
	Publisher        publisher.Publisher
	MessageHandler   interfaces.InboundMessageHandler
}

// GoQueueOptionFunc used for option chaining
type GoQueueOptionFunc func(opt *GoQueueOption)

func DefaultGoQueueOption() *GoQueueOption {
	return &GoQueueOption{
		NumberOfConsumer: 1,
	}
}

func WithNumberOfConsumer(n int) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.NumberOfConsumer = n
	}
}

func WithConsumer(c consumer.Consumer) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.Consumer = c
	}
}

func WithPublisher(p publisher.Publisher) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.Publisher = p
	}
}

func WithMessageHandler(h interfaces.InboundMessageHandler) GoQueueOptionFunc {
	return func(opt *GoQueueOption) {
		opt.MessageHandler = h
	}
}

package goqueue

type Option struct {
	// number of consumer/worker/goroutine that will be spawned in one goqueue instance
	NumberOfConsumer int
	Consumer         Consumer
	Publisher        Publisher
	MessageHandler   InboundMessageHandler
}

// OptionFunc used for option chaining
type OptionFunc func(opt *Option)

func defaultOption() *Option {
	return &Option{
		NumberOfConsumer: 1,
	}
}

func WithNumberOfConsumer(n int) OptionFunc {
	return func(opt *Option) {
		opt.NumberOfConsumer = n
	}
}

func WithConsumer(c Consumer) OptionFunc {
	return func(opt *Option) {
		opt.Consumer = c
	}
}

func WithPublisher(p Publisher) OptionFunc {
	return func(opt *Option) {
		opt.Publisher = p
	}
}

func WithMessageHandler(h InboundMessageHandler) OptionFunc {
	return func(opt *Option) {
		opt.MessageHandler = h
	}
}

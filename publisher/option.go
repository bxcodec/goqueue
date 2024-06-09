package publisher

import "github.com/bxcodec/goqu"

// Option define the option property
type Option struct {
	PublisherID string
	Middlewares []goqu.PublisherMiddlewareFunc
}

// OptionFunc used for option chaining
type OptionFunc func(opt *Option)

func WithPublisherID(id string) OptionFunc {
	return func(opt *Option) {
		opt.PublisherID = id
	}
}

func WithMiddlewares(middlewares ...goqu.PublisherMiddlewareFunc) OptionFunc {
	return func(opt *Option) {
		opt.Middlewares = middlewares
	}
}

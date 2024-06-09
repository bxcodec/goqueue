package publisher

import (
	"github.com/bxcodec/goqueue"
	headerVal "github.com/bxcodec/goqueue/headers/value"
)

const (
	DefaultContentType = headerVal.ContentTypeJSON
)

// Option define the option property
type Option struct {
	PublisherID string
	Middlewares []goqueue.PublisherMiddlewareFunc
}

// OptionFunc used for option chaining
type OptionFunc func(opt *Option)

func WithPublisherID(id string) OptionFunc {
	return func(opt *Option) {
		opt.PublisherID = id
	}
}

func WithMiddlewares(middlewares ...goqueue.PublisherMiddlewareFunc) OptionFunc {
	return func(opt *Option) {
		opt.Middlewares = middlewares
	}
}

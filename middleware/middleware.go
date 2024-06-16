package middleware

import (
	"github.com/bxcodec/goqueue/interfaces"
)

// ApplyHandlerMiddleware applies a series of middlewares to an inbound message handler function.
// It takes an inbound message handler function and a variadic list of middlewares as input.
// Each middleware is applied to the handler function in the order they are provided.
// The resulting function is returned as the final handler function with all the middlewares applied.
func ApplyHandlerMiddleware(h interfaces.InboundMessageHandlerFunc,
	middlewares ...interfaces.InboundMessageHandlerMiddlewareFunc) interfaces.InboundMessageHandlerFunc {
	for _, middleware := range middlewares {
		h = middleware(h)
	}
	return h
}

// ApplyPublisherMiddleware applies a series of middlewares to a publisher function.
// It takes a publisher function and a variadic list of publisher middleware functions as input.
// Each middleware function is applied to the publisher function in the order they are provided.
// The resulting publisher function is returned.
func ApplyPublisherMiddleware(p interfaces.PublisherFunc,
	middlewares ...interfaces.PublisherMiddlewareFunc) interfaces.PublisherFunc {
	for _, middleware := range middlewares {
		p = middleware(p)
	}
	return p
}

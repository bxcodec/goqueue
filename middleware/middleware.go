package middleware

import (
	"context"

	"github.com/bxcodec/goqu"
	"github.com/sirupsen/logrus"
)

// ApplyHandlerMiddleware applies a series of middleware functions to an inbound message handler function.
// It takes an inbound message handler function `h` and a variadic list of middleware functions `middleware`.
// Each middleware function is applied to the handler function in the order they are provided.
// The resulting handler function with all the middleware applied is returned.
func ApplyHandlerMiddleware(h goqu.InboundMessageHandlerFunc, middleware ...goqu.InboundMessageHandlerMiddlewareFunc) goqu.InboundMessageHandlerFunc {
	for _, middleware := range middleware {
		h = middleware(h)
	}
	return h
}

// ApplyPublisherMiddleware applies the given publisher middleware functions to the provided publisher function.
// It iterates over the middleware functions and applies them in the order they are provided.
// The resulting publisher function is returned.
func ApplyPublisherMiddleware(p goqu.PublisherFunc, middleware ...goqu.PublisherMiddlewareFunc) goqu.PublisherFunc {
	for _, middleware := range middleware {
		p = middleware(p)
	}
	return p
}

// HelloWorldMiddlewareExecuteAfterHandler returns an inbound message handler middleware function.
// It wraps the provided `next` inbound message handler function and executes some additional logic after it.
// The additional logic includes logging any error that occurred during the execution of the `next` function
// and logging a message indicating that the middleware has been executed.
func HelloWorldMiddlewareExecuteAfterHandler() goqu.InboundMessageHandlerMiddlewareFunc {
	return func(next goqu.InboundMessageHandlerFunc) goqu.InboundMessageHandlerFunc {
		return func(ctx context.Context, m goqu.InboundMessage) (err error) {
			err = next(ctx, m)
			if err != nil {
				logrus.Error("Error: ", err, "processing to sent the error to Sentry")
			}
			logrus.Info("hello-world-last-middleware executed")
			return err
		}
	}
}

// HelloWorldMiddlewareExecuteBeforeHandler returns a middleware function that logs a message before executing the handler.
func HelloWorldMiddlewareExecuteBeforeHandler() goqu.InboundMessageHandlerMiddlewareFunc {
	return func(next goqu.InboundMessageHandlerFunc) goqu.InboundMessageHandlerFunc {
		return func(ctx context.Context, m goqu.InboundMessage) (err error) {
			logrus.Info("hello-world-first-middleware executed")
			return next(ctx, m)
		}
	}
}

// HelloWorldMiddlewareExecuteAfterPublisher returns a PublisherMiddlewareFunc that executes after the publisher function.
// It logs any error that occurs during publishing and logs a message indicating that the last middleware has been executed.
func HelloWorldMiddlewareExecuteAfterPublisher() goqu.PublisherMiddlewareFunc {
	return func(next goqu.PublisherFunc) goqu.PublisherFunc {
		return func(ctx context.Context, m goqu.Message) (err error) {
			err = next(ctx, m)
			if err != nil {
				logrus.Error("got error while publishing the message: ", err)
				return err
			}
			logrus.Info("hello-world-last-middleware executed")
			return nil
		}
	}
}

// HelloWorldMiddlewareExecuteBeforePublisher is a function that returns a PublisherMiddlewareFunc.
// It wraps the provided PublisherFunc with a middleware that logs a message before executing the next function.
func HelloWorldMiddlewareExecuteBeforePublisher() goqu.PublisherMiddlewareFunc {
	return func(next goqu.PublisherFunc) goqu.PublisherFunc {
		return func(ctx context.Context, e goqu.Message) (err error) {
			logrus.Info("hello-world-first-middleware executed")
			return next(ctx, e)
		}
	}
}

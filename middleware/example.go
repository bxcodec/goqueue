package middleware

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/bxcodec/goqueue/interfaces"
)

// HelloWorldMiddlewareExecuteAfterInboundMessageHandler returns an inbound message handler middleware function.
// This middleware function executes after the inbound message handler and performs additional tasks.
// It logs any errors that occur during the execution of the next handler and provides an opportunity to handle them.
// You can customize the error handling logic by adding your own error handler,
// such as sending errors to Sentry or other error tracking tools.
// After error handling, it logs a message indicating that the hello-world-last-middleware has been executed.
// The function signature follows the `interfaces.InboundMessageHandlerMiddlewareFunc` type.
func HelloWorldMiddlewareExecuteAfterInboundMessageHandler() interfaces.InboundMessageHandlerMiddlewareFunc {
	return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
		return func(ctx context.Context, m interfaces.InboundMessage) (err error) {
			err = next(ctx, m)
			if err != nil {
				log.Error().Err(err).Msg("Error: add your custom error handler here, eg send to Sentry or other error tracking tools")
			}
			log.Info().Msg("hello-world-last-middleware executed")
			return err
		}
	}
}

// HelloWorldMiddlewareExecuteBeforeInboundMessageHandler is a function that returns an inbound message handler middleware.
// This middleware logs a message and then calls the next inbound message handler in the chain.
func HelloWorldMiddlewareExecuteBeforeInboundMessageHandler() interfaces.InboundMessageHandlerMiddlewareFunc {
	return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
		return func(ctx context.Context, m interfaces.InboundMessage) (err error) {
			log.Info().Msg("hello-world-first-middleware executed")
			return next(ctx, m)
		}
	}
}

// HelloWorldMiddlewareExecuteAfterPublisher is a function that returns a PublisherMiddlewareFunc.
// It wraps the provided PublisherFunc with additional functionality to be executed after publishing a message.
func HelloWorldMiddlewareExecuteAfterPublisher() interfaces.PublisherMiddlewareFunc {
	return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
		return func(ctx context.Context, m interfaces.Message) (err error) {
			err = next(ctx, m)
			if err != nil {
				log.Error().Err(err).Msg("got error while publishing the message")
				return err
			}
			log.Info().Msg("hello-world-last-middleware executed")
			return nil
		}
	}
}

// HelloWorldMiddlewareExecuteBeforePublisher is a function that returns a PublisherMiddlewareFunc.
// It wraps the provided PublisherFunc with a middleware that logs a message before executing
// the next middleware or the actual publisher function.
func HelloWorldMiddlewareExecuteBeforePublisher() interfaces.PublisherMiddlewareFunc {
	return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
		return func(ctx context.Context, e interfaces.Message) (err error) {
			log.Info().Msg("hello-world-first-middleware executed")
			return next(ctx, e)
		}
	}
}

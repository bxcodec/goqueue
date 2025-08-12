package middleware

import (
	"context"
	"errors"

	goqueueErrors "github.com/bxcodec/goqueue/errors"
	"github.com/bxcodec/goqueue/interfaces"
)

// mapError maps generic errors to specific goqueue error types
func mapError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, goqueueErrors.ErrInvalidMessageFormat):
		return goqueueErrors.ErrInvalidMessageFormat
	case errors.Is(err, goqueueErrors.ErrEncodingFormatNotSupported):
		return goqueueErrors.ErrEncodingFormatNotSupported
	default:
		return goqueueErrors.Error{
			Code:    goqueueErrors.UnKnownError,
			Message: err.Error(),
		}
	}
}

// PublisherDefaultErrorMapper returns a middleware function that maps publisher errors to specific error types.
// It takes a next PublisherFunc as input and returns a new PublisherFunc that performs error mapping.
// If an error occurs during publishing, it will be mapped to a specific error type based on the error code.
// The mapped error will be returned, or nil if no error occurred.
func PublisherDefaultErrorMapper() interfaces.PublisherMiddlewareFunc {
	return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
		return func(ctx context.Context, e interfaces.Message) (err error) {
			err = next(ctx, e)
			return mapError(err)
		}
	}
}

// InboundMessageHandlerDefaultErrorMapper returns a middleware function that maps specific errors to predefined error types.
// It takes the next inbound message handler function as input and returns a new inbound message handler function.
// The returned function checks if an error occurred during the execution of the next handler function.
// If an error is found, it maps the error to a predefined error type and returns it.
// If no error is found, it returns nil.
func InboundMessageHandlerDefaultErrorMapper() interfaces.InboundMessageHandlerMiddlewareFunc {
	return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
		return func(ctx context.Context, m interfaces.InboundMessage) (err error) {
			err = next(ctx, m)
			return mapError(err)
		}
	}
}

package interfaces

import "context"

// PublisherHandler is an interface that defines the behavior of a message publisher.
//
//go:generate mockery --name PublisherHandler
type PublisherHandler interface {
	Publish(ctx context.Context, m Message) (err error)
}

// PublisherFunc is a function type that represents a publisher function.
// It takes a context and a message as input parameters and returns an error.
type PublisherFunc func(ctx context.Context, m Message) (err error)

// Publish sends the given message using the provided context.
// It calls the underlying PublisherFunc to perform the actual publishing.
// If an error occurs during publishing, it is returned.
func (f PublisherFunc) Publish(ctx context.Context, m Message) (err error) {
	return f(ctx, m)
}

// PublisherMiddlewareFunc is a function type that represents a publisher middleware function.
// It takes a next PublisherFunc as input parameter and returns a PublisherFunc.
type PublisherMiddlewareFunc func(next PublisherFunc) PublisherFunc

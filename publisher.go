package goqueue

import "context"

// Publisher represents an interface for publishing messages.
type Publisher interface {
	PublisherHandler
	Close(ctx context.Context) (err error)
}

// PublisherHandler is an interface that defines the behavior of a message publisher.
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

type PublisherMiddlewareFunc func(next PublisherFunc) PublisherFunc

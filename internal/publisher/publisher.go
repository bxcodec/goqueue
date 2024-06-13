package publisher

import (
	"context"

	"github.com/bxcodec/goqueue/interfaces"
)

// Publisher represents an interface for publishing messages.
//
//go:generate mockery --name Publisher
type Publisher interface {
	interfaces.PublisherHandler
	Close(ctx context.Context) (err error)
}

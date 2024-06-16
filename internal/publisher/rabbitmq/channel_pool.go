package rabbitmq

import (
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/multierr"
)

// ChannelPool represents a pool of AMQP channels used for publishing messages.
// It provides a way to manage and reuse AMQP channels efficiently.
type ChannelPool struct {
	conn    *amqp.Connection
	mutex   sync.Mutex
	pool    chan *amqp.Channel
	maxSize int
}

// NewChannelPool creates a new ChannelPool instance.
// It takes an AMQP connection and the maximum size of the pool as parameters.
// It returns a pointer to the newly created ChannelPool.
func NewChannelPool(conn *amqp.Connection, maxSize int) *ChannelPool {
	return &ChannelPool{
		conn:    conn,
		maxSize: maxSize,
		pool:    make(chan *amqp.Channel, maxSize),
	}
}

// Get returns a channel from the pool. If there are available channels in the pool, it returns one of them.
// Otherwise, it creates a new channel from the underlying connection and returns it.
func (cp *ChannelPool) Get() (*amqp.Channel, error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	select {
	case ch := <-cp.pool:
		return ch, nil
	default:
		return cp.conn.Channel()
	}
}

// Return returns a channel back to the channel pool.
// If the pool is full, the channel is closed.
func (cp *ChannelPool) Return(ch *amqp.Channel) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	select {
	case cp.pool <- ch:
		// channel returned to pool
	default:
		// pool is full, close the channel
		err := ch.Close()
		if err != nil {
			log.Printf("Error closing RabbitMQ channel: %s", err)
		}
	}
}

// Close closes the ChannelPool and all its associated channels.
// It returns an error if there was an error closing any of the channels.
func (cp *ChannelPool) Close() (err error) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	var errs = make([]error, 0)
	close(cp.pool)
	for ch := range cp.pool {
		err := ch.Close()
		if err != nil {
			log.Printf("Error closing channel: %s", err)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		log.Printf("Error closing channels: %v", errs)
		return multierr.Combine(errs...)
	}
	return nil
}

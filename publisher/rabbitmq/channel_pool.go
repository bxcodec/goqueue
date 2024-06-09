package rabbitmq

import (
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/multierr"
)

type ChannelPool struct {
	conn    *amqp.Connection
	mutex   sync.Mutex
	pool    chan *amqp.Channel
	maxSize int
}

func NewChannelPool(conn *amqp.Connection, maxSize int) *ChannelPool {
	return &ChannelPool{
		conn:    conn,
		maxSize: maxSize,
		pool:    make(chan *amqp.Channel, maxSize),
	}
}

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

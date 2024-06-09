package rabbitmq

import (
	"context"
	"time"

	"github.com/bxcodec/goqueue"
	headerKey "github.com/bxcodec/goqueue/headers/key"
	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/middleware"
	"github.com/bxcodec/goqueue/publisher"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
)

const (
	DefaultChannelPoolSize        = 5
	ExchangeType           string = "topic"
	DefaultContentType     string = "application/json"
)

type rabbitMQ struct {
	channelPool *ChannelPool
	option      *publisher.Option
}

var defaultOption = func() *publisher.Option {
	return &publisher.Option{
		Middlewares: []goqueue.PublisherMiddlewareFunc{},
		PublisherID: uuid.New().String(),
	}
}

func NewPublisher(
	conn *amqp.Connection,
	opts ...publisher.OptionFunc,
) goqueue.Publisher {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	channelPool := NewChannelPool(conn, DefaultChannelPoolSize)
	ch, err := channelPool.Get()
	if err != nil {
		logrus.Fatal(err)
	}
	defer channelPool.Return(ch)

	return &rabbitMQ{
		channelPool: channelPool,
		option:      opt,
	}
}

func (r *rabbitMQ) Publish(ctx context.Context, m goqueue.Message) (err error) {
	if m.ContentType == "" {
		m.ContentType = publisher.DefaultContentType
	}
	publishFunc := middleware.ApplyPublisherMiddleware(
		r.buildPublisher(),
		r.option.Middlewares...,
	)
	return publishFunc(ctx, m)
}

func (r *rabbitMQ) buildPublisher() goqueue.PublisherFunc {
	return func(ctx context.Context, m goqueue.Message) (err error) {
		id := m.ID
		if id == "" {
			id = uuid.New().String()
		}

		timestamp := m.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		defaultHeaders := map[string]interface{}{
			headerKey.AppID:              r.option.PublisherID,
			headerKey.MessageID:          id,
			headerKey.PublishedTimestamp: timestamp.Format(time.RFC3339),
			headerKey.RetryCount:         0,
			headerKey.ContentType:        string(m.ContentType),
			headerKey.QueueServiceAgent:  string(headerVal.RabbitMQ),
			headerKey.SchemaVer:          headerVal.GoquMessageSchemaVersionV1,
		}

		headers := amqp.Table{}
		for key, value := range defaultHeaders {
			headers[key] = value
		}
		for key, value := range m.Headers {
			headers[key] = value
		}

		m.Headers = headers
		m.ServiceAgent = headerVal.RabbitMQ
		m.Timestamp = timestamp
		m.ID = id
		encoder, ok := goqueue.GetGoQueueEncoding(m.ContentType)
		if !ok {
			encoder = goqueue.DefaultEncoding
		}

		data, err := encoder.Encode(ctx, m)
		if err != nil {
			return err
		}

		ch, err := r.channelPool.Get()
		if err != nil {
			return err
		}
		defer r.channelPool.Return(ch)
		return ch.PublishWithContext(
			ctx,
			m.Topic,  // exchange
			m.Action, // routing-key
			false,    // mandatory
			false,    // immediate
			amqp.Publishing{
				Headers:     headers,
				ContentType: string(m.ContentType),
				Body:        data,
				Timestamp:   timestamp,
				AppId:       r.option.PublisherID,
			},
		)
	}
}

// Close will close the connection
func (r *rabbitMQ) Close(_ context.Context) (err error) {
	return r.channelPool.Close()
}

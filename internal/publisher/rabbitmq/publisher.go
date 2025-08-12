package rabbitmq

import (
	"context"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"

	"github.com/bxcodec/goqueue"
	"github.com/bxcodec/goqueue/errors"
	headerKey "github.com/bxcodec/goqueue/headers/key"
	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/internal/publisher"
	"github.com/bxcodec/goqueue/middleware"
	publisherOpts "github.com/bxcodec/goqueue/options/publisher"
)

const (
	DefaultChannelPoolSize = 5
)

type rabbitMQ struct {
	channelPool *ChannelPool
	option      *publisherOpts.PublisherOption
}

// NewPublisher creates a new instance of the publisher.Publisher interface
// using the provided options. It returns a publisher.Publisher implementation
// that utilizes RabbitMQ as the underlying message broker.
//
// The function accepts a variadic parameter `opts` of type
// `publisherOpts.PublisherOptionFunc`, which allows the caller to provide
// custom configuration options for the publisher.
//
// Example usage:
//
//	publisher := NewPublisher(
//					publisherOpts.PublisherPlatformRabbitMQ,
//					publisherOpts.WithRabbitMQPublisherConfig(&publisherOpts.RabbitMQPublisherConfig{
//						Conn:                     rmqConn,
//						PublisherChannelPoolSize: 5,
//						}),
//					publisherOpts.WithPublisherID("publisher_id"),
//					publisherOpts.WithMiddlewares(
//						middleware.HelloWorldMiddlewareExecuteBeforePublisher(),
//						middleware.HelloWorldMiddlewareExecuteAfterPublisher(),
//					),
//
// )
//
// The returned publisher can be used to publish messages to the configured
// RabbitMQ exchange and routing key.
func NewPublisher(
	opts ...publisherOpts.PublisherOptionFunc,
) publisher.Publisher {
	opt := publisherOpts.DefaultPublisherOption()
	for _, o := range opts {
		o(opt)
	}

	conn := opt.RabbitMQPublisherConfig.Conn
	channelPool := NewChannelPool(conn, opt.RabbitMQPublisherConfig.PublisherChannelPoolSize)
	ch, err := channelPool.Get()
	if err != nil {
		log.Fatal().Err(err).Msg("error getting channel from pool")
	}
	defer channelPool.Return(ch)

	return &rabbitMQ{
		channelPool: channelPool,
		option:      opt,
	}
}

// Publish sends a message to the RabbitMQ exchange.
// It applies the default content type if not specified in the message.
// It also applies any registered middlewares before publishing the message.
func (r *rabbitMQ) Publish(ctx context.Context, m interfaces.Message) (err error) {
	if m.ContentType == "" {
		m.ContentType = publisherOpts.DefaultContentType
	}
	publishFunc := middleware.ApplyPublisherMiddleware(
		r.buildPublisher(),
		r.option.Middlewares...,
	)
	return publishFunc(ctx, m)
}

func (r *rabbitMQ) buildPublisher() interfaces.PublisherFunc {
	return func(ctx context.Context, m interfaces.Message) (err error) {
		id := m.ID
		if id == "" {
			id = uuid.New().String()
		}

		timestamp := m.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		defaultHeaders := map[string]any{
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
			return errors.ErrEncodingFormatNotSupported
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

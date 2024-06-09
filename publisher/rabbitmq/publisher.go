package rabbitmq

import (
	"context"
	"time"

	"github.com/bxcodec/goqu"
	headerKey "github.com/bxcodec/goqu/headers/key"
	headerVal "github.com/bxcodec/goqu/headers/value"
	"github.com/bxcodec/goqu/middleware"
	"github.com/bxcodec/goqu/publisher"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type rabbitMQ struct {
	publisherChannel *amqp.Channel
	option           *publisher.Option
}

var defaultOption = func() *publisher.Option {
	return &publisher.Option{
		Middlewares: []goqu.PublisherMiddlewareFunc{},
		PublisherID: uuid.New().String(),
	}
}

func New(
	publisherChannel *amqp.Channel,
	opts ...publisher.OptionFunc,
) goqu.Publisher {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	return &rabbitMQ{
		publisherChannel: publisherChannel,
		option:           opt,
	}
}

func (r *rabbitMQ) Publish(ctx context.Context, m goqu.Message) (err error) {
	publishFunc := middleware.ApplyPublisherMiddleware(
		r.buildPublisher(),
		r.option.Middlewares...,
	)
	return publishFunc(ctx, m)
}

func (r *rabbitMQ) buildPublisher() goqu.PublisherFunc {
	return func(ctx context.Context, m goqu.Message) (err error) {
		data, err := goqu.GetGoquEncoding(m.ContentType).Encode(ctx, m)
		if err != nil {
			return err
		}

		id := m.ID
		if id == "" {
			id = uuid.New().String()
		}

		timestamp := m.Timestamp
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		defaultHeaders := map[string]interface{}{
			headerKey.MessageID:          id,
			headerKey.PublishedTimestamp: timestamp.Format(time.RFC3339),
			headerKey.RetryCount:         0,
			headerKey.ContentType:        m.ContentType,
			headerKey.QueueServiceAgent:  headerVal.RabbitMQ,
			headerKey.SchemaVer:          headerVal.GoquMessageSchemaVersionV1,
		}

		headers := amqp.Table{}
		for key, value := range defaultHeaders {
			headers[key] = value
		}
		for key, value := range m.Headers {
			headers[key] = value
		}

		return r.publisherChannel.PublishWithContext(
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
	return r.publisherChannel.Close()
}

package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"github.com/bxcodec/goqu"
	"github.com/bxcodec/goqu/consumer"
	headerKey "github.com/bxcodec/goqu/headers/key"
	headerVal "github.com/bxcodec/goqu/headers/value"
	"github.com/bxcodec/goqu/middleware"
)

// rabbitMQ is the subscriber handler for rabbitmq
type rabbitMQ struct {
	client          *amqp.Connection
	consumerChannel *amqp.Channel
	requeueChannel  *amqp.Channel //if want requeue support to another queue
	option          *consumer.Option
	tagName         string
	msgReceiver     <-chan amqp.Delivery
}

var defaultOption = func() *consumer.Option {
	return &consumer.Option{
		Middlewares: []goqu.InboundMessageHandlerMiddlewareFunc{},
	}
}

// New will initialize the rabbitMQ subscriber
func NewConsumer(
	client *amqp.Connection,
	consumerChannel *amqp.Channel,
	requeueChannel *amqp.Channel,
	opts ...consumer.OptionFunc) goqu.Consumer {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}
	rmqHandler := &rabbitMQ{
		client:          client,
		consumerChannel: consumerChannel,
		requeueChannel:  requeueChannel,
		option:          opt,
	}
	rmqHandler.initConsumer()
	return rmqHandler
}

// initConsumer initializes the consumer for the RabbitMQ instance.
// It sets the prefetch count, creates a consumer channel, and starts consuming messages from the queue.
// The consumer tag is generated using the queue name and a unique identifier.
// It uses manual ack to avoid message loss and allows multiple consumers to access the queue with round-robin message distribution.
// If an error occurs during the initialization, it logs the error and exits the program.
func (r *rabbitMQ) initConsumer() {
	r.tagName = fmt.Sprintf("%s-%s", r.option.QueueName, uuid.New().String())
	err := r.consumerChannel.Qos(r.option.BatchMessageSize, 0, false)
	if err != nil {
		logrus.Fatal("error when setting the prefetch count, ", err)
	}

	receiver, err := r.consumerChannel.Consume(
		// name of the queue
		r.option.QueueName,
		// consumerTag (unique consumer tag)
		r.tagName,
		// noAck
		//  - false = manual ack,
		//  - true = auto ack. We use manual ack to avoid message loss
		false,
		// exclusive:
		//  - true = only this consumer,
		//  - false = multiple consumer can access the queue and message distribution will be round-robin
		false,
		// noLocal //by default not supported by rabbitMQ from the library notes
		//https://pkg.go.dev/github.com/rabbitmq/amqp091-go@v1.9.0#Channel.Consume
		false,
		// noWait
		false,
		// queue arguments
		// TODO(bxcodec): to support custom queue arguments on consumer initialization
		// https://github.com/bxcodec/goqu/issues/1
		nil,
	)
	if err != nil {
		logrus.Fatal(err, "error consuming message")
	}
	r.msgReceiver = receiver
}

// Consume consumes messages from a RabbitMQ queue.
// It takes a context, an inbound message handler, and metadata as input parameters.
// The method continuously listens for messages from the queue and handles them using the provided handler.
// If the context is canceled, the method stops consuming messages and returns.
// The method returns an error if there was an issue consuming messages.
func (r *rabbitMQ) Consume(ctx context.Context,
	h goqu.InboundMessageHandler,
	meta map[string]interface{}) (err error) {
	logrus.WithFields(logrus.Fields{
		"queue_name":    r.option.QueueName,
		"consumer_meta": meta,
	}).Info("starting the worker")

	for {
		select {
		case <-ctx.Done():
			logrus.WithFields(logrus.Fields{
				"queue_name":    r.option.QueueName,
				"consumer_meta": meta,
			}).Info("stopping the worker")
			return
		case receivedMsg := <-r.msgReceiver:
			msg := &goqu.Message{
				ID:           extractHeaderString(receivedMsg.Headers, headerKey.AppID),
				Timestamp:    extractHeaderTime(receivedMsg.Headers, headerKey.PublishedTimestamp),
				Action:       receivedMsg.RoutingKey,
				Topic:        receivedMsg.Exchange,
				ContentType:  headerVal.ContentType(extractHeaderString(receivedMsg.Headers, headerKey.ContentType)),
				Headers:      receivedMsg.Headers,
				Data:         receivedMsg.Body,
				ServiceAgent: headerVal.GoquServiceAgent(extractHeaderString(receivedMsg.Headers, headerKey.QueueServiceAgent)),
			}
			msg.SetSchemaVersion(extractHeaderString(receivedMsg.Headers, headerKey.SchemaVer))
			m := goqu.InboundMessage{
				Message:    *msg,
				RetryCount: extractHeaderInt(receivedMsg.Headers, headerKey.RetryCount),
				Metadata: map[string]interface{}{
					"app-id":           receivedMsg.AppId,
					"consumer-tag":     receivedMsg.ConsumerTag,
					"content-encoding": receivedMsg.ContentEncoding,
					"content-type":     receivedMsg.ContentType,
					"correlation-id":   receivedMsg.CorrelationId,
					"receivedMsg-mode": receivedMsg.DeliveryMode,
					"receivedMsg-tag":  receivedMsg.DeliveryTag,
					"expiration":       receivedMsg.Expiration,
					"message-count":    receivedMsg.MessageCount,
					"priority":         receivedMsg.Priority,
					"redelivered":      receivedMsg.Redelivered,
					"reply-to":         receivedMsg.ReplyTo,
					"type":             receivedMsg.Type,
					"user-id":          receivedMsg.UserId,
				},
				Ack: func(ctx context.Context) (err error) {
					err = receivedMsg.Ack(false)
					return
				},
				Nack: func(ctx context.Context) (err error) {
					//  receivedMsg.Nack(true) => will redelivered again instantly (same with receivedMsg.reject)
					//  receivedMsg.Nack(false) => will put the message to dead letter queue (same with receivedMsg.reject)
					err = receivedMsg.Nack(false, true)
					return
				},
				MoveToDeadLetterQueue: func(ctx context.Context) (err error) {
					//  receivedMsg.Nack(true) => will redelivered again instantly (same with receivedMsg.reject)
					//  receivedMsg.Nack(false) => will put the message to dead letter queue (same with receivedMsg.reject)
					err = receivedMsg.Nack(false, false)
					return
				},
				Requeue: func(ctx context.Context, delayFn goqu.DelayFn) (err error) {
					if delayFn == nil {
						delayFn = goqu.DefaultDelayFn
					}
					retries := extractHeaderInt(receivedMsg.Headers, headerKey.RetryCount)
					retries++
					go func() {
						delay := delayFn(retries)
						time.Sleep(time.Duration(delay) * time.Second)
						headers := receivedMsg.Headers
						headers[string(headerKey.RetryCount)] = retries
						requeueErr := r.requeueChannel.PublishWithContext( //nolint:staticcheck
							ctx,
							receivedMsg.Exchange,
							receivedMsg.RoutingKey,
							false,
							false,
							amqp.Publishing{
								Headers:     headers,
								ContentType: receivedMsg.ContentType,
								Body:        receivedMsg.Body,
								Timestamp:   time.Now(),
								AppId:       r.tagName,
							},
						)
						if err != nil {
							logrus.Error("failed to requeue the message, got err: ", requeueErr)
						}
					}()
					err = receivedMsg.Ack(false)
					return
				},
			}

			logrus.WithFields(logrus.Fields{
				"consumer_meta": meta,
				"message_id":    msg.ID,
				"topic":         msg.Topic,
				"action":        msg.Action,
				"timestamp":     msg.Timestamp,
			}).Info("message received")
			handleCtx := middleware.ApplyHandlerMiddleware(h.HandleMessage, r.option.Middlewares...)
			err := handleCtx(ctx, m)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"consumer_meta": meta,
					"message_id":    msg.ID,
					"topic":         msg.Topic,
					"action":        msg.Action,
					"timestamp":     msg.Timestamp,
				}).Error("error handling message, ", err)
			}
		}
	}
}

func extractHeaderString(headers amqp.Table, key string) string {
	val, ok := headers[key]
	if !ok {
		return ""
	}

	res, _ := val.(string)
	return res
}

func extractHeaderInt(headers amqp.Table, key string) int64 {
	val, ok := headers[key]
	if !ok {
		return 0
	}

	var res int64
	_, err := fmt.Sscanf(fmt.Sprintf("%v", val), "%d", &res)
	if err != nil {
		return 0
	}
	return res
}

func extractHeaderTime(headers amqp.Table, key string) time.Time {
	val, ok := headers[key]
	if !ok {
		return time.Now()
	}

	switch res := val.(type) {
	case time.Time:
		return res
	case *time.Time:
		return *res
	case string:
		parsedTime, err := time.Parse(time.RFC3339, res)
		if err != nil {
			return time.Now()
		}
		return parsedTime
	default:
		return time.Now()
	}
}

// Stop stops the RabbitMQ consumer.
// It cancels the consumer channel and closes the channel connection.
// If an error occurs during the cancellation or closing process, it is returned.
func (r *rabbitMQ) Stop(_ context.Context) (err error) {
	err = r.consumerChannel.Cancel(r.tagName, false)
	if err != nil {
		return
	}

	err = r.consumerChannel.Close()
	if err != nil {
		return
	}
	return
}

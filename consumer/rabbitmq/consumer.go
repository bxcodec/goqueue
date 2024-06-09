package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"github.com/bxcodec/goqueue"
	"github.com/bxcodec/goqueue/consumer"
	"github.com/bxcodec/goqueue/errors"
	headerKey "github.com/bxcodec/goqueue/headers/key"
	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/middleware"
)

// rabbitMQ is the subscriber handler for rabbitmq
type rabbitMQ struct {
	consumerChannel *amqp.Channel
	requeueChannel  *amqp.Channel //if want requeue support to another queue
	option          *consumer.Option
	tagName         string
	msgReceiver     <-chan amqp.Delivery
}

var defaultOption = func() *consumer.Option {
	return &consumer.Option{
		Middlewares:           []goqueue.InboundMessageHandlerMiddlewareFunc{},
		BatchMessageSize:      consumer.DefaultBatchMessageSize,
		MaxRetryFailedMessage: consumer.DefaultMaxRetryFailedMessage,
	}
}

// New will initialize the rabbitMQ subscriber
func NewConsumer(
	consumerChannel *amqp.Channel,
	requeueChannel *amqp.Channel,
	opts ...consumer.OptionFunc) goqueue.Consumer {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	rmqHandler := &rabbitMQ{
		consumerChannel: consumerChannel,
		requeueChannel:  requeueChannel,
		option:          opt,
	}
	if len(opt.ActionsPatternSubscribed) > 0 && opt.TopicName != "" {
		rmqHandler.initQueue()
	}
	rmqHandler.initConsumer()
	return rmqHandler
}

func (r *rabbitMQ) initQueue() {
	// QueueDeclare declares a queue on the consumer's channel with the specified parameters.
	// It takes the queue name, durable, autoDelete, exclusive, noWait, and arguments as arguments.
	// Returns an instance of amqp.Queue and any error encountered.
	// Please refer to the RabbitMQ documentation for more information on the parameters.
	// https://www.rabbitmq.com/amqp-0-9-1-reference.html#queue.declare
	// (from bxcodec: please raise a PR if you need a custom queue argument)
	_, err := r.consumerChannel.QueueDeclare(
		r.option.QueueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		logrus.Fatal("error declaring the queue, ", err)
	}

	for _, eventType := range r.option.ActionsPatternSubscribed {
		err = r.consumerChannel.QueueBind(
			r.option.QueueName,
			eventType,
			r.option.TopicName,
			false,
			nil,
		)
		if err != nil {
			logrus.Fatal("error binding the queue, ", err)
		}
	}
}

// initConsumer initializes the consumer for the RabbitMQ instance.
// It sets the prefetch count, creates a consumer channel, and starts consuming messages from the queue.
// The consumer tag is generated using the queue name and a unique identifier.
// It uses manual ack to avoid message loss and allows multiple consumers to access the queue with round-robin message distribution.
// If an error occurs during the initialization, it logs the error and exits the program.
func (r *rabbitMQ) initConsumer() {
	r.tagName = fmt.Sprintf("%s-%s", r.option.QueueName, uuid.New().String())
	if r.option.ConsumerID != "" {
		r.tagName = r.option.ConsumerID
	}

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
		// https://github.com/bxcodec/goqueue/issues/1
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
	h goqueue.InboundMessageHandler,
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
			msg, err := buildMessage(meta, receivedMsg)
			if err != nil {
				if err == errors.ErrInvalidMessageFormat {
					nackErr := receivedMsg.Nack(false, false) // nack with requeue false
					if nackErr != nil {
						logrus.WithFields(logrus.Fields{
							"consumer_meta": meta,
							"error":         nackErr,
						}).Error("failed to nack the message")
					}
				}
				continue
			}

			retryCount := extractHeaderInt(receivedMsg.Headers, headerKey.RetryCount)
			if retryCount > r.option.MaxRetryFailedMessage {
				logrus.WithFields(logrus.Fields{
					"consumer_meta": meta,
					"message_id":    msg.ID,
					"topic":         msg.Topic,
					"action":        msg.Action,
					"timestamp":     msg.Timestamp,
				}).Error("max retry failed message reached, moving message to dead letter queue")
				err = receivedMsg.Nack(false, false)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"consumer_meta": meta,
						"error":         err,
					}).Error("failed to nack the message")
				}
				continue
			}

			m := goqueue.InboundMessage{
				Message:    msg,
				RetryCount: retryCount,
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
				PutToBackOfQueueWithDelay: r.requeueMessage(meta, receivedMsg),
			}

			logrus.WithFields(logrus.Fields{
				"consumer_meta": meta,
				"message_id":    msg.ID,
				"topic":         msg.Topic,
				"action":        msg.Action,
				"timestamp":     msg.Timestamp,
			}).Info("message received")

			handleCtx := middleware.ApplyHandlerMiddleware(h.HandleMessage, r.option.Middlewares...)
			err = handleCtx(ctx, m)
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

func buildMessage(consumerMeta map[string]interface{}, receivedMsg amqp.Delivery) (msg goqueue.Message, err error) {
	err = json.Unmarshal(receivedMsg.Body, &msg)
	if err != nil {
		logrus.Error("failed to unmarshal the message, got err: ", err)
		logrus.WithFields(logrus.Fields{
			"consumer_meta": consumerMeta,
			"error":         err,
		}).Error("failed to unmarshal the message, removing the message due to wrong message format")
		return msg, errors.ErrInvalidMessageFormat
	}

	if msg.ID == "" {
		msg.ID = extractHeaderString(receivedMsg.Headers, headerKey.AppID)
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = extractHeaderTime(receivedMsg.Headers, headerKey.PublishedTimestamp)
	}
	if msg.Action == "" {
		msg.Action = receivedMsg.RoutingKey
	}
	if msg.Topic == "" {
		msg.Topic = receivedMsg.Exchange
	}
	if msg.ContentType == "" {
		msg.ContentType = headerVal.ContentType(extractHeaderString(receivedMsg.Headers, headerKey.ContentType))
	}
	if msg.Headers == nil {
		msg.Headers = receivedMsg.Headers
	}
	if msg.Data == "" || msg.Data == nil {
		logrus.WithFields(logrus.Fields{
			"consumer_meta": consumerMeta,
			"msg":           msg,
		}).Error("message data is empty, removing the message due to wrong message format")
		return msg, errors.ErrInvalidMessageFormat
	}

	msg.SetSchemaVersion(extractHeaderString(receivedMsg.Headers, headerKey.SchemaVer))
	return msg, nil
}

func (r *rabbitMQ) requeueMessage(consumerMeta map[string]interface{},
	receivedMsg amqp.Delivery) func(ctx context.Context, delayFn goqueue.DelayFn) (err error) {
	return func(ctx context.Context, delayFn goqueue.DelayFn) (err error) {
		if delayFn == nil {
			delayFn = goqueue.DefaultDelayFn
		}
		retries := extractHeaderInt(receivedMsg.Headers, headerKey.RetryCount)
		retries++
		delay := delayFn(retries)
		time.Sleep(time.Duration(delay) * time.Second)
		headers := receivedMsg.Headers
		headers[string(headerKey.RetryCount)] = retries
		requeueErr := r.requeueChannel.PublishWithContext(
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
		if requeueErr != nil {
			logrus.WithFields(logrus.Fields{
				"consumer_meta": consumerMeta,
				"error":         requeueErr,
			}).Error("failed to requeue the message")
			err = receivedMsg.Nack(false, false) // move to DLQ instead (depend on the RMQ server configuration)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"consumer_meta": consumerMeta,
					"error":         err,
				}).Error("failed to nack the message")
				return err
			}
			return requeueErr
		}

		// requeed successfully
		// ack the message
		err = receivedMsg.Ack(false)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"consumer_meta": consumerMeta,
				"error":         err,
			}).Error("failed to ack the message")
			return err
		}
		return nil
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

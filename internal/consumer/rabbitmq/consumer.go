package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"

	"github.com/bxcodec/goqueue/errors"
	headerKey "github.com/bxcodec/goqueue/headers/key"
	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/internal/consumer"
	"github.com/bxcodec/goqueue/middleware"
	consumerOpts "github.com/bxcodec/goqueue/options/consumer"
)

// rabbitMQ is the subscriber handler for rabbitmq
type rabbitMQ struct {
	consumerChannel             *amqp.Channel
	requeueChannel              *amqp.Channel
	option                      *consumerOpts.ConsumerOption
	tagName                     string
	msgReceiver                 <-chan amqp.Delivery
	retryExchangeName           string
	retryDeadLetterExchangeName string
}

// New will initialize the rabbitMQ subscriber
func NewConsumer(
	opts ...consumerOpts.ConsumerOptionFunc) consumer.Consumer {
	opt := consumerOpts.DefaultConsumerOption()
	for _, o := range opts {
		o(opt)
	}

	rmqHandler := &rabbitMQ{
		consumerChannel:             opt.RabbitMQConsumerConfig.ConsumerChannel,
		requeueChannel:              opt.RabbitMQConsumerConfig.ReQueueChannel,
		option:                      opt,
		retryExchangeName:           fmt.Sprintf("%s__retry_exchange", opt.QueueName),
		retryDeadLetterExchangeName: fmt.Sprintf("%s__retry_dlx", opt.QueueName),
	}
	if opt.RabbitMQConsumerConfig.QueueDeclareConfig != nil &&
		opt.RabbitMQConsumerConfig.QueueBindConfig != nil {
		rmqHandler.initQueue()
	}

	rmqHandler.initConsumer()
	rmqHandler.initRetryModule()
	return rmqHandler
}

func (r *rabbitMQ) initQueue() {
	// QueueDeclare declares a queue on the consumer's channel with the specified parameters.
	// It takes the queue name, durable, autoDelete, exclusive, noWait, and arguments as arguments.
	// Returns an instance of amqp.Queue and any error encountered.
	// Please refer to the RabbitMQ documentation for more information on the parameters.
	// https://www.rabbitmq.com/amqp-0-9-1-reference.html#queue.declare
	_, err := r.consumerChannel.QueueDeclare(
		r.option.QueueName,
		r.option.RabbitMQConsumerConfig.QueueDeclareConfig.Durable,
		r.option.RabbitMQConsumerConfig.QueueDeclareConfig.AutoDelete,
		r.option.RabbitMQConsumerConfig.QueueDeclareConfig.Exclusive,
		r.option.RabbitMQConsumerConfig.QueueDeclareConfig.NoWait,
		r.option.RabbitMQConsumerConfig.QueueDeclareConfig.Args,
	)
	if err != nil {
		logrus.Fatal("error declaring the queue, ", err)
	}

	for _, eventType := range r.option.RabbitMQConsumerConfig.QueueBindConfig.RoutingKeys {
		err = r.consumerChannel.QueueBind(
			r.option.QueueName,
			eventType,
			r.option.RabbitMQConsumerConfig.QueueBindConfig.ExchangeName,
			r.option.RabbitMQConsumerConfig.QueueBindConfig.NoWait,
			r.option.RabbitMQConsumerConfig.QueueBindConfig.Args,
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

// initRetryModule initializes the retry module for the RabbitMQ consumer.
// It declares the retry exchange, dead letter exchange, and retry queues.
// It also binds the dead letter exchange to the original queue and the retry queues to the retry exchange.
func (r *rabbitMQ) initRetryModule() {
	// declare retry exchange
	err := r.consumerChannel.ExchangeDeclare(
		r.retryExchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logrus.Fatal("error declaring the retry exchange, ", err)
	}

	// declare dead letter exchange
	err = r.consumerChannel.ExchangeDeclare(
		r.retryDeadLetterExchangeName,
		"fanout",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		logrus.Fatal("error declaring the retry dead letter exchange, ", err)
	}

	// bind dead letter exchange to original queue
	err = r.consumerChannel.QueueBind(
		r.option.QueueName,
		"",
		r.retryDeadLetterExchangeName,
		false,
		nil,
	)
	if err != nil {
		logrus.Fatal("error binding the dead letter exchange to the original queue, ", err)
	}

	// declare retry queue
	for i := int64(1); i <= r.option.MaxRetryFailedMessage; i++ {
		// declare retry queue
		_, err = r.consumerChannel.QueueDeclare(
			getRetryRoutingKey(r.option.QueueName, i), // queue name and routing key is the same for retry queue
			true,
			false,
			false,
			false,
			amqp.Table{
				"x-dead-letter-exchange": r.retryDeadLetterExchangeName,
			},
		)
		if err != nil {
			logrus.Fatal("error declaring the retry queue, ", err)
		}

		// bind retry queue to retry exchange
		err = r.consumerChannel.QueueBind(
			getRetryRoutingKey(r.option.QueueName, i), // queue name and routing key is the same for retry queue
			getRetryRoutingKey(r.option.QueueName, i), // queue name and routing key is the same for retry queue
			r.retryExchangeName,
			false,
			nil,
		)
		if err != nil {
			logrus.Fatal("error binding the retry queue, ", err)
		}
	}
}

// Consume consumes messages from a RabbitMQ queue and handles them using the provided message handler.
// It takes a context, an inbound message handler, and a map of metadata as input parameters.
// The function continuously listens for messages from the queue and processes them until the context is canceled.
// If the context is canceled, the function stops consuming messages and returns.
// For each received message, the function builds an inbound message, extracts the retry count, and checks if the maximum retry count has been reached.
// If the maximum retry count has been reached, the message is moved to the dead letter queue.
// Otherwise, the message is passed to the message handler for processing.
// The message handler is responsible for handling the message and returning an error if any.
// If an error occurs while handling the message, it is logged.
// The function provides methods for acknowledging, rejecting, and moving messages to the dead letter queue.
// These methods can be used by the message handler to control the message processing flow.
// The function also logs information about the received message, such as the message ID, topic, action, and timestamp.
// It applies any configured middlewares to the message handler before calling it.
// The function returns an error if any occurred during message handling or if the context was canceled.
func (r *rabbitMQ) Consume(ctx context.Context,
	h interfaces.InboundMessageHandler,
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
			m := interfaces.InboundMessage{
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
					//  receivedMsg.Nack(false, true) => will redelivered again instantly (same with receivedMsg.reject)
					//  receivedMsg.Nack(false, false) => will put the message to dead letter queue (same with receivedMsg.reject)
					err = receivedMsg.Nack(false, true)
					return
				},
				MoveToDeadLetterQueue: func(ctx context.Context) (err error) {
					//  receivedMsg.Nack(false, true) => will redelivered again instantly (same with receivedMsg.reject)
					//  receivedMsg.Nack(false, false) => will put the message to dead letter queue (same with receivedMsg.reject)
					err = receivedMsg.Nack(false, false)
					return
				},
				RetryWithDelayFn: r.requeueMessageWithDLQ(meta, msg, receivedMsg),
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

func buildMessage(consumerMeta map[string]interface{}, receivedMsg amqp.Delivery) (msg interfaces.Message, err error) {
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

func (r *rabbitMQ) requeueMessageWithDLQ(consumerMeta map[string]interface{}, msg interfaces.Message,
	receivedMsg amqp.Delivery) func(ctx context.Context, delayFn interfaces.DelayFn) (err error) {
	return func(ctx context.Context, delayFn interfaces.DelayFn) (err error) {
		if delayFn == nil {
			delayFn = interfaces.DefaultDelayFn
		}
		retries := extractHeaderInt(receivedMsg.Headers, headerKey.RetryCount)
		retries++
		delayInSeconds := delayFn(retries)
		routingKeyPrefixForRetryQueue := getRetryRoutingKey(r.option.QueueName, retries)
		headers := receivedMsg.Headers
		headers[headerKey.OriginalTopicName] = msg.Topic
		headers[headerKey.OriginalActionName] = msg.Action
		headers[headerKey.RetryCount] = retries
		requeueErr := r.requeueChannel.PublishWithContext(
			ctx,
			r.retryExchangeName,
			routingKeyPrefixForRetryQueue,
			false,
			false,
			amqp.Publishing{
				Headers:     headers,
				ContentType: receivedMsg.ContentType,
				Body:        receivedMsg.Body,
				Timestamp:   time.Now(),
				AppId:       r.tagName,
				Expiration:  fmt.Sprintf("%d", delayInSeconds*10000),
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

func getRetryRoutingKey(queueName string, retry int64) string {
	return fmt.Sprintf("%s__retry.%d", queueName, retry)
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

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bxcodec/goqueue"
	"github.com/bxcodec/goqueue/consumer"
	rmqConsumer "github.com/bxcodec/goqueue/consumer/rabbitmq"
	"github.com/bxcodec/goqueue/middleware"
	"github.com/bxcodec/goqueue/publisher"
	rmqPublisher "github.com/bxcodec/goqueue/publisher/rabbitmq"
)

func initExchange(ch *amqp.Channel, exchangeName string) error {
	return ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
}

func main() {
	rmqDSN := "amqp://rabbitmq:rabbitmq@localhost:5672/"
	rmqConn, err := amqp.Dial(rmqDSN)
	if err != nil {
		panic(err)
	}

	rmqPub := rmqPublisher.NewPublisher(rmqConn,
		publisher.WithPublisherID("publisher_id"),
		publisher.WithMiddlewares(
			middleware.HelloWorldMiddlewareExecuteBeforePublisher(),
			middleware.HelloWorldMiddlewareExecuteAfterPublisher()),
	)

	publisherChannel, err := rmqConn.Channel()
	if err != nil {
		panic(err)
	}

	defer publisherChannel.Close()
	initExchange(publisherChannel, "goqueue")

	consumerChannel, err := rmqConn.Channel()
	if err != nil {
		panic(err)
	}
	defer consumerChannel.Close()

	rmqConsumer := rmqConsumer.NewConsumer(
		publisherChannel,
		consumerChannel,
		consumer.WithMiddlewares(
			middleware.HelloWorldMiddlewareExecuteAfterInboundMessageHandler(),
			middleware.HelloWorldMiddlewareExecuteBeforeInboundMessageHandler()),
		consumer.WithQueueName("consumer_queue"),
		consumer.WithConsumerID("consumer_id"),
		consumer.WithBatchMessageSize(1),
		consumer.WithMaxRetryFailedMessage(3),
		consumer.WithActionsPatternSubscribed("goqueue.payments.#", "goqueue.users.#"),
		consumer.WithTopicName("goqueue"),
	)

	queueSvc := goqueue.NewQueueService(
		goqueue.WithPublisher(rmqPub),
		goqueue.WithConsumer(rmqConsumer),
		goqueue.WithMessageHandler(handler()),
	)

	go func() {
		for i := 0; i < 10; i++ {
			data := map[string]interface{}{
				"message": fmt.Sprintf("Hello World %d", i),
			}
			jbyt, _ := json.Marshal(data)
			err := queueSvc.Publish(context.Background(), goqueue.Message{
				Data:   data,
				Action: "goqueue.payments.create",
				Topic:  "goqueue",
			})
			if err != nil {
				panic(err)
			}
			fmt.Println("Message Sent: ", string(jbyt))
		}
	}()

	// change to context.Background() if you want to run it forever
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = queueSvc.Start(ctx)
	if err != nil {
		panic(err)
	}
}

func handler() goqueue.InboundMessageHandlerFunc {
	return func(ctx context.Context, m goqueue.InboundMessage) (err error) {
		data := m.Data
		jbyt, _ := json.Marshal(data)
		fmt.Println("Message Received: ", string(jbyt))
		return m.Ack(ctx)
	}
}

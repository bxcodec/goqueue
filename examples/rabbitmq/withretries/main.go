package main

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bxcodec/goqueue"
	"github.com/bxcodec/goqueue/consumer"
	"github.com/bxcodec/goqueue/interfaces"
	"github.com/bxcodec/goqueue/middleware"
	"github.com/bxcodec/goqueue/options"
	consumerOpts "github.com/bxcodec/goqueue/options/consumer"
	publisherOpts "github.com/bxcodec/goqueue/options/publisher"
	"github.com/bxcodec/goqueue/publisher"
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

	rmqPub := publisher.NewPublisher(
		publisherOpts.PublisherPlatformRabbitMQ,
		publisherOpts.WithRabbitMQPublisherConfig(&publisherOpts.RabbitMQPublisherConfig{
			Conn:                     rmqConn,
			PublisherChannelPoolSize: 5,
		}),
		publisherOpts.WithPublisherID("publisher_id"),
		publisherOpts.WithMiddlewares(
			middleware.HelloWorldMiddlewareExecuteBeforePublisher(),
			middleware.HelloWorldMiddlewareExecuteAfterPublisher(),
		),
	)

	requeueChannel, err := rmqConn.Channel()
	if err != nil {
		panic(err)
	}

	defer requeueChannel.Close()
	initExchange(requeueChannel, "goqueue")

	consumerChannel, err := rmqConn.Channel()
	if err != nil {
		panic(err)
	}
	defer consumerChannel.Close()
	rmqConsumer := consumer.NewConsumer(
		consumerOpts.ConsumerPlatformRabbitMQ,
		consumerOpts.WithRabbitMQConsumerConfig(consumerOpts.RabbitMQConfigWithDefaultTopicFanOutPattern(
			consumerChannel,
			requeueChannel,
			"goqueue",                      // exchange name
			[]string{"goqueue.payments.#"}, // routing keys pattern
		)),
		consumerOpts.WithConsumerID("consumer_id"),
		consumerOpts.WithMiddlewares(
			middleware.HelloWorldMiddlewareExecuteAfterInboundMessageHandler(),
			middleware.HelloWorldMiddlewareExecuteBeforeInboundMessageHandler(),
		),
		consumerOpts.WithMaxRetryFailedMessage(5),
		consumerOpts.WithBatchMessageSize(1),
		consumerOpts.WithQueueName("consumer_queue"),
	)

	queueSvc := goqueue.NewQueueService(
		options.WithConsumer(rmqConsumer),
		options.WithPublisher(rmqPub),
		options.WithMessageHandler(handler()),
	)
	go func() {
		for i := 0; i < 10; i++ {
			data := map[string]interface{}{
				"message": fmt.Sprintf("Hello World %d", i),
			}
			jbyt, _ := json.Marshal(data)
			err := queueSvc.Publish(context.Background(), interfaces.Message{
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
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()
	err = queueSvc.Start(context.Background())
	if err != nil {
		panic(err)
	}
}

func handler() interfaces.InboundMessageHandlerFunc {
	return func(ctx context.Context, m interfaces.InboundMessage) (err error) {
		fmt.Printf("Message: %+v\n", m)
		// something happend, we need to requeue the message
		return m.PutToBackOfQueueWithDelay(ctx, interfaces.ExponentialBackoffDelayFn)
	}
}

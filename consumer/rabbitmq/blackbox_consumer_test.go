package rabbitmq_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/goqueue"
	"github.com/bxcodec/goqueue/consumer"
	rmq "github.com/bxcodec/goqueue/consumer/rabbitmq"
	headerKey "github.com/bxcodec/goqueue/headers/key"
	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/middleware"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	rabbitMQTestQueueName = "testqueuesubscriber"
	testExchange          = "goqueue-exchange-test"
	testAction            = "goqueue.action.test"
	testActionRequeue     = "goqueue.action.testrequeue"
)

type rabbitMQTestSuite struct {
	suite.Suite
	rmqURL          string
	conn            *amqp.Connection
	publishChannel  *amqp.Channel
	consumerChannel *amqp.Channel
	queue           *amqp.Queue
}

func TestSuiteRabbitMQConsumer(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip the Test Suite for RabbitMQ Consumer")
	}

	rmqURL := os.Getenv("RABBITMQ_TEST_URL")
	if rmqURL == "" {
		rmqURL = "amqp://test:test@localhost:5672/test"
	}

	rabbitMQTestSuite := &rabbitMQTestSuite{
		rmqURL: rmqURL,
	}
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{})
	rabbitMQTestSuite.initConnection(t)
	suite.Run(t, rabbitMQTestSuite)
}

func (s *rabbitMQTestSuite) BeforeTest(_, _ string) {
}

func (s *rabbitMQTestSuite) AfterTest(_, _ string) {
	_, err := s.consumerChannel.QueuePurge(rabbitMQTestQueueName, true) // force purge the queue after test
	s.Require().NoError(err)
}

func (s *rabbitMQTestSuite) TearDownSuite() {
	err := s.publishChannel.Close()
	s.Require().NoError(err)
	_, err = s.consumerChannel.QueuePurge(rabbitMQTestQueueName, true) // force purge the queue after test suite done
	s.Require().NoError(err)
	err = s.consumerChannel.Close()
	s.Require().NoError(err)
	err = s.conn.Close()
	s.Require().NoError(err)
}

func (s *rabbitMQTestSuite) initConnection(t *testing.T) {
	var err error
	s.conn, err = amqp.Dial(s.rmqURL)
	require.NoError(t, err)
	s.publishChannel, err = s.conn.Channel()
	require.NoError(t, err)
	s.consumerChannel, err = s.conn.Channel()
	require.NoError(t, err)

	err = s.publishChannel.ExchangeDeclare(
		testExchange, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	require.NoError(t, err)
}

func (s *rabbitMQTestSuite) initQueueForTesting(t *testing.T, exchangePattern ...string) {
	q, err := s.consumerChannel.QueueDeclare(
		rabbitMQTestQueueName, // name
		true,                  // durable
		false,                 // auto_delete
		false,                 // exclusive
		false,                 // no-wait
		nil,                   // arguments
	)
	require.NoError(t, err)
	s.queue = &q

	for _, patternRoutingKey := range exchangePattern {
		logrus.Printf("binding queue %s to exchange %s with routing key %s",
			q.Name, testExchange, patternRoutingKey)

		err = s.consumerChannel.QueueBind(
			rabbitMQTestQueueName, // queue name
			patternRoutingKey,     // routing key
			testExchange,          // exchange
			false,
			nil)

		require.NoError(t, err)
	}
}

func (s *rabbitMQTestSuite) getMockData(action string) (res goqueue.Message) {
	res = goqueue.Message{
		Action: action,
		Topic:  testExchange,
		Data: map[string]interface{}{
			"message": "hello-world-test",
		},
		Timestamp: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	return res
}

func (s *rabbitMQTestSuite) seedPublish(contentType string, action string) {
	mockData := s.getMockData(action)
	jsonData, err := json.Marshal(mockData)
	s.Require().NoError(err)
	err = s.publishChannel.PublishWithContext(
		context.Background(),
		testExchange, // exchange
		action,       // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			Headers: amqp.Table{
				headerKey.PublishedTimestamp: mockData.Timestamp.Format(time.RFC3339),
				headerKey.MessageID:          uuid.New().String(),
				headerKey.RetryCount:         1,
				headerKey.ContentType:        contentType,
				headerKey.SchemaVer:          string(headerVal.GoquMessageSchemaVersionV1),
				headerKey.QueueServiceAgent:  string(headerVal.RabbitMQ),
			},
			ContentType: contentType,
			Body:        jsonData,
		})
	s.Require().NoError(err)
}

// func (s *rabbitMQTestSuite) TestConsumerWithoutExchangePatternProvided() {
// 	s.initQueueForTesting(s.T(), "goqueue.action.#")
// 	s.seedPublish(string(headerVal.ContentTypeJSON), testAction)
// 	rmqSubs := rmq.NewConsumer(s.conn,
// 		s.consumerChannel,
// 		s.publishChannel,
// 		consumer.WithBatchMessageSize(1),
// 		consumer.WithMiddlewares(middleware.HelloWorldMiddlewareExecuteAfterInboundMessageHandler()),
// 		consumer.WithQueueName("testqueuesubscriber"),
// 	)

// 	msgHandler := handler(s.T(), s.getMockData(testAction))
// 	queueSvc := goqueue.NewQueueService(
// 		goqueue.WithConsumer(rmqSubs),
// 		goqueue.WithMessageHandler(msgHandler),
// 	)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3) // increase this context if want to test a long running worker
// 	defer cancel()

// 	err := queueSvc.Start(ctx)
// 	s.Require().NoError(err)
// }

// func (s *rabbitMQTestSuite) TestConsumerWithExchangePatternProvided() {
// 	s.seedPublish(string(headerVal.ContentTypeJSON), testAction)

// 	rmqSubs := rmq.NewConsumer(s.conn,
// 		s.consumerChannel,
// 		s.publishChannel,
// 		consumer.WithBatchMessageSize(1),
// 		consumer.WithMiddlewares(middleware.HelloWorldMiddlewareExecuteAfterInboundMessageHandler()),
// 		consumer.WithQueueName(rabbitMQTestQueueName),
// 		consumer.WithActionsPatternSubscribed("goqueue.action.#"), //exchange pattern provided in constructor
// 		consumer.WithTopicName(testExchange),                      //exchange name provided in constructor
// 	)

// 	msgHandler := handler(s.T(), s.getMockData(testAction))
// 	queueSvc := goqueue.NewQueueService(
// 		goqueue.WithConsumer(rmqSubs),
// 		goqueue.WithMessageHandler(msgHandler),
// 	)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3) // increase this context if want to test a long running worker
// 	defer cancel()

// 	err := queueSvc.Start(ctx)
// 	s.Require().NoError(err)
// }

// func handler(t *testing.T, expected goqueue.Message) goqueue.InboundMessageHandlerFunc {
// 	return func(ctx context.Context, m goqueue.InboundMessage) (err error) {
// 		fmt.Println(">>>>>> MASUK MESSAGE OKGAS")
// 		switch m.ContentType {
// 		case headerVal.ContentTypeText:
// 			assert.Equal(t, expected.Data, m.Data)
// 		case headerVal.ContentTypeJSON:
// 			expectedJSON, err := json.Marshal(expected.Data)
// 			require.NoError(t, err)
// 			actualJSON, err := json.Marshal(m.Data)
// 			require.NoError(t, err)
// 			assert.JSONEq(t, string(expectedJSON), string(actualJSON))
// 		}

// 		assert.EqualValues(t, 9, m.RetryCount)
// 		err = m.Ack(ctx)
// 		assert.NoError(t, err)

// 		assert.Equal(t, headerVal.GoquMessageSchemaVersionV1, m.GetSchemaVersion())
// 		return err
// 	}
// }

func (s *rabbitMQTestSuite) TestRequeueWithouthExchangePatternProvided() {
	s.initQueueForTesting(s.T(), "goqueue.action.#")
	s.seedPublish(string(headerVal.ContentTypeJSON), testActionRequeue)
	rmqSubs := rmq.NewConsumer(s.conn,
		s.consumerChannel,
		s.publishChannel,
		consumer.WithBatchMessageSize(1),
		consumer.WithMiddlewares(middleware.HelloWorldMiddlewareExecuteAfterInboundMessageHandler()),
		consumer.WithQueueName(rabbitMQTestQueueName),
		consumer.WithMaxRetryFailedMessage(2),
	)

	msgHandlerRequeue := handlerRequeue(s.T())
	queueSvc := goqueue.NewQueueService(
		goqueue.WithConsumer(rmqSubs),
		goqueue.WithMessageHandler(msgHandlerRequeue),
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5) // increase this context if want to test a long running worker
	defer cancel()

	err := queueSvc.Start(ctx)
	s.Require().NoError(err)
}

func handlerRequeue(t *testing.T) goqueue.InboundMessageHandlerFunc {
	return func(ctx context.Context, m goqueue.InboundMessage) (err error) {
		fmt.Println(">>>>>> REQUEUE MESSAGE")
		delayFn := func(retries int64) int64 {
			assert.Equal(t, int64(m.RetryCount)+1, retries) // because the retry++ is done before this delayfn is called
			return m.RetryCount
		}

		// fmt.Println(">>>>>> REQUEUE MESSAGE")
		// logrus.Info(">>>>>> REQUEUE MESSAGE")
		err = m.PutToBackOfQueueWithDelay(ctx, delayFn)
		assert.NoError(t, err)
		return
	}
}

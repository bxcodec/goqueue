package rabbitmq_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bxcodec/goqueue"
	headerKey "github.com/bxcodec/goqueue/headers/key"
	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/interfaces"
	rmq "github.com/bxcodec/goqueue/internal/publisher/rabbitmq"
	"github.com/bxcodec/goqueue/options"
	publisherOpts "github.com/bxcodec/goqueue/options/publisher"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	rabbitMQTestQueueName = "testqueuepublisher"
	testExchange          = "goqueue-exchange-test-publish"
	testAction            = "goqueue.action.test"
)

type rabbitMQTestSuite struct {
	suite.Suite
	rmqURL          string
	conn            *amqp.Connection
	publishChannel  *amqp.Channel
	consumerChannel *amqp.Channel
}

func TestSuiteRabbitMQPublisher(t *testing.T) {
	if testing.Short() {
		t.Skip("Skip the Test Suite for RabbitMQ Publisher")
	}

	time.Sleep(5 * time.Second) // wait for the rabbitmq to be ready

	rmqURL := os.Getenv("RABBITMQ_TEST_URL")
	if rmqURL == "" {
		rmqURL = "amqp://test:test@localhost:5672/test"
	}

	rabbitMQTestSuite := &rabbitMQTestSuite{
		rmqURL: rmqURL,
	}
	log.Logger = log.With().Caller().Logger()
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

	for _, patternRoutingKey := range exchangePattern {
		log.Info().
			Str("queue_name", q.Name).
			Str("exchange", testExchange).
			Str("routing_key", patternRoutingKey).
			Msg("binding queue to exchange")

		err = s.consumerChannel.QueueBind(
			rabbitMQTestQueueName, // queue name
			patternRoutingKey,     // routing key
			testExchange,          // exchange
			false,
			nil)

		require.NoError(t, err)
	}
}
func (s *rabbitMQTestSuite) getMockData(action string, identifier string) (res interfaces.Message) {
	res = interfaces.Message{
		Action:      action,
		ID:          identifier,
		Topic:       testExchange,
		ContentType: headerVal.ContentTypeJSON,
		Data: map[string]interface{}{
			"message": "hello-world-test",
		},
		Timestamp: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	return res
}

func (s *rabbitMQTestSuite) TestPublisher() {
	s.initQueueForTesting(s.T(), testAction)
	publisher := rmq.NewPublisher(
		publisherOpts.WithPublisherID("test-publisher"),
		publisherOpts.WithRabbitMQPublisherConfig(&publisherOpts.RabbitMQPublisherConfig{
			PublisherChannelPoolSize: 10,
			Conn:                     s.conn,
		}),
	)

	queueSvc := goqueue.NewQueueService(
		options.WithPublisher(publisher),
	)
	var err error
	totalPublishedMessage := 10
	for i := 0; i < totalPublishedMessage; i++ {
		err = queueSvc.Publish(context.Background(), s.getMockData(testAction, fmt.Sprintf("test-id-%d", i)))
		s.Require().NoError(err)
	}

	// assertion event
	// Ensure the published message is correct
	msgs, err := s.consumerChannel.Consume(
		rabbitMQTestQueueName, // queue
		"testing-consumer",    // consumer
		false,                 // noAck
		false,                 // exclusive
		false,                 // noLocal
		false,                 // noWait
		nil,                   // arguments
	)
	s.Require().NoError(err)

	done := make(chan bool)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*6) // increase this context if want to test a long running worker
	defer cancel()
	go func() {
		totalMessages := 0
		for {
			select {
			case <-ctx.Done():
				done <- true
			case d := <-msgs:
				var content map[string]interface{}
				inErr := json.Unmarshal(d.Body, &content)

				/*
					headerKey.MessageID:          id,
					headerKey.PublishedTimestamp: timestamp.Format(time.RFC3339),
					headerKey.RetryCount:         0,
					headerKey.ContentType:        m.ContentType,
					headerKey.QueueServiceAgent:  headerVal.RabbitMQ,
					headerKey.SchemaVer:          headerVal.GoquMessageSchemaVersionV1,
				*/
				s.Require().NoError(inErr)
				s.Require().Contains(content, "data")
				s.Require().Equal(string(headerVal.ContentTypeJSON), d.ContentType)
				s.Require().Equal(string(headerVal.ContentTypeJSON), d.Headers[headerKey.ContentType])
				s.Require().Contains(d.Headers, headerKey.MessageID)
				s.Require().Contains(d.Headers, headerKey.PublishedTimestamp)
				s.Require().Contains(d.Headers, headerKey.RetryCount)
				s.Require().Contains(d.Headers, headerKey.ContentType)
				s.Require().Contains(d.Headers, headerKey.QueueServiceAgent)
				s.Require().Contains(d.Headers, headerKey.SchemaVer)
				s.Require().Equal("test-publisher", d.Headers[headerKey.AppID])
				inErr = d.Ack(false)
				s.Require().NoError(inErr)
				totalMessages++
				if totalMessages == totalPublishedMessage {
					done <- true
				}
			}
		}
	}()

	log.Info().
		Msg("waiting for the message to be consumed")
	<-done

	err = publisher.Close(context.Background())
	s.Require().NoError(err)
}

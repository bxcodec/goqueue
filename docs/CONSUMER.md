# 📨 GoQueue Consumer System

The GoQueue Consumer system provides a robust, scalable way to consume and process messages from various queue platforms. It handles the complexity of message acknowledgment, retries, and error handling while providing a simple, unified interface.

## 📖 Table of Contents

- [📨 GoQueue Consumer System](#-goqueue-consumer-system)
  - [📖 Table of Contents](#-table-of-contents)
  - [🎯 Overview](#-overview)
  - [🏗️ Architecture](#️-architecture)
  - [🚀 Quick Start](#-quick-start)
  - [⚙️ Configuration](#️-configuration)
  - [📝 Message Handling](#-message-handling)
  - [🔄 Retry Mechanisms](#-retry-mechanisms)
  - [🛡️ Error Handling](#️-error-handling)
  - [📊 Monitoring & Observability](#-monitoring--observability)
  - [🧪 Testing](#-testing-consumers)
  - [💡 Best Practices](#-best-practices)
  - [🔧 Troubleshooting](#-troubleshooting)

---

## 🎯 Overview

The Consumer system in GoQueue provides:

- **🔄 Automatic Retries** with configurable strategies
- **⚡ Concurrent Processing** with controllable parallelism
- **🛡️ Error Handling** with dead letter queue support
- **📊 Built-in Observability** with logging and metrics hooks
- **🔌 Middleware Support** for extending functionality
- **🎛️ Flexible Configuration** for different use cases

## 🏗️ Architecture

```
┌─────────────────────┐
│   Queue Platform    │
│    (RabbitMQ)       │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│    GoQueue          │
│    Consumer         │
│                     │
│ ┌─────────────────┐ │
│ │   Middleware    │ │
│ │     Chain       │ │
│ └─────────────────┘ │
│ ┌─────────────────┐ │
│ │   Message       │ │
│ │   Handler       │ │
│ └─────────────────┘ │
│ ┌─────────────────┐ │
│ │   Retry         │ │
│ │   Logic         │ │
│ └─────────────────┘ │
└─────────────────────┘
```

---

## 🚀 Quick Start

### Basic Consumer Setup

```go
package main

import (
    "context"
    "log"

    "github.com/bxcodec/goqueue/consumer"
    "github.com/bxcodec/goqueue/interfaces"
    consumerOpts "github.com/bxcodec/goqueue/options/consumer"
    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    // Connect to RabbitMQ
    conn, err := amqp.Dial("amqp://localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create channels
    consumerChannel, _ := conn.Channel()
    requeueChannel, _ := conn.Channel()

    // Create consumer
    consumer := consumer.NewConsumer(
        consumerOpts.ConsumerPlatformRabbitMQ,
        consumerOpts.WithQueueName("user-events"),
        consumerOpts.WithMaxRetryFailedMessage(3),
        consumerOpts.WithBatchMessageSize(10),
        consumerOpts.WithRabbitMQConsumerConfig(
            consumerOpts.RabbitMQConfigWithDefaultTopicFanOutPattern(
                consumerChannel,
                requeueChannel,
                "user-exchange",
                []string{"user.created", "user.updated"},
            ),
        ),
    )

    // Define message handler
    handler := func(ctx context.Context, m interfaces.InboundMessage) error {
        log.Printf("Processing message: %+v", m.Data)

        // Your business logic here
        if err := processMessage(m); err != nil {
            // Retry with exponential backoff
            return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
        }

        // Acknowledge successful processing
        return m.Ack(ctx)
    }

    // Start consuming
    metadata := map[string]interface{}{
        "consumer_id": "user-service-01",
        "version":     "1.0.0",
    }

    err = consumer.Consume(context.Background(),
        interfaces.InboundMessageHandlerFunc(handler), metadata)
    if err != nil {
        log.Fatal(err)
    }
}
```

---

## ⚙️ Configuration

### Consumer Options

```go
consumer := consumer.NewConsumer(
    consumerOpts.ConsumerPlatformRabbitMQ,

    // Basic Configuration
    consumerOpts.WithQueueName("my-queue"),
    consumerOpts.WithConsumerID("service-01"),
    consumerOpts.WithBatchMessageSize(50),
    consumerOpts.WithMaxRetryFailedMessage(5),

    // Middleware
    consumerOpts.WithMiddlewares(
        middleware.LoggingMiddleware(),
        middleware.MetricsMiddleware(),
        middleware.AuthenticationMiddleware(),
    ),

    // Platform-specific configuration
    consumerOpts.WithRabbitMQConsumerConfig(&consumerOpts.RabbitMQConsumerConfig{
        ConsumerChannel: channel,
        ReQueueChannel:  requeueChannel,
        QueueDeclareConfig: &consumerOpts.RabbitMQQueueDeclareConfig{
            Durable:    true,
            AutoDelete: false,
            Exclusive:  false,
            NoWait:     false,
            Args:       nil,
        },
        QueueBindConfig: &consumerOpts.RabbitMQQueueBindConfig{
            RoutingKeys:  []string{"user.*", "order.created"},
            ExchangeName: "main-exchange",
            NoWait:       false,
            Args:         nil,
        },
    }),
)
```

### Configuration Options Explained

| Option                  | Description                                  | Default        |
| ----------------------- | -------------------------------------------- | -------------- |
| `QueueName`             | Name of the queue to consume from            | Required       |
| `ConsumerID`            | Unique identifier for this consumer instance | Auto-generated |
| `BatchMessageSize`      | Number of messages to prefetch               | 1              |
| `MaxRetryFailedMessage` | Maximum retry attempts                       | 3              |
| `Middlewares`           | List of middleware functions                 | Empty          |

---

## 📝 Message Handling

### Message Structure

```go
type InboundMessage struct {
    Message                     // Embedded message data
    RetryCount int64           // Current retry attempt
    Metadata   map[string]any  // Platform-specific metadata

    // Control functions
    Ack                  func(ctx context.Context) error
    Nack                 func(ctx context.Context) error
    MoveToDeadLetterQueue func(ctx context.Context) error
    RetryWithDelayFn     func(ctx context.Context, delayFn DelayFn) error
}

type Message struct {
    ID          string                 `json:"id"`
    Topic       string                 `json:"topic"`
    Action      string                 `json:"action"`
    Data        interface{}           `json:"data"`
    Headers     map[string]interface{} `json:"headers"`
    Timestamp   time.Time             `json:"timestamp"`
    ContentType string                `json:"contentType"`
}
```

### Handler Patterns

#### 1. Simple Handler

```go
func simpleHandler(ctx context.Context, m interfaces.InboundMessage) error {
    log.Printf("Received: %s - %s", m.Action, m.ID)

    // Process message
    return processBusinessLogic(m.Data)
}
```

#### 2. Handler with Error Handling

```go
func errorHandler(ctx context.Context, m interfaces.InboundMessage) error {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Panic recovered: %v", r)
            m.MoveToDeadLetterQueue(ctx)
        }
    }()

    // Validate message
    if err := validateMessage(m); err != nil {
        log.Printf("Invalid message: %v", err)
        return m.MoveToDeadLetterQueue(ctx)
    }

    // Process with retries
    if err := processMessage(m); err != nil {
        if isRetryableError(err) {
            return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
        }
        return m.MoveToDeadLetterQueue(ctx)
    }

    return m.Ack(ctx)
}
```

#### 3. Route-based Handler

```go
func routeHandler(ctx context.Context, m interfaces.InboundMessage) error {
    switch m.Action {
    case "user.created":
        return handleUserCreated(ctx, m)
    case "user.updated":
        return handleUserUpdated(ctx, m)
    case "user.deleted":
        return handleUserDeleted(ctx, m)
    default:
        log.Printf("Unknown action: %s", m.Action)
        return m.MoveToDeadLetterQueue(ctx)
    }
}

func handleUserCreated(ctx context.Context, m interfaces.InboundMessage) error {
    var user User
    if err := json.Unmarshal(m.Data.([]byte), &user); err != nil {
        return m.MoveToDeadLetterQueue(ctx)
    }

    if err := userService.Create(ctx, user); err != nil {
        return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
    }

    return m.Ack(ctx)
}
```

---

## 🔄 Retry Mechanisms

### Built-in Delay Functions

```go
// Exponential backoff: 1s, 2s, 4s, 8s, 16s...
return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)

// Linear backoff: 1s, 2s, 3s, 4s, 5s...
return m.RetryWithDelayFn(ctx, interfaces.LinearBackoffDelayFn)

// Fixed delay: 5s, 5s, 5s, 5s...
return m.RetryWithDelayFn(ctx, func(retryCount int64) int64 {
    return 5 // 5 seconds
})
```

### Custom Delay Functions

```go
// Custom exponential with jitter
func customDelayFn(retryCount int64) int64 {
    baseDelay := time.Duration(retryCount) * time.Second
    jitter := time.Duration(rand.Int63n(1000)) * time.Millisecond
    return int64((baseDelay + jitter) / time.Millisecond)
}

// Fibonacci backoff
func fibonacciDelayFn(retryCount int64) int64 {
    fib := fibonacci(retryCount)
    return fib * 1000 // Convert to milliseconds
}

// Usage
return m.RetryWithDelayFn(ctx, customDelayFn)
```

### Conditional Retries

```go
func smartRetryHandler(ctx context.Context, m interfaces.InboundMessage) error {
    err := processMessage(m)
    if err == nil {
        return m.Ack(ctx)
    }

    switch {
    case isTemporaryError(err):
        // Retry temporary errors
        return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
    case isValidationError(err):
        // Don't retry validation errors
        log.Printf("Validation error: %v", err)
        return m.MoveToDeadLetterQueue(ctx)
    case isRateLimitError(err):
        // Longer delay for rate limits
        return m.RetryWithDelayFn(ctx, func(retryCount int64) int64 {
            return 60 * 1000 // 60 seconds
        })
    default:
        // Unknown error, move to DLQ
        return m.MoveToDeadLetterQueue(ctx)
    }
}
```

---

## 🛡️ Error Handling

### Error Categories

```go
type ErrorType int

const (
    ErrorTypeTemporary ErrorType = iota
    ErrorTypePermanent
    ErrorTypeValidation
    ErrorTypeRateLimit
    ErrorTypeAuth
)

func categorizeError(err error) ErrorType {
    switch {
    case isNetworkError(err):
        return ErrorTypeTemporary
    case isValidationError(err):
        return ErrorTypeValidation
    case isAuthError(err):
        return ErrorTypeAuth
    case isRateLimitError(err):
        return ErrorTypeRateLimit
    default:
        return ErrorTypePermanent
    }
}
```

### Error Handler Middleware

```go
func ErrorHandlingMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            defer func() {
                if r := recover(); r != nil {
                    log.Error().
                        Interface("panic", r).
                        Str("message_id", m.ID).
                        Msg("Panic in message handler")

                    // Send panic to monitoring
                    sendPanicToMonitoring(r, m)

                    // Move to DLQ
                    m.MoveToDeadLetterQueue(ctx)
                }
            }()

            err := next(ctx, m)
            if err != nil {
                // Log error with context
                log.Error().
                    Err(err).
                    Str("message_id", m.ID).
                    Str("action", m.Action).
                    Int64("retry_count", m.RetryCount).
                    Msg("Message processing failed")

                // Send to error tracking
                sendToErrorTracking(err, m)
            }

            return err
        }
    }
}
```

---

## 📊 Monitoring & Observability

### Metrics Collection

```go
func MetricsMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            start := time.Now()

            // Increment processed counter
            messagesProcessed.WithLabelValues(m.Topic, m.Action).Inc()

            err := next(ctx, m)

            // Record duration
            duration := time.Since(start).Seconds()
            processingDuration.WithLabelValues(m.Topic, m.Action).Observe(duration)

            // Record result
            if err != nil {
                messagesErrors.WithLabelValues(m.Topic, m.Action).Inc()
            } else {
                messagesSuccess.WithLabelValues(m.Topic, m.Action).Inc()
            }

            return err
        }
    }
}
```

### Health Checks

```go
type ConsumerHealth struct {
    consumer    consumer.Consumer
    lastMessage time.Time
    mu          sync.RWMutex
}

func (h *ConsumerHealth) HealthCheck() error {
    h.mu.RLock()
    defer h.mu.RUnlock()

    // Check if we've received messages recently
    if time.Since(h.lastMessage) > 5*time.Minute {
        return errors.New("no messages received in last 5 minutes")
    }

    return nil
}

func (h *ConsumerHealth) UpdateLastMessage() {
    h.mu.Lock()
    h.lastMessage = time.Now()
    h.mu.Unlock()
}
```

---

## 🧪 Testing Consumers

### Unit Testing

```go
func TestMessageHandler(t *testing.T) {
    // Create test message
    msg := interfaces.InboundMessage{
        Message: interfaces.Message{
            ID:     "test-123",
            Topic:  "test",
            Action: "test.action",
            Data:   map[string]interface{}{"key": "value"},
        },
        RetryCount: 0,
        Ack: func(ctx context.Context) error {
            return nil
        },
        Nack: func(ctx context.Context) error {
            return nil
        },
        MoveToDeadLetterQueue: func(ctx context.Context) error {
            return nil
        },
        RetryWithDelayFn: func(ctx context.Context, delayFn interfaces.DelayFn) error {
            return nil
        },
    }

    // Test handler
    err := myHandler(context.Background(), msg)
    assert.NoError(t, err)
}
```

### Integration Testing

```go
func TestConsumerIntegration(t *testing.T) {
    // Setup test infrastructure
    testContainer := setupRabbitMQContainer(t)
    defer testContainer.Cleanup()

    // Create consumer
    consumer := consumer.NewConsumer(
        consumerOpts.ConsumerPlatformRabbitMQ,
        consumerOpts.WithQueueName("test-queue"),
    )

    // Test message handling
    processed := make(chan bool, 1)
    handler := func(ctx context.Context, m interfaces.InboundMessage) error {
        processed <- true
        return m.Ack(ctx)
    }

    // Start consumer
    go consumer.Consume(context.Background(),
        interfaces.InboundMessageHandlerFunc(handler), nil)

    // Publish test message
    publishTestMessage(t, "test-queue", testMessage)

    // Wait for processing
    select {
    case <-processed:
        // Success
    case <-time.After(5 * time.Second):
        t.Fatal("Message not processed within timeout")
    }
}
```

---

## 💡 Best Practices

### 1. **Graceful Shutdown**

```go
func gracefulShutdown(consumer consumer.Consumer) {
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)

    <-c
    log.Info().Msg("Shutting down consumer...")

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := consumer.Stop(ctx); err != nil {
        log.Error().Err(err).Msg("Error stopping consumer")
    }
}
```

### 2. **Idempotent Processing**

```go
func idempotentHandler(ctx context.Context, m interfaces.InboundMessage) error {
    // Check if message already processed
    if isProcessed(m.ID) {
        log.Info().Str("message_id", m.ID).Msg("Message already processed")
        return m.Ack(ctx)
    }

    // Process message
    if err := processMessage(m); err != nil {
        return err
    }

    // Mark as processed
    markAsProcessed(m.ID)

    return m.Ack(ctx)
}
```

### 3. **Circuit Breaker**

```go
func circuitBreakerHandler(ctx context.Context, m interfaces.InboundMessage) error {
    if circuitBreaker.IsOpen() {
        // Circuit is open, reject message
        return m.RetryWithDelayFn(ctx, func(retryCount int64) int64 {
            return 60 * 1000 // Wait 1 minute
        })
    }

    err := processMessage(m)
    if err != nil {
        circuitBreaker.RecordFailure()
        return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
    }

    circuitBreaker.RecordSuccess()
    return m.Ack(ctx)
}
```

---

## 🔧 Troubleshooting

### Common Issues

1. **Messages Not Being Consumed**

   - Check queue name and bindings
   - Verify consumer is running
   - Check connection status

2. **High Memory Usage**

   - Reduce batch size
   - Implement message pooling
   - Check for memory leaks in handlers

3. **Slow Processing**

   - Profile handler performance
   - Check database connections
   - Consider parallel processing

4. **Messages Going to DLQ**
   - Check error logs
   - Validate message format
   - Review retry logic

### Debug Configuration

```go
consumer := consumer.NewConsumer(
    consumerOpts.ConsumerPlatformRabbitMQ,
    consumerOpts.WithQueueName("debug-queue"),
    consumerOpts.WithMiddlewares(
        DebugMiddleware(),
        LoggingMiddleware(),
    ),
)
```

---

The GoQueue Consumer system provides a powerful foundation for building robust, scalable message processing applications. By following these patterns and best practices, you can build reliable systems that handle failures gracefully and scale with your needs.

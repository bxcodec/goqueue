# 🔌 GoQueue Middleware System

The GoQueue middleware system provides a powerful and flexible way to extend the functionality of your message processing pipeline. Middleware allows you to inject custom logic before and after message publishing/consuming operations without modifying the core business logic.

## 📖 Table of Contents

- [🔌 GoQueue Middleware System](#-goqueue-middleware-system)
  - [📖 Table of Contents](#-table-of-contents)
  - [🎯 Overview](#-overview)
  - [🏗️ How It Works](#️-how-it-works)
  - [📝 Types of Middleware](#-types-of-middleware)
  - [🚀 Basic Usage](#-basic-usage)
  - [🛠️ Creating Custom Middleware](#️-creating-custom-middleware)
  - [📊 Built-in Middleware Examples](#-built-in-middleware-examples)
  - [🎨 Advanced Patterns](#-advanced-patterns)
  - [📈 Performance Considerations](#-performance-considerations)
  - [🧪 Testing Middleware](#-testing-middleware)
  - [💡 Best Practices](#-best-practices)
  - [🔧 Troubleshooting](#-troubleshooting)

---

## 🎯 Overview

Middleware in GoQueue follows the **decorator pattern**, wrapping your core message handlers with additional functionality. This allows you to:

- **🔍 Monitor & Log** message processing
- **📊 Collect Metrics** and performance data
- **🛡️ Handle Authentication** and authorization
- **⚡ Implement Rate Limiting**
- **🔄 Add Retry Logic** with custom strategies
- **🔍 Trace Requests** across services
- **🧪 Test Message Flows** with mock data

## 🏗️ How It Works

GoQueue middleware uses function composition to create a pipeline of operations:

```
Request → Middleware 1 → Middleware 2 → Handler → Middleware 2 → Middleware 1 → Response
```

Each middleware can:

1. **Inspect/modify** the message before processing
2. **Execute custom logic** before calling the next handler
3. **Process the result** after the handler executes
4. **Handle errors** and implement retry logic

---

## 📝 Types of Middleware

### 1. **Consumer Middleware** (`InboundMessageHandlerMiddlewareFunc`)

Processes incoming messages before they reach your business logic.

```go
type InboundMessageHandlerMiddlewareFunc func(
    next InboundMessageHandlerFunc,
) InboundMessageHandlerFunc
```

### 2. **Publisher Middleware** (`PublisherMiddlewareFunc`)

Processes outgoing messages before they are sent to the queue.

```go
type PublisherMiddlewareFunc func(
    next PublisherFunc,
) PublisherFunc
```

---

## 🚀 Basic Usage

### Adding Middleware to Consumer

```go
package main

import (
    "github.com/bxcodec/goqueue/consumer"
    "github.com/bxcodec/goqueue/middleware"
    consumerOpts "github.com/bxcodec/goqueue/options/consumer"
)

func main() {
    consumer := consumer.NewConsumer(
        consumerOpts.ConsumerPlatformRabbitMQ,
        consumerOpts.WithQueueName("user-events"),
        consumerOpts.WithMiddlewares(
            // Middleware executes in order
            middleware.HelloWorldMiddlewareExecuteBeforeInboundMessageHandler(),
            LoggingMiddleware(),
            MetricsMiddleware(),
            middleware.HelloWorldMiddlewareExecuteAfterInboundMessageHandler(),
        ),
    )
}
```

### Adding Middleware to Publisher

```go
package main

import (
    "github.com/bxcodec/goqueue/publisher"
    "github.com/bxcodec/goqueue/middleware"
    publisherOpts "github.com/bxcodec/goqueue/options/publisher"
)

func main() {
    pub := publisher.NewPublisher(
        publisherOpts.PublisherPlatformRabbitMQ,
        publisherOpts.WithMiddlewares(
            middleware.HelloWorldMiddlewareExecuteBeforePublisher(),
            ValidationMiddleware(),
            CompressionMiddleware(),
            middleware.HelloWorldMiddlewareExecuteAfterPublisher(),
        ),
    )
}
```

---

## 🛠️ Creating Custom Middleware

### Consumer Middleware Example

```go
package main

import (
    "context"
    "time"

    "github.com/rs/zerolog/log"
    "github.com/bxcodec/goqueue/interfaces"
)

// LoggingMiddleware logs message processing details
func LoggingMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            start := time.Now()

            // Log before processing
            log.Info().
                Str("message_id", m.ID).
                Str("action", m.Action).
                Str("topic", m.Topic).
                Msg("Processing message")

            // Call the next handler
            err := next(ctx, m)

            // Log after processing
            duration := time.Since(start)
            logEvent := log.Info().
                Str("message_id", m.ID).
                Dur("duration", duration)

            if err != nil {
                logEvent = log.Error().
                    Str("message_id", m.ID).
                    Dur("duration", duration).
                    Err(err)
                logEvent.Msg("Message processing failed")
            } else {
                logEvent.Msg("Message processed successfully")
            }

            return err
        }
    }
}

// MetricsMiddleware collects processing metrics
func MetricsMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            start := time.Now()

            // Increment counter
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

// AuthenticationMiddleware validates message authentication
func AuthenticationMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            // Extract auth token from headers
            authToken, ok := m.Headers["authorization"].(string)
            if !ok {
                log.Error().Str("message_id", m.ID).Msg("Missing authorization header")
                return errors.New("unauthorized: missing auth token")
            }

            // Validate token
            user, err := validateToken(authToken)
            if err != nil {
                log.Error().
                    Str("message_id", m.ID).
                    Err(err).
                    Msg("Invalid authorization token")
                return fmt.Errorf("unauthorized: %w", err)
            }

            // Add user to context
            ctx = context.WithValue(ctx, "user", user)

            return next(ctx, m)
        }
    }
}
```

### Publisher Middleware Example

```go
// ValidationMiddleware validates outgoing messages
func ValidationMiddleware() interfaces.PublisherMiddlewareFunc {
    return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
        return func(ctx context.Context, m interfaces.Message) error {
            // Validate required fields
            if m.Topic == "" {
                return errors.New("validation error: topic is required")
            }

            if m.Action == "" {
                return errors.New("validation error: action is required")
            }

            if m.Data == nil {
                return errors.New("validation error: data is required")
            }

            // Validate data structure based on action
            if err := validateMessageData(m.Action, m.Data); err != nil {
                return fmt.Errorf("validation error: %w", err)
            }

            log.Info().
                Str("topic", m.Topic).
                Str("action", m.Action).
                Msg("Message validation passed")

            return next(ctx, m)
        }
    }
}

// CompressionMiddleware compresses large message payloads
func CompressionMiddleware() interfaces.PublisherMiddlewareFunc {
    return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
        return func(ctx context.Context, m interfaces.Message) error {
            // Serialize data to check size
            dataBytes, err := json.Marshal(m.Data)
            if err != nil {
                return fmt.Errorf("compression middleware: %w", err)
            }

            // Compress if data is large (> 1KB)
            if len(dataBytes) > 1024 {
                compressed, err := compressData(dataBytes)
                if err != nil {
                    return fmt.Errorf("compression failed: %w", err)
                }

                // Update message with compressed data
                m.Data = compressed
                if m.Headers == nil {
                    m.Headers = make(map[string]interface{})
                }
                m.Headers["compression"] = "gzip"

                log.Info().
                    Int("original_size", len(dataBytes)).
                    Int("compressed_size", len(compressed)).
                    Msg("Message compressed")
            }

            return next(ctx, m)
        }
    }
}
```

---

## 📊 Built-in Middleware Examples

### 1. **Error Handling Middleware**

```go
func ErrorHandlingMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            defer func() {
                if r := recover(); r != nil {
                    log.Error().
                        Interface("panic", r).
                        Str("message_id", m.ID).
                        Msg("Panic recovered in message handler")

                    // Send to dead letter queue
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
                    Interface("data", m.Data).
                    Msg("Message processing error")

                // Send to monitoring system
                sendToErrorTracking(err, m)
            }

            return err
        }
    }
}
```

### 2. **Rate Limiting Middleware**

```go
func RateLimitingMiddleware(limit int, window time.Duration) interfaces.InboundMessageHandlerMiddlewareFunc {
    limiter := make(map[string]*time.Timer)
    counter := make(map[string]int)
    mu := sync.RWMutex{}

    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            key := fmt.Sprintf("%s:%s", m.Topic, m.Action)

            mu.Lock()
            count := counter[key]

            if count >= limit {
                mu.Unlock()
                log.Warn().
                    Str("key", key).
                    Int("limit", limit).
                    Msg("Rate limit exceeded")

                // Requeue with delay
                return m.RetryWithDelayFn(ctx, func(retryCount int64) int64 {
                    return 30 // 30 second delay
                })
            }

            counter[key]++

            // Reset counter after window
            if limiter[key] == nil {
                limiter[key] = time.AfterFunc(window, func() {
                    mu.Lock()
                    delete(counter, key)
                    delete(limiter, key)
                    mu.Unlock()
                })
            }

            mu.Unlock()

            return next(ctx, m)
        }
    }
}
```

### 3. **Tracing Middleware**

```go
func TracingMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            // Extract trace context from headers
            traceID := extractTraceID(m.Headers)
            if traceID == "" {
                traceID = generateTraceID()
            }

            // Create span
            ctx, span := tracer.Start(ctx, "message.process",
                trace.WithAttributes(
                    attribute.String("message.id", m.ID),
                    attribute.String("message.topic", m.Topic),
                    attribute.String("message.action", m.Action),
                    attribute.String("trace.id", traceID),
                ),
            )
            defer span.End()

            // Add trace context to message
            if m.Headers == nil {
                m.Headers = make(map[string]interface{})
            }
            m.Headers["trace_id"] = traceID

            err := next(ctx, m)

            if err != nil {
                span.SetStatus(codes.Error, err.Error())
                span.RecordError(err)
            } else {
                span.SetStatus(codes.Ok, "Message processed successfully")
            }

            return err
        }
    }
}
```

---

## 🎨 Advanced Patterns

### 1. **Conditional Middleware**

```go
func ConditionalMiddleware(condition func(interfaces.InboundMessage) bool,
                          middleware interfaces.InboundMessageHandlerMiddlewareFunc) interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            if condition(m) {
                // Apply middleware
                return middleware(next)(ctx, m)
            }
            // Skip middleware
            return next(ctx, m)
        }
    }
}

// Usage
middleware := ConditionalMiddleware(
    func(m interfaces.InboundMessage) bool {
        return m.Topic == "payments" // Only apply to payment messages
    },
    AuthenticationMiddleware(),
)
```

### 2. **Circuit Breaker Middleware**

```go
func CircuitBreakerMiddleware(threshold int, timeout time.Duration) interfaces.InboundMessageHandlerMiddlewareFunc {
    var (
        failures    int64
        lastFailure time.Time
        mu          sync.RWMutex
    )

    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            mu.RLock()
            currentFailures := failures
            lastFail := lastFailure
            mu.RUnlock()

            // Check if circuit is open
            if currentFailures >= int64(threshold) && time.Since(lastFail) < timeout {
                log.Warn().
                    Int64("failures", currentFailures).
                    Dur("timeout", timeout).
                    Msg("Circuit breaker is open, rejecting message")

                return fmt.Errorf("circuit breaker open: too many failures")
            }

            err := next(ctx, m)

            mu.Lock()
            if err != nil {
                failures++
                lastFailure = time.Now()
            } else {
                // Reset on success
                failures = 0
            }
            mu.Unlock()

            return err
        }
    }
}
```

### 3. **Batching Middleware**

```go
func BatchingMiddleware(batchSize int, flushInterval time.Duration) interfaces.InboundMessageHandlerMiddlewareFunc {
    type batchItem struct {
        ctx context.Context
        msg interfaces.InboundMessage
        result chan error
    }

    batch := make([]batchItem, 0, batchSize)
    mu := sync.Mutex{}

    // Flush timer
    ticker := time.NewTicker(flushInterval)
    go func() {
        for range ticker.C {
            mu.Lock()
            if len(batch) > 0 {
                processBatch(batch)
                batch = batch[:0]
            }
            mu.Unlock()
        }
    }()

    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            result := make(chan error, 1)

            mu.Lock()
            batch = append(batch, batchItem{ctx, m, result})

            if len(batch) >= batchSize {
                processBatch(batch)
                batch = batch[:0]
            }
            mu.Unlock()

            return <-result
        }
    }
}
```

---

## 📈 Performance Considerations

### 1. **Middleware Order Matters**

- Place **fast, filtering** middleware first (auth, validation)
- Place **expensive operations** last (database calls, external APIs)
- Consider **early returns** to avoid unnecessary processing

### 2. **Memory Management**

```go
// ❌ Bad: Creates new logger for each message
func BadLoggingMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            logger := log.With().Str("message_id", m.ID).Logger() // New logger each time
            // ...
        }
    }
}

// ✅ Good: Reuse logger with context
func GoodLoggingMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            log.Info().Str("message_id", m.ID).Msg("Processing") // Direct usage
            // ...
        }
    }
}
```

### 3. **Async Operations**

```go
func AsyncMetricsMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    metricsChan := make(chan MetricEvent, 1000)

    // Background worker
    go func() {
        for metric := range metricsChan {
            sendToMetricsSystem(metric)
        }
    }()

    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            start := time.Now()
            err := next(ctx, m)

            // Non-blocking metrics send
            select {
            case metricsChan <- MetricEvent{
                Duration: time.Since(start),
                Topic:    m.Topic,
                Action:   m.Action,
                Error:    err,
            }:
            default:
                // Drop metric if channel is full
            }

            return err
        }
    }
}
```

---

## 🧪 Testing Middleware

### Unit Testing

```go
func TestLoggingMiddleware(t *testing.T) {
    // Create test message
    msg := interfaces.InboundMessage{
        Message: interfaces.Message{
            ID:     "test-123",
            Topic:  "test-topic",
            Action: "test.action",
            Data:   map[string]interface{}{"key": "value"},
        },
    }

    // Mock handler
    called := false
    handler := func(ctx context.Context, m interfaces.InboundMessage) error {
        called = true
        return nil
    }

    // Apply middleware
    middleware := LoggingMiddleware()
    wrappedHandler := middleware(handler)

    // Execute
    err := wrappedHandler(context.Background(), msg)

    // Assertions
    assert.NoError(t, err)
    assert.True(t, called)
}
```

### Integration Testing

```go
func TestMiddlewareChain(t *testing.T) {
    // Create consumer with middleware
    consumer := consumer.NewConsumer(
        consumerOpts.ConsumerPlatformRabbitMQ,
        consumerOpts.WithMiddlewares(
            LoggingMiddleware(),
            MetricsMiddleware(),
            AuthenticationMiddleware(),
        ),
    )

    // Test message processing
    // ... integration test logic
}
```

---

## 💡 Best Practices

### 1. **Keep Middleware Focused**

Each middleware should have a single responsibility:

```go
// ✅ Good: Single responsibility
func LoggingMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc { ... }
func MetricsMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc { ... }

// ❌ Bad: Multiple responsibilities
func LoggingAndMetricsMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc { ... }
```

### 2. **Handle Errors Gracefully**

```go
func ResilientMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            // Always call next, even if middleware logic fails
            defer func() {
                if r := recover(); r != nil {
                    log.Error().Interface("panic", r).Msg("Middleware panic recovered")
                }
            }()

            // Middleware logic here...
            // Don't block the main flow on auxiliary operations

            return next(ctx, m)
        }
    }
}
```

### 3. **Use Context for Request-Scoped Data**

```go
func ContextEnrichmentMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            // Add request ID to context
            requestID := generateRequestID()
            ctx = context.WithValue(ctx, "request_id", requestID)

            // Add to message headers for downstream services
            if m.Headers == nil {
                m.Headers = make(map[string]interface{})
            }
            m.Headers["request_id"] = requestID

            return next(ctx, m)
        }
    }
}
```

### 4. **Configuration via Options**

```go
type LoggingConfig struct {
    Level     string
    Fields    []string
    Structured bool
}

func LoggingMiddleware(config LoggingConfig) interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            // Use config for customizable behavior
            // ...
        }
    }
}
```

---

## 🔧 Troubleshooting

### Common Issues

1. **Middleware Not Executing**

   - Check middleware order
   - Ensure middleware returns a function
   - Verify middleware is properly registered

2. **Performance Degradation**

   - Profile middleware execution time
   - Check for blocking operations
   - Consider async processing for heavy operations

3. **Context Cancellation**

   - Always respect context cancellation
   - Use `ctx.Done()` in long-running operations

4. **Memory Leaks**
   - Avoid storing message references
   - Clean up resources in defer statements
   - Use object pools for frequently created objects

### Debug Middleware

```go
func DebugMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            log.Debug().
                Str("middleware", "debug").
                Str("message_id", m.ID).
                Interface("headers", m.Headers).
                Interface("data", m.Data).
                Msg("Message received in debug middleware")

            err := next(ctx, m)

            log.Debug().
                Str("middleware", "debug").
                Str("message_id", m.ID).
                Err(err).
                Msg("Message processed in debug middleware")

            return err
        }
    }
}
```

---

The middleware system in GoQueue provides powerful extensibility while maintaining clean separation of concerns. By following these patterns and best practices, you can build robust, maintainable message processing pipelines that scale with your application needs.

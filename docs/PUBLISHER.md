# 📤 GoQueue Publisher System

The GoQueue Publisher system provides a robust, high-performance way to publish messages to various queue platforms. It handles connection management, message serialization, error handling, and provides extensibility through middleware.

## 📖 Table of Contents

- [📤 GoQueue Publisher System](#-goqueue-publisher-system)
  - [📖 Table of Contents](#-table-of-contents)
  - [🎯 Overview](#-overview)
  - [🏗️ Architecture](#️-architecture)
  - [🚀 Quick Start](#-quick-start)
  - [⚙️ Configuration](#️-configuration)
  - [📝 Message Publishing](#-message-publishing)
  - [🔌 Middleware System](#-middleware-system)
  - [📊 Connection Management](#-connection-management)
  - [🛡️ Error Handling](#️-error-handling)
  - [📈 Performance Optimization](#-performance-optimization)
  - [🎨 Advanced Usage](#-advanced-usage)
  - [🧪 Testing Publishers](#-testing-publishers)
  - [📊 Monitoring & Observability](#-monitoring--observability)
  - [💡 Best Practices](#-best-practices)
  - [🔧 Troubleshooting](#-troubleshooting)

---

## 🎯 Overview

The Publisher system in GoQueue provides:

- **🚀 High Performance** with connection pooling and batching
- **🔌 Middleware Support** for extending functionality
- **🛡️ Error Handling** with retry and circuit breaker patterns
- **📊 Built-in Observability** with logging and metrics hooks
- **⚡ Async Publishing** with optional delivery confirmations
- **🎛️ Flexible Configuration** for different use cases

## 🏗️ Architecture

```
┌─────────────────────┐
│   Application       │
│     Code            │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│    GoQueue          │
│    Publisher        │
│                     │
│ ┌─────────────────┐ │
│ │   Middleware    │ │
│ │     Chain       │ │
│ └─────────────────┘ │
│ ┌─────────────────┐ │
│ │   Serializer    │ │
│ └─────────────────┘ │
│ ┌─────────────────┐ │
│ │  Connection     │ │
│ │     Pool        │ │
│ └─────────────────┘ │
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   Queue Platform    │
│    (RabbitMQ)       │
└─────────────────────┘
```

---

## 🚀 Quick Start

### Basic Publisher Setup

```go
package main

import (
    "context"
    "log"

    "github.com/bxcodec/goqueue/publisher"
    "github.com/bxcodec/goqueue/interfaces"
    publisherOpts "github.com/bxcodec/goqueue/options/publisher"
    amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
    // Connect to RabbitMQ
    conn, err := amqp.Dial("amqp://localhost:5672/")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    // Create publisher
    pub := publisher.NewPublisher(
        publisherOpts.PublisherPlatformRabbitMQ,
        publisherOpts.WithRabbitMQPublisherConfig(&publisherOpts.RabbitMQPublisherConfig{
            Conn:                     conn,
            PublisherChannelPoolSize: 10,
        }),
        publisherOpts.WithPublisherID("user-service"),
    )
    defer pub.Close(context.Background())

    // Publish a message
    message := interfaces.Message{
        ID:      "msg-123",
        Topic:   "users",
        Action:  "user.created",
        Data: map[string]interface{}{
            "user_id": 12345,
            "email":   "user@example.com",
            "name":    "John Doe",
        },
        Headers: map[string]interface{}{
            "source":      "user-service",
            "correlation": "req-456",
        },
    }

    err = pub.Publish(context.Background(), message)
    if err != nil {
        log.Printf("Failed to publish message: %v", err)
    } else {
        log.Printf("Message published successfully: %s", message.ID)
    }
}
```

---

## ⚙️ Configuration

### Publisher Options

```go
pub := publisher.NewPublisher(
    publisherOpts.PublisherPlatformRabbitMQ,

    // Basic Configuration
    publisherOpts.WithPublisherID("service-01"),

    // Middleware
    publisherOpts.WithMiddlewares(
        middleware.ValidationMiddleware(),
        middleware.CompressionMiddleware(),
        middleware.MetricsMiddleware(),
        middleware.LoggingMiddleware(),
    ),

    // Platform-specific configuration
    publisherOpts.WithRabbitMQPublisherConfig(&publisherOpts.RabbitMQPublisherConfig{
        Conn:                     connection,
        PublisherChannelPoolSize: 20,
        ExchangeName:            "main-exchange",
        Mandatory:               false,
        Immediate:               false,
        DefaultHeaders: map[string]interface{}{
            "version": "1.0",
            "service": "my-service",
        },
    }),
)
```

### Configuration Options Explained

| Option                     | Description                                   | Default        |
| -------------------------- | --------------------------------------------- | -------------- |
| `PublisherID`              | Unique identifier for this publisher instance | Auto-generated |
| `PublisherChannelPoolSize` | Number of channels in the connection pool     | 5              |
| `ExchangeName`             | Default exchange for publishing               | ""             |
| `Mandatory`                | Return unroutable messages                    | false          |
| `Immediate`                | Return undeliverable messages                 | false          |
| `DefaultHeaders`           | Headers added to all messages                 | Empty          |

---

## 📝 Message Publishing

### Message Structure

```go
type Message struct {
    ID          string                 `json:"id"`          // Unique message identifier
    Topic       string                 `json:"topic"`       // Message topic/exchange
    Action      string                 `json:"action"`      // Action type/routing key
    Data        interface{}           `json:"data"`        // Message payload
    Headers     map[string]interface{} `json:"headers"`     // Additional metadata
    Timestamp   time.Time             `json:"timestamp"`   // Message timestamp
    ContentType string                `json:"contentType"` // Content type (JSON, etc.)
}
```

### Publishing Patterns

#### 1. Simple Publishing

```go
func publishUserEvent(pub publisher.Publisher, userID int, action string) error {
    message := interfaces.Message{
        ID:     generateMessageID(),
        Topic:  "users",
        Action: action,
        Data: map[string]interface{}{
            "user_id": userID,
            "timestamp": time.Now(),
        },
    }

    return pub.Publish(context.Background(), message)
}
```

#### 2. Batch Publishing

```go
func publishBatch(pub publisher.Publisher, messages []interfaces.Message) error {
    ctx := context.Background()

    for _, msg := range messages {
        if err := pub.Publish(ctx, msg); err != nil {
            return fmt.Errorf("failed to publish message %s: %w", msg.ID, err)
        }
    }

    return nil
}
```

#### 3. Publishing with Context

```go
func publishWithTimeout(pub publisher.Publisher, msg interfaces.Message) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return pub.Publish(ctx, msg)
}
```

#### 4. Publishing with Correlation

```go
func publishWithCorrelation(pub publisher.Publisher, correlationID string, data interface{}) error {
    message := interfaces.Message{
        ID:     generateMessageID(),
        Topic:  "events",
        Action: "data.processed",
        Data:   data,
        Headers: map[string]interface{}{
            "correlation_id": correlationID,
            "reply_to":      "response-queue",
        },
    }

    return pub.Publish(context.Background(), message)
}
```

### Message Builders

```go
type MessageBuilder struct {
    message interfaces.Message
}

func NewMessageBuilder() *MessageBuilder {
    return &MessageBuilder{
        message: interfaces.Message{
            ID:        generateMessageID(),
            Timestamp: time.Now(),
            Headers:   make(map[string]interface{}),
        },
    }
}

func (b *MessageBuilder) Topic(topic string) *MessageBuilder {
    b.message.Topic = topic
    return b
}

func (b *MessageBuilder) Action(action string) *MessageBuilder {
    b.message.Action = action
    return b
}

func (b *MessageBuilder) Data(data interface{}) *MessageBuilder {
    b.message.Data = data
    return b
}

func (b *MessageBuilder) Header(key string, value interface{}) *MessageBuilder {
    b.message.Headers[key] = value
    return b
}

func (b *MessageBuilder) Build() interfaces.Message {
    return b.message
}

// Usage
message := NewMessageBuilder().
    Topic("users").
    Action("user.created").
    Data(userData).
    Header("source", "user-service").
    Build()
```

---

## 🔌 Middleware System

### Built-in Middleware

#### 1. Validation Middleware

```go
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

            // Validate data structure
            if err := validateMessageData(m.Action, m.Data); err != nil {
                return fmt.Errorf("validation error: %w", err)
            }

            return next(ctx, m)
        }
    }
}
```

#### 2. Compression Middleware

```go
func CompressionMiddleware(threshold int) interfaces.PublisherMiddlewareFunc {
    return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
        return func(ctx context.Context, m interfaces.Message) error {
            // Serialize data to check size
            dataBytes, err := json.Marshal(m.Data)
            if err != nil {
                return fmt.Errorf("compression middleware: %w", err)
            }

            // Compress if data is large
            if len(dataBytes) > threshold {
                compressed, err := compressData(dataBytes)
                if err != nil {
                    return fmt.Errorf("compression failed: %w", err)
                }

                // Update message
                m.Data = base64.StdEncoding.EncodeToString(compressed)
                if m.Headers == nil {
                    m.Headers = make(map[string]interface{})
                }
                m.Headers["encoding"] = "gzip+base64"
                m.Headers["original_size"] = len(dataBytes)

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

#### 3. Metrics Middleware

```go
func MetricsMiddleware() interfaces.PublisherMiddlewareFunc {
    return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
        return func(ctx context.Context, m interfaces.Message) error {
            start := time.Now()

            // Increment counter
            messagesPublished.WithLabelValues(m.Topic, m.Action).Inc()

            err := next(ctx, m)

            // Record duration
            duration := time.Since(start).Seconds()
            publishDuration.WithLabelValues(m.Topic, m.Action).Observe(duration)

            // Record result
            if err != nil {
                publishErrors.WithLabelValues(m.Topic, m.Action).Inc()
            } else {
                publishSuccess.WithLabelValues(m.Topic, m.Action).Inc()
            }

            return err
        }
    }
}
```

#### 4. Encryption Middleware

```go
func EncryptionMiddleware(key []byte) interfaces.PublisherMiddlewareFunc {
    return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
        return func(ctx context.Context, m interfaces.Message) error {
            // Encrypt sensitive data
            if shouldEncrypt(m) {
                encrypted, err := encryptData(m.Data, key)
                if err != nil {
                    return fmt.Errorf("encryption failed: %w", err)
                }

                m.Data = encrypted
                if m.Headers == nil {
                    m.Headers = make(map[string]interface{})
                }
                m.Headers["encrypted"] = true
                m.Headers["algorithm"] = "AES-256-GCM"
            }

            return next(ctx, m)
        }
    }
}
```

---

## 📊 Connection Management

### Connection Pooling

```go
type ConnectionPool struct {
    channels chan *amqp.Channel
    conn     *amqp.Connection
    size     int
    mu       sync.RWMutex
}

func NewConnectionPool(conn *amqp.Connection, size int) *ConnectionPool {
    pool := &ConnectionPool{
        channels: make(chan *amqp.Channel, size),
        conn:     conn,
        size:     size,
    }

    // Pre-create channels
    for i := 0; i < size; i++ {
        ch, err := conn.Channel()
        if err != nil {
            log.Error().Err(err).Msg("Failed to create channel")
            continue
        }
        pool.channels <- ch
    }

    return pool
}

func (p *ConnectionPool) GetChannel() (*amqp.Channel, error) {
    select {
    case ch := <-p.channels:
        return ch, nil
    case <-time.After(1 * time.Second):
        return nil, errors.New("timeout getting channel from pool")
    }
}

func (p *ConnectionPool) ReturnChannel(ch *amqp.Channel) {
    select {
    case p.channels <- ch:
    default:
        // Pool is full, close the channel
        ch.Close()
    }
}
```

### Connection Health Monitoring

```go
type HealthMonitor struct {
    conn     *amqp.Connection
    callback func(bool)
    mu       sync.RWMutex
    healthy  bool
}

func NewHealthMonitor(conn *amqp.Connection, callback func(bool)) *HealthMonitor {
    monitor := &HealthMonitor{
        conn:     conn,
        callback: callback,
        healthy:  true,
    }

    go monitor.monitor()
    return monitor
}

func (h *HealthMonitor) monitor() {
    for {
        select {
        case <-h.conn.NotifyClose(make(chan *amqp.Error)):
            h.setHealthy(false)
            h.callback(false)
        case <-time.After(30 * time.Second):
            // Periodic health check
            if h.conn.IsClosed() {
                h.setHealthy(false)
                h.callback(false)
            } else if !h.IsHealthy() {
                h.setHealthy(true)
                h.callback(true)
            }
        }
    }
}

func (h *HealthMonitor) IsHealthy() bool {
    h.mu.RLock()
    defer h.mu.RUnlock()
    return h.healthy
}

func (h *HealthMonitor) setHealthy(healthy bool) {
    h.mu.Lock()
    h.healthy = healthy
    h.mu.Unlock()
}
```

---

## 🛡️ Error Handling

### Retry Strategies

```go
type RetryConfig struct {
    MaxAttempts int
    InitialDelay time.Duration
    MaxDelay     time.Duration
    Multiplier   float64
}

func RetryMiddleware(config RetryConfig) interfaces.PublisherMiddlewareFunc {
    return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
        return func(ctx context.Context, m interfaces.Message) error {
            var err error
            delay := config.InitialDelay

            for attempt := 0; attempt < config.MaxAttempts; attempt++ {
                err = next(ctx, m)
                if err == nil {
                    return nil
                }

                // Check if error is retryable
                if !isRetryableError(err) {
                    return err
                }

                // Don't retry on last attempt
                if attempt == config.MaxAttempts-1 {
                    break
                }

                // Wait before retry
                select {
                case <-ctx.Done():
                    return ctx.Err()
                case <-time.After(delay):
                }

                // Exponential backoff
                delay = time.Duration(float64(delay) * config.Multiplier)
                if delay > config.MaxDelay {
                    delay = config.MaxDelay
                }

                log.Warn().
                    Err(err).
                    Int("attempt", attempt+1).
                    Dur("delay", delay).
                    Str("message_id", m.ID).
                    Msg("Retrying message publish")
            }

            return fmt.Errorf("failed to publish after %d attempts: %w",
                config.MaxAttempts, err)
        }
    }
}
```

### Circuit Breaker

```go
type CircuitBreaker struct {
    threshold    int
    timeout      time.Duration
    failures     int64
    lastFailure  time.Time
    state        string // "closed", "open", "half-open"
    mu           sync.RWMutex
}

func CircuitBreakerMiddleware(threshold int, timeout time.Duration) interfaces.PublisherMiddlewareFunc {
    cb := &CircuitBreaker{
        threshold: threshold,
        timeout:   timeout,
        state:     "closed",
    }

    return func(next interfaces.PublisherFunc) interfaces.PublisherFunc {
        return func(ctx context.Context, m interfaces.Message) error {
            if !cb.AllowRequest() {
                return errors.New("circuit breaker is open")
            }

            err := next(ctx, m)

            if err != nil {
                cb.RecordFailure()
            } else {
                cb.RecordSuccess()
            }

            return err
        }
    }
}

func (cb *CircuitBreaker) AllowRequest() bool {
    cb.mu.RLock()
    defer cb.mu.RUnlock()

    switch cb.state {
    case "closed":
        return true
    case "open":
        return time.Since(cb.lastFailure) > cb.timeout
    case "half-open":
        return true
    default:
        return false
    }
}

func (cb *CircuitBreaker) RecordFailure() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures++
    cb.lastFailure = time.Now()

    if cb.failures >= int64(cb.threshold) {
        cb.state = "open"
        log.Warn().Msg("Circuit breaker opened")
    }
}

func (cb *CircuitBreaker) RecordSuccess() {
    cb.mu.Lock()
    defer cb.mu.Unlock()

    cb.failures = 0
    cb.state = "closed"
}
```

---

## 📈 Performance Optimization

### Async Publishing

```go
type AsyncPublisher struct {
    publisher publisher.Publisher
    queue     chan PublishRequest
    workers   int
}

type PublishRequest struct {
    Message  interfaces.Message
    Context  context.Context
    Response chan error
}

func NewAsyncPublisher(pub publisher.Publisher, workers int) *AsyncPublisher {
    ap := &AsyncPublisher{
        publisher: pub,
        queue:     make(chan PublishRequest, 1000),
        workers:   workers,
    }

    // Start workers
    for i := 0; i < workers; i++ {
        go ap.worker()
    }

    return ap
}

func (ap *AsyncPublisher) worker() {
    for req := range ap.queue {
        err := ap.publisher.Publish(req.Context, req.Message)
        req.Response <- err
        close(req.Response)
    }
}

func (ap *AsyncPublisher) PublishAsync(ctx context.Context, msg interfaces.Message) <-chan error {
    response := make(chan error, 1)

    req := PublishRequest{
        Message:  msg,
        Context:  ctx,
        Response: response,
    }

    select {
    case ap.queue <- req:
        return response
    case <-ctx.Done():
        response <- ctx.Err()
        close(response)
        return response
    }
}

// Usage
errChan := asyncPublisher.PublishAsync(ctx, message)
go func() {
    if err := <-errChan; err != nil {
        log.Error().Err(err).Msg("Async publish failed")
    }
}()
```

### Message Batching

```go
type BatchPublisher struct {
    publisher   publisher.Publisher
    batchSize   int
    flushTime   time.Duration
    batch       []interfaces.Message
    mu          sync.Mutex
    lastFlush   time.Time
}

func NewBatchPublisher(pub publisher.Publisher, batchSize int, flushTime time.Duration) *BatchPublisher {
    bp := &BatchPublisher{
        publisher: pub,
        batchSize: batchSize,
        flushTime: flushTime,
        batch:     make([]interfaces.Message, 0, batchSize),
        lastFlush: time.Now(),
    }

    // Auto-flush timer
    go bp.autoFlush()

    return bp
}

func (bp *BatchPublisher) Publish(ctx context.Context, msg interfaces.Message) error {
    bp.mu.Lock()
    defer bp.mu.Unlock()

    bp.batch = append(bp.batch, msg)

    if len(bp.batch) >= bp.batchSize {
        return bp.flush(ctx)
    }

    return nil
}

func (bp *BatchPublisher) flush(ctx context.Context) error {
    if len(bp.batch) == 0 {
        return nil
    }

    batch := make([]interfaces.Message, len(bp.batch))
    copy(batch, bp.batch)
    bp.batch = bp.batch[:0]
    bp.lastFlush = time.Now()

    // Publish batch
    for _, msg := range batch {
        if err := bp.publisher.Publish(ctx, msg); err != nil {
            return err
        }
    }

    return nil
}

func (bp *BatchPublisher) autoFlush() {
    ticker := time.NewTicker(bp.flushTime)
    defer ticker.Stop()

    for range ticker.C {
        bp.mu.Lock()
        if time.Since(bp.lastFlush) >= bp.flushTime && len(bp.batch) > 0 {
            bp.flush(context.Background())
        }
        bp.mu.Unlock()
    }
}
```

---

## 🎨 Advanced Usage

### Message Routing

```go
type Router struct {
    routes map[string]publisher.Publisher
    fallback publisher.Publisher
}

func NewRouter(fallback publisher.Publisher) *Router {
    return &Router{
        routes:   make(map[string]publisher.Publisher),
        fallback: fallback,
    }
}

func (r *Router) AddRoute(pattern string, pub publisher.Publisher) {
    r.routes[pattern] = pub
}

func (r *Router) Publish(ctx context.Context, msg interfaces.Message) error {
    // Find matching route
    for pattern, pub := range r.routes {
        if matched, _ := filepath.Match(pattern, msg.Topic); matched {
            return pub.Publish(ctx, msg)
        }
    }

    // Use fallback
    return r.fallback.Publish(ctx, msg)
}

// Usage
router := NewRouter(defaultPublisher)
router.AddRoute("user.*", userPublisher)
router.AddRoute("order.*", orderPublisher)
router.AddRoute("payment.*", paymentPublisher)
```

### Priority Publishing

```go
type PriorityPublisher struct {
    high   publisher.Publisher
    normal publisher.Publisher
    low    publisher.Publisher
}

func (pp *PriorityPublisher) Publish(ctx context.Context, msg interfaces.Message) error {
    priority := getPriority(msg)

    switch priority {
    case "high":
        return pp.high.Publish(ctx, msg)
    case "low":
        return pp.low.Publish(ctx, msg)
    default:
        return pp.normal.Publish(ctx, msg)
    }
}

func getPriority(msg interfaces.Message) string {
    if priority, ok := msg.Headers["priority"].(string); ok {
        return priority
    }

    // Default priority based on action
    switch {
    case strings.HasPrefix(msg.Action, "alert."):
        return "high"
    case strings.HasPrefix(msg.Action, "analytics."):
        return "low"
    default:
        return "normal"
    }
}
```

---

## 🧪 Testing Publishers

### Unit Testing

```go
func TestPublisher(t *testing.T) {
    // Create mock publisher
    mockPub := &mockPublisher{}

    // Test message
    msg := interfaces.Message{
        ID:     "test-123",
        Topic:  "test",
        Action: "test.action",
        Data:   map[string]interface{}{"key": "value"},
    }

    // Test publishing
    err := mockPub.Publish(context.Background(), msg)
    assert.NoError(t, err)

    // Verify mock was called
    mockPub.AssertExpectations(t)
}

type mockPublisher struct {
    mock.Mock
}

func (m *mockPublisher) Publish(ctx context.Context, msg interfaces.Message) error {
    args := m.Called(ctx, msg)
    return args.Error(0)
}

func (m *mockPublisher) Close(ctx context.Context) error {
    args := m.Called(ctx)
    return args.Error(0)
}
```

### Integration Testing

```go
func TestPublisherIntegration(t *testing.T) {
    // Setup test infrastructure
    testContainer := setupRabbitMQContainer(t)
    defer testContainer.Cleanup()

    // Create publisher
    pub := publisher.NewPublisher(
        publisherOpts.PublisherPlatformRabbitMQ,
        publisherOpts.WithRabbitMQPublisherConfig(&publisherOpts.RabbitMQPublisherConfig{
            Conn: testContainer.Connection,
        }),
    )
    defer pub.Close(context.Background())

    // Test message
    msg := interfaces.Message{
        ID:     "integration-test-123",
        Topic:  "test-topic",
        Action: "test.action",
        Data:   map[string]interface{}{"test": true},
    }

    // Publish message
    err := pub.Publish(context.Background(), msg)
    assert.NoError(t, err)

    // Verify message was published
    verifyMessagePublished(t, testContainer, msg)
}
```

---

## 📊 Monitoring & Observability

### Metrics Collection

```go
var (
    publishedMessages = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "goqueue_messages_published_total",
            Help: "Total number of messages published",
        },
        []string{"topic", "action", "status"},
    )

    publishDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "goqueue_publish_duration_seconds",
            Help: "Time spent publishing messages",
        },
        []string{"topic", "action"},
    )
)

func init() {
    prometheus.MustRegister(publishedMessages)
    prometheus.MustRegister(publishDuration)
}
```

### Health Checks

```go
type PublisherHealth struct {
    publisher   publisher.Publisher
    lastPublish time.Time
    mu          sync.RWMutex
}

func (h *PublisherHealth) HealthCheck() error {
    // Test publish
    testMsg := interfaces.Message{
        ID:     "health-check",
        Topic:  "health",
        Action: "ping",
        Data:   map[string]interface{}{"timestamp": time.Now()},
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    return h.publisher.Publish(ctx, testMsg)
}
```

---

## 💡 Best Practices

### 1. **Message Design**

```go
// ✅ Good: Clear, structured message
message := interfaces.Message{
    ID:     generateUUIDv4(),
    Topic:  "user-events",
    Action: "user.profile.updated",
    Data: UserProfileUpdatedEvent{
        UserID:    12345,
        Changes:   []string{"email", "name"},
        UpdatedBy: "admin",
        Timestamp: time.Now(),
    },
    Headers: map[string]interface{}{
        "version":       "1.0",
        "source":        "user-service",
        "correlation":   request.ID,
        "content-type":  "application/json",
    },
}

// ❌ Bad: Unclear, unstructured message
message := interfaces.Message{
    Topic:  "events",
    Action: "update",
    Data:   "user123|email@example.com|John Doe",
}
```

### 2. **Error Handling**

```go
func publishWithRetry(pub publisher.Publisher, msg interfaces.Message) error {
    const maxRetries = 3
    var err error

    for i := 0; i < maxRetries; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        err = pub.Publish(ctx, msg)
        cancel()

        if err == nil {
            return nil
        }

        // Check if error is retryable
        if !isRetryableError(err) {
            return err
        }

        // Exponential backoff
        time.Sleep(time.Duration(1<<uint(i)) * time.Second)
    }

    return fmt.Errorf("failed after %d retries: %w", maxRetries, err)
}
```

### 3. **Resource Management**

```go
func usePublisher() error {
    pub := publisher.NewPublisher(/* config */)
    defer func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := pub.Close(ctx); err != nil {
            log.Error().Err(err).Msg("Error closing publisher")
        }
    }()

    // Use publisher
    return pub.Publish(context.Background(), message)
}
```

### 4. **Schema Versioning**

```go
type VersionedMessage struct {
    interfaces.Message
    SchemaVersion string `json:"schema_version"`
}

func publishVersionedMessage(pub publisher.Publisher, data interface{}, version string) error {
    msg := VersionedMessage{
        Message: interfaces.Message{
            ID:     generateMessageID(),
            Topic:  "versioned-events",
            Action: "data.updated",
            Data:   data,
        },
        SchemaVersion: version,
    }

    return pub.Publish(context.Background(), msg.Message)
}
```

---

## 🔧 Troubleshooting

### Common Issues

1. **Connection Failures**

   - Check network connectivity
   - Verify credentials
   - Check connection pool size

2. **Publishing Timeouts**

   - Increase context timeout
   - Check queue server performance
   - Monitor connection health

3. **Message Loss**

   - Enable publisher confirms
   - Use mandatory/immediate flags
   - Implement retry logic

4. **Memory Leaks**
   - Close publishers properly
   - Monitor connection pools
   - Check for goroutine leaks

### Debug Tools

```go
func DebugPublisher(pub publisher.Publisher) publisher.Publisher {
    return &debugPublisher{
        wrapped: pub,
    }
}

type debugPublisher struct {
    wrapped publisher.Publisher
}

func (d *debugPublisher) Publish(ctx context.Context, msg interfaces.Message) error {
    start := time.Now()

    log.Debug().
        Str("message_id", msg.ID).
        Str("topic", msg.Topic).
        Str("action", msg.Action).
        Interface("data", msg.Data).
        Msg("Publishing message")

    err := d.wrapped.Publish(ctx, msg)

    log.Debug().
        Str("message_id", msg.ID).
        Dur("duration", time.Since(start)).
        Err(err).
        Msg("Message published")

    return err
}

func (d *debugPublisher) Close(ctx context.Context) error {
    return d.wrapped.Close(ctx)
}
```

---

The GoQueue Publisher system provides a robust foundation for building reliable, high-performance message publishing applications. By following these patterns and best practices, you can build scalable systems that handle failures gracefully and deliver messages reliably.

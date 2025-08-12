# 📚 GoQueue Documentation

Welcome to the comprehensive documentation for GoQueue - the universal Go message queue library. This documentation provides in-depth guides for each component of the system.

## 📖 Documentation Index

### 🔌 [Middleware System](MIDDLEWARE.md)

Learn how to extend GoQueue's functionality using the powerful middleware system.

**What you'll learn:**

- How middleware works in GoQueue
- Creating custom middleware for consumers and publishers
- Built-in middleware examples (logging, metrics, validation, compression)
- Advanced patterns (conditional middleware, circuit breakers, batching)
- Performance considerations and best practices
- Testing middleware components

**Key Topics:**

- Consumer and Publisher middleware
- Error handling middleware
- Rate limiting and circuit breaker patterns
- Tracing and observability middleware
- Custom middleware development

---

### 📨 [Consumer System](CONSUMER.md)

Master the art of consuming and processing messages reliably.

**What you'll learn:**

- Setting up and configuring consumers
- Message handling patterns and best practices
- Retry mechanisms and error handling strategies
- Monitoring and observability for consumers
- Performance tuning and optimization

**Key Topics:**

- Message acknowledgment strategies
- Retry patterns and dead letter queues
- Concurrent processing and scaling
- Health checks and graceful shutdown
- Testing consumer logic

---

### 📤 [Publisher System](PUBLISHER.md)

Build robust, high-performance message publishing systems.

**What you'll learn:**

- Publisher configuration and setup
- Message structure and design patterns
- Connection management and pooling
- Error handling and retry strategies
- Performance optimization techniques

**Key Topics:**

- Message builders and serialization
- Async publishing and batching
- Connection health monitoring
- Circuit breaker patterns
- Metrics and observability

---

### 🔄 [RabbitMQ Retry Architecture](RABBITMQ-RETRY.md)

Deep dive into GoQueue's sophisticated retry mechanism for RabbitMQ.

**What you'll learn:**

- How the retry architecture works internally
- Queue topology and message flow
- Configuration options and strategies
- Monitoring retry operations
- Troubleshooting retry issues
- Performance considerations

**Key Topics:**

- Dead letter exchange patterns
- TTL-based retry delays
- Exponential backoff strategies
- Retry queue management
- Failure analysis and debugging

---

## 🚀 Quick Start Guide

If you're new to GoQueue, start with these steps:

1. **📦 Installation**

   ```bash
   go get -u github.com/bxcodec/goqueue
   ```

2. **🎯 Choose Your Platform**

   - Currently supported: RabbitMQ
   - Coming soon: Google Pub/Sub, AWS SQS, Apache Kafka

3. **📖 Read the Basics**

   - Start with the main [README](../README.md)
   - Review the [examples](../examples/) directory
   - Check out the [Consumer](CONSUMER.md) and [Publisher](PUBLISHER.md) docs

4. **🔧 Advanced Features**
   - Explore [Middleware](MIDDLEWARE.md) for extensibility
   - Learn about [RabbitMQ Retry](RABBITMQ-RETRY.md) for resilience

## 🎯 Use Case Guides

### Event-Driven Architecture

```go
// Publisher side
publisher.Publish(ctx, interfaces.Message{
    Topic:  "user-events",
    Action: "user.created",
    Data:   userData,
})

// Consumer side
func handleUserEvent(ctx context.Context, m interfaces.InboundMessage) error {
    switch m.Action {
    case "user.created":
        return handleUserCreated(ctx, m)
    case "user.updated":
        return handleUserUpdated(ctx, m)
    }
}
```

### Microservices Communication

```go
// Service A publishes
publisher.Publish(ctx, interfaces.Message{
    Topic:  "orders",
    Action: "order.placed",
    Data:   orderData,
    Headers: map[string]interface{}{
        "correlation_id": requestID,
        "reply_to":      "order-responses",
    },
})

// Service B consumes and processes
func processOrder(ctx context.Context, m interfaces.InboundMessage) error {
    // Process order
    if err := orderService.Process(m.Data); err != nil {
        return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
    }
    return m.Ack(ctx)
}
```

### Background Job Processing

```go
// Job publisher
publisher.Publish(ctx, interfaces.Message{
    Topic:  "background-jobs",
    Action: "email.send",
    Data: EmailJob{
        To:      "user@example.com",
        Subject: "Welcome!",
        Body:    emailBody,
    },
})

// Job worker
func processEmailJob(ctx context.Context, m interfaces.InboundMessage) error {
    var job EmailJob
    if err := json.Unmarshal(m.Data, &job); err != nil {
        return m.MoveToDeadLetterQueue(ctx)
    }

    if err := emailService.Send(job); err != nil {
        return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
    }

    return m.Ack(ctx)
}
```

## 🛠️ Development Patterns

### Repository Pattern Integration

```go
type UserEventHandler struct {
    userRepo    UserRepository
    emailSvc    EmailService
    logger      *log.Logger
}

func (h *UserEventHandler) HandleMessage(ctx context.Context, m interfaces.InboundMessage) error {
    switch m.Action {
    case "user.created":
        return h.handleUserCreated(ctx, m)
    case "user.deleted":
        return h.handleUserDeleted(ctx, m)
    }
    return nil
}

func (h *UserEventHandler) handleUserCreated(ctx context.Context, m interfaces.InboundMessage) error {
    var event UserCreatedEvent
    if err := json.Unmarshal(m.Data, &event); err != nil {
        return m.MoveToDeadLetterQueue(ctx)
    }

    // Business logic with repository
    user, err := h.userRepo.FindByID(ctx, event.UserID)
    if err != nil {
        return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
    }

    // Send welcome email
    if err := h.emailSvc.SendWelcome(ctx, user.Email); err != nil {
        return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
    }

    return m.Ack(ctx)
}
```

### Domain-Driven Design Integration

```go
type OrderDomainHandler struct {
    orderAggregate OrderAggregate
    eventBus       EventBus
}

func (h *OrderDomainHandler) HandleMessage(ctx context.Context, m interfaces.InboundMessage) error {
    // Convert to domain event
    domainEvent, err := h.toDomainEvent(m)
    if err != nil {
        return m.MoveToDeadLetterQueue(ctx)
    }

    // Process through domain aggregate
    events, err := h.orderAggregate.Handle(ctx, domainEvent)
    if err != nil {
        if errors.Is(err, domain.ErrRetryable) {
            return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
        }
        return m.MoveToDeadLetterQueue(ctx)
    }

    // Publish resulting events
    for _, event := range events {
        if err := h.eventBus.Publish(ctx, event); err != nil {
            return m.RetryWithDelayFn(ctx, interfaces.ExponentialBackoffDelayFn)
        }
    }

    return m.Ack(ctx)
}
```

## 📊 Monitoring and Observability

### Health Checks

```go
type QueueHealthCheck struct {
    consumer  consumer.Consumer
    publisher publisher.Publisher
}

func (h *QueueHealthCheck) Check(ctx context.Context) error {
    // Test publishing
    testMsg := interfaces.Message{
        ID:     "health-check",
        Topic:  "health",
        Action: "ping",
        Data:   map[string]interface{}{"timestamp": time.Now()},
    }

    if err := h.publisher.Publish(ctx, testMsg); err != nil {
        return fmt.Errorf("publisher health check failed: %w", err)
    }

    // Additional consumer health checks...
    return nil
}
```

### Metrics Integration

```go
func MetricsMiddleware() interfaces.InboundMessageHandlerMiddlewareFunc {
    return func(next interfaces.InboundMessageHandlerFunc) interfaces.InboundMessageHandlerFunc {
        return func(ctx context.Context, m interfaces.InboundMessage) error {
            start := time.Now()

            messagesProcessed.WithLabelValues(m.Topic, m.Action).Inc()

            err := next(ctx, m)

            duration := time.Since(start).Seconds()
            processingDuration.WithLabelValues(m.Topic, m.Action).Observe(duration)

            if err != nil {
                messageErrors.WithLabelValues(m.Topic, m.Action).Inc()
            }

            return err
        }
    }
}
```

## 🤝 Contributing to Documentation

We welcome contributions to improve our documentation! Here's how you can help:

### 📝 Writing Guidelines

- Use clear, concise language
- Provide practical examples
- Include code samples that work
- Add troubleshooting sections
- Keep content up-to-date

### 🐛 Reporting Issues

- Documentation bugs or inaccuracies
- Missing information or examples
- Unclear explanations
- Broken code samples

### 💡 Suggestions

- New use case examples
- Additional patterns and best practices
- Performance optimization tips
- Integration guides

## 📞 Getting Help

- **📖 Documentation**: You're here! Check the component-specific docs above
- **💬 Discussions**: [GitHub Discussions](https://github.com/bxcodec/goqueue/discussions)
- **🐛 Issues**: [GitHub Issues](https://github.com/bxcodec/goqueue/issues)
- **📧 Email**: [maintainer@example.com](mailto:maintainer@example.com)

---

**Happy queueing! 🚀**

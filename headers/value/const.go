package value

// GoquServiceAgent represents a service agent used in the Goqu library.
type GoquServiceAgent string

const (
	RabbitMQ GoquServiceAgent = "goqueue/rabbitmq"
	SQS      GoquServiceAgent = "goqueue/sqs"
	SNS      GoquServiceAgent = "goqueue/sns"
)

// ContentType represents the type of content in an the message
type ContentType string

const (
	ContentTypeJSON ContentType = "application/json"
	ContentTypeText ContentType = "text/plain"
	ContentTypeXML  ContentType = "application/xml"
	ContentTypeHTML ContentType = "text/html"
)

const (
	GoquMessageSchemaVersionV1 = "1.0.0"
)

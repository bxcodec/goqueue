package options

// Platform represents a platform in the system.
type Platform string

const (
	PlatformRabbitMQ     Platform = "rabbitmq"
	PlatformGooglePubSub Platform = "google_pubsub"
	PlatformSNS          Platform = "sns"
	PlatformSQS          Platform = "sqs"
)

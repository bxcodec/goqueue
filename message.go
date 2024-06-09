package goqueue

import (
	"time"

	headerVal "github.com/bxcodec/goqueue/headers/value"
)

// Message represents a message that will be published to the queue
// It contains the message ID, action, topic, data, content type, timestamp, headers, and service agent.
// The schema version is set by the SetSchemaVersion method.
// The message is used to publish messages to the queue.
// Read the concept of message publishing in the documentation, here: TODO(bxcodec): Add link to the documentation
type Message struct {
	ID            string                     `json:"id"`
	Action        string                     `json:"action"`
	Topic         string                     `json:"topic"`
	Data          any                        `json:"data"`
	ContentType   headerVal.ContentType      `json:"contentType"`
	Timestamp     time.Time                  `json:"timestamp"`
	Headers       map[string]interface{}     `json:"headers"`
	ServiceAgent  headerVal.GoquServiceAgent `json:"serviceAgent"`
	schemaVersion string
}

func (m *Message) SetSchemaVersion(v string) {
	m.schemaVersion = v
}

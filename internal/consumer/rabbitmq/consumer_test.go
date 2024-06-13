package rabbitmq

import (
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func TestExtractHeaderInt(t *testing.T) {
	headers := amqp.Table{
		"key1": "123",
		"key2": 456,
		"key3": "not_an_int",
	}

	result1 := extractHeaderInt(headers, "key1")
	if result1 != 123 {
		t.Errorf("Expected 123, but got %d", result1)
	}

	result2 := extractHeaderInt(headers, "key2")
	if result2 != 456 {
		t.Errorf("Expected 456, but got %d", result2)
	}

	result3 := extractHeaderInt(headers, "key3")
	if result3 != 0 {
		t.Errorf("Expected 0, but got %d", result3)
	}

	result4 := extractHeaderInt(headers, "nonexistent_key")
	if result4 != 0 {
		t.Errorf("Expected 0, but got %d", result4)
	}
}

func TestExtractHeaderTime(t *testing.T) {
	headers := amqp.Table{
		"key1": "2022-01-01T12:00:00Z",
		"key2": time.Date(2022, 1, 1, 12, 0, 0, 0, time.UTC),
		"key3": "not_a_time",
	}

	result1 := extractHeaderTime(headers, "key1")
	expected1, _ := time.Parse(time.RFC3339, "2022-01-01T12:00:00Z")
	if !result1.Equal(expected1) {
		t.Errorf("Expected %v, but got %v", expected1, result1)
	}

	result2 := extractHeaderTime(headers, "key2")
	if !result2.Equal(headers["key2"].(time.Time)) {
		t.Errorf("Expected %v, but got %v", headers["key2"].(time.Time), result2)
	}

	result3 := extractHeaderTime(headers, "key3")
	if result3.IsZero() {
		t.Errorf("Expected %v, but got zero date", time.Now())
	}

	result4 := extractHeaderTime(headers, "nonexistent_key")
	if result4.IsZero() {
		t.Errorf("Expected %v, but got zero date", time.Now())
	}
}
func TestExtractHeaderString(t *testing.T) {
	headers := amqp.Table{
		"key1": "value1",
		"key2": "value2",
		"key3": 123,
	}

	result1 := extractHeaderString(headers, "key1")
	if result1 != "value1" {
		t.Errorf("Expected 'value1', but got '%s'", result1)
	}

	result2 := extractHeaderString(headers, "key2")
	if result2 != "value2" {
		t.Errorf("Expected 'value2', but got '%s'", result2)
	}

	result3 := extractHeaderString(headers, "key3")
	if result3 != "" {
		t.Errorf("Expected '', but got '%s'", result3)
	}

	result4 := extractHeaderString(headers, "nonexistent_key")
	if result4 != "" {
		t.Errorf("Expected '', but got '%s'", result4)
	}
}

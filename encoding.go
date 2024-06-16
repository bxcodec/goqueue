package goqueue

import (
	"context"
	"encoding/json"
	"sync"

	headerVal "github.com/bxcodec/goqueue/headers/value"
	"github.com/bxcodec/goqueue/interfaces"
)

// EncoderFn is a function type that encodes a message into a byte slice.
// It takes a context and a message as input and returns the encoded data and an error (if any).
type EncoderFn func(ctx context.Context, m interfaces.Message) (data []byte, err error)

// DecoderFn is a function type that decodes a byte slice into a Message.
// It takes a context and a byte slice as input and returns a Message and an error.
type DecoderFn func(ctx context.Context, data []byte) (m interfaces.Message, err error)

var (
	// JSONEncoder is an implementation of the EncoderFn interface
	// that encodes a Message into JSON format.
	JSONEncoder EncoderFn = func(ctx context.Context, m interfaces.Message) (data []byte, err error) {
		return json.Marshal(m)
	}
	// JSONDecoder is a DecoderFn implementation that decodes JSON data into a Message.
	JSONDecoder DecoderFn = func(ctx context.Context, data []byte) (m interfaces.Message, err error) {
		err = json.Unmarshal(data, &m)
		return
	}

	DefaultEncoder EncoderFn = JSONEncoder
	DefaultDecoder DecoderFn = JSONDecoder
)

var (
	// goQueueEncodingMap is a concurrent-safe map used for encoding in GoQueue.
	goQueueEncodingMap = sync.Map{}
)

// AddGoQueueEncoding stores the given encoding for the specified content type in the goQueueEncodingMap.
// The goQueueEncodingMap is a concurrent-safe map that maps content types to encodings.
// The content type is specified by the `contentType` parameter, and the encoding is specified by the `encoding` parameter.
// This function is typically used to register custom encodings for specific content types in the GoQueue library.
func AddGoQueueEncoding(contentType headerVal.ContentType, encoding *Encoding) {
	goQueueEncodingMap.Store(contentType, encoding)
}

// GetGoQueueEncoding returns the encoding associated with the given content type.
// It looks up the encoding in the goQueueEncodingMap and returns it along with a boolean value indicating if the encoding was found.
// If the encoding is not found, it returns nil and false.
func GetGoQueueEncoding(contentType headerVal.ContentType) (res *Encoding, ok bool) {
	if encoding, ok := goQueueEncodingMap.Load(contentType); ok {
		return encoding.(*Encoding), ok
	}
	return nil, false
}

// Encoding represents an encoding configuration for a specific content type.
type Encoding struct {
	ContentType headerVal.ContentType // The content type associated with this encoding.
	Encode      EncoderFn             // The encoding function used to encode data.
	Decode      DecoderFn             // The decoding function used to decode data.
}

var (
	// JSONEncoding represents the encoding configuration for JSON.
	JSONEncoding = &Encoding{
		ContentType: headerVal.ContentTypeJSON,
		Encode:      JSONEncoder,
		Decode:      JSONDecoder,
	}
	DefaultEncoding = JSONEncoding
)

func init() {
	AddGoQueueEncoding(JSONEncoding.ContentType, JSONEncoding)
}

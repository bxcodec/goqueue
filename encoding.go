package goqu

import (
	"context"
	"encoding/json"
	"sync"

	headerVal "github.com/bxcodec/goqu/headers/value"
)

type EncoderFn func(ctx context.Context, m Message) (data []byte, err error)
type DecoderFn func(ctx context.Context, data []byte) (m Message, err error)

var (
	// JSONEncoder is an implementation of the EncoderFn interface
	// that encodes a Message into JSON format.
	JSONEncoder EncoderFn = func(ctx context.Context, m Message) (data []byte, err error) {
		return json.Marshal(m)
	}
	// JSONDecoder is a DecoderFn implementation that decodes JSON data into a Message.
	JSONDecoder DecoderFn = func(ctx context.Context, data []byte) (m Message, err error) {
		err = json.Unmarshal(data, &m)
		return
	}
)

var (
	GoquEncodingMap = sync.Map{}
)

func AddGoquEncoding(contentType headerVal.ContentType, encoding *Encoding) {
	GoquEncodingMap.Store(contentType, encoding)
}

func GetGoquEncoding(contentType headerVal.ContentType) *Encoding {
	if encoding, ok := GoquEncodingMap.Load(contentType); ok {
		return encoding.(*Encoding)
	}
	return nil
}

type Encoding struct {
	ContentType headerVal.ContentType
	Encode      EncoderFn
	Decode      DecoderFn
}

var (
	JSONEncoding = &Encoding{
		ContentType: headerVal.ContentTypeJSON,
		Encode:      JSONEncoder,
		Decode:      JSONDecoder,
	}
)

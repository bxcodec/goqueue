package errors

const (
	InvalidMessageFormatCode   = "INVALID_MESSAGE_FORMAT"
	EncodingFormatNotSupported = "ENCODING_FORMAT_NOT_SUPPORTED"
	UnKnownError               = "UNKNOWN_ERROR"
)

// Error represents an error with a code and a message.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

var (
	// ErrInvalidMessageFormat is an error that occurs when attempting to unmarshal a message with an invalid format.
	ErrInvalidMessageFormat = Error{
		Code:    InvalidMessageFormatCode,
		Message: "failed to unmarshal the message, removing the message due to wrong message format"}
	// ErrEncodingFormatNotSupported is an error that indicates the encoding format is not supported.
	ErrEncodingFormatNotSupported = Error{
		Code:    EncodingFormatNotSupported,
		Message: "encoding format not supported. Please register the encoding format before using it"}

	// ErrUnknownError is an error that indicates an unknown error occurred.
	ErrUnknownError = Error{
		Code:    UnKnownError,
		Message: "an unknown error occurred"}
)

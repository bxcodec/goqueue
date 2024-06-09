package errors

const (
	InvalidMessageFormatCode = "INVALID_MESSAGE_FORMAT"
)

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e Error) Error() string {
	return e.Message
}

var (
	ErrInvalidMessageFormat = Error{
		Code:    InvalidMessageFormatCode,
		Message: "failed to unmarshal the message, removing the message due to wrong message format"}
)

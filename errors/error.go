package errors

import "fmt"

type ErrorCode int

const (
	GeneralError ErrorCode = iota
	FormatError
	ParseError
	ServiceError
)

var (
	ErrChannelUnavailable = &Error{Code: ServiceError, Message: "channel unavailable"}
	ErrChannelMessage     = &Error{Code: ServiceError, Message: "fail to get message"}
	ErrConnection         = &Error{Code: ServiceError, Message: "fail to connect"}
	ErrMessagePublishing  = &Error{Code: ServiceError, Message: "error on message publishing"}
	ErrJobRegister        = &Error{Code: ServiceError, Message: "error on job register"}
	ErrJobEnqueue         = &Error{Code: ServiceError, Message: "error on enqueue job"}
	ErrEmptyURL           = &Error{Code: FormatError, Message: "needs a non-empty url"}
	ErrMessageSerialize   = &Error{Code: ParseError, Message: "error on message serialize"}
)

type Error struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[Error %d] %s: %s", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[Error %d] %s", e.Code, e.Message)
}

func (e *Error) WithValue(Err error) *Error {
	e.Err = Err
	return e
}

package types

import "fmt"

// Equivalent to aws-sdk-go Error interface
type OpseeError interface {
	error

	// Returns the short phrase depicting the classification of the error.
	Code() string

	// Returns the error details message.
	Message() string

	// Returns the original error if one was set.  Nil is returned if not set.
	OrigErr() error
}

// Error error
func (o Error) Code() string {
	return o.ErrorCode
}

func (o Error) Message() string {
	return o.ErrorMessage
}

func (o Error) Error() string {
	return fmt.Sprintf("ErrorCode: %s, ErrorMessage: %s", o.ErrorCode, o.ErrorMessage)
}

func (o Error) OrigErr() error {
	return fmt.Errorf(o.Error())
}

func NewError(errorCode string, errorMessage string) *Error {
	return &Error{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}

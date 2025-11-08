package server

import (
	"fmt"
)

// StatusError is an error with an associated HTTP status code
type StatusError interface {
	error
	Status() int
}

// statusError is a simple implementation of StatusError
type statusError struct {
	err    error
	status int
}

// Error implements the error interface
func (s statusError) Error() string {
	return s.err.Error()
}

// Status implements the StatusError interface
func (s statusError) Status() int {
	return s.status
}

// Unwrap returns the underlying error
func (s statusError) Unwrap() error {
	return s.err
}

// Error creates a new status error with a message and status code
func Error(status int, format string, a ...any) statusError {
	return statusError{
		err:    fmt.Errorf(format, a...),
		status: status,
	}
}

// NewError creates a new status error with an existing status code and error
func NewError(status int, err error) statusError {
	return statusError{
		err:    err,
		status: status,
	}
}

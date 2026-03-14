package apperr

import "fmt"

// Validation represents a client-caused error: bad input, business rule violation,
// or invalid state. Maps to HTTP 400.
type Validation struct {
	Message string
}

func (e *Validation) Error() string {
	return e.Message
}

// NewValidation creates a validation error with a formatted message.
func NewValidation(format string, args ...interface{}) *Validation {
	return &Validation{Message: fmt.Sprintf(format, args...)}
}

// NotFound represents a missing resource. Maps to HTTP 404.
type NotFound struct {
	Message string
}

func (e *NotFound) Error() string {
	return e.Message
}

// NewNotFound creates a not-found error with a formatted message.
func NewNotFound(format string, args ...interface{}) *NotFound {
	return &NotFound{Message: fmt.Sprintf(format, args...)}
}

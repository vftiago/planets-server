package errors

import (
	"errors"
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	// ErrorTypeNotFound indicates a resource was not found
	ErrorTypeNotFound ErrorType = "not_found"
	// ErrorTypeValidation indicates invalid input data
	ErrorTypeValidation ErrorType = "validation"
	// ErrorTypeConflict indicates a conflict with existing data
	ErrorTypeConflict ErrorType = "conflict"
	// ErrorTypeUnauthorized indicates authentication failure
	ErrorTypeUnauthorized ErrorType = "unauthorized"
	// ErrorTypeForbidden indicates insufficient permissions
	ErrorTypeForbidden ErrorType = "forbidden"
	// ErrorTypeInternal indicates an internal server error
	ErrorTypeInternal ErrorType = "internal"
	// ErrorTypeMethodNotAllowed indicates an unsupported HTTP method
	ErrorTypeMethodNotAllowed ErrorType = "method_not_allowed"
	// ErrorTypeExternal indicates an external service error
	ErrorTypeExternal ErrorType = "external"
)

// AppError is the base error type for application errors
type AppError struct {
	Type    ErrorType
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NotFoundf creates a not found error with formatting
func NotFoundf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// Validation creates a validation error
func Validation(message string) error {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

// Validationf creates a validation error with formatting
func Validationf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: fmt.Sprintf(format, args...),
	}
}

// WrapValidation wraps an error as a validation error
func WrapValidation(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

// Conflictf creates a conflict error with formatting
func Conflictf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: fmt.Sprintf(format, args...),
	}
}

// WrapInternal wraps an error as an internal error
func WrapInternal(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

// Unauthorized creates an unauthorized error
func Unauthorized(message string) error {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

// MethodNotAllowed creates a method not allowed error
func MethodNotAllowed(method string) error {
	return &AppError{
		Type:    ErrorTypeMethodNotAllowed,
		Message: fmt.Sprintf("method %s not allowed", method),
	}
}

// External creates an external service error
func External(message string) error {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: message,
	}
}

// WrapExternal wraps an error as an external service error
func WrapExternal(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: message,
		Err:     err,
	}
}

// GetType returns the error type of an error
func GetType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return ErrorTypeInternal
}

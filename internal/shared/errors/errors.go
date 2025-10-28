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

// NotFound creates a not found error
func NotFound(message string) error {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

// NotFoundf creates a not found error with formatting
func NotFoundf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

// WrapNotFound wraps an error as a not found error
func WrapNotFound(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
		Err:     err,
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

// Conflict creates a conflict error
func Conflict(message string) error {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
	}
}

// Conflictf creates a conflict error with formatting
func Conflictf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: fmt.Sprintf(format, args...),
	}
}

// WrapConflict wraps an error as a conflict error
func WrapConflict(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeConflict,
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

// Unauthorizedf creates an unauthorized error with formatting
func Unauthorizedf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: fmt.Sprintf(format, args...),
	}
}

// WrapUnauthorized wraps an error as an unauthorized error
func WrapUnauthorized(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
		Err:     err,
	}
}

// Forbidden creates a forbidden error
func Forbidden(message string) error {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
	}
}

// Forbiddenf creates a forbidden error with formatting
func Forbiddenf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: fmt.Sprintf(format, args...),
	}
}

// WrapForbidden wraps an error as a forbidden error
func WrapForbidden(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
		Err:     err,
	}
}

// Internal creates an internal error
func Internal(message string) error {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
	}
}

// Internalf creates an internal error with formatting
func Internalf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeInternal,
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

// External creates an external service error
func External(message string) error {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: message,
	}
}

// Externalf creates an external service error with formatting
func Externalf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: fmt.Sprintf(format, args...),
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

// IsNotFound checks if an error is a not found error
func IsNotFound(err error) bool {
	return GetType(err) == ErrorTypeNotFound
}

// IsValidation checks if an error is a validation error
func IsValidation(err error) bool {
	return GetType(err) == ErrorTypeValidation
}

// IsConflict checks if an error is a conflict error
func IsConflict(err error) bool {
	return GetType(err) == ErrorTypeConflict
}

// IsUnauthorized checks if an error is an unauthorized error
func IsUnauthorized(err error) bool {
	return GetType(err) == ErrorTypeUnauthorized
}

// IsForbidden checks if an error is a forbidden error
func IsForbidden(err error) bool {
	return GetType(err) == ErrorTypeForbidden
}

// IsInternal checks if an error is an internal error
func IsInternal(err error) bool {
	return GetType(err) == ErrorTypeInternal
}

// IsExternal checks if an error is an external service error
func IsExternal(err error) bool {
	return GetType(err) == ErrorTypeExternal
}

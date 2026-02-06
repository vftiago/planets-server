package errors

import (
	"errors"
	"fmt"
)

type ErrorType string

const (
	ErrorTypeNotFound         ErrorType = "not_found"
	ErrorTypeValidation       ErrorType = "validation"
	ErrorTypeConflict         ErrorType = "conflict"
	ErrorTypeUnauthorized     ErrorType = "unauthorized"
	ErrorTypeForbidden        ErrorType = "forbidden"
	ErrorTypeInternal         ErrorType = "internal"
	ErrorTypeMethodNotAllowed ErrorType = "method_not_allowed"
	ErrorTypeExternal         ErrorType = "external"
)

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

func NotFoundf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: fmt.Sprintf(format, args...),
	}
}

func Validation(message string) error {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
	}
}

func Validationf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: fmt.Sprintf(format, args...),
	}
}

func WrapValidation(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Err:     err,
	}
}

func Conflictf(format string, args ...interface{}) error {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: fmt.Sprintf(format, args...),
	}
}

func WrapInternal(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: message,
		Err:     err,
	}
}

func Unauthorized(message string) error {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

func Forbidden(message string) error {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
	}
}

func MethodNotAllowed(method string) error {
	return &AppError{
		Type:    ErrorTypeMethodNotAllowed,
		Message: fmt.Sprintf("method %s not allowed", method),
	}
}

func External(message string) error {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: message,
	}
}

func WrapExternal(message string, err error) error {
	return &AppError{
		Type:    ErrorTypeExternal,
		Message: message,
		Err:     err,
	}
}

func GetType(err error) ErrorType {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type
	}
	return ErrorTypeInternal
}

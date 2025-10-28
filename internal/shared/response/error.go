package response

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"planets-server/internal/shared/errors"
)

// ErrorResponse represents the JSON error response sent to clients
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Error logs an error and sends a JSON error response to the client
// This should be the only place where errors are logged in the application
func Error(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	errorType := errors.GetType(err)
	statusCode := mapErrorTypeToStatusCode(errorType)

	// Log the error with context
	logError(logger, r, err, errorType, statusCode)

	// Send JSON error response
	sendErrorResponse(w, errorType, err.Error(), statusCode)
}

// ErrorWithMessage logs an error and sends a JSON error response with a custom client message
// Use this when you want to show a different message to the client than the internal error
func ErrorWithMessage(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error, clientMessage string) {
	errorType := errors.GetType(err)
	statusCode := mapErrorTypeToStatusCode(errorType)

	// Log the actual error with context
	logError(logger, r, err, errorType, statusCode)

	// Send JSON error response with custom message
	sendErrorResponse(w, errorType, clientMessage, statusCode)
}

// mapErrorTypeToStatusCode maps error types to HTTP status codes
func mapErrorTypeToStatusCode(errorType errors.ErrorType) int {
	switch errorType {
	case errors.ErrorTypeNotFound:
		return http.StatusNotFound
	case errors.ErrorTypeValidation:
		return http.StatusBadRequest
	case errors.ErrorTypeConflict:
		return http.StatusConflict
	case errors.ErrorTypeUnauthorized:
		return http.StatusUnauthorized
	case errors.ErrorTypeForbidden:
		return http.StatusForbidden
	case errors.ErrorTypeExternal:
		return http.StatusServiceUnavailable
	case errors.ErrorTypeInternal:
		fallthrough
	default:
		return http.StatusInternalServerError
	}
}

// logError logs the error with appropriate level and context
func logError(logger *slog.Logger, r *http.Request, err error, errorType errors.ErrorType, statusCode int) {
	// Create logger with request context
	logCtx := logger.With(
		"method", r.Method,
		"path", r.URL.Path,
		"remote_addr", r.RemoteAddr,
		"error_type", errorType,
		"status_code", statusCode,
	)

	// Log at appropriate level based on error type
	switch errorType {
	case errors.ErrorTypeNotFound:
		// Not found errors are expected, log at debug level
		logCtx.Debug("Resource not found", "error", err)
	case errors.ErrorTypeValidation:
		// Validation errors are client errors, log at debug level
		logCtx.Debug("Validation error", "error", err)
	case errors.ErrorTypeUnauthorized, errors.ErrorTypeForbidden:
		// Auth errors might indicate security issues, log at warn level
		logCtx.Warn("Authorization error", "error", err)
	case errors.ErrorTypeConflict:
		// Conflict errors are expected in some cases, log at info level
		logCtx.Info("Conflict error", "error", err)
	case errors.ErrorTypeExternal:
		// External service errors should be investigated, log at error level
		logCtx.Error("External service error", "error", err)
	case errors.ErrorTypeInternal:
		fallthrough
	default:
		// Internal errors are unexpected and serious, log at error level
		logCtx.Error("Internal server error", "error", err)
	}
}

// sendErrorResponse sends a JSON error response to the client
func sendErrorResponse(w http.ResponseWriter, errorType errors.ErrorType, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   string(errorType),
		Message: message,
		Code:    statusCode,
	}

	// If JSON encoding fails, there's not much we can do at this point
	// The status code has already been sent
	_ = json.NewEncoder(w).Encode(response)
}

// Success sends a JSON success response to the client
func Success(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		// If JSON encoding fails, there's not much we can do at this point
		// The status code has already been sent
		_ = json.NewEncoder(w).Encode(data)
	}
}

package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"planets-server/internal/utils"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// sendErrorResponse sends a JSON error response
func sendErrorResponse(w http.ResponseWriter, statusCode int, errorType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := ErrorResponse{
		Error:   errorType,
		Message: message,
		Code:    statusCode,
	}
	
	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		slog.Error("Failed to encode error response", 
			"error", err,
			"status_code", statusCode,
			"error_type", errorType)
	}
}

// redirectWithError redirects to frontend with error parameters
func redirectWithError(w http.ResponseWriter, r *http.Request, errorType, message string) {
	frontendURL := utils.GetEnv("FRONTEND_URL", "http://localhost:3000")
	errorURL := fmt.Sprintf("%s/auth/error?error=%s&message=%s", 
		frontendURL, errorType, message)
	
	slog.Debug("Redirecting to frontend with error",
		"frontend_url", frontendURL,
		"error_type", errorType,
		"message", message)
	
	http.Redirect(w, r, errorURL, http.StatusTemporaryRedirect)
}

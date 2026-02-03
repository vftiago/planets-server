package handlers

import (
	"fmt"
	"net/http"

	"planets-server/internal/shared/config"
)

// redirectWithError redirects to frontend with error parameters
func redirectWithError(w http.ResponseWriter, r *http.Request, errorType, message string) {
	cfg := config.GlobalConfig
	errorURL := fmt.Sprintf("%s/auth/error?error=%s&message=%s",
		cfg.Frontend.URL, errorType, message)

	http.Redirect(w, r, errorURL, http.StatusTemporaryRedirect)
}
